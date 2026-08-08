package main

// TableColumn identifies a specific column in a table.
type TableColumn struct {
	Table    string
	Column   string
	DataType string
}

// ColumnChange represents a column whose type changed between before and after.
type ColumnChange struct {
	Table   string
	Column  string
	OldType string
	NewType string
}

// SchemaDiff captures the structural differences between two schema snapshots.
type SchemaDiff struct {
	AddedTables    []string
	DroppedTables  []string
	AddedColumns   []TableColumn
	DroppedColumns []TableColumn
	ChangedColumns []ColumnChange
}

// DiffSchemas compares before and after schema snapshots and returns
// the structural differences.
func DiffSchemas(before, after []SnapshotColumn) SchemaDiff {
	var diff SchemaDiff

	// Build lookup maps
	beforeMap := buildColumnMap(before) // "table.column" -> DataType
	afterMap := buildColumnMap(after)
	beforeTables := buildTableSet(before)
	afterTables := buildTableSet(after)

	// Detect added/dropped tables
	for t := range afterTables {
		if !beforeTables[t] {
			diff.AddedTables = append(diff.AddedTables, t)
		}
	}
	for t := range beforeTables {
		if !afterTables[t] {
			diff.DroppedTables = append(diff.DroppedTables, t)
		}
	}

	// Detect added/dropped/changed columns
	for key, afterType := range afterMap {
		table, column := splitKey(key)
		if beforeType, ok := beforeMap[key]; ok {
			if beforeType != afterType {
				diff.ChangedColumns = append(diff.ChangedColumns, ColumnChange{
					Table:   table,
					Column:  column,
					OldType: beforeType,
					NewType: afterType,
				})
			}
		} else {
			diff.AddedColumns = append(diff.AddedColumns, TableColumn{
				Table:    table,
				Column:   column,
				DataType: afterType,
			})
		}
	}
	for key := range beforeMap {
		if _, ok := afterMap[key]; !ok {
			table, column := splitKey(key)
			diff.DroppedColumns = append(diff.DroppedColumns, TableColumn{
				Table:  table,
				Column: column,
			})
		}
	}

	return diff
}

func buildColumnMap(cols []SnapshotColumn) map[string]string {
	m := make(map[string]string)
	for _, c := range cols {
		m[c.TableName+"."+c.ColumnName] = c.DataType
	}
	return m
}

func buildTableSet(cols []SnapshotColumn) map[string]bool {
	m := make(map[string]bool)
	for _, c := range cols {
		m[c.TableName] = true
	}
	return m
}

func splitKey(key string) (string, string) {
	for i, ch := range key {
		if ch == '.' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}
