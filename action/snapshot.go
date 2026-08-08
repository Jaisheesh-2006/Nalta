package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// SnapshotColumn represents a column captured from INFORMATION_SCHEMA.
type SnapshotColumn struct {
	TableName  string
	ColumnName string
	DataType   string
}

// SnapshotSchemas captures "before" and "after" schema states.
//
// Two modes:
//   - Ephemeral mode (default): spins up a throwaway MySQL container,
//     applies migrations via golang-migrate, snapshots, and tears down.
//   - External DB mode: if cfg.DSN is provided, uses that DSN directly
//     (for CI environments with pre-existing MySQL services).
//
// Implements the single-container migrate-forward strategy from
// architecture.md §3.2.
func SnapshotSchemas(cfg *ActionConfig) (before, after []SnapshotColumn, err error) {
	if cfg.DSN != "" {
		return snapshotWithExternalDB(cfg)
	}
	return snapshotWithEphemeralContainer(cfg)
}

// snapshotWithExternalDB uses a pre-existing MySQL instance pointed to by
// cfg.DSN. Applies before-migrations, snapshots, then applies
// after-migrations on top and snapshots again.
func snapshotWithExternalDB(cfg *ActionConfig) (before, after []SnapshotColumn, err error) {
	slog.Info("using external DB", "dsn", redactDSN(cfg.DSN))

	// Apply base-branch migrations
	if cfg.BeforeMigrations != "" {
		if err := applyMigrations(cfg.DSN, cfg.BeforeMigrations); err != nil {
			return nil, nil, fmt.Errorf("applying base migrations: %w", err)
		}
		slog.Info("base migrations applied", "path", cfg.BeforeMigrations)
	}

	// Snapshot "before"
	before, err = snapshotSchema(cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("snapshotting before: %w", err)
	}
	slog.Info("before snapshot captured", "columns", len(before))

	// Apply PR-branch migrations on top
	if cfg.AfterMigrations != "" {
		if err := applyMigrations(cfg.DSN, cfg.AfterMigrations); err != nil {
			return nil, nil, fmt.Errorf("applying PR migrations: %w", err)
		}
		slog.Info("PR migrations applied", "path", cfg.AfterMigrations)
	}

	// Snapshot "after"
	after, err = snapshotSchema(cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("snapshotting after: %w", err)
	}
	slog.Info("after snapshot captured", "columns", len(after))

	return before, after, nil
}

// snapshotWithEphemeralContainer implements the full single-container
// migrate-forward strategy from architecture.md §3.2:
//
//  1. Start a throwaway MySQL 8.0 container
//  2. Wait for it to accept connections
//  3. Apply base-branch migrations → snapshot "before"
//  4. Apply PR-branch migrations on top → snapshot "after"
//  5. Stop + remove container (even on error)
func snapshotWithEphemeralContainer(cfg *ActionConfig) (before, after []SnapshotColumn, err error) {
	// 1. Start ephemeral MySQL container
	container, err := startEphemeralMySQL()
	if err != nil {
		return nil, nil, fmt.Errorf("starting ephemeral MySQL: %w", err)
	}
	// Always clean up the container, even on error or panic.
	defer func() {
		if cleanupErr := stopEphemeralMySQL(container.id); cleanupErr != nil {
			slog.Warn("failed to clean up ephemeral container",
				"container", container.id[:12], "error", cleanupErr)
		}
	}()

	slog.Info("ephemeral MySQL started",
		"container", container.id[:12], "port", container.port)

	// 2. Wait for MySQL to accept connections
	if err := waitForMySQL(container.dsn, 60*time.Second); err != nil {
		return nil, nil, fmt.Errorf("waiting for ephemeral MySQL: %w", err)
	}
	slog.Info("ephemeral MySQL ready")

	// 3. Apply base-branch migrations → snapshot "before"
	if cfg.BeforeMigrations != "" {
		if err := applyMigrations(container.dsn, cfg.BeforeMigrations); err != nil {
			return nil, nil, fmt.Errorf("applying base migrations: %w", err)
		}
		slog.Info("base migrations applied", "path", cfg.BeforeMigrations)
	}

	before, err = snapshotSchema(container.dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("snapshotting before: %w", err)
	}
	slog.Info("before snapshot captured", "columns", len(before))

	// 4. Apply PR-branch migrations on top → snapshot "after"
	//    golang-migrate tracks version in schema_migrations, so calling Up()
	//    with the after-migrations dir only applies new versions beyond the
	//    base tip.
	if cfg.AfterMigrations != "" {
		if err := applyMigrations(container.dsn, cfg.AfterMigrations); err != nil {
			return nil, nil, fmt.Errorf("applying PR migrations: %w", err)
		}
		slog.Info("PR migrations applied", "path", cfg.AfterMigrations)
	}

	after, err = snapshotSchema(container.dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("snapshotting after: %w", err)
	}
	slog.Info("after snapshot captured", "columns", len(after))

	return before, after, nil
}

// ---------- Ephemeral Container Management ----------

// ephemeralContainer holds identifiers for a running container.
type ephemeralContainer struct {
	id   string // Docker container ID (full SHA)
	name string // Human-readable container name
	port int    // Host port mapped to container's 3306
	dsn  string // go-sql-driver/mysql DSN for this container
}

