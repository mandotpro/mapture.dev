// Package updater upgrades the current mapture installation in place.
package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const (
	repoOwner = "mandotpro"
	repoName  = "mapture.dev"
	moduleRef = "github.com/mandotpro/mapture.dev/cmd/mapture"
)

// Channel represents a public release lane.
type Channel string

// Channel selects which public release lane should be used for install or upgrade.
const (
	ChannelAuto   Channel = ""
	ChannelStable Channel = "stable"
	ChannelCanary Channel = "canary"
)

type installMethod string

const (
	installMethodUnknown  installMethod = "unknown"
	installMethodHomebrew installMethod = "homebrew"
	installMethodGo       installMethod = "go-install"
	installMethodDirect   installMethod = "direct-binary"
)

// Options configures the self-update flow.
type Options struct {
	RequestedChannel Channel
	CurrentVersion   string
	BuildInfo        *debug.BuildInfo
	Stdout           io.Writer
	Stderr           io.Writer
}

type runtimeInfo struct {
	Version         string
	ModuleVersion   string
	Channel         Channel
	InstallMethod   installMethod
	ExecutablePath  string
	HomebrewFormula string
}

type release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var pseudoVersionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+-\d{14}-[0-9a-f]+$`)

var (
	httpClientFactory = func() *http.Client {
		return &http.Client{Timeout: 45 * time.Second}
	}
	osExecutable  = os.Executable
	evalSymlinks  = filepath.EvalSymlinks
	commandRunner = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, name, args...)
	}
)

// Run upgrades the current mapture installation using the requested or detected channel.
func Run(ctx context.Context, opts Options) error {
	out := opts.Stdout
	if out == nil {
		out = io.Discard
	}

	current, err := detectRuntime(opts.CurrentVersion, opts.BuildInfo)
	if err != nil {
		return err
	}

	targetChannel := opts.RequestedChannel
	if targetChannel == ChannelAuto {
		targetChannel = current.Channel
		if targetChannel == ChannelAuto {
			targetChannel = ChannelStable
		}
	}
	if targetChannel != ChannelStable && targetChannel != ChannelCanary {
		return fmt.Errorf("unsupported update channel %q", targetChannel)
	}

	writef(out, "mapture update: detected %s install (%s)\n", current.InstallMethod, current.ExecutablePath)
	writef(out, "mapture update: current version %s, channel %s\n", displayVersion(current.Version), targetChannel)

	switch current.InstallMethod {
	case installMethodHomebrew:
		if targetChannel != current.Channel {
			return fmt.Errorf("homebrew channel switches are not automatic; uninstall %s and install the %s formula instead", current.HomebrewFormula, targetChannel)
		}
		return upgradeViaHomebrew(ctx, current, out)
	case installMethodGo:
		return upgradeViaGoInstall(ctx, targetChannel, out)
	default:
		return upgradeDirectBinary(ctx, current, targetChannel, out)
	}
}

func detectRuntime(currentVersion string, info *debug.BuildInfo) (runtimeInfo, error) {
	exe, err := osExecutable()
	if err != nil {
		return runtimeInfo{}, fmt.Errorf("resolve executable: %w", err)
	}
	if resolved, resolveErr := evalSymlinks(exe); resolveErr == nil {
		exe = resolved
	}

	moduleVersion := ""
	if info != nil {
		moduleVersion = info.Main.Version
	}

	method, formula := detectInstallMethod(exe)
	channel := detectChannel(currentVersion, moduleVersion, method, formula)
	if channel == ChannelAuto {
		channel = ChannelStable
	}

	return runtimeInfo{
		Version:         currentVersion,
		ModuleVersion:   moduleVersion,
		Channel:         channel,
		InstallMethod:   method,
		ExecutablePath:  exe,
		HomebrewFormula: formula,
	}, nil
}

func detectInstallMethod(executablePath string) (installMethod, string) {
	slashed := filepath.ToSlash(executablePath)
	switch {
	case strings.Contains(slashed, "/Cellar/mapture-canary/"):
		return installMethodHomebrew, "mapture-canary"
	case strings.Contains(slashed, "/Cellar/mapture/"):
		return installMethodHomebrew, "mapture"
	}

	if matchesGoBin(executablePath) {
		return installMethodGo, ""
	}

	return installMethodDirect, ""
}

func matchesGoBin(executablePath string) bool {
	goBinary, err := exec.LookPath("go")
	if err != nil || goBinary == "" {
		return false
	}

	gobin, err := goEnv("GOBIN")
	if err == nil && gobin != "" {
		if sameFilepath(filepath.Join(gobin, binaryNameFor(runtime.GOOS)), executablePath) {
			return true
		}
	}

	gopath, err := goEnv("GOPATH")
	if err != nil || gopath == "" {
		return false
	}
	for _, entry := range filepath.SplitList(gopath) {
		if sameFilepath(filepath.Join(entry, "bin", binaryNameFor(runtime.GOOS)), executablePath) {
			return true
		}
	}
	return false
}

func goEnv(key string) (string, error) {
	cmd := commandRunner(context.Background(), "go", "env", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func sameFilepath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}

func detectChannel(version string, moduleVersion string, method installMethod, formula string) Channel {
	switch formula {
	case "mapture-canary":
		return ChannelCanary
	case "mapture":
		return ChannelStable
	}

	candidate := firstNonEmpty(version, moduleVersion)
	if strings.Contains(candidate, "canary") {
		return ChannelCanary
	}
	if pseudoVersionPattern.MatchString(candidate) {
		return ChannelCanary
	}
	if looksLikeStableSemver(candidate) {
		return ChannelStable
	}
	if method == installMethodGo && candidate != "" {
		return ChannelCanary
	}
	return ChannelAuto
}

func looksLikeStableSemver(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	value = strings.TrimPrefix(value, "v")
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func upgradeViaHomebrew(ctx context.Context, current runtimeInfo, out io.Writer) error {
	formula := current.HomebrewFormula
	if formula == "" {
		return errors.New("homebrew install detected without formula name")
	}

	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("brew is not available in PATH; reinstall with curl or replace the binary manually")
	}

	writef(out, "mapture update: running brew upgrade %s\n", formula)
	cmd := commandRunner(ctx, "brew", "upgrade", formula)
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

func upgradeViaGoInstall(ctx context.Context, channel Channel, out io.Writer) error {
	if _, err := exec.LookPath("go"); err != nil {
		return upgradeDirectBinary(ctx, runtimeInfo{ExecutablePath: mustExecutablePath()}, channel, out)
	}

	ref := "@latest"
	if channel == ChannelCanary {
		ref = "@main"
	}

	writef(out, "mapture update: running go install %s%s\n", moduleRef, ref)
	cmd := commandRunner(ctx, "go", "install", moduleRef+ref)
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if err == nil {
		return nil
	}

	if channel == ChannelCanary {
		writef(out, "mapture update: retrying canary install with GOPROXY=direct\n")
		retry := commandRunner(ctx, "go", "install", moduleRef+ref)
		retry.Env = append(os.Environ(), "GOPROXY=direct")
		retry.Stdout = out
		retry.Stderr = out
		return retry.Run()
	}

	return err
}

func mustExecutablePath() string {
	exe, err := osExecutable()
	if err != nil {
		return ""
	}
	if resolved, resolveErr := evalSymlinks(exe); resolveErr == nil {
		return resolved
	}
	return exe
}

func upgradeDirectBinary(ctx context.Context, current runtimeInfo, channel Channel, out io.Writer) error {
	release, err := fetchRelease(ctx, channel)
	if err != nil {
		return err
	}

	asset, err := release.findAsset(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	targetVersion := release.binaryVersion(runtime.GOOS, runtime.GOARCH)
	if current.Version != "" && targetVersion != "" && current.Version == targetVersion {
		writef(out, "mapture update: already at %s\n", targetVersion)
		return nil
	}

	writef(out, "mapture update: downloading %s\n", asset.Name)
	binaryBytes, mode, err := downloadAndExtractBinary(ctx, asset)
	if err != nil {
		return err
	}

	if current.ExecutablePath == "" {
		return errors.New("could not determine current executable path")
	}

	if err := replaceExecutable(current.ExecutablePath, binaryBytes, mode); err != nil {
		return err
	}

	if targetVersion == "" {
		targetVersion = release.TagName
	}
	writef(out, "mapture update: updated to %s\n", targetVersion)
	if runtime.GOOS == "windows" {
		writef(out, "mapture update: restart mapture to finish the staged replacement\n")
	}
	return nil
}

func fetchRelease(ctx context.Context, channel Channel) (*release, error) {
	var apiURL string
	switch channel {
	case ChannelCanary:
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/canary", repoOwner, repoName)
	case ChannelStable:
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	default:
		return nil, fmt.Errorf("unsupported channel %q", channel)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "mapture-self-update")
	if token := firstNonEmpty(os.Getenv("GH_TOKEN"), os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClientFactory().Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s release metadata: %w", channel, err)
	}
	defer closeSilently(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("fetch %s release metadata: github returned %s: %s", channel, resp.Status, strings.TrimSpace(string(body)))
	}

	var release release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release metadata: %w", err)
	}
	return &release, nil
}

func (r *release) findAsset(goos, goarch string) (*asset, error) {
	if len(r.Assets) == 0 {
		return nil, fmt.Errorf("release %s does not publish prebuilt assets for %s/%s yet", r.TagName, goos, goarch)
	}
	suffix := fmt.Sprintf("_%s_%s", goos, goarch)
	for idx := range r.Assets {
		asset := &r.Assets[idx]
		if !strings.HasPrefix(asset.Name, "mapture_") {
			continue
		}
		if !strings.Contains(asset.Name, suffix) {
			continue
		}
		if strings.HasSuffix(asset.Name, ".tar.gz") || strings.HasSuffix(asset.Name, ".zip") {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("no release asset for %s/%s in %s", goos, goarch, r.TagName)
}

func (r *release) binaryVersion(goos, goarch string) string {
	asset, err := r.findAsset(goos, goarch)
	if err != nil {
		return ""
	}
	name := asset.Name
	name = strings.TrimPrefix(name, "mapture_")
	name = strings.TrimSuffix(name, ".tar.gz")
	name = strings.TrimSuffix(name, ".zip")
	name = strings.TrimSuffix(name, fmt.Sprintf("_%s_%s", goos, goarch))
	return strings.ReplaceAll(name, "_", "+")
}

func downloadAndExtractBinary(ctx context.Context, asset *asset) ([]byte, os.FileMode, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "mapture-self-update")
	if token := firstNonEmpty(os.Getenv("GH_TOKEN"), os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClientFactory().Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("download %s: %w", asset.Name, err)
	}
	defer closeSilently(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("download %s: github returned %s", asset.Name, resp.Status)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	switch {
	case strings.HasSuffix(asset.Name, ".tar.gz"):
		return extractBinaryFromTarGz(payload)
	case strings.HasSuffix(asset.Name, ".zip"):
		return extractBinaryFromZip(payload)
	default:
		return nil, 0, fmt.Errorf("unsupported archive format for %s", asset.Name)
	}
}

func extractBinaryFromTarGz(payload []byte) ([]byte, os.FileMode, error) {
	gz, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	defer closeSilently(gz)

	reader := tar.NewReader(gz)
	target := binaryNameFor(runtime.GOOS)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		if filepath.Base(header.Name) != target {
			continue
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, 0, err
		}
		mode := os.FileMode(header.Mode)
		if mode == 0 {
			mode = 0o755
		}
		return data, mode, nil
	}
	return nil, 0, errors.New("mapture binary not found in tar archive")
}

func extractBinaryFromZip(payload []byte) ([]byte, os.FileMode, error) {
	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return nil, 0, err
	}
	target := binaryNameFor(runtime.GOOS)
	for _, file := range reader.File {
		if filepath.Base(file.Name) != target {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			return nil, 0, err
		}
		defer closeSilently(handle)
		data, err := io.ReadAll(handle)
		if err != nil {
			return nil, 0, err
		}
		mode := file.Mode()
		if mode == 0 {
			mode = 0o755
		}
		return data, mode, nil
	}
	return nil, 0, errors.New("mapture binary not found in zip archive")
}

func replaceExecutable(executablePath string, binary []byte, mode os.FileMode) error {
	if mode == 0 {
		mode = 0o755
	}
	dir := filepath.Dir(executablePath)
	tmp, err := os.CreateTemp(dir, "mapture-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer removeSilently(tmpPath)

	if _, err := tmp.Write(binary); err != nil {
		closeSilently(tmp)
		return fmt.Errorf("write temp binary: %w", err)
	}
	if err := tmp.Chmod(mode); err != nil {
		closeSilently(tmp)
		return fmt.Errorf("chmod temp binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp binary: %w", err)
	}

	if runtime.GOOS == "windows" {
		return stageWindowsReplacement(executablePath, tmpPath)
	}

	if err := os.Rename(tmpPath, executablePath); err != nil {
		return fmt.Errorf("replace current binary: %w", err)
	}
	return nil
}

func stageWindowsReplacement(executablePath, tmpPath string) error {
	script, err := os.CreateTemp(filepath.Dir(executablePath), "mapture-update-*.cmd")
	if err != nil {
		return fmt.Errorf("create windows helper: %w", err)
	}
	scriptPath := script.Name()
	defer closeSilently(script)

	content := fmt.Sprintf("@echo off\r\nping 127.0.0.1 -n 3 >NUL\r\nmove /Y %q %q >NUL\r\ndel /Q %q >NUL\r\n", tmpPath, executablePath, scriptPath)
	if _, err := script.WriteString(content); err != nil {
		return fmt.Errorf("write windows helper: %w", err)
	}
	if err := script.Close(); err != nil {
		return fmt.Errorf("close windows helper: %w", err)
	}

	cmd := commandRunner(context.Background(), "cmd", "/C", "start", "", "/min", scriptPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch windows helper: %w", err)
	}
	return nil
}

func binaryNameFor(goos string) string {
	if goos == "windows" {
		return "mapture.exe"
	}
	return "mapture"
}

func displayVersion(version string) string {
	if strings.TrimSpace(version) == "" {
		return "unknown"
	}
	return version
}

func writef(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	_, _ = fmt.Fprintf(w, format, args...)
}

func closeSilently(closer io.Closer) {
	if closer == nil {
		return
	}
	_ = closer.Close()
}

func removeSilently(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}
