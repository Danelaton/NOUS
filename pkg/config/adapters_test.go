package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveHermesHomePrefersHERMES_HOME(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HERMES_HOME", tmp)
	if got := ResolveHermesHome(); got != tmp {
		t.Fatalf("ResolveHermesHome() = %q, want %q", got, tmp)
	}
}

func TestResolveHermesHomePrefersLocalAppDataOverAppData(t *testing.T) {
	local := t.TempDir()
	roaming := t.TempDir()

	t.Setenv("HERMES_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("LOCALAPPDATA", local)
	t.Setenv("APPDATA", roaming)

	for _, d := range []string{
		filepath.Join(local, "hermes"),
		filepath.Join(roaming, "hermes"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	want := filepath.Join(local, "hermes")
	if got := ResolveHermesHome(); got != want {
		t.Fatalf("ResolveHermesHome() = %q, want %q", got, want)
	}
}

func TestResolveHermesHomeFallsBackToHomeDotHermes(t *testing.T) {
	t.Setenv("HERMES_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("LOCALAPPDATA", "")
	t.Setenv("APPDATA", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}
	want := filepath.Join(home, ".hermes")
	if got := ResolveHermesHome(); got != want {
		t.Fatalf("ResolveHermesHome() = %q, want %q", got, want)
	}
}
