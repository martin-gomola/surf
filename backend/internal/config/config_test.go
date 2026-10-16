package config

import (
	"path/filepath"
	"testing"
)

func TestCompatibilityGenerationAcceptsPublishedLegacyToken(t *testing.T) {
	original := CompatibilityVersion
	defer func() { CompatibilityVersion = original }()
	CompatibilityVersion = "1"
	if CompatibilityGeneration() != 1 || WireCompatibilityVersion() != "20260831-1" {
		t.Fatalf("generation=%d wire=%q", CompatibilityGeneration(), WireCompatibilityVersion())
	}
	for _, test := range []struct {
		value, legacy string
		want          int
		ok            bool
	}{
		{"1", "", 1, true},
		{"", "20260831-1", 1, true},
		{"2", "20260831-1", 2, true},
		{"", "unknown", 0, false},
	} {
		got, ok := ParseClientCompatibility(test.value, test.legacy)
		if got != test.want || ok != test.ok {
			t.Errorf("ParseClientCompatibility(%q, %q)=(%d,%t), want (%d,%t)",
				test.value, test.legacy, got, ok, test.want, test.ok)
		}
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"SURF_HOME", "BIND_ADDR", "PORT", "CHROME", "START_URL",
		"PROFILE", "VW", "VH",
		"SURF_SERVER_NAME", "SURF_PUBLIC_ADDRESS", "DOWNLOADS", "UPLOADS",
		"SURF_TUNNEL_HOST",
		"CHROME_NO_SANDBOX",
		"SURF_BROWSER_VISIBLE",
		"SURF_DASHBOARD_ADDR", "SURF_DASHBOARD_PUBLIC_URL",
		"SURF_ADAPTIVE_VIDEO",
		"SURF_SUSPEND_SILENT_AUDIO",
		"STREAM_SCALE", "STREAM_BITRATE", "STREAM_QUANTIZER", "STREAM_MAX_FPS",
	} {
		t.Setenv(key, "")
	}
}

func loadConfig(t *testing.T) *Config {
	t.Helper()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestLoadDefaults(t *testing.T) {
	clearConfigEnv(t)
	home := t.TempDir()
	t.Setenv("SURF_HOME", home)
	t.Setenv("CHROME_NO_SANDBOX", "0")
	cfg := loadConfig(t)
	if cfg.Port != 18080 {
		t.Fatalf("Port=%d, want 18080", cfg.Port)
	}
	if cfg.Profile != filepath.Join(home, "profile") || cfg.DownloadsDir != filepath.Join(home, "downloads") || cfg.UploadsDir != filepath.Join(home, "uploads") {
		t.Fatalf("data dirs = %q %q %q", cfg.Profile, cfg.DownloadsDir, cfg.UploadsDir)
	}
	if cfg.ChromePath != "" {
		t.Fatalf("ChromePath=%q, want managed runtime selection", cfg.ChromePath)
	}
	if cfg.ChromeNoSandbox {
		t.Fatal("ChromeNoSandbox=true")
	}
	if cfg.BrowserVisible {
		t.Fatal("BrowserVisible=true, want headless default")
	}
	if cfg.DashboardAddr != "" || cfg.DashboardPublicURL != "" {
		t.Fatalf("dashboard defaults = %q %q, want disabled", cfg.DashboardAddr, cfg.DashboardPublicURL)
	}
	if cfg.StreamBitrateK != 16000 {
		t.Fatalf("StreamBitrateK=%d, want 16000", cfg.StreamBitrateK)
	}
	if cfg.StreamQuantizer != 12 {
		t.Fatalf("StreamQuantizer=%d, want 12", cfg.StreamQuantizer)
	}
	if cfg.StreamMaxFPS != 60 {
		t.Fatalf("StreamMaxFPS=%d, want 60", cfg.StreamMaxFPS)
	}
	if cfg.AdaptiveVideo {
		t.Fatal("AdaptiveVideo=true, want fixed native 60 FPS by default")
	}
	if cfg.SuspendSilentAudio {
		t.Fatal("SuspendSilentAudio=true, want uninterrupted audio by default")
	}
}

func TestDashboardCanBeEnabled(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SURF_DASHBOARD_ADDR", "0.0.0.0:18081")
	t.Setenv("SURF_DASHBOARD_PUBLIC_URL", "http://127.0.0.1:18081/")
	cfg := loadConfig(t)
	if cfg.DashboardAddr != "0.0.0.0:18081" || cfg.DashboardPublicURL != "http://127.0.0.1:18081/" {
		t.Fatalf("dashboard config = %q %q", cfg.DashboardAddr, cfg.DashboardPublicURL)
	}
}

func TestVisibleBrowserCanBeEnabled(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SURF_BROWSER_VISIBLE", "1")
	if !loadConfig(t).BrowserVisible {
		t.Fatal("BrowserVisible=false after explicit enable")
	}
}

func TestStreamMaxFPSIsBounded(t *testing.T) {
	for _, test := range []struct {
		value string
		want  int
	}{
		{"30", 30},
		{"5", 15},
		{"90", 60},
	} {
		clearConfigEnv(t)
		t.Setenv("STREAM_MAX_FPS", test.value)
		if got := loadConfig(t).StreamMaxFPS; got != test.want {
			t.Errorf("STREAM_MAX_FPS=%q loaded %d, want %d", test.value, got, test.want)
		}
	}
}

func TestSilentAudioSuspensionCanBeEnabled(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SURF_SUSPEND_SILENT_AUDIO", "1")
	if !loadConfig(t).SuspendSilentAudio {
		t.Fatal("SuspendSilentAudio=false after explicit enable")
	}
}

func TestAdaptiveVideoCanBeEnabled(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SURF_ADAPTIVE_VIDEO", "1")
	if !loadConfig(t).AdaptiveVideo {
		t.Fatal("AdaptiveVideo=false after explicit enable")
	}
}

func TestLoadDoesNotRequirePassword(t *testing.T) {
	clearConfigEnv(t)
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}