// startEphemeralMySQL starts a throwaway MySQL 8.0 container on a random
// free port and returns its metadata. The caller MUST call
// stopEphemeralMySQL when done.
func startEphemeralMySQL() (*ephemeralContainer, error) {
	// Find a free host port to avoid collisions when multiple
	// PRs run concurrently on self-hosted runners.
	port, err := findFreePort()
	if err != nil {
		return nil, fmt.Errorf("finding free port: %w", err)
	}

	name := generateContainerName()

	cmd := exec.Command("docker", "run", "-d",
		"--name", name,
		"-e", "MYSQL_ROOT_PASSWORD=root",
		"-e", "MYSQL_DATABASE=guard_db",
		"-p", fmt.Sprintf("%d:3306", port),
		"mysql:8.0",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker run failed: %s: %w",
			strings.TrimSpace(string(out)), err)
	}

	containerID := strings.TrimSpace(string(out))
	dsn := fmt.Sprintf("root:root@tcp(127.0.0.1:%d)/guard_db", port)

	return &ephemeralContainer{
		id:   containerID,
		name: name,
		port: port,
		dsn:  dsn,
	}, nil
}

// stopEphemeralMySQL stops and removes the container. This is safe to call
// even if the container has already been removed.
func stopEphemeralMySQL(containerID string) error {
	// Stop (with a short timeout to avoid hanging)
	stop := exec.Command("docker", "stop", "-t", "5", containerID)
	if out, err := stop.CombinedOutput(); err != nil {
		slog.Debug("docker stop output", "output", strings.TrimSpace(string(out)))
		// Continue to rm even if stop fails (container may already be stopped)
	}

	// Remove
	rm := exec.Command("docker", "rm", "-f", containerID)
	if out, err := rm.CombinedOutput(); err != nil {
		return fmt.Errorf("docker rm failed: %s: %w",
			strings.TrimSpace(string(out)), err)
	}

	slog.Info("ephemeral container removed", "container", containerID[:12])
	return nil
}

// waitForMySQL polls the MySQL instance until it accepts connections or
// the timeout expires. MySQL typically needs 10-15s to initialise.
func waitForMySQL(dsn string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}

		err = db.Ping()
		db.Close()
		if err == nil {
			return nil
		}

		lastErr = err
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("MySQL not ready after %v: %w", timeout, lastErr)
}

// ---------- Migration Application ----------

// applyMigrations uses golang-migrate to apply all pending migrations
// from the given directory to the MySQL instance at dsn.
//
// If all migrations have already been applied (ErrNoChange), this is
// treated as success — it's the expected case when the before-migrations
// and after-migrations overlap.
func applyMigrations(dsn, migrationsPath string) error {
	sourceURL, err := migrationsSourceURL(migrationsPath)
	if err != nil {
		return err
	}

	// golang-migrate's MySQL driver expects "mysql://" prefix and
	// multiStatements=true for migration files with multiple statements.
	migrateDSN := buildMigrateDSN(dsn)

	m, err := migrate.New(sourceURL, migrateDSN)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	slog.Debug("migrations applied", "version", version, "dirty", dirty)

	return nil
}

// migrationsSourceURL converts a filesystem path to a golang-migrate
// file:// source URL, handling cross-platform path differences.
func migrationsSourceURL(migrationsPath string) (string, error) {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return "", fmt.Errorf("resolving migrations path: %w", err)
	}

	// Convert to forward slashes for URL compatibility.
	urlPath := filepath.ToSlash(absPath)

	// Ensure the path starts with "/" for a valid file:// URL.
	// On Linux absPath already starts with "/"; on Windows it starts
	// with "C:/..." which needs a leading "/" → "file:///C:/...".
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}

	return "file://" + urlPath, nil
}

// buildMigrateDSN wraps a go-sql-driver/mysql DSN with the "mysql://"
// scheme and multiStatements parameter that golang-migrate requires.
func buildMigrateDSN(dsn string) string {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return "mysql://" + dsn + sep + "multiStatements=true"
}

// ---------- Schema Snapshotting ----------

// snapshotSchema connects to the MySQL instance at dsn and captures all
// columns from INFORMATION_SCHEMA.COLUMNS for the current database.
func snapshotSchema(dsn string) ([]SnapshotColumn, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to DB: %w", err)
	}
	defer db.Close()

	return snapshotCurrentSchema(db)
}

// snapshotCurrentSchema queries INFORMATION_SCHEMA.COLUMNS for all
// columns in the current database, ordered by table and position.
func snapshotCurrentSchema(db *sql.DB) ([]SnapshotColumn, error) {
	rows, err := db.Query(`
		SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE
		FROM   INFORMATION_SCHEMA.COLUMNS
		WHERE  TABLE_SCHEMA = DATABASE()
		ORDER  BY TABLE_NAME, ORDINAL_POSITION
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []SnapshotColumn
	for rows.Next() {
		var c SnapshotColumn
		if err := rows.Scan(&c.TableName, &c.ColumnName, &c.DataType); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

// ---------- Helpers ----------

// findFreePort asks the OS for an available TCP port by binding to port 0,
// reading the assigned port, and immediately releasing it.
func findFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("listening on :0: %w", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port, nil
}

// generateContainerName creates a unique container name to avoid
// collisions when multiple guard runs happen concurrently.
func generateContainerName() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(99999))
	return fmt.Sprintf("schema-guard-%05d", n.Int64())
}

// redactDSN masks the password portion of a DSN for safe logging.
// "user:password@tcp(host:port)/db" → "user:***@tcp(host:port)/db"
func redactDSN(dsn string) string {
	atIdx := strings.Index(dsn, "@")
	if atIdx < 0 {
		return "***"
	}
	colonIdx := strings.Index(dsn[:atIdx], ":")
	if colonIdx < 0 {
		return "***"
	}
	return dsn[:colonIdx+1] + "***" + dsn[atIdx:]
}
