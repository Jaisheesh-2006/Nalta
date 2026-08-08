# schema-context-mcp

> Connect your database schema to AI agents — with human-authored context, sensitivity flags, and drift detection.

## Components

| Component | What it does |
|---|---|
| **schema-context-mcp** (`server/`) | Go binary that introspects a live MySQL database, merges schema with `context.yaml`, and serves the result to AI agents over MCP. |
| **schema-guard-action** (`action/`) | GitHub Action that detects migration changes in PRs and warns when documented or sensitive columns are affected. |
| **eval harness** (`eval/`) | Python test suite (DeepEval) that validates LLM responses are grounded in `context.yaml`. |

## Quick Start

```bash
# 1. Start MySQL with seeded schema
docker compose up -d

# 2. Build and run the MCP server
go build -o schema-mcp ./server
export DSN="cosmo:cosmo@tcp(localhost:3306)/cosmo_db"
./schema-mcp --context ./examples/context.yaml

# The server is now listening on stdio.
# Connect any MCP client (e.g., Claude Desktop) to interact.
```

## Project Structure

```
├── server/           # Component A — MCP server
├── action/           # Component B — GitHub Action
├── internal/         # Shared context.yaml parser
├── eval/             # Component C — Python eval harness
├── examples/         # Demo context.yaml + migrations
└── .github/          # CI workflows
```

## Confirmations

- Env var and flag precedence: `--dsn` and `--context` flags are implemented and fall back to `DSN` and `CONTEXT_FILE` env vars; flags take precedence. See `server/config.go`.
- MCP SDK: the project uses `mark3labs/mcp-go` (listed in `go.mod`) and the MCP server is implemented in `server/mcp.go`.

## License

MIT — see [LICENSE](LICENSE).
