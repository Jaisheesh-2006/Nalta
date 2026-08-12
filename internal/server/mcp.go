package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StartMCPServer creates and starts an MCP server over stdio transport,
// exposing the merged schema as a resource and explain_column as a tool.
func StartMCPServer(merged []MergedTable) error {
	s := server.NewMCPServer(
		"schema-context-mcp",
		"1.0.0",
	)

	// Register the schema://full resource
	schemaResource := mcp.NewResource(
		"schema://full",
		"Full merged database schema with context annotations",
		mcp.WithResourceDescription("Returns every table and column with DB metadata and context.yaml annotations merged together."),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(schemaResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		payload := struct {
			Tables []MergedTable `json:"tables"`
		}{Tables: merged}

		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshalling schema: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "schema://full",
				MIMEType: "application/json",
				Text:     string(data),
			},
		}, nil
	})

	// Register the explain_column tool
	explainTool := mcp.NewTool(
		"explain_column",
		mcp.WithDescription("Look up a single column's metadata, type info, and context annotations."),
		mcp.WithString("table", mcp.Required(), mcp.Description("Table name")),
		mcp.WithString("column", mcp.Required(), mcp.Description("Column name")),
	)

	s.AddTool(explainTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tableName, err := request.RequireString("table")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		columnName, err := request.RequireString("column")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Find the table
		var foundTable *MergedTable
		for i := range merged {
			if merged[i].Name == tableName {
				foundTable = &merged[i]
				break
			}
		}
		if foundTable == nil {
			return mcp.NewToolResultError(fmt.Sprintf("table '%s' not found", tableName)), nil
		}

		// Find the column
		var foundColumn *MergedColumn
		for i := range foundTable.Columns {
			if foundTable.Columns[i].Name == columnName {
				foundColumn = &foundTable.Columns[i]
				break
			}
		}
		if foundColumn == nil {
			return mcp.NewToolResultError(fmt.Sprintf("column '%s' not found in table '%s'", columnName, tableName)), nil
		}

		// Build response
		response := struct {
			Table  string       `json:"table"`
			Column MergedColumn `json:"column"`
		}{
			Table:  tableName,
			Column: *foundColumn,
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshalling column: %w", err)
		}

		return mcp.NewToolResultText(string(data)), nil
	})

	// Start stdio server
	return server.ServeStdio(s)
}
