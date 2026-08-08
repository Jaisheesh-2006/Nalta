package main

import (
	"testing"

	"github.com/Jaisheesh-2006/schema-context-mcp/internal/contextfile"
)

func TestDiffSchemas_DroppedColumn(t *testing.T) {
	before := []SnapshotColumn{
		{TableName: "users", ColumnName: "id", DataType: "int"},
		{TableName: "users", ColumnName: "email", DataType: "varchar"},
	}
	after := []SnapshotColumn{
		{TableName: "users", ColumnName: "id", DataType: "int"},
	}

	diff := DiffSchemas(before, after)

	if len(diff.DroppedColumns) != 1 {
		t.Fatalf("expected 1 dropped column, got %d", len(diff.DroppedColumns))
	}
	if diff.DroppedColumns[0].Column != "email" {
		t.Errorf("expected dropped column 'email', got %q", diff.DroppedColumns[0].Column)
	}
}

func TestDiffSchemas_AddedTable(t *testing.T) {
	before := []SnapshotColumn{
		{TableName: "users", ColumnName: "id", DataType: "int"},
	}
	after := []SnapshotColumn{
		{TableName: "users", ColumnName: "id", DataType: "int"},
		{TableName: "posts", ColumnName: "id", DataType: "int"},
	}

	diff := DiffSchemas(before, after)

	if len(diff.AddedTables) != 1 {
		t.Fatalf("expected 1 added table, got %d", len(diff.AddedTables))
	}
	if diff.AddedTables[0] != "posts" {
		t.Errorf("expected added table 'posts', got %q", diff.AddedTables[0])
	}
}

func TestDiffSchemas_NoChanges(t *testing.T) {
	cols := []SnapshotColumn{
		{TableName: "users", ColumnName: "id", DataType: "int"},
	}

	diff := DiffSchemas(cols, cols)

	if len(diff.AddedTables) != 0 || len(diff.DroppedTables) != 0 ||
		len(diff.AddedColumns) != 0 || len(diff.DroppedColumns) != 0 ||
		len(diff.ChangedColumns) != 0 {
		t.Error("expected no changes for identical schemas")
	}
}

func TestCrossReference_SensitiveDropped(t *testing.T) {
	diff := SchemaDiff{
		DroppedColumns: []TableColumn{
			{Table: "ingredients", Column: "toxicity_class"},
		},
	}
	ctx := &contextfile.ContextFile{
		Version: "1",
		Tables: map[string]contextfile.TableContext{
			"ingredients": {
				Description: "Ingredients",
				Columns: map[string]contextfile.ColumnContext{
					"toxicity_class": {Description: "Safety tier", Sensitive: true},
				},
			},
		},
	}

	findings := CrossReference(diff, ctx)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != SeverityWarning {
		t.Errorf("expected warning severity, got %v", findings[0].Severity)
	}
	if !findings[0].Sensitive {
		t.Error("expected finding to be marked sensitive")
	}
}

func TestFormatComment_NoIssues(t *testing.T) {
	comment := FormatComment(nil, SchemaDiff{})
	if comment == "" {
		t.Error("expected non-empty comment")
	}
	if !contains(comment, "✅") {
		t.Error("expected green checkmark for no-issues case")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstr(s, substr)
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
