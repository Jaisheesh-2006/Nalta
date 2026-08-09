//go:build integration

// Package main — Integration tests for the schema-guard action pipeline.
//
// These tests require Docker to be running and are excluded from the
// default `go test ./...` run. Execute them explicitly with:
//
//	go test -tags=integration -v -count=1 ./action/
//
// They exercise the full pipeline: ephemeral container → migrations →
// snapshot → diff → crossref → comment formatting.
package main

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/Jaisheesh-2006/schema-context-mcp/internal/contextfile"
)

// skipIfNoDocker skips the test if Docker is not available on the host.
func skipIfNoDocker(t *testing.T) {
	t.Helper()
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker not available — skipping integration test")
	}
}

// TestE2E_FullPipeline runs the complete schema-guard pipeline using
// the Synthea example migrations. The "before" directory has migrations
// 001–016 and the "after" directory has 001–017 (017 drops patients.PASSPORT
// and adds an audit_log table).
//
// Expected findings:
//   - patients.PASSPORT DROPPED → ⚠️ Warning (PII column, documented)
//   - audit_log columns ADDED  → ℹ️ Info (undocumented new table)
func TestE2E_FullPipeline(t *testing.T) {
	skipIfNoDocker(t)

	cfg := &ActionConfig{
		ContextFile:      "../examples/synthea/context.yaml",
		BeforeMigrations: "../examples/synthea/migrations",
		AfterMigrations:  "../examples/synthea/test-pr-migrations",
		DryRun:           true,
	}

	// ---- Step 1: Snapshot schemas via ephemeral container ----
	t.Log("Starting ephemeral MySQL container and applying migrations...")
	before, after, err := SnapshotSchemas(cfg)
	if err != nil {
		t.Fatalf("SnapshotSchemas failed: %v", err)
	}

	t.Logf("Before: %d columns, After: %d columns", len(before), len(after))

	if len(before) == 0 {
		t.Fatal("before snapshot has 0 columns — migrations likely failed")
	}
	if len(after) == 0 {
		t.Fatal("after snapshot has 0 columns — migrations likely failed")
	}

	// ---- Step 2: Compute diff ----
	diff := DiffSchemas(before, after)

	t.Logf("Diff: +%d tables, -%d tables, +%d cols, -%d cols, ~%d cols",
		len(diff.AddedTables), len(diff.DroppedTables),
		len(diff.AddedColumns), len(diff.DroppedColumns),
		len(diff.ChangedColumns))

	// Verify PASSPORT was detected as dropped
	passportDropped := false
	for _, col := range diff.DroppedColumns {
		if col.Table == "patients" && col.Column == "PASSPORT" {
			passportDropped = true
			break
		}
	}
	if !passportDropped {
		t.Error("expected patients.PASSPORT to appear in DroppedColumns")
	}

	// Verify audit_log was detected as a new table
	auditLogAdded := false
	for _, table := range diff.AddedTables {
		if table == "audit_log" {
			auditLogAdded = true
			break
		}
	}
	if !auditLogAdded {
		t.Error("expected audit_log to appear in AddedTables")
	}

	// ---- Step 3: Cross-reference with context.yaml ----
	ctx, err := contextfile.Load(cfg.ContextFile)
	if err != nil {
		t.Fatalf("failed to load context.yaml: %v", err)
	}

	findings := CrossReference(diff, ctx)

	t.Logf("Findings: %d total", len(findings))
	for _, f := range findings {
		t.Logf("  [%s] %s.%s — %s (PII=%v, Sensitive=%v)",
			f.Severity, f.Table, f.Column, f.Change, f.PII, f.Sensitive)
	}

	// Verify PASSPORT finding is a warning with PII flag
	passportWarning := false
	for _, f := range findings {
		if f.Table == "patients" && f.Column == "PASSPORT" {
			if f.Severity != SeverityWarning {
				t.Errorf("PASSPORT finding should be warning, got %v", f.Severity)
			}
			if !f.PII {
				t.Error("PASSPORT finding should have PII=true")
			}
			passportWarning = true
			break
		}
	}
	if !passportWarning {
		t.Error("expected a finding for patients.PASSPORT")
	}

	// Verify audit_log columns appear as info (undocumented)
	auditLogInfoCount := 0
	for _, f := range findings {
		if f.Table == "audit_log" && f.Severity == SeverityInfo {
			auditLogInfoCount++
		}
	}
	if auditLogInfoCount == 0 {
		t.Error("expected info findings for undocumented audit_log columns")
	}

	// ---- Step 4: Format comment ----
	comment := FormatComment(findings, diff)

	t.Logf("Generated comment (%d chars):\n%s", len(comment), comment)

	// Verify the comment contains the dedup marker
	if !strings.Contains(comment, commentMarker) {
		t.Error("comment missing dedup marker")
	}

	// Verify PII badge appears for PASSPORT
	if !strings.Contains(comment, "🔒") {
		t.Error("comment missing PII badge 🔒 for PASSPORT column")
	}

	// Verify the "Attention Required" section exists (warnings present)
	if !strings.Contains(comment, "Attention Required") {
		t.Error("comment missing Attention Required section")
	}

	// Verify the "Other Changes" section exists (info items present)
	if !strings.Contains(comment, "Other Changes") {
		t.Error("comment missing Other Changes section")
	}

	// Verify audit_log appears in the comment
	if !strings.Contains(comment, "audit_log") {
		t.Error("comment missing audit_log table reference")
	}
}

// TestE2E_NoChanges verifies that identical before/after migrations
// produce a clean "no issues" report.
func TestE2E_NoChanges(t *testing.T) {
	skipIfNoDocker(t)

	cfg := &ActionConfig{
		ContextFile:      "../examples/synthea/context.yaml",
		BeforeMigrations: "../examples/synthea/migrations",
		AfterMigrations:  "../examples/synthea/migrations", // Same dir = no changes
		DryRun:           true,
	}

	before, after, err := SnapshotSchemas(cfg)
	if err != nil {
		t.Fatalf("SnapshotSchemas failed: %v", err)
	}

	diff := DiffSchemas(before, after)

	if len(diff.AddedTables) != 0 || len(diff.DroppedTables) != 0 ||
		len(diff.AddedColumns) != 0 || len(diff.DroppedColumns) != 0 ||
		len(diff.ChangedColumns) != 0 {
		t.Errorf("expected no diff for identical migrations, got: "+
			"+%d tables, -%d tables, +%d cols, -%d cols, ~%d cols",
			len(diff.AddedTables), len(diff.DroppedTables),
			len(diff.AddedColumns), len(diff.DroppedColumns),
			len(diff.ChangedColumns))
	}

	comment := FormatComment(nil, diff)

	if !strings.Contains(comment, "✅") {
		t.Error("expected ✅ checkmark for no-changes case")
	}
	if !strings.Contains(comment, commentMarker) {
		t.Error("no-changes comment missing dedup marker")
	}
}
