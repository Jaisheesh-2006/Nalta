#!/usr/bin/env bash
set -euo pipefail

# Run a docker-backed integration run: start MySQL, build and run the server.
# This is a simple orchestrator for local manual verification.

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "Starting docker compose (MySQL)..."
docker compose up -d

echo "Building server..."
go build -o schema-mcp ./server

export DSN="cosmo:cosmo@tcp(localhost:3306)/cosmo_db"
export CONTEXT_FILE="./examples/context.yaml"

echo "Running server in background (logs -> integration/server.log)..."
mkdir -p integration
./schema-mcp --dsn "$DSN" --context "$CONTEXT_FILE" > integration/server.log 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

echo "Waiting 3s for server to start..."
sleep 3

if ps -p $SERVER_PID > /dev/null; then
  echo "Server appears to be running. Tail of logs:"
  head -n 200 integration/server.log || true
else
  echo "Server process not running; check integration/server.log for details:" >&2
  cat integration/server.log >&2 || true
  exit 1
fi

echo
echo "Manual next steps:"
echo " - Use an MCP client to call resource 'schema://full' and tool 'explain_column'."
echo " - When finished, kill the server: kill $SERVER_PID"
