# Developer Work Division Plan

> **Scaffolding is done.** `go build ./...` passes, all tests pass, every file has real code.
> This plan divides the **remaining implementation work** across 3 developers who can work independently.

---

## The One Rule

`internal/contextfile/` is **frozen**. The types, parser, and validation are committed and tested. If any developer needs to change this package, they must notify the other two **before** committing. This is the single shared surface — everything else is independently owned.

---

## Ownership Map

```
Developer 1 (Server)        Developer 2 (Action)        Developer 3 (Eval + CI)
──────────────────          ──────────────────          ────────────────────────
server/main.go              action/main.go              eval/conftest.py
server/config.go            action/config.go            eval/test_grounding.py
server/introspect.go        action/snapshot.go           eval/test_sensitive.py
server/merge.go             action/diff.go              eval/test_undocumented.py
server/mcp.go               action/crossref.go          eval/testcases.yaml
server/merge_test.go        action/comment.go           eval/requirements.txt
                            action/diff_test.go         .github/workflows/eval.yml
                                                        .github/workflows/schema-guard.yml
```

**Shared (frozen):** `internal/contextfile/` — nobody touches without group notification.
**Shared (infrequent):** `go.mod`, `docker-compose.yml`, `examples/` — edit only if adding a dependency or fixture.

---

## Developer 1: MCP Server (`server/`)

### What's already scaffolded
- Config loading with flag/env precedence ✅
- INFORMATION_SCHEMA queries (columns + FKs) ✅
- Full merge logic with all 4 cases ✅
- MCP server with resource + tool registered ✅
- Merge tests for all 4 cases ✅

### Remaining work

| Task | Priority | Detail |
|---|---|---|
| **Integration test against real MySQL** | P0 | Run `docker compose up`, build server, connect, verify `schema://full` returns 4 tables with correct FK references |
| **Edge case handling in introspect** | P1 | Handle MySQL views, tables with no columns, connection timeouts gracefully |
| **Introspection timestamp log** | P1 | Log `INFO schema introspected at <time> tables=N columns=N` at startup (per arch §2.5) |
| **Structured logging** | P2 | Replace `slog.Info/Warn/Error` calls with proper log levels based on `--log-level` config |
| **MCP error responses** | P2 | Ensure `explain_column` returns proper MCP error format (not Go panic) for bad inputs |
| **Add `explain_table` tool** | P3 | Optional convenience tool — return all columns for a table without the full payload |

### Interface contract (what Dev 3 needs from you)
- The server binary must accept `--dsn` and `--context` flags
- It must communicate over **stdio** (stdin/stdout JSON-RPC)
- The `schema://full` resource must return `{"tables": [...]}`
- The `explain_column` tool must accept `{table, column}` and return `{table, column: {...}}`

---

## Developer 2: GitHub Action (`action/`)

### What's already scaffolded
- Config parsing with CLI flags + env vars ✅
- Schema snapshot from INFORMATION_SCHEMA ✅
- Full diff logic (added/dropped/changed tables and columns) ✅
- Cross-reference against context.yaml with severity levels ✅
- PR comment formatting (markdown with ⚠️/ℹ️ tables + 🔒 PII badges) ✅
- GitHub API posting via `go-github` ✅
- Diff tests (dropped column, added table, no changes, sensitive finding) ✅

### Remaining work

| Task | Priority | Detail |
|---|---|---|
| **Ephemeral container lifecycle** | P0 | Implement the single-container migrate-forward strategy in `snapshot.go` — currently a stub that queries an existing DB |
| **`golang-migrate` integration** | P0 | Wire up `golang-migrate/migrate` to apply base then PR migrations against the ephemeral container |
| **GitHub Action input parsing** | P1 | Parse inputs from `action.yml` format (not just CLI flags) — `INPUT_CONTEXT_PATH`, `INPUT_DSN`, etc. |
| **Comment deduplication** | P2 | If the action runs multiple times on the same PR, update the existing comment instead of posting duplicates |
| **PII badge in cross-reference** | P2 | The `crossref.go` already flags PII — verify the 🔒 badge renders correctly in the markdown |
| **Integration test with a real PR** | P2 | Open a test PR modifying `examples/migrations/`, verify the Action posts the expected comment |

### Interface contract (what Dev 3 needs from you)
- The action binary must accept `--dry-run` for local testing (prints to stdout)
- The `action.yml` inputs must be: `context-path`, `dsn`, `github-token`

---

## Developer 3: Eval Harness + CI (`eval/` + `.github/`)

### What's already scaffolded
- Session-scoped pytest fixtures (server subprocess + placeholder client) ✅
- Test stubs for all 4 test cases (skipped with `@pytest.mark.skip`) ✅
- `testcases.yaml` with question/expected definitions ✅
- CI workflows for both guard and eval ✅

### Remaining work

| Task | Priority | Detail |
|---|---|---|
| **Wire MCP client in conftest.py** | P0 | Replace `PlaceholderClient` with actual MCP client that connects to the server subprocess over stdio — needs Dev 1's binary |
| **LLM integration** | P0 | Connect an LLM (via `EVAL_MODEL` env var — default `gpt-4o`) as an MCP client that can call `explain_column` and read `schema://full` |
| **Remove `@skip` decorators** | P1 | Once the client is wired, remove all skip markers so tests actually run |
| **Add HallucinationMetric tests** | P1 | Test that the LLM doesn't fabricate meanings for undocumented columns (threshold 0.3) |
| **CI secrets configuration** | P1 | Document which secrets need to be set: `OPENAI_API_KEY` for eval, `GITHUB_TOKEN` for guard |
| **Guard workflow integration test** | P2 | Verify `schema-guard.yml` triggers correctly on migration file changes |
| **Eval result reporting** | P3 | Parse DeepEval output and post a summary as a GitHub check annotation |

### Interface contract (what you need from Dev 1)
- A built server binary at a known path (default `./schema-mcp`)
- The server must start and be ready to accept MCP requests within ~2 seconds
- You depend on Dev 1's binary being stable before removing `@skip`

---

## Timeline

```
Day 1       │  Scaffolding committed (DONE)
            │  All 3 devs pull and verify: go build ./..., go test ./...
            │
Days 2-5    │  Dev 1: server integration + edge cases
            │  Dev 2: ephemeral container + migrate integration
            │  Dev 3: MCP client wiring + LLM connection
            │
Day 3-4     │  🔄 MID-PHASE SYNC: verify JSON field names match
            │     Dev 1 produces a sample schema://full output
            │     Dev 2 confirms crossref.go can parse it
            │     Dev 3 confirms eval tests can consume it
            │
Days 6-7    │  Dev 1: polish + explain_table tool
            │  Dev 2: comment dedup + real PR test
            │  Dev 3: remove @skip, run full eval suite
            │
Day 8       │  Integration: all 3 verify against same examples/ fixtures
            │  Done when: build + test + eval all pass on same commit
```

---

## Quick Reference: How to test independently

```bash
# Developer 1 (needs Docker)
docker compose up -d
go build -o schema-mcp ./server
export DSN="cosmo:cosmo@tcp(localhost:3306)/cosmo_db"
./schema-mcp --context ./examples/context.yaml
# → server running on stdio, test with any MCP client

# Developer 2 (needs Docker)
go build -o schema-guard ./action
./schema-guard --context ./examples/context.yaml --dry-run
# → prints markdown comment to stdout

# Developer 3 (needs Python 3.11+, needs Dev 1's binary)
pip install -r eval/requirements.txt
go build -o schema-mcp ./server
pytest eval/ -v --collect-only
# → shows test collection (all skipped until client is wired)
```
