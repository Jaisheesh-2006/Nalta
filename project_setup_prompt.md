# Prompt for AI-Agent — Design architecture.md for schema-context-mcp



---

I'm building **schema-context-mcp**, a Go open-source project with two independent components sharing one contract file. I need you to design the full technical architecture and write it up as `architecture.md`. Don't write implementation code yet — this is the design pass two people will build from in parallel, so it needs to be precise enough that neither of us second-guesses the other's half.

## What the project is

**`context.yaml`** — a human-authored file, committed to the repo next to migrations, describing tables/columns in plain English and flagging sensitive ones.

**Component A — `schema-context-mcp`** (Go binary): connects to a live Postgres database, introspects its schema via `information_schema`, merges that with `context.yaml`, and serves the combined result to AI agents over MCP — as a Resource (full merged schema+context) and a Tool (`explain_column(table, column)`).

**Component B — `schema-guard-action`** (GitHub Action): runs on pull requests touching migration files. Takes a schema snapshot before and after the migration, diffs them, cross-references changed columns against `context.yaml`, and if a documented or `sensitive: true` column changed, posts a PR comment warning that context/downstream agents may need updating.

**Component C — eval harness** (v1.1): a thin script that connects an LLM to the running MCP server, asks fixed test questions, and checks the agent's answers are grounded in `context.yaml` (correct definitions, refuses to leak `sensitive: true` values, doesn't hallucinate meaning for undocumented columns).

Two people are building A and B in parallel over about a week. `context.yaml`'s shape is the only shared surface between them — everything else must be independently buildable and testable.

## What I need `architecture.md` to nail down, precisely

### 1. `context.yaml` — final schema
Exact field names and types, what's required vs optional at the table level and column level, how nesting works, a full realistic example (use a cosmetics/ingredient-safety domain — tables like `ingredients`, `products`, `affiliates` — not generic `users`/`orders`), and validation rules (what makes a `context.yaml` invalid and how that should fail).

### 2. `schema-context-mcp` internals
- Exact flow: config/flags → connect to Postgres → introspect schema → parse `context.yaml` → merge into one in-memory model → start MCP server → expose Resource + Tool.
- The merge logic specifically: what happens when a DB column has no matching entry in `context.yaml` (must not silently drop it — agents still need to see it exists, just without meaning attached), and what happens when `context.yaml` references a table/column that no longer exists in the DB (must warn, not crash).
- Exact shape of the MCP Resource payload and the `explain_column` Tool's input/output.
- Package/file layout inside `/server`.

### 3. `schema-guard-action` internals
- Exact flow: triggered on PR → obtain "before" schema snapshot and "after" schema snapshot → diff → cross-reference diff against `context.yaml` → build PR comment → post via GitHub API.
- How "before" and "after" snapshots are actually obtained (be concrete — e.g., running migrations against ephemeral DBs, or diffing migration files directly) and the trade-offs of the approach you pick.
- Exact PR comment format/template.
- Package/file layout inside `/action`.

### 4. The integration seam
State explicitly: the only coupling between A and B is that both parse the same `context.yaml` schema. Specify whether that parsing logic is a shared internal Go package (`/internal/contextfile`) used by both, or duplicated — and justify the choice.

### 5. Eval harness wiring
How it connects to the MCP server, where test cases live, how pass/fail is determined (rule-based vs LLM-judge, and for which cases each applies), and how/where this plugs into CI alongside `schema-guard-action`.

### 6. Repo layout
Full folder tree for the whole repo (`/server`, `/action`, `/internal`, `/examples`, `/eval`, config files, etc.) with a one-line purpose per folder.

### 7. Local dev / demo setup
What `docker-compose.yml` needs to spin up (Postgres seeded with the ingredient-safety example schema) so a stranger can try the whole loop — server + Action — in under two minutes.

## Constraints to respect

- No custom SQL parser — if parsing is needed anywhere, name the existing library to use.
- No RBAC/masking middleware, no hosted platform, no multi-database support in v1.
- Postgres only.
- Everything must be independently buildable by two people working in parallel after this document is done — flag anywhere the design forces sequential dependency, and resolve it if you can.

Ask me clarifying questions first if anything about the merge logic, snapshot strategy, or eval scoring is ambiguous — don't guess silently on the parts that determine how the two components integrate.