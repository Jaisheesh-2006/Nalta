package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Jaisheesh-2006/nalta/internal/contextfile"
	"github.com/Jaisheesh-2006/nalta/internal/server"
)

func main() {
	cfg, err := server.LoadConfig()
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
	dbSchema, err := server.IntrospectSchema(context.Background(), db)
	if err != nil {
		slog.Error("failed to introspect schema", "error", err)
		os.Exit(1)
	}
	slog.Info("schema introspected", "tables", len(dbSchema))

	// Merge DB schema with context.yaml
	merged := server.Merge(dbSchema, ctx)
	slog.Info("schema merged", "tables", len(merged))

	// Dump column if requested
	if cfg.DumpColumn != "" {
		parts := strings.SplitN(cfg.DumpColumn, ":", 2)
		if len(parts) != 2 {
			slog.Error("invalid --dump-column format, expected table:column", "val", cfg.DumpColumn)
			os.Exit(1)
		}
		out, err := server.ExplainColumnJSON(merged, parts[0], parts[1])
		if err != nil {
			slog.Error("failed to explain column", "error", err)
			os.Exit(1)
		}
		fmt.Println(out)
		os.Exit(0)
	}

	// Dump schema if requested
	if cfg.DumpSchema != "" {
		wrapped := struct {
			Tables []server.MergedTable `json:"tables"`
		}{Tables: merged}
		b, err := json.MarshalIndent(wrapped, "", "  ")
		if err != nil {
			slog.Error("failed to marshal merged schema", "error", err)
			os.Exit(1)
		}
		if cfg.DumpSchema == "-" {
			fmt.Println(string(b))
		} else {
			if err := os.WriteFile(cfg.DumpSchema, b, 0644); err != nil {
				slog.Error("failed to write schema", "error", err, "file", cfg.DumpSchema)
				os.Exit(1)
			}
		}
		os.Exit(0)
	}

	// Start MCP server
	if err := server.StartMCPServer(merged); err != nil {
		slog.Error("MCP server failed", "error", err)
		os.Exit(1)
	}
}
