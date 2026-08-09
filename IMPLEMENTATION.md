# Implementation Plan — Developer 1 (MCP Server)

This document captures the step-by-step implementation plan for Developer 1 (MCP server), and confirms environment / dependency facts discovered in the repository.

## Confirmations

- Env vars / flag precedence: `--dsn` and `--context` flags are implemented in `server/config.go` and fall back to `DSN` and `CONTEXT_FILE` environment variables. Flags take precedence over env vars. See `server/config.go` for the implementation and semantics.
- MCP SDK: `mark3labs/mcp-go` is declared as a dependency in `go.mod` (present as `github.com/mark3labs/mcp-go v0.57.0`) and is used in `server/mcp.go` to register resources and tools.

Files referenced:

- `server/config.go` — flag + env precedence and validation
- `go.mod` — contains `github.com/mark3labs/mcp-go` requirement
- `server/mcp.go` — MCP resource/tool registration using `mark3labs/mcp-go`

## Step-by-step implementation workflow (Developer 1)

1. Prep & tooling
   - Add small `Makefile` targets to build and run integration tests: `build-server`, `run-integration`.
   - Add minimal CI stub (optional) to run integration locally.

2. Introspection core
   - Implement robust INFORMATION_SCHEMA queries in `server/introspect.go` to enumerate tables, columns, types, defaults, nullability, and foreign keys.
   - Return typed Go structs representing DB schema.

3. Introspect edge cases
   - Detect and skip or mark views.
   - Handle tables with zero columns.
   - Add DB connect timeout and simple retry/backoff logic.

4. Merge logic hardening
   - Ensure `server/merge.go` merges the DB metadata with `internal/contextfile` outputs into `MergedTable` and `MergedColumn` values.
   - Add tests in `server/merge_test.go` for undocumented columns, context-only entries (warnings), and FK wiring.

5. Structured logging + startup log
   - Respect `--log-level` / `LOG_LEVEL` and use structured logging consistently.
   - Log `INFO schema introspected at <time> tables=N columns=M` at startup.

6. MCP surface: resource & tools
   - Implement `schema://full` resource returning the merged model as JSON (MIME `application/json`).
   - Implement `explain_column` tool accepting `{table, column}` and returning `{table, column: {...}}` or an MCP-friendly error object.
   - Add unit tests for correct result and error shapes.

7. Optional: `explain_table` tool
   - Return the list of columns for a single table (lighter payload than `schema://full`).

8. Integration test (Docker-backed)
   - Create `integration/run_integration.sh` or Makefile target that:
     1. Runs `docker compose up -d` (uses repository `docker-compose.yml`).
     2. Builds server: `go build -o schema-mcp ./server`.
     3. Runs `./schema-mcp --dsn "$DSN" --context ./examples/context.yaml` in background.
     4. Uses a small test client to call `schema://full` and `explain_column` and assert expected shapes.

9. Startup readiness
   - Ensure server is ready within ~2s for the eval harness; add a simple readiness probe in the integration script.

10. Sample outputs & docs
    - Add `examples/schema_full_sample.json` and `examples/explain_column_sample.json` for Dev 2/3.
    - Document MCP error shapes in `docs/mcp_contract.md` or `IMPLEMENTATION.md`.

11. Cross-team sync & PR
    - Share sample JSON and run quick compatibility checks with Dev 2 and Dev 3.
    - Run `gofmt`/`go vet`, open PR with tests and integration results.

## Quick run commands (local)

```bash
docker compose up -d
go build -o schema-mcp ./server
export DSN="cosmo:cosmo@tcp(localhost:3306)/cosmo_db"
./schema-mcp --context ./examples/context.yaml
```

## Contacts / next actions

- Next I will scaffold the Docker-backed integration test and add sample JSON outputs unless you prefer I start implementing `server/introspect.go` changes directly.
