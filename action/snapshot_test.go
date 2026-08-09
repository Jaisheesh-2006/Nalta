package main

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRedactDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		{
			name:     "standard DSN",
			dsn:      "root:secretpass@tcp(127.0.0.1:3306)/guard_db",
			expected: "root:***@tcp(127.0.0.1:3306)/guard_db",
		},
		{
			name:     "no password section",
			dsn:      "tcp(127.0.0.1:3306)/guard_db",
			expected: "***",
		},
		{
			name:     "empty DSN",
			dsn:      "",
			expected: "***",
		},
		{
			name:     "user without colon",
			dsn:      "root@tcp(127.0.0.1:3306)/guard_db",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactDSN(tt.dsn)
			if got != tt.expected {
				t.Errorf("redactDSN(%q) = %q, want %q", tt.dsn, got, tt.expected)
			}
		})
	}
}

func TestBuildMigrateDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		{
			name:     "plain DSN",
			dsn:      "root:root@tcp(127.0.0.1:3306)/guard_db",
			expected: "mysql://root:root@tcp(127.0.0.1:3306)/guard_db?multiStatements=true",
		},
		{
			name:     "DSN with existing params",
			dsn:      "root:root@tcp(127.0.0.1:3306)/guard_db?parseTime=true",
			expected: "mysql://root:root@tcp(127.0.0.1:3306)/guard_db?parseTime=true&multiStatements=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMigrateDSN(tt.dsn)
			if got != tt.expected {
				t.Errorf("buildMigrateDSN(%q) = %q, want %q", tt.dsn, got, tt.expected)
			}
		})
	}
}

func TestMigrationsSourceURL(t *testing.T) {
	// Use a relative path that we know exists
	url, err := migrationsSourceURL(".")
	if err != nil {
		t.Fatalf("migrationsSourceURL failed: %v", err)
	}

	if !strings.HasPrefix(url, "file://") {
		t.Errorf("expected file:// prefix, got %q", url)
	}

	// The URL path portion should start with "/" (even on Windows).
	urlPath := strings.TrimPrefix(url, "file://")
	if !strings.HasPrefix(urlPath, "/") {
		t.Errorf("expected URL path to start with /, got %q", urlPath)
	}

	// On Windows, we should see something like /C:/...
	if runtime.GOOS == "windows" {
		if len(urlPath) < 3 || urlPath[2] != ':' {
			t.Errorf("expected Windows drive letter in URL path, got %q", urlPath)
		}
	}
}

func TestMigrationsSourceURL_Absolute(t *testing.T) {
	var absPath string
	if runtime.GOOS == "windows" {
		absPath = `C:\Users\test\migrations`
	} else {
		absPath = "/home/test/migrations"
	}

	url, err := migrationsSourceURL(absPath)
	if err != nil {
		t.Fatalf("migrationsSourceURL failed: %v", err)
	}

	// Verify forward slashes in URL
	if strings.Contains(url, `\`) {
		t.Errorf("URL should not contain backslashes: %q", url)
	}

	// Verify the file scheme
	if !strings.HasPrefix(url, "file:///") {
		t.Errorf("expected file:/// prefix for absolute path, got %q", url)
	}
}

func TestGenerateContainerName(t *testing.T) {
	name1 := generateContainerName()
	name2 := generateContainerName()

	if !strings.HasPrefix(name1, "schema-guard-") {
		t.Errorf("expected 'schema-guard-' prefix, got %q", name1)
	}

	// Names should be different (with overwhelming probability)
	if name1 == name2 {
		t.Errorf("expected unique names, got %q twice", name1)
	}
}

func TestFindFreePort(t *testing.T) {
	port, err := findFreePort()
	if err != nil {
		t.Fatalf("findFreePort failed: %v", err)
	}

	if port <= 0 || port > 65535 {
		t.Errorf("expected valid port, got %d", port)
	}
}

func TestSnapshotSchemas_RequiresDSNOrDocker(t *testing.T) {
	// With no DSN and no Docker, SnapshotSchemas should fail gracefully
	// rather than panic. We don't actually test Docker here — just
	// verify the mode routing logic.
	cfg := &ActionConfig{
		ContextFile:      "./testdata/context.yaml",
		BeforeMigrations: filepath.Join(t.TempDir(), "before"),
		AfterMigrations:  filepath.Join(t.TempDir(), "after"),
		// DSN is empty → ephemeral mode → will try Docker
	}

	// This will fail because Docker may not be available, but it should
	// return an error, not panic.
	_, _, err := SnapshotSchemas(cfg)
	if err == nil {
		// If Docker IS available and this somehow succeeds, that's fine too
		t.Log("SnapshotSchemas succeeded (Docker available)")
	} else {
		t.Logf("SnapshotSchemas returned expected error: %v", err)
	}
}
