package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// ActionConfig holds all configuration for the schema-guard GitHub Action.
type ActionConfig struct {
	ContextFile      string
	BeforeMigrations string
	AfterMigrations  string
	DSN              string
	GitHubToken      string
	Repo             string // "owner/repo"
	PRNumber         int
	DryRun           bool
}

// LoadActionConfig parses configuration from CLI flags and environment variables.
func LoadActionConfig() (*ActionConfig, error) {
	cfg := &ActionConfig{}

	flag.StringVar(&cfg.ContextFile, "context", envOrDefault("CONTEXT_FILE", "./context.yaml"), "Path to context.yaml")
	flag.StringVar(&cfg.BeforeMigrations, "before-migrations", os.Getenv("BEFORE_MIGRATIONS"), "Path to base branch migrations")
	flag.StringVar(&cfg.AfterMigrations, "after-migrations", os.Getenv("AFTER_MIGRATIONS"), "Path to PR branch migrations")
	flag.StringVar(&cfg.DSN, "dsn", os.Getenv("DSN"), "MySQL DSN for ephemeral DB")
	flag.StringVar(&cfg.GitHubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	flag.StringVar(&cfg.Repo, "repo", os.Getenv("GITHUB_REPOSITORY"), "GitHub repository (owner/repo)")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Print comment to stdout instead of posting to GitHub")

	prStr := os.Getenv("PR_NUMBER")
	flag.Parse()

	if prStr != "" {
		pr, err := strconv.Atoi(prStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PR_NUMBER: %w", err)
		}
		cfg.PRNumber = pr
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
