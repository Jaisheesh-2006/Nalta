# MCP Contract

This document defines the shapes of the MCP resources and tools exposed by `schema-context-mcp`, including their error responses. This serves as the contract between the Server (Dev 1), the Action (Dev 2), and the Eval Harness (Dev 3).

## Transport
- **Protocol**: MCP (Model Context Protocol) over stdio.
- **Format**: JSON-RPC 2.0.

## Resources

### `schema://full`
Returns the entire merged database schema with `context.yaml` annotations.

**MIME Type**: `application/json`

**Shape**:
```json
{
  "tables": [
    {
      "name": "string",
      "description": "string",
      "sensitive": "boolean",
      "columns": [
        {
          "name": "string",
          "data_type": "string",
          "nullable": "boolean",
          "default_value": "string",
          "references": {
            "table": "string",
            "column": "string"
          } | null,
          "description": "string",
          "sensitive": "boolean",
          "pii": "boolean",
          "documented": "boolean"
        }
      ]
    }
  ]
}
```
*Note: See `examples/schema_full_sample.json` for a complete payload.*

## Tools

### `explain_column`
Looks up a single column's metadata, type info, and context annotations.

**Inputs**:
- `table` (string, required): The exact name of the database table.
- `column` (string, required): The exact name of the column in the table.

**Success Response** (Text):
```json
{
  "table": "string",
  "column": {
    "name": "string",
    "data_type": "string",
    "nullable": "boolean",
    "default_value": "string",
    "references": {
      "table": "string",
      "column": "string"
    } | null,
    "description": "string",
    "sensitive": "boolean",
    "pii": "boolean",
    "documented": "boolean"
  }
}
```

**Error Responses**:
If the table or column is not found, the tool returns an MCP ToolResultError (which clients handle gracefully, instead of a panic).

- **Table not found**: `table 'xyz' not found`
- **Column not found**: `column 'xyz' not found in table 'abc'`

### `explain_table`
Looks up a single table's merged metadata and columns.

**Inputs**:
- `table` (string, required): The exact name of the database table.

**Success Response** (Text):
Matches the shape of a single table object from `schema://full`.

**Error Responses**:
- **Table not found**: `table 'xyz' not found`
