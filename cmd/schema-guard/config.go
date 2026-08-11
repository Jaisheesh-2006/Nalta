package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
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

// LoadActionConfig parses configuration from CLI flags, environment
// variables, and GitHub Action INPUT_* variables.
//
// Precedence (highest to lowest):
//
//	CLI flag > ENV var > INPUT_* env var > default
func LoadActionConfig() (*ActionConfig, error) {
	cfg := &ActionConfig{}

	flag.StringVar(&cfg.ContextFile, "context",
		envOrInputOrDefault("CONTEXT_FILE", "INPUT_CONTEXT-PATH", "./context.yaml"),
		"Path to context.yaml")
	flag.StringVar(&cfg.BeforeMigrations, "before-migrations",
		envOrInputOrDefault("BEFORE_MIGRATIONS", "INPUT_BEFORE-MIGRATIONS", ""),
		"Path to base branch migrations")
	flag.StringVar(&cfg.AfterMigrations, "after-migrations",
		envOrInputOrDefault("AFTER_MIGRATIONS", "INPUT_AFTER-MIGRATIONS", ""),
		"Path to PR branch migrations")
	flag.StringVar(&cfg.DSN, "dsn",
		envOrInputOrDefault("DSN", "INPUT_DSN", ""),
		"MySQL DSN (if empty, spins up ephemeral container)")
	flag.StringVar(&cfg.GitHubToken, "github-token",
		envOrInputOrDefault("GITHUB_TOKEN", "INPUT_GITHUB-TOKEN", ""),
		"GitHub API token")
	flag.StringVar(&cfg.Repo, "repo",
		envOrDefault("GITHUB_REPOSITORY", ""),
		"GitHub repository (owner/repo)")
	flag.BoolVar(&cfg.DryRun, "dry-run", false,
		"Print comment to stdout instead of posting to GitHub")

	flag.Parse()

	// Resolve PR number: explicit env > auto-detect from GITHUB_REF
	if pr, err := resolvePRNumber(); err == nil {
		cfg.PRNumber = pr
	}

	return cfg, nil
}

// Validate checks that required fields are present for the current mode.
// In dry-run mode, GitHub-specific fields are not required.
func (cfg *ActionConfig) Validate() error {
	if cfg.DryRun {
		return nil
	}

	if cfg.GitHubToken == "" {
		return fmt.Errorf("--github-token or GITHUB_TOKEN is required (use --dry-run for local testing)")
	}
	if cfg.Repo == "" {
		return fmt.Errorf("--repo or GITHUB_REPOSITORY is required")
	}
	if cfg.PRNumber == 0 {
		return fmt.Errorf("PR number is required (set PR_NUMBER env var or run in a pull_request workflow)")
	}

	return nil
}

// resolvePRNumber attempts to determine the PR number from environment
// variables. It checks (in order):
//  1. PR_NUMBER env var (explicit)
//  2. GITHUB_REF (format: refs/pull/<number>/merge)
//  3. GITHUB_EVENT_PATH (JSON file with .pull_request.number)
func resolvePRNumber() (int, error) {
	// 1. Explicit PR_NUMBER env var
	if prStr := os.Getenv("PR_NUMBER"); prStr != "" {
		pr, err := strconv.Atoi(prStr)
		if err != nil {
			return 0, fmt.Errorf("invalid PR_NUMBER %q: %w", prStr, err)
		}
		return pr, nil
	}

	// 2. GITHUB_REF: refs/pull/42/merge
	if ref := os.Getenv("GITHUB_REF"); strings.HasPrefix(ref, "refs/pull/") {
		parts := strings.Split(ref, "/")
		if len(parts) >= 3 {
			pr, err := strconv.Atoi(parts[2])
			if err == nil {
				return pr, nil
			}
		}
	}

	// 3. GITHUB_EVENT_PATH: JSON file containing .pull_request.number
	if eventPath := os.Getenv("GITHUB_EVENT_PATH"); eventPath != "" {
		pr, err := prNumberFromEventFile(eventPath)
		if err == nil {
			return pr, nil
		}
	}

	return 0, fmt.Errorf("could not detect PR number")
}

// prNumberFromEventFile reads the GitHub event JSON file and extracts
// .pull_request.number. This is the most reliable method in CI.
func prNumberFromEventFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("reading event file: %w", err)
	}

	var event struct {
		PullRequest struct {
			Number int `json:"number"`
		} `json:"pull_request"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return 0, fmt.Errorf("parsing event JSON: %w", err)
	}

	if event.PullRequest.Number == 0 {
		return 0, fmt.Errorf("no pull_request.number in event payload")
	}

	return event.PullRequest.Number, nil
}

// envOrInputOrDefault resolves a value with the precedence:
// ENV var > GitHub Action INPUT_* var > fallback default.
func envOrInputOrDefault(envKey, inputKey, fallback string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	if v := os.Getenv(inputKey); v != "" {
		return v
	}
	return fallback
}

// envOrDefault returns the env var value if set, otherwise the fallback.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
