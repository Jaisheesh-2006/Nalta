package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v60/github"
)

// FormatComment renders the PR comment markdown from findings and diff.
func FormatComment(findings []Finding, diff SchemaDiff) string {
	var warnings, infos []Finding
	for _, f := range findings {
		if f.Severity == SeverityWarning {
			warnings = append(warnings, f)
		} else {
			infos = append(infos, f)
		}
	}

	// No issues at all
	if len(warnings) == 0 && len(infos) == 0 {
		return "## ✅ Schema Guard Report\n\nMigration changes detected. No documented or sensitive columns were affected.\n"
	}

	var sb strings.Builder
	sb.WriteString("## 🔍 Schema Guard Report\n\n")
	sb.WriteString("**Migration changes detected in this PR.**\n\n")

	if len(warnings) > 0 {
		sb.WriteString("### ⚠️ Attention Required\n\n")
		sb.WriteString("| Change | Table | Column | Concern |\n")
		sb.WriteString("|--------|-------|--------|---------|\n")
		for _, f := range warnings {
			badge := ""
			if f.PII {
				badge = "🔒 "
			}
			sb.WriteString(fmt.Sprintf("| %s | `%s` | `%s` | %s**%s** |\n",
				f.Change, f.Table, f.Column, badge, f.Reason))
		}
		sb.WriteString("\n")
	}

	if len(infos) > 0 {
		sb.WriteString("### ℹ️ Other Changes\n\n")
		sb.WriteString("| Change | Table | Column | Status |\n")
		sb.WriteString("|--------|-------|--------|--------|\n")
		for _, f := range infos {
			sb.WriteString(fmt.Sprintf("| %s | `%s` | `%s` | %s |\n",
				f.Change, f.Table, f.Column, f.Reason))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n*Posted by schema-guard-action. Update `context.yaml` if needed.*\n")
	return sb.String()
}

// PostComment posts a PR comment via the GitHub API.
func PostComment(token, repo string, prNumber int, body string) error {
	if token == "" {
		return fmt.Errorf("GitHub token is required")
	}

	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format, expected owner/repo: %q", repo)
	}
	owner, repoName := parts[0], parts[1]

	client := github.NewClient(nil).WithAuthToken(token)

	_, _, err := client.Issues.CreateComment(
		context.Background(),
		owner,
		repoName,
		prNumber,
		&github.IssueComment{Body: github.String(body)},
	)
	return err
}
