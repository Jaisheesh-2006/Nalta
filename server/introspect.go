package main

import (
	"database/sql"
	"fmt"
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
func IntrospectSchema(db *sql.DB) ([]TableSchema, error) {
	columns, err := queryColumns(db)
	if err != nil {
		return nil, fmt.Errorf("querying columns: %w", err)
	}

	fks, err := queryForeignKeys(db)
	if err != nil {
		return nil, fmt.Errorf("querying foreign keys: %w", err)
	}

	return buildTableSchemas(columns, fks), nil
}

func queryColumns(db *sql.DB) ([]DBColumn, error) {
	rows, err := db.Query(`
		SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COALESCE(COLUMN_DEFAULT, '')
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
		if err := rows.Scan(&c.TableName, &c.ColumnName, &c.DataType, &nullable, &c.DefaultValue); err != nil {
			return nil, err
		}
		c.IsNullable = nullable == "YES"
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

func buildTableSchemas(columns []DBColumn, fks []DBForeignKey) []TableSchema {
	// Build FK lookup: "table.column" -> DBForeignKey
	fkMap := make(map[string]DBForeignKey)
	for _, fk := range fks {
		key := fk.TableName + "." + fk.ColumnName
		fkMap[key] = fk
	}

	// Group columns by table
	tableMap := make(map[string]*TableSchema)
	var tableOrder []string

	for _, col := range columns {
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
