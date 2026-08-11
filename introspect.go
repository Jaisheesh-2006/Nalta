package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DBColumn represents a single column from INFORMATION_SCHEMA.COLUMNS.
type DBColumn struct {
	TableName    string
	ColumnName   string
	DataType     string
	IsNullable   bool
	DefaultValue string
}

// DBForeignKey represents a foreign key relationship from INFORMATION_SCHEMA.KEY_COLUMN_USAGE.
type DBForeignKey struct {
	TableName            string
	ColumnName           string
	ReferencedTableName  string
	ReferencedColumnName string
}

// TableSchema groups columns and foreign keys for a single table.
type TableSchema struct {
	Name        string
	Columns     []DBColumn
	ForeignKeys map[string]DBForeignKey // keyed by column name
}

// IntrospectSchema queries INFORMATION_SCHEMA to get the full database schema
// including column metadata and foreign key relationships.
// Accepts a context for timeout/cancellation. Retries once on transient errors.
func IntrospectSchema(ctx context.Context, db *sql.DB) ([]TableSchema, error) {
	const maxAttempts = 2
	var columns []DBColumn
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		columns, err = queryColumns(db)
		if err == nil {
			break
		}
		if attempt < maxAttempts-1 {
			// Brief backoff before retry — respect context cancellation
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("querying columns: %w", err)
	}

	fks, err := queryForeignKeys(db)
	if err != nil {
		return nil, fmt.Errorf("querying foreign keys: %w", err)
	}

	// Filter out views — only return BASE TABLEs
	tables, err := queryTableTypes(db)
	if err != nil {
		return nil, fmt.Errorf("querying table types: %w", err)
	}

	return buildTableSchemas(columns, fks, tables), nil
}

func queryColumns(db *sql.DB) ([]DBColumn, error) {
	rows, err := db.Query(`
		SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA
		FROM   INFORMATION_SCHEMA.COLUMNS
		WHERE  TABLE_SCHEMA = DATABASE()
		ORDER  BY TABLE_NAME, ORDINAL_POSITION
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []DBColumn
	for rows.Next() {
		var c DBColumn
		var nullable string
		var defaultVal sql.NullString
		var extra sql.NullString
		if err := rows.Scan(&c.TableName, &c.ColumnName, &c.DataType, &nullable, &defaultVal, &extra); err != nil {
			return nil, err
		}
		c.IsNullable = nullable == "YES"
		c.DefaultValue = defaultVal.String

		// Normalize MySQL type aliases to standard SQL names
		if c.DataType == "int" {
			c.DataType = "integer"
		}
		// COLUMN_DEFAULT doesn't capture AUTO_INCREMENT; detect it from EXTRA
		if c.DefaultValue == "" && extra.String == "auto_increment" {
			c.DefaultValue = "auto_increment"
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func queryForeignKeys(db *sql.DB) ([]DBForeignKey, error) {
	rows, err := db.Query(`
		SELECT TABLE_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME
		FROM   INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE  TABLE_SCHEMA = DATABASE()
		  AND  REFERENCED_TABLE_NAME IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []DBForeignKey
	for rows.Next() {
		var fk DBForeignKey
		if err := rows.Scan(&fk.TableName, &fk.ColumnName, &fk.ReferencedTableName, &fk.ReferencedColumnName); err != nil {
			return nil, err
		}
		fks = append(fks, fk)
	}
	return fks, rows.Err()
}

func queryTableTypes(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query(`
		SELECT TABLE_NAME, TABLE_TYPE
		FROM   INFORMATION_SCHEMA.TABLES
		WHERE  TABLE_SCHEMA = DATABASE()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make(map[string]string)
	for rows.Next() {
		var name, ttype string
		if err := rows.Scan(&name, &ttype); err != nil {
			return nil, err
		}
		types[name] = ttype
	}
	return types, rows.Err()
}

func buildTableSchemas(columns []DBColumn, fks []DBForeignKey, tableTypes map[string]string) []TableSchema {
	// Build FK lookup: "table.column" -> DBForeignKey
	fkMap := make(map[string]DBForeignKey)
	for _, fk := range fks {
		key := fk.TableName + "." + fk.ColumnName
		fkMap[key] = fk
	}

	// Group columns by table, skipping VIEWs
	tableMap := make(map[string]*TableSchema)
	var tableOrder []string

	for _, col := range columns {
		// Skip views
		if t, ok := tableTypes[col.TableName]; ok && t != "BASE TABLE" {
			continue
		}
		ts, exists := tableMap[col.TableName]
		if !exists {
			ts = &TableSchema{
				Name:        col.TableName,
				ForeignKeys: make(map[string]DBForeignKey),
			}
			tableMap[col.TableName] = ts
			tableOrder = append(tableOrder, col.TableName)
		}
		ts.Columns = append(ts.Columns, col)

		// Attach FK if this column has one
		fkKey := col.TableName + "." + col.ColumnName
		if fk, ok := fkMap[fkKey]; ok {
			ts.ForeignKeys[col.ColumnName] = fk
		}
	}

	// Preserve discovery order
	result := make([]TableSchema, 0, len(tableOrder))
	for _, name := range tableOrder {
		result = append(result, *tableMap[name])
	}
	return result
}
