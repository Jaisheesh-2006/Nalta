package main

import (
	"log/slog"

	"github.com/Jaisheesh-2006/nalta/internal/contextfile"
)

// ForeignKey represents a foreign key reference from the DB.
type ForeignKey struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

// MergedColumn combines DB metadata with context.yaml annotations.
type MergedColumn struct {
	Name         string      `json:"name"`
	DataType     string      `json:"data_type"`
	Nullable     bool        `json:"nullable"`
	DefaultValue string      `json:"default_value"`
	References   *ForeignKey `json:"references"`
	Description  string      `json:"description"`
	Sensitive    bool        `json:"sensitive"`
	PII          bool        `json:"pii"`
	Documented   bool        `json:"documented"`
}

// MergedTable combines DB table metadata with context.yaml annotations.
type MergedTable struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Sensitive   bool           `json:"sensitive"`
	Columns     []MergedColumn `json:"columns"`
}

// Merge combines introspected DB schema with parsed context.yaml.
// Implements the four merge rules from architecture.md §2.4.
func Merge(dbSchema []TableSchema, ctx *contextfile.ContextFile) []MergedTable {
	var result []MergedTable

	// Build a set of DB table names for stale-reference detection
	dbTableSet := make(map[string]bool)
	for _, ts := range dbSchema {
		dbTableSet[ts.Name] = true
	}

	// Warn about context.yaml tables not in DB
	for tName := range ctx.Tables {
		if !dbTableSet[tName] {
			slog.Warn("context.yaml references table not in database — skipping",
				"table", tName)
		}
	}

	for _, ts := range dbSchema {
		mt := MergedTable{
			Name: ts.Name,
		}

		// Merge table-level context
		if tc, ok := ctx.Tables[ts.Name]; ok {
			mt.Description = tc.Description
			mt.Sensitive = tc.Sensitive
		}

		// Build a set of DB column names for this table
		dbColSet := make(map[string]bool)
		for _, col := range ts.Columns {
			dbColSet[col.ColumnName] = true
		}

		// Warn about context.yaml columns not in DB
		if tc, ok := ctx.Tables[ts.Name]; ok {
			for cName := range tc.Columns {
				if !dbColSet[cName] {
					slog.Warn("context.yaml references column not in database — skipping",
						"table", ts.Name, "column", cName)
				}
			}
		}

		// Merge each DB column
		for _, col := range ts.Columns {
			mc := MergedColumn{
				Name:         col.ColumnName,
				DataType:     col.DataType,
				Nullable:     col.IsNullable,
				DefaultValue: col.DefaultValue,
			}

			// Attach FK reference if present
			if fk, ok := ts.ForeignKeys[col.ColumnName]; ok {
				mc.References = &ForeignKey{
					Table:  fk.ReferencedTableName,
					Column: fk.ReferencedColumnName,
				}
			}

			// Merge context.yaml annotations
			if tc, ok := ctx.Tables[ts.Name]; ok {
				if cc, ok := tc.Columns[col.ColumnName]; ok {
					mc.Description = cc.Description
					mc.Sensitive = cc.Sensitive
					mc.PII = cc.PII
					mc.Documented = true
				}
			}

			mt.Columns = append(mt.Columns, mc)
		}

		result = append(result, mt)
	}

	return result
}
