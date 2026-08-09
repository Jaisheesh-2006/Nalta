package main

import (
	"encoding/json"
	"fmt"
)

// ExplainTableJSON returns a JSON string describing the named table in merged.
// Returns an error if the table is not found.
func ExplainTableJSON(merged []MergedTable, table string) (string, error) {
	for _, t := range merged {
		if t.Name == table {
			b, err := json.Marshal(t)
			if err != nil {
				return "", fmt.Errorf("marshal table %q: %w", table, err)
			}
			return string(b), nil
		}
	}
	return "", fmt.Errorf("table %q not found in merged schema", table)
}

// ExplainColumnJSON returns a JSON string describing a specific column within
// the named table in merged. Returns an error if the table or column is not found.
func ExplainColumnJSON(merged []MergedTable, table, column string) (string, error) {
	for _, t := range merged {
		if t.Name != table {
			continue
		}
		for _, c := range t.Columns {
			if c.Name != column {
				continue
			}
			result := struct {
				Table  string       `json:"table"`
				Column MergedColumn `json:"column"`
			}{
				Table:  table,
				Column: c,
			}
			b, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("marshal column %q.%q: %w", table, column, err)
			}
			return string(b), nil
		}
		return "", fmt.Errorf("column %q not found in table %q", column, table)
	}
	return "", fmt.Errorf("table %q not found in merged schema", table)
}
