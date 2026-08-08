package main

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all configuration for the MCP server.
type Config struct {
	DSN         string // MySQL data source name
	ContextFile string // Path to context.yaml
	LogLevel    string // Log level (debug, info, warn, error)
}

// LoadConfig reads configuration from CLI flags and environment variables.
// Flags take precedence over env vars.
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.DSN, "dsn", envOrDefault("DSN", ""), "MySQL DSN (prefer DSN env var to avoid credential leakage in ps)")
	flag.StringVar(&cfg.ContextFile, "context", envOrDefault("CONTEXT_FILE", "./context.yaml"), "Path to context.yaml")
	flag.StringVar(&cfg.LogLevel, "log-level", envOrDefault("LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
	flag.Parse()

	if cfg.DSN == "" {
		return nil, fmt.Errorf("DSN is required (use --dsn flag or DSN env var)")
	}

	return cfg, nil
}

// envOrDefault returns the value of an environment variable, or a default if unset.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
