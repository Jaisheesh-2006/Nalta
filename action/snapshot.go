package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// SnapshotColumn represents a column captured from INFORMATION_SCHEMA.
type SnapshotColumn struct {
	TableName  string
	ColumnName string
	DataType   string
}

// SnapshotSchemas captures "before" and "after" schema states using the
// single-container migrate-forward strategy from architecture.md §3.2.
//
// TODO: Implement the full ephemeral container lifecycle.
// For now, this is a stub that connects to an existing DB.
func SnapshotSchemas(cfg *ActionConfig) (before []SnapshotColumn, after []SnapshotColumn, err error) {
	if cfg.DSN == "" {
		return nil, nil, fmt.Errorf("DSN required for schema snapshots")
	}

	// In the full implementation, this will:
	// 1. Spin up ephemeral MySQL container
	// 2. Apply base migrations → snapshot "before"
	// 3. Apply PR migrations on top → snapshot "after"
	// 4. Tear down container
	//
	// For scaffolding, we query an existing DB as a placeholder.
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to DB: %w", err)
	}
	defer db.Close()

	after, err = snapshotCurrentSchema(db)
	if err != nil {
		return nil, nil, err
	}

	// before is empty for scaffolding — treated as "all columns are new"
	return nil, after, nil
}

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
