package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Jaisheesh-2006/nalta/internal/contextfile"
)

func main() {
	cfg, err := LoadActionConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Validate required fields for the current mode
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	// Load context.yaml
	ctx, err := contextfile.Load(cfg.ContextFile)
	if err != nil {
		slog.Error("failed to load context file", "error", err)
		os.Exit(1)
	}

	// Snapshot before and after schemas
	before, after, err := SnapshotSchemas(cfg)
	if err != nil {
		slog.Error("failed to snapshot schemas", "error", err)
		os.Exit(1)
	}

	// Compute diff
	diff := DiffSchemas(before, after)
	slog.Info("schema diff computed",
		"added_tables", len(diff.AddedTables),
		"dropped_tables", len(diff.DroppedTables),
		"added_columns", len(diff.AddedColumns),
		"dropped_columns", len(diff.DroppedColumns),
		"changed_columns", len(diff.ChangedColumns),
	)

	// Cross-reference with context.yaml
	findings := CrossReference(diff, ctx)

	// Format the PR comment
	comment := FormatComment(findings, diff)

	if cfg.DryRun {
		fmt.Println(comment)
		return
	}

	// Post to GitHub (updates existing comment if one exists)
	if err := PostOrUpdateComment(cfg.GitHubToken, cfg.Repo, cfg.PRNumber, comment); err != nil {
		slog.Error("failed to post comment", "error", err)
		os.Exit(1)
	}

	slog.Info("schema guard report posted")
}
