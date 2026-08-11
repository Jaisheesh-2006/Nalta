# Nalta

> Connect your database schema to AI agents — with human-authored context, sensitivity flags, and drift detection.

## Project Overview

**Nalta** is a toolset designed to bridge the gap between your raw database structure and human knowledge. It provides AI agents with a semantically enriched understanding of a MySQL database schema while enforcing data governance during development.

*This project is built following cloud-native documentation guidelines.*

### What it is

The project consists of tools that allow you to define a `context.yaml` file alongside your database schema. This YAML file adds plain-English descriptions and flags for `sensitive` business data or `pii` (Personally Identifiable Information). By introspecting a live database and merging it with this context, AI agents can safely and intelligently interact with your data without leaking business secrets. 

## What has already been implemented

The core architecture has been fully built and tested:

- **Component A: MCP Server (`/`)**
  - A Go-based Model Context Protocol (MCP) server that connects to a live MySQL database, introspects its `INFORMATION_SCHEMA`, and merges it with your `context.yaml`. 
  - Exposes the `schema://full` resource and `explain_column` tool via stdio for AI agents (like Claude Desktop).
- **Component B: Schema Guard Action (`cmd/schema-guard/`)**
  - A GitHub Action that triggers on Pull Requests containing database migrations.
  - Automatically spins up an ephemeral MySQL instance, diffs the schema changes, and cross-references them against `context.yaml`.
  - Posts PR comments to warn developers if sensitive columns are dropped, altered, or if new columns lack documentation.
- **Component C: Evaluation Harness (`eval/`)**
  - A Python-based testing suite utilizing DeepEval to ensure LLMs correctly interpret the schema context.
  - Automatically spins up the MCP server and tests that LLMs respect the `sensitive` and `pii` flags without hallucinating undocumented definitions.

## What is yet to come

*(More features coming soon)*

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.24+
- Python 3.11+ (for the eval harness)
- (Optional) Any MCP Client like Claude Desktop

### Step 1: Boot the database with example data
Spin up the local MySQL container pre-seeded with our example schema:
```bash
docker compose up -d
```
Wait about 10 seconds for the database to become healthy.

### Step 2: Install and run Nalta
Install the tool globally using `go install`:
```bash
go install github.com/Jaisheesh-2006/nalta@latest
```

Then you can run it with the provided example `context.yaml` and connection string:
```bash
export DSN="cosmo:cosmo@tcp(localhost:3306)/cosmo_db"
nalta --context ./examples/context.yaml
```

*The server is now running and communicating over standard input/output (stdio). You can point your MCP client to this executable.*

### Step 3: Configure Claude Desktop (Optional)
To use Nalta with Claude Desktop, add the following to your `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "nalta": {
      "command": "nalta",
      "args": ["--context", "/absolute/path/to/context.yaml"],
      "env": {
        "DSN": "cosmo:cosmo@tcp(localhost:3306)/cosmo_db"
      }
    }
  }
}
```

### Step 4: Run the Eval Harness (Optional)
If you want to run the DeepEval test suite locally to test an LLM's adherence:
```bash
python -m venv .venv
# source .venv/bin/activate on Linux/Mac
.venv\Scripts\activate # Windows
pip install -r eval/requirements.txt
export EVAL_MODEL="gpt-4o"
export OPENAI_API_KEY="your-api-key"
pytest eval/ -v
```

## Documentation

For a deeper dive into the system design and how to use it, please see the following documentation:
- [Architecture & Design Details](architecture.md)
- [MCP Contract Documentation](docs/mcp_contract.md)

## License

MIT — see [LICENSE](LICENSE) for more details.
