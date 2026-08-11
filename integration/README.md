Integration: Docker-backed server run

This folder contains a small helper to run a local MySQL instance (via `docker compose`), build the MCP server, and run it for manual integration checks.

Usage (Linux / macOS / WSL):

```bash
./integration/run_integration.sh
```

What it does:
- `docker compose up -d` using repository `docker-compose.yml` (starts MySQL seeded with `examples/migrations`).
- `go build -o nalta .` and runs the built binary with `DSN` and `CONTEXT_FILE` set for the example DB.
- Writes runtime logs to `integration/server.log` and prints the server PID.

Manual checks:
- Use an MCP-capable client to call resource `schema://full` and tool `explain_column`.
- Compare responses to `examples/schema_full_sample.json` and `examples/explain_column_sample.json`.

Stopping:
- Kill the server process printed by the script, and run `docker compose down` to stop MySQL.
