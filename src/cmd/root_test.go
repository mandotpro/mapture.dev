package cmd

import (
	"strings"
	"testing"
)

func TestLoadProjectSuccess(t *testing.T) {
	t.Parallel()

	configPath, cfg, cat, err := loadProject("../../examples/demo")
	if err != nil {
		t.Fatalf("loadProject returned error: %v", err)
	}
	if !strings.HasSuffix(configPath, "examples/demo/mapture.yaml") {
		t.Fatalf("unexpected config path: %s", configPath)
	}
	if cfg.Catalog.Dir != "./architecture" {
		t.Fatalf("unexpected catalog dir: %s", cfg.Catalog.Dir)
	}
	if len(cat.Teams) != 2 || len(cat.Domains) != 2 || len(cat.Events) != 1 {
		t.Fatalf("unexpected catalog sizes: teams=%d domains=%d events=%d", len(cat.Teams), len(cat.Domains), len(cat.Events))
	}
}

func TestLoadProjectRejectsBrokenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "bad config role", path: "../../examples/invalid/bad-config-role", wantErr: "random"},
		{name: "duplicate team", path: "../../examples/invalid/duplicate-team", wantErr: "duplicate team id"},
		{name: "unknown domain owner", path: "../../examples/invalid/unknown-domain-owner", wantErr: "unknown team"},
		{name: "invalid event status", path: "../../examples/invalid/invalid-event-status", wantErr: "random"},
		{name: "missing teams file", path: "../../examples/invalid/missing-teams-file", wantErr: "read"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, err := loadProject(tc.path)
			if err == nil {
				t.Fatalf("expected error for %s", tc.path)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected %q in error, got %v", tc.wantErr, err)
			}
		})
	}
}
