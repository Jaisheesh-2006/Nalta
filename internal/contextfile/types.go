package contextfile

// ContextFile represents the top-level structure of context.yaml.
type ContextFile struct {
	Version string                  `yaml:"version"`
	Tables  map[string]TableContext `yaml:"tables"`
}

// TableContext holds human-authored metadata for a single database table.
type TableContext struct {
	Description string                   `yaml:"description"`
	Sensitive   bool                     `yaml:"sensitive"`
	Columns     map[string]ColumnContext `yaml:"columns"`
}

// ColumnContext holds human-authored metadata for a single database column.
type ColumnContext struct {
	Description string `yaml:"description"`
	Sensitive   bool   `yaml:"sensitive"`
	PII         bool   `yaml:"pii"`
}
