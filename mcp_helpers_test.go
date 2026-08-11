package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExplainTableJSON_and_ExplainColumnJSON(t *testing.T) {
	merged := []MergedTable{
		{
			Name:        "ingredients",
			Description: "Raw chemical or natural ingredients used in product formulas.",
			Columns: []MergedColumn{
				{Name: "id", DataType: "integer", Nullable: false, DefaultValue: "auto_increment", Description: "Auto-generated primary key.", Documented: true},
				{Name: "toxicity_class", DataType: "text", Nullable: true, DefaultValue: "", Description: "Internal safety tier: 'safe', 'restricted', or 'banned'.", Sensitive: true, Documented: true},
			},
		},
	}

	// Table JSON
	tj, err := ExplainTableJSON(merged, "ingredients")
	require.NoError(t, err)
	var tjObj MergedTable
	require.NoError(t, json.Unmarshal([]byte(tj), &tjObj))
	require.Equal(t, "ingredients", tjObj.Name)

	// Column JSON
	cj, err := ExplainColumnJSON(merged, "ingredients", "toxicity_class")
	require.NoError(t, err)
	var cjObj struct {
		Table  string       `json:"table"`
		Column MergedColumn `json:"column"`
	}
	require.NoError(t, json.Unmarshal([]byte(cj), &cjObj))
	require.Equal(t, "ingredients", cjObj.Table)
	require.Equal(t, "toxicity_class", cjObj.Column.Name)
}
