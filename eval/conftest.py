"""
Pytest fixtures for the eval harness.

Both fixtures are session-scoped to avoid per-test server restarts
and redundant LLM API costs (architecture.md §5.2).
"""
import subprocess
import os
import pytest


@pytest.fixture(scope="session")
def mcp_server():
    """Start the Go MCP server as a subprocess (stdio transport)."""
    server_bin = os.environ.get("SERVER_BIN", "./schema-mcp")
    dsn = os.environ.get("DSN", "cosmo:cosmo@tcp(localhost:3306)/cosmo_db")
    context_path = os.environ.get("CONTEXT_FILE", "./examples/context.yaml")

    proc = subprocess.Popen(
        [server_bin, "--dsn", dsn, "--context", context_path],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    yield proc

    proc.terminate()
    proc.wait(timeout=5)


@pytest.fixture(scope="session")
def mcp_client(mcp_server):
    """
    Connect an LLM as an MCP client to the running server.

    TODO: Wire up actual MCP client + LLM connection.
    For scaffolding, this returns a placeholder object.
    """

    class PlaceholderClient:
        def ask(self, question: str) -> str:
            raise NotImplementedError(
                "MCP client not yet wired — implement in Phase 2"
            )

    return PlaceholderClient()
