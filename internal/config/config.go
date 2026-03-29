package config

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"
)

// Config holds all runtime configuration.
type Config struct {
	Port           string
	APIKey         string
	Debug          bool
	Headless       bool
	LinxoEmail     string
	LinxoPass      string
	ChromeBin      string
	BrowserTimeout time.Duration
}

// Load reads configuration from environment variables and CLI flags.
func Load() *Config {
	fs := flag.NewFlagSet("linxo-reader", flag.ContinueOnError)
	debug := fs.Bool("debug", false, "Launch browser visibly (non-headless)")
	port := fs.String("port", "8080", "HTTP listen port")
	_ = fs.Parse(os.Args[1:])

	apiKey := envOr("API_KEY", "")
	if apiKey == "" {
		log.Fatal("API_KEY env var is required")
	}

	linxoEmail := envOr("LINXO_EMAIL", "")
	if linxoEmail == "" {
		log.Fatal("LINXO_EMAIL env var is required")
	}

	linxoPass := envOr("LINXO_PASSWORD", "")
	if linxoPass == "" {
		log.Fatal("LINXO_PASSWORD env var is required")
	}

	return &Config{
		Port:           *port,
		APIKey:         apiKey,
		Debug:          *debug,
		Headless:       !*debug,
		LinxoEmail:     linxoEmail,
		LinxoPass:      linxoPass,
		ChromeBin:      envOr("CHROME_BIN", ""),
		BrowserTimeout: envOrDuration("BROWSER_TIMEOUT", 3*time.Minute),
	}
}

func envOr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func envOrDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
