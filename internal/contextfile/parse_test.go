package contextfile

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "context.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_ValidYAML(t *testing.T) {
	yaml := `
version: "1"
tables:
  users:
    description: "App users"
    columns:
      id:
        description: "Primary key"
      email:
        description: "User email"
        pii: true
`
	cf, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cf.Version != "1" {
		t.Errorf("expected version 1, got %s", cf.Version)
	}
	if len(cf.Tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(cf.Tables))
	}
	if !cf.Tables["users"].Columns["email"].PII {
		t.Error("expected email column to be PII")
	}
}

func TestLoad_MissingVersion(t *testing.T) {
	yaml := `
tables:
  users:
    description: "App users"
    columns:
      id:
        description: "Primary key"
`
	_, err := Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestLoad_EmptyTables(t *testing.T) {
	yaml := `
version: "1"
tables: {}
`
	_, err := Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("expected error for empty tables")
	}
}

func TestLoad_MissingColumnDescription(t *testing.T) {
	yaml := `
version: "1"
tables:
  users:
    description: "App users"
    columns:
      id:
        sensitive: true
`
	_, err := Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("expected error for missing column description")
	}
}

func TestLoad_InvalidIdentifier(t *testing.T) {
	yaml := `
version: "1"
tables:
  "123-bad":
    description: "Bad table name"
    columns:
      id:
        description: "Primary key"
`
	_, err := Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("expected error for invalid identifier")
	}
}
