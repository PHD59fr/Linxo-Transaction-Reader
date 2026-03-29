package config

import (
	"os"
	"testing"
	"time"
)

func setEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

func TestLoad_AllDefaults(t *testing.T) {
	setEnv(t, map[string]string{
		"API_KEY":        "test-key",
		"LINXO_EMAIL":    "user@test.com",
		"LINXO_PASSWORD": "pass123",
	})

	cfg := Load()

	if cfg.APIKey != "test-key" {
		t.Errorf("expected APIKey=%q, got %q", "test-key", cfg.APIKey)
	}
	if cfg.LinxoEmail != "user@test.com" {
		t.Errorf("expected LinxoEmail=%q, got %q", "user@test.com", cfg.LinxoEmail)
	}
	if cfg.LinxoPass != "pass123" {
		t.Errorf("expected LinxoPass=%q, got %q", "pass123", cfg.LinxoPass)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected Port=%q, got %q", "8080", cfg.Port)
	}
	if cfg.Headless != true {
		t.Error("expected Headless=true by default")
	}
	if cfg.BrowserTimeout != 3*time.Minute {
		t.Errorf("expected BrowserTimeout=3m, got %v", cfg.BrowserTimeout)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setEnv(t, map[string]string{
		"API_KEY":         "my-key",
		"LINXO_EMAIL":     "a@b.com",
		"LINXO_PASSWORD":  "pw",
		"CHROME_BIN":      "/usr/bin/chromium",
		"BROWSER_TIMEOUT": "30s",
	})

	cfg := Load()

	if cfg.ChromeBin != "/usr/bin/chromium" {
		t.Errorf("expected ChromeBin=%q, got %q", "/usr/bin/chromium", cfg.ChromeBin)
	}
	if cfg.BrowserTimeout != 30*time.Second {
		t.Errorf("expected BrowserTimeout=30s, got %v", cfg.BrowserTimeout)
	}
}

func TestLoad_InvalidTimeoutFallsBack(t *testing.T) {
	setEnv(t, map[string]string{
		"API_KEY":         "k",
		"LINXO_EMAIL":     "e",
		"LINXO_PASSWORD":  "p",
		"BROWSER_TIMEOUT": "not-a-duration",
	})

	cfg := Load()

	if cfg.BrowserTimeout != 3*time.Minute {
		t.Errorf("expected fallback to 3m, got %v", cfg.BrowserTimeout)
	}
}

// envOr tests ---------------------------------------------------------------

func TestEnvOr_Present(t *testing.T) {
	t.Setenv("TEST_ENV_OR", "value")

	got := envOr("TEST_ENV_OR", "fallback")
	if got != "value" {
		t.Errorf("expected %q, got %q", "value", got)
	}
}

func TestEnvOr_Missing(t *testing.T) {
	key := "TEST_ENV_OR_MISSING"
	os.Unsetenv(key) //nolint:errcheck

	got := envOr(key, "fallback")
	if got != "fallback" {
		t.Errorf("expected %q, got %q", "fallback", got)
	}
}

func TestEnvOr_EmptyString(t *testing.T) {
	t.Setenv("TEST_ENV_OR_EMPTY", "   ")

	got := envOr("TEST_ENV_OR_EMPTY", "fallback")
	if got != "fallback" {
		t.Errorf("expected %q for whitespace-only value, got %q", "fallback", got)
	}
}

// envOrDuration tests -------------------------------------------------------

func TestEnvOrDuration_Valid(t *testing.T) {
	t.Setenv("TEST_DUR", "5m30s")

	got := envOrDuration("TEST_DUR", time.Hour)
	if got != 5*time.Minute+30*time.Second {
		t.Errorf("expected 5m30s, got %v", got)
	}
}

func TestEnvOrDuration_Invalid(t *testing.T) {
	t.Setenv("TEST_DUR_BAD", "nope")

	got := envOrDuration("TEST_DUR_BAD", time.Hour)
	if got != time.Hour {
		t.Errorf("expected fallback to 1h, got %v", got)
	}
}

func TestEnvOrDuration_Missing(t *testing.T) {
	key := "TEST_DUR_MISSING"
	os.Unsetenv(key) //nolint:errcheck

	got := envOrDuration(key, 10*time.Second)
	if got != 10*time.Second {
		t.Errorf("expected fallback, got %v", got)
	}
}
