package main

import (
	"github.com/Jaisheesh-2006/nalta/internal/contextfile"
)

// FindingSeverity indicates how critical a schema change is.
type FindingSeverity string

const (
	SeverityWarning FindingSeverity = "warning" // documented or sensitive column affected
	SeverityInfo    FindingSeverity = "info"    // undocumented column changed
)

// Finding represents a single cross-reference result.
type Finding struct {
	Change    string          // "ADDED", "DROPPED", "TYPE CHANGED"
	Table     string
	Column    string
	Severity  FindingSeverity
	Reason    string
	Sensitive bool
	PII       bool
}

// CrossReference checks each entry in the schema diff against context.yaml
// and produces findings for anything documented or sensitive.
func CrossReference(diff SchemaDiff, ctx *contextfile.ContextFile) []Finding {
	var findings []Finding

	// Check dropped columns
	for _, col := range diff.DroppedColumns {
		f := Finding{Change: "DROPPED", Table: col.Table, Column: col.Column}
		if cc, ok := lookupColumn(ctx, col.Table, col.Column); ok {
			f.Sensitive = cc.Sensitive
			f.PII = cc.PII
			if cc.Sensitive {
				f.Severity = SeverityWarning
				f.Reason = "Sensitive column removed — context.yaml and downstream agents reference this."
			} else if cc.PII {
				f.Severity = SeverityWarning
				f.Reason = "PII column removed — verify compliance implications."
			} else {
				f.Severity = SeverityWarning
				f.Reason = "Documented column removed — update context.yaml."
			}
		} else {
			f.Severity = SeverityInfo
			f.Reason = "Undocumented column removed."
		}
		findings = append(findings, f)
	}

	// Check changed columns
	for _, col := range diff.ChangedColumns {
		f := Finding{Change: "TYPE CHANGED", Table: col.Table, Column: col.Column}
		if cc, ok := lookupColumn(ctx, col.Table, col.Column); ok {
			f.Sensitive = cc.Sensitive
			f.PII = cc.PII
			f.Severity = SeverityWarning
			f.Reason = "Documented column type changed — verify context.yaml still accurate."
		} else {
			f.Severity = SeverityInfo
			f.Reason = "Undocumented column type changed."
		}
		findings = append(findings, f)
	}

	// Check added columns (informational)
	for _, col := range diff.AddedColumns {
		if _, ok := lookupColumn(ctx, col.Table, col.Column); !ok {
			findings = append(findings, Finding{
				Change:   "ADDED",
				Table:    col.Table,
				Column:   col.Column,
				Severity: SeverityInfo,
				Reason:   "Not yet documented in context.yaml.",
			})
		}
	}

	return findings
}

func lookupColumn(ctx *contextfile.ContextFile, table, column string) (*contextfile.ColumnContext, bool) {
	tc, ok := ctx.Tables[table]
	if !ok {
		return nil, false
	}
	cc, ok := tc.Columns[column]
	if !ok {
		return nil, false
	}
	return &cc, true
}
