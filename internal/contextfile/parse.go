package contextfile

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// validIdentifier matches valid MySQL identifiers.
var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Load reads a context.yaml file from disk, parses it, and validates it.
// Returns a typed ContextFile or a validation error.
func Load(path string) (*ContextFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading context file: %w", err)
	}

	var cf ContextFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("parsing context YAML: %w", err)
	}

	if err := validate(&cf); err != nil {
		return nil, fmt.Errorf("validating context file: %w", err)
	}

	return &cf, nil
}

// validate checks all the rules from architecture.md §1.3.
func validate(cf *ContextFile) error {
	if cf.Version != "1" {
		return fmt.Errorf("unsupported version %q (expected \"1\")", cf.Version)
	}

	if len(cf.Tables) == 0 {
		return fmt.Errorf("tables must be a non-empty map")
	}

	for tName, table := range cf.Tables {
		if !validIdentifier.MatchString(tName) {
			return fmt.Errorf("invalid table name %q: must match %s", tName, validIdentifier.String())
		}
		if table.Description == "" {
			return fmt.Errorf("table %q: description is required", tName)
		}

		for cName, col := range table.Columns {
			if !validIdentifier.MatchString(cName) {
				return fmt.Errorf("table %q, column %q: invalid identifier", tName, cName)
			}
			if col.Description == "" {
				return fmt.Errorf("table %q, column %q: description is required", tName, cName)
			}
		}
	}

	return nil
}
