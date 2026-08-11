package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/go-github/v60/github"
)

// commentMarker is a hidden HTML tag embedded in every Schema Guard
// comment. It allows PostOrUpdateComment to find and update an existing
// comment instead of posting duplicates on re-runs.
const commentMarker = "<!-- schema-guard-action -->"

// FormatComment renders the PR comment markdown from findings and diff.
// The comment includes the hidden marker for deduplication.
func FormatComment(findings []Finding, diff SchemaDiff) string {
	var warnings, infos []Finding
	for _, f := range findings {
		if f.Severity == SeverityWarning {
			warnings = append(warnings, f)
		} else {
			infos = append(infos, f)
		}
	}

	var sb strings.Builder

	// Always start with the hidden marker for deduplication
	sb.WriteString(commentMarker + "\n")

	// No issues at all
	if len(warnings) == 0 && len(infos) == 0 {
		sb.WriteString("## ✅ Schema Guard Report\n\nMigration changes detected. No documented or sensitive columns were affected.\n")
		return sb.String()
	}

	sb.WriteString("## 🔍 Schema Guard Report\n\n")
	sb.WriteString("**Migration changes detected in this PR.**\n\n")

	if len(warnings) > 0 {
		sb.WriteString("### ⚠️ Attention Required\n\n")
		sb.WriteString("| Change | Table | Column | Concern |\n")
		sb.WriteString("|--------|-------|--------|--------|\n")
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

// PostOrUpdateComment posts a new PR comment or updates an existing one.
// It searches for a previous Schema Guard comment (identified by the
// hidden HTML marker) and edits it in place if found. This prevents
// duplicate comments when the action re-runs on force-pushes or retries.
func PostOrUpdateComment(token, repo string, prNumber int, body string) error {
	if token == "" {
		return fmt.Errorf("GitHub token is required")
	}

	owner, repoName, err := parseRepo(repo)
	if err != nil {
		return err
	}

	client := github.NewClient(nil).WithAuthToken(token)
	ctx := context.Background()

	// Search for an existing Schema Guard comment
	existing, err := findExistingComment(ctx, client, owner, repoName, prNumber)
	if err != nil {
		// Non-fatal: if we can't search, fall back to creating a new comment.
		slog.Warn("failed to search for existing comment, will create new",
			"error", err)
	}

	if existing != nil {
		// Update the existing comment in place
		_, _, err := client.Issues.EditComment(
			ctx, owner, repoName,
			existing.GetID(),
			&github.IssueComment{Body: github.String(body)},
		)
		if err != nil {
			return fmt.Errorf("updating existing comment: %w", err)
		}
		slog.Info("updated existing schema guard comment",
			"comment_id", existing.GetID())
		return nil
	}

	// No existing comment found — create a new one
	_, _, err = client.Issues.CreateComment(
		ctx, owner, repoName, prNumber,
		&github.IssueComment{Body: github.String(body)},
	)
	if err != nil {
		return fmt.Errorf("creating comment: %w", err)
	}
	slog.Info("created new schema guard comment")
	return nil
}

// findExistingComment paginates through PR comments looking for one
// that contains the schema-guard-action HTML marker.
func findExistingComment(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	prNumber int,
) (*github.IssueComment, error) {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		comments, resp, err := client.Issues.ListComments(
			ctx, owner, repo, prNumber, opts,
		)
		if err != nil {
			return nil, fmt.Errorf("listing comments: %w", err)
		}

		for _, c := range comments {
			if strings.Contains(c.GetBody(), commentMarker) {
				return c, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil, nil // No existing comment found
}

// parseRepo splits an "owner/repo" string into its two parts.
func parseRepo(repo string) (owner, repoName string, err error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format, expected owner/repo: %q", repo)
	}
	return parts[0], parts[1], nil
}
