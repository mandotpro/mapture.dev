package updater

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectChannel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		version       string
		moduleVersion string
		method        installMethod
		formula       string
		want          Channel
	}{
		{
			name:    "homebrew canary formula",
			method:  installMethodHomebrew,
			formula: "mapture-canary",
			want:    ChannelCanary,
		},
		{
			name:    "homebrew stable formula",
			method:  installMethodHomebrew,
			formula: "mapture",
			want:    ChannelStable,
		},
		{
			name:    "embedded canary version",
			version: "0.0.0-canary.20260411+sha.abcdef0",
			method:  installMethodDirect,
			want:    ChannelCanary,
		},
		{
			name:          "go pseudo version defaults to canary",
			moduleVersion: "v0.0.0-20260411101010-abcdef123456",
			method:        installMethodGo,
			want:          ChannelCanary,
		},
		{
			name:    "stable semver release",
			version: "v0.3.0",
			method:  installMethodDirect,
			want:    ChannelStable,
		},
		{
			name:    "unknown direct binary defaults to auto",
			version: "0.0.0-dev+sha.abcdef0",
			method:  installMethodDirect,
			want:    ChannelAuto,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detectChannel(tc.version, tc.moduleVersion, tc.method, tc.formula)
			if got != tc.want {
				t.Fatalf("detectChannel() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReleaseBinaryVersion(t *testing.T) {
	t.Parallel()

	release := &release{
		TagName: "canary",
		Assets: []asset{
			{Name: "mapture_0.0.0-canary.20260411_sha.abcdef0_darwin_arm64.tar.gz"},
		},
	}

	got := release.binaryVersion("darwin", "arm64")
	if got != "0.0.0-canary.20260411+sha.abcdef0" {
		t.Fatalf("binaryVersion() = %q", got)
	}
}

func TestFindAssetMatchesPlatform(t *testing.T) {
	t.Parallel()

	release := &release{
		TagName: "v0.3.0",
		Assets: []asset{
			{Name: "mapture_v0.3.0_linux_amd64.tar.gz"},
			{Name: "mapture_v0.3.0_darwin_arm64.tar.gz"},
		},
	}

	asset, err := release.findAsset("darwin", "arm64")
	if err != nil {
		t.Fatalf("findAsset returned error: %v", err)
	}
	if asset.Name != "mapture_v0.3.0_darwin_arm64.tar.gz" {
		t.Fatalf("findAsset returned %q", asset.Name)
	}
}

func TestFindAssetWithoutPublishedAssets(t *testing.T) {
	t.Parallel()

	release := &release{TagName: "mapture-v0.2.0"}
	_, err := release.findAsset("darwin", "arm64")
	if err == nil || err.Error() != "release mapture-v0.2.0 does not publish prebuilt assets for darwin/arm64 yet" {
		t.Fatalf("findAsset returned %v", err)
	}
}

func TestDetectInstallMethod(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
		want installMethod
	}{
		{
			name: "homebrew canary",
			path: "/opt/homebrew/Cellar/mapture-canary/0.0.0/bin/mapture",
			want: installMethodHomebrew,
		},
		{
			name: "homebrew stable",
			path: "/opt/homebrew/Cellar/mapture/0.3.0/bin/mapture",
			want: installMethodHomebrew,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, _ := detectInstallMethod(tc.path)
			if got != tc.want {
				t.Fatalf("detectInstallMethod() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBinaryNameFor(t *testing.T) {
	t.Parallel()

	if got := binaryNameFor("windows"); got != "mapture.exe" {
		t.Fatalf("binaryNameFor(windows) = %q", got)
	}
	if got := binaryNameFor("darwin"); got != "mapture" {
		t.Fatalf("binaryNameFor(darwin) = %q", got)
	}
}

func TestSameFilepath(t *testing.T) {
	t.Parallel()

	left := filepath.Join("a", "..", "tmp", binaryNameFor(runtime.GOOS))
	right := filepath.Join("tmp", binaryNameFor(runtime.GOOS))
	if !sameFilepath(left, right) {
		t.Fatal("expected paths to normalize to the same path")
	}
}

func TestInspectReportsHumanReadableRuntime(t *testing.T) {

	originalExecutable := osExecutable
	originalEvalSymlinks := evalSymlinks
	defer func() {
		osExecutable = originalExecutable
		evalSymlinks = originalEvalSymlinks
	}()

	osExecutable = func() (string, error) {
		return "/opt/homebrew/Cellar/mapture-canary/0.0.0/bin/mapture", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return path, nil
	}

	details, err := Inspect("0.0.0-canary.20260412+sha.1bd3598", nil)
	if err != nil {
		t.Fatalf("Inspect returned error: %v", err)
	}
	if details.Channel != ChannelCanary {
		t.Fatalf("Channel = %q, want %q", details.Channel, ChannelCanary)
	}
	if details.InstallMethod != "homebrew" {
		t.Fatalf("InstallMethod = %q", details.InstallMethod)
	}
}

func TestCheckVersionReportsOutdatedCanary(t *testing.T) {
	originalExecutable := osExecutable
	originalEvalSymlinks := evalSymlinks
	originalFetch := fetchReleaseFn
	defer func() {
		osExecutable = originalExecutable
		evalSymlinks = originalEvalSymlinks
		fetchReleaseFn = originalFetch
	}()

	osExecutable = func() (string, error) {
		return "/opt/homebrew/Cellar/mapture-canary/0.0.0/bin/mapture", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return path, nil
	}
	fetchReleaseFn = func(_ context.Context, channel Channel) (*release, error) {
		switch channel {
		case ChannelStable:
			return &release{TagName: "v0.3.0"}, nil
		case ChannelCanary:
			return &release{
				TagName: "canary",
				Assets: []asset{
					{Name: "mapture_0.0.0-canary.20260412_sha.1bd3598_darwin_arm64.tar.gz"},
				},
			}, nil
		default:
			t.Fatalf("unexpected channel: %q", channel)
			return nil, nil
		}
	}

	status, err := CheckVersion(context.Background(), "0.0.0-canary.20260409+sha.c649dd6", nil)
	if err != nil {
		t.Fatalf("CheckVersion returned error: %v", err)
	}
	if !status.UpdateAvailable {
		t.Fatal("expected update to be available")
	}
	if status.LatestForChannel != "0.0.0-canary.20260412+sha.1bd3598" {
		t.Fatalf("LatestForChannel = %q", status.LatestForChannel)
	}
}

func TestCheckVersionReportsCurrentStable(t *testing.T) {
	originalExecutable := osExecutable
	originalEvalSymlinks := evalSymlinks
	originalFetch := fetchReleaseFn
	defer func() {
		osExecutable = originalExecutable
		evalSymlinks = originalEvalSymlinks
		fetchReleaseFn = originalFetch
	}()

	osExecutable = func() (string, error) {
		return "/usr/local/bin/mapture", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return path, nil
	}
	fetchReleaseFn = func(_ context.Context, channel Channel) (*release, error) {
		switch channel {
		case ChannelStable:
			return &release{TagName: "v0.3.0"}, nil
		case ChannelCanary:
			return &release{
				TagName: "canary",
				Assets: []asset{
					{Name: "mapture_0.0.0-canary.20260412_sha.1bd3598_darwin_arm64.tar.gz"},
				},
			}, nil
		default:
			t.Fatalf("unexpected channel: %q", channel)
			return nil, nil
		}
	}

	status, err := CheckVersion(context.Background(), "v0.3.0", nil)
	if err != nil {
		t.Fatalf("CheckVersion returned error: %v", err)
	}
	if status.UpdateAvailable {
		t.Fatal("expected stable version to be current")
	}
	if status.LatestStable != "v0.3.0" {
		t.Fatalf("LatestStable = %q", status.LatestStable)
	}
}
