package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEnvOrInputOrDefault_EnvTakesPrecedence(t *testing.T) {
	t.Setenv("TEST_ENV_KEY", "from-env")
	t.Setenv("INPUT_TEST-KEY", "from-input")

	got := envOrInputOrDefault("TEST_ENV_KEY", "INPUT_TEST-KEY", "default")
	if got != "from-env" {
		t.Errorf("expected 'from-env', got %q", got)
	}
}

func TestEnvOrInputOrDefault_InputFallback(t *testing.T) {
	// TEST_ENV_KEY is not set (t.Setenv not called for it)
	t.Setenv("INPUT_TEST-KEY", "from-input")

	got := envOrInputOrDefault("TEST_ENV_KEY_UNSET_"+t.Name(), "INPUT_TEST-KEY", "default")
	if got != "from-input" {
		t.Errorf("expected 'from-input', got %q", got)
	}
}

func TestEnvOrInputOrDefault_DefaultFallback(t *testing.T) {
	got := envOrInputOrDefault(
		"UNSET_ENV_"+t.Name(),
		"UNSET_INPUT_"+t.Name(),
		"my-default",
	)
	if got != "my-default" {
		t.Errorf("expected 'my-default', got %q", got)
	}
}

func TestValidate_DryRunSkipsValidation(t *testing.T) {
	cfg := &ActionConfig{DryRun: true}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error in dry-run mode, got: %v", err)
	}
}

func TestValidate_MissingGitHubToken(t *testing.T) {
	cfg := &ActionConfig{
		DryRun:   false,
		Repo:     "owner/repo",
		PRNumber: 42,
		// GitHubToken is empty
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing GitHub token")
	}
	if !contains(err.Error(), "github-token") {
		t.Errorf("error should mention github-token, got: %v", err)
	}
}

func TestValidate_MissingRepo(t *testing.T) {
	cfg := &ActionConfig{
		DryRun:      false,
		GitHubToken: "ghp_xxx",
		PRNumber:    42,
		// Repo is empty
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
	if !contains(err.Error(), "GITHUB_REPOSITORY") {
		t.Errorf("error should mention GITHUB_REPOSITORY, got: %v", err)
	}
}

func TestValidate_MissingPRNumber(t *testing.T) {
	cfg := &ActionConfig{
		DryRun:      false,
		GitHubToken: "ghp_xxx",
		Repo:        "owner/repo",
		// PRNumber is 0
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing PR number")
	}
	if !contains(err.Error(), "PR number") {
		t.Errorf("error should mention PR number, got: %v", err)
	}
}

func TestValidate_AllPresent(t *testing.T) {
	cfg := &ActionConfig{
		DryRun:      false,
		GitHubToken: "ghp_xxx",
		Repo:        "owner/repo",
		PRNumber:    42,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error with all fields present, got: %v", err)
	}
}

func TestResolvePRNumber_FromEnv(t *testing.T) {
	t.Setenv("PR_NUMBER", "99")
	t.Setenv("GITHUB_REF", "")
	t.Setenv("GITHUB_EVENT_PATH", "")

	pr, err := resolvePRNumber()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr != 99 {
		t.Errorf("expected PR 99, got %d", pr)
	}
}

func TestResolvePRNumber_FromGitHubRef(t *testing.T) {
	t.Setenv("PR_NUMBER", "")
	t.Setenv("GITHUB_REF", "refs/pull/42/merge")
	t.Setenv("GITHUB_EVENT_PATH", "")

	pr, err := resolvePRNumber()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr != 42 {
		t.Errorf("expected PR 42, got %d", pr)
	}
}

func TestResolvePRNumber_FromEventFile(t *testing.T) {
	t.Setenv("PR_NUMBER", "")
	t.Setenv("GITHUB_REF", "")

	// Create a temporary event JSON file
	event := map[string]interface{}{
		"pull_request": map[string]interface{}{
			"number": 77,
		},
	}
	data, _ := json.Marshal(event)
	eventFile := filepath.Join(t.TempDir(), "event.json")
	os.WriteFile(eventFile, data, 0644)

	t.Setenv("GITHUB_EVENT_PATH", eventFile)

	pr, err := resolvePRNumber()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr != 77 {
		t.Errorf("expected PR 77, got %d", pr)
	}
}

func TestResolvePRNumber_NoneAvailable(t *testing.T) {
	t.Setenv("PR_NUMBER", "")
	t.Setenv("GITHUB_REF", "refs/heads/main")
	t.Setenv("GITHUB_EVENT_PATH", "")

	_, err := resolvePRNumber()
	if err == nil {
		t.Fatal("expected error when no PR number is available")
	}
}

func TestResolvePRNumber_InvalidEnvValue(t *testing.T) {
	t.Setenv("PR_NUMBER", "not-a-number")
	t.Setenv("GITHUB_REF", "")
	t.Setenv("GITHUB_EVENT_PATH", "")

	_, err := resolvePRNumber()
	if err == nil {
		t.Fatal("expected error for invalid PR_NUMBER")
	}
}
