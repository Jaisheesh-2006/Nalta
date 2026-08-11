package main

import (
	"testing"

	"github.com/Jaisheesh-2006/nalta/internal/contextfile"
)

func TestMerge_FullyMatched(t *testing.T) {
	dbSchema := []TableSchema{
		{
			Name: "users",
			Columns: []DBColumn{
				{TableName: "users", ColumnName: "id", DataType: "int", IsNullable: false},
				{TableName: "users", ColumnName: "email", DataType: "varchar", IsNullable: true},
			},
			ForeignKeys: map[string]DBForeignKey{},
		},
	}
	ctx := &contextfile.ContextFile{
		Version: "1",
		Tables: map[string]contextfile.TableContext{
			"users": {
				Description: "Application users",
				Columns: map[string]contextfile.ColumnContext{
					"id":    {Description: "Primary key"},
					"email": {Description: "User email", PII: true},
				},
			},
		},
	}

	result := Merge(dbSchema, ctx)

	if len(result) != 1 {
		t.Fatalf("expected 1 table, got %d", len(result))
	}
	if result[0].Description != "Application users" {
		t.Errorf("expected table description, got %q", result[0].Description)
	}
	if len(result[0].Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result[0].Columns))
	}
	for _, col := range result[0].Columns {
		if !col.Documented {
			t.Errorf("column %q should be documented", col.Name)
		}
	}
}

func TestMerge_UndocumentedColumn(t *testing.T) {
	dbSchema := []TableSchema{
		{
			Name: "users",
			Columns: []DBColumn{
				{TableName: "users", ColumnName: "id", DataType: "int"},
				{TableName: "users", ColumnName: "created_at", DataType: "datetime"},
			},
			ForeignKeys: map[string]DBForeignKey{},
		},
	}
	ctx := &contextfile.ContextFile{
		Version: "1",
		Tables: map[string]contextfile.TableContext{
			"users": {
				Description: "Application users",
				Columns: map[string]contextfile.ColumnContext{
					"id": {Description: "Primary key"},
				},
			},
		},
	}

	result := Merge(dbSchema, ctx)
	createdAt := result[0].Columns[1]

	if createdAt.Documented {
		t.Error("created_at should not be documented")
	}
	if createdAt.Description != "" {
		t.Errorf("expected empty description for undocumented column, got %q", createdAt.Description)
	}
	if createdAt.DataType != "datetime" {
		t.Errorf("expected data type preserved, got %q", createdAt.DataType)
	}
}

func TestMerge_StaleContextColumn(t *testing.T) {
	dbSchema := []TableSchema{
		{
			Name: "users",
			Columns: []DBColumn{
				{TableName: "users", ColumnName: "id", DataType: "int"},
			},
			ForeignKeys: map[string]DBForeignKey{},
		},
	}
	ctx := &contextfile.ContextFile{
		Version: "1",
		Tables: map[string]contextfile.TableContext{
			"users": {
				Description: "Application users",
				Columns: map[string]contextfile.ColumnContext{
					"id":        {Description: "Primary key"},
					"old_field": {Description: "This column was removed"},
				},
			},
		},
	}

	result := Merge(dbSchema, ctx)

	// old_field should NOT appear in merged output — only DB columns appear
	if len(result[0].Columns) != 1 {
		t.Errorf("expected 1 column (DB only), got %d", len(result[0].Columns))
	}
}

func TestMerge_StaleContextTable(t *testing.T) {
	dbSchema := []TableSchema{
		{
			Name: "users",
			Columns: []DBColumn{
				{TableName: "users", ColumnName: "id", DataType: "int"},
			},
			ForeignKeys: map[string]DBForeignKey{},
		},
	}
	ctx := &contextfile.ContextFile{
		Version: "1",
		Tables: map[string]contextfile.TableContext{
			"users": {
				Description: "Application users",
				Columns: map[string]contextfile.ColumnContext{
					"id": {Description: "Primary key"},
				},
			},
			"deleted_table": {
				Description: "This table no longer exists",
				Columns:     map[string]contextfile.ColumnContext{},
			},
		},
	}

	result := Merge(dbSchema, ctx)

	// deleted_table should NOT appear in output
	if len(result) != 1 {
		t.Errorf("expected 1 table (DB only), got %d", len(result))
	}
}
