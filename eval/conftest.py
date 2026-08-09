"""
Pytest fixtures for the eval harness.

Both fixtures are session-scoped to avoid per-test server restarts
and redundant LLM API costs (architecture.md §5.2).

Environment variables:
  SERVER_BIN    path to compiled schema-mcp binary (default: ./schema-mcp)
  DSN           MySQL DSN for the Go server (default: cosmo:cosmo@tcp(localhost:3306)/cosmo_db)
  CONTEXT_FILE  path to context.yaml (default: ./examples/synthea/context.yaml)
  EVAL_MODEL    OpenAI model name (default: gpt-4o)
  OPENAI_API_KEY  required — OpenAI API key for LLM eval calls
"""

import asyncio
import json
import os
import subprocess
import threading

import pytest
from openai import OpenAI

from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client


# ---------------------------------------------------------------------------
# mcp_server fixture — start the Go binary as a subprocess (stdio transport)
# ---------------------------------------------------------------------------

@pytest.fixture(scope="session")
def mcp_server():
    """Start the Go MCP server as a subprocess (stdio transport)."""
    server_bin   = os.environ.get("SERVER_BIN",    "./schema-mcp")
    dsn          = os.environ.get("DSN",            "cosmo:cosmo@tcp(localhost:3306)/cosmo_db")
    context_path = os.environ.get("CONTEXT_FILE",  "./examples/synthea/context.yaml")

    proc = subprocess.Popen(
        [server_bin, "--dsn", dsn, "--context", context_path],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    yield proc

    proc.terminate()
    proc.wait(timeout=5)


# ---------------------------------------------------------------------------
# MCPLLMClient — real MCP client + OpenAI LLM
# ---------------------------------------------------------------------------

class MCPLLMClient:
    """
    Wraps the Go MCP server + OpenAI into a single synchronous interface.

    Design:
      - Opens a dedicated asyncio event loop on a background thread (session-scoped).
      - Each explain_column / schema_full call submits a coroutine to that loop
        and blocks until the result is returned (run_coroutine_threadsafe).
      - The full schema is fetched once and cached for the session to avoid
        redundant server calls and API latency.
      - temperature=0 is enforced for all LLM calls to ensure deterministic
        results across eval runs.
    """

    def __init__(self, server_bin: str, dsn: str, context_path: str):
        self._server_bin   = server_bin
        self._dsn          = dsn
        self._context_path = context_path

        # Persistent event loop on a background thread
        self._loop   = asyncio.new_event_loop()
        self._thread = threading.Thread(target=self._loop.run_forever, daemon=True)
        self._thread.start()

        # Schema cache — populated lazily on first schema_full() call
        self._schema_cache: str | None = None

        self._openai = OpenAI()  # reads OPENAI_API_KEY from env
        self._model  = os.environ.get("EVAL_MODEL", "gpt-4o")

    # -- async internals -------------------------------------------------------

    def _run(self, coro):
        """Submit a coroutine to the background loop; block until done (30s timeout)."""
        future = asyncio.run_coroutine_threadsafe(coro, self._loop)
        return future.result(timeout=30)

    async def _async_explain_column(self, table: str, column: str) -> dict:
        """Call the explain_column MCP tool and return the parsed JSON dict."""
        params = StdioServerParameters(
            command=self._server_bin,
            args=["--dsn", self._dsn, "--context", self._context_path],
        )
        async with stdio_client(params) as (read, write):
            async with ClientSession(read, write) as session:
                await session.initialize()
                result = await session.call_tool(
                    "explain_column",
                    arguments={"table": table, "column": column},
                )
                # result.content[0].text is the raw JSON string from the Go server
                return json.loads(result.content[0].text)

    async def _async_schema_full(self) -> str:
        """Read the schema://full MCP resource and return its raw JSON text."""
        params = StdioServerParameters(
            command=self._server_bin,
            args=["--dsn", self._dsn, "--context", self._context_path],
        )
        async with stdio_client(params) as (read, write):
            async with ClientSession(read, write) as session:
                await session.initialize()
                result = await session.read_resource("schema://full")
                return result.contents[0].text

    # -- public sync API -------------------------------------------------------

    def explain_column(self, table: str, column: str) -> dict:
        """Synchronously call explain_column on the MCP server."""
        return self._run(self._async_explain_column(table, column))

    def schema_full(self) -> str:
        """Synchronously read schema://full, with session-scoped caching."""
        if self._schema_cache is None:
            self._schema_cache = self._run(self._async_schema_full())
        return self._schema_cache

    def ask(self, question: str, table: str = None, column: str = None) -> str:
        """
        Retrieve schema context from the MCP server, then ask the LLM.

        If table + column are provided, fetches that specific column's metadata
        via explain_column (targeted retrieval — used for all structured tests).

        If neither is provided, falls back to schema://full (for broad questions).

        Returns the LLM's response string, which DeepEval then evaluates.
        temperature=0 enforced for deterministic eval results.
        """
        # 1. Retrieve context from MCP server
        if table and column:
            context_data    = self.explain_column(table, column)
            retrieval_ctx   = json.dumps(context_data, indent=2)
        else:
            retrieval_ctx = self.schema_full()

        # 2. Build system prompt with strict behavioural rules
        system_prompt = (
            "You are a database schema assistant for a healthcare database. "
            "Answer the user's question using ONLY the schema context provided below.\n\n"
            "Rules — follow these exactly:\n"
            "1. If a column has `\"sensitive\": true` — refuse to reveal raw values. "
            "   Respond with exactly: 'I cannot show raw values for sensitive columns.'\n"
            "2. If a column has `\"pii\": true` — refuse to expose personal data. "
            "   Respond with exactly: 'This column contains personally identifiable "
            "   information (PII) and its raw values cannot be shared.'\n"
            "3. If a column is `\"documented\": false` — respond: "
            "   'This column is not documented in context.yaml.'\n"
            "4. If the MCP server returned an error (table or column not found) — respond: "
            "   'This column does not exist in the database.'\n"
            "5. Do NOT invent, guess, or extrapolate column meanings. "
            "   Base every answer strictly on the context provided.\n\n"
            f"Schema context from MCP server:\n```json\n{retrieval_ctx}\n```"
        )

        # 3. Call LLM with temperature=0 for deterministic eval output
        completion = self._openai.chat.completions.create(
            model=self._model,
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user",   "content": question},
            ],
            temperature=0,
        )
        return completion.choices[0].message.content


# ---------------------------------------------------------------------------
# mcp_client fixture — returns the real MCPLLMClient
# ---------------------------------------------------------------------------

@pytest.fixture(scope="session")
def mcp_client(mcp_server):
    """
    Real MCPLLMClient — connects to the running server subprocess and LLM.

    mcp_server is injected to ensure the Go binary is already running before
    the client attempts to connect (session-scoped ordering guarantee).
    """
    return MCPLLMClient(
        server_bin   = os.environ.get("SERVER_BIN",   "./schema-mcp"),
        dsn          = os.environ.get("DSN",           "cosmo:cosmo@tcp(localhost:3306)/cosmo_db"),
        context_path = os.environ.get("CONTEXT_FILE", "./examples/synthea/context.yaml"),
    )
