package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Jaisheesh-2006/schema-context-mcp/internal/contextfile"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Load and validate context.yaml
	ctx, err := contextfile.Load(cfg.ContextFile)
	if err != nil {
		slog.Error("failed to load context file", "error", err)
		os.Exit(1)
	}
	slog.Info("context file loaded", "tables", len(ctx.Tables))

	// Connect to MySQL
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	// Introspect schema
	dbSchema, err := IntrospectSchema(db)
	if err != nil {
		slog.Error("failed to introspect schema", "error", err)
		os.Exit(1)
	}
	slog.Info("schema introspected", "tables", len(dbSchema))

	// Merge DB schema with context.yaml
	merged := Merge(dbSchema, ctx)
	slog.Info("schema merged", "tables", len(merged))

	// Start MCP server
	if err := StartMCPServer(merged); err != nil {
		slog.Error("MCP server failed", "error", err)
		os.Exit(1)
	}
}
