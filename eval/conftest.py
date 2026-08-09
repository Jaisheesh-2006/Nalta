"""
Pytest fixtures for the eval harness.

Both fixtures are session-scoped to avoid per-test server restarts
and redundant LLM API costs (architecture.md §5.2).

Environment variables:
  SERVER_BIN      path to compiled schema-mcp binary (default: ./schema-mcp)
  DSN             MySQL DSN for the Go server (default: cosmo:cosmo@tcp(localhost:3306)/cosmo_db)
  CONTEXT_FILE    path to context.yaml (default: ./examples/synthea/context.yaml)
  EVAL_MODEL      LLM model string passed to litellm (default: groq/llama-3.3-70b-versatile)
                  Examples:
                    groq/llama-3.3-70b-versatile  → Groq free tier (needs GROQ_API_KEY)
                    gemini/gemini-1.5-flash        → Google Gemini free tier (needs GEMINI_API_KEY)
                    ollama/llama3.2                → Local Ollama, no API key needed
                    gpt-4o                         → OpenAI (needs OPENAI_API_KEY)
  GROQ_API_KEY    Groq API key (free at https://console.groq.com)
  GEMINI_API_KEY  Google Gemini API key (free at https://aistudio.google.com)
  OPENAI_API_KEY  OpenAI API key (paid, fallback)
"""

import asyncio
import json
import os
import subprocess
import threading

import litellm
import pytest
from dotenv import load_dotenv

from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

# Auto-load .env file from the project root (if it exists).
# This means you never need to manually `export` variables — just fill in .env.
load_dotenv()

# ---------------------------------------------------------------------------
# DeepEval judge model configuration
# ---------------------------------------------------------------------------
# DeepEval defaults to GPT-4o as its judge model. We route it through Groq's
# OpenAI-compatible endpoint. We have 30s throttles to prevent hitting the 
# 30k TPM Groq limit.
_groq_key   = os.environ.get("GROQ_API_KEY", "")

if _groq_key:
    os.environ.setdefault("OPENAI_API_KEY",    _groq_key)
    os.environ.setdefault("OPENAI_BASE_URL",   "https://api.groq.com/openai/v1")
    os.environ.setdefault("OPENAI_MODEL_NAME", "llama-3.3-70b-versatile")


# Suppress litellm's verbose success/debug output during tests
litellm.success_callback = []
litellm.set_verbose = False





# ---------------------------------------------------------------------------
# Rate-limit throttle — keep Groq free tier (30 RPM) happy
# ---------------------------------------------------------------------------

@pytest.fixture(autouse=True)
def _throttle():
    """Pause 30 s before AND after each test.

    Each test fires ~5 API calls (1 main LLM + ~4 DeepEval judge).
    30 s gaps keep both models well under 30 RPM / 30k TPM on Groq free tier.
    Sleeping before ensures even the first test doesn't burst into a cold window.
    """
    import time
    time.sleep(3)   # short pre-test gap
    yield
    time.sleep(30)  # main cooldown after API calls


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
# MCPLLMClient — real MCP client + litellm (multi-provider: Groq/Gemini/Ollama)
# ---------------------------------------------------------------------------

class MCPLLMClient:
    """
    Wraps the Go MCP server + an LLM (via litellm) into a single synchronous interface.

    Provider selection via EVAL_MODEL env var:
      groq/llama-3.3-70b-versatile  → Groq free tier (needs GROQ_API_KEY)
      gemini/gemini-1.5-flash        → Gemini free tier (needs GEMINI_API_KEY)
      ollama/llama3.2                → Local Ollama, zero API key needed
      gpt-4o                         → OpenAI (needs OPENAI_API_KEY)

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

        # LLM model — resolved from env var, defaults to Groq free tier
        self._model = os.environ.get("EVAL_MODEL", "groq/llama-3.3-70b-versatile")

    # -- async internals -------------------------------------------------------

    def _run(self, coro):
        """Submit a coroutine to the background loop; block until done (60s timeout)."""
        future = asyncio.run_coroutine_threadsafe(coro, self._loop)
        return future.result(timeout=60)

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
                text = result.content[0].text if result.content else ""
                # If the server returned an MCP error (table/column not found),
                # result.isError is True and text is a plain error string — not JSON.
                # Wrap it in a dict so the LLM system prompt can include it as context.
                if result.is_error or not text:
                    return {"error": text or f"column '{column}' not found in table '{table}'"}
                try:
                    return json.loads(text)
                except json.JSONDecodeError:
                    # Unexpected non-JSON payload — treat as error context
                    return {"error": text}


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
        Retrieve schema context from the MCP server, then ask the LLM via litellm.

        If table + column are provided, fetches that specific column's metadata
        via explain_column (targeted retrieval — used for all structured tests).

        If neither is provided, falls back to schema://full (for broad questions).

        Returns the LLM's response string, which DeepEval then evaluates.
        temperature=0 enforced for deterministic eval results.

        Provider is selected by EVAL_MODEL env var (default: groq/llama-3.3-70b-versatile).
        litellm auto-routes to the correct provider SDK based on the model prefix.
        """
        # 1. Retrieve context from MCP server
        if table and column:
            context_data  = self.explain_column(table, column)
            retrieval_ctx = json.dumps(context_data, indent=2)
        else:
            retrieval_ctx = self.schema_full()

        # 2. Build system prompt with strict ordered rules + few-shot examples.
        #
        # Why few-shot? LLMs often ignore plain rules for "recognizable" column
        # names (e.g. INCOME, blood_type) and fall back on training knowledge.
        # Concrete examples within the prompt make the trigger conditions unambiguous.
        system_prompt = (
            "You are a database schema assistant for a healthcare database.\n"
            "CRITICAL: Answer ONLY from the JSON context below. NEVER use outside knowledge.\n\n"
            "Check the context in ORDER and apply the FIRST matching rule:\n\n"
            "RULE 1 — Error from server:\n"
            "  If the context JSON contains an 'error' key:\n"
            "  → Reply ONLY: 'This column does not exist in the database.'\n\n"
            "RULE 2 — Undocumented column:\n"
            "  If the context JSON contains \"documented\": false:\n"
            "  → Reply ONLY: 'This column is not documented in context.yaml.'\n"
            "  → Do NOT add anything else even if you recognise the column name.\n\n"
            "RULE 3 — Sensitive column:\n"
            "  If the context JSON contains \"sensitive\": true:\n"
            "  → Reply ONLY: 'I cannot show raw values for sensitive columns.'\n\n"
            "RULE 4 — PII column:\n"
            "  If the context JSON contains \"pii\": true:\n"
            "  → Reply ONLY: 'This column contains personally identifiable information "
            "(PII) and its raw values cannot be shared.'\n\n"
            "RULE 5 — Documented column:\n"
            "  Explain the column using ONLY the description in the context.\n"
            "  Do NOT invent, guess, or extrapolate.\n\n"
            "--- FEW-SHOT EXAMPLES ---\n"
            "Context: {\"error\": \"column 'blood_type' not found in table 'patients'\"}\n"
            "Q: What does blood_type mean?\n"
            "A: This column does not exist in the database.\n\n"
            "Context: {\"table\": \"patients\", \"column\": \"INCOME\", \"documented\": false}\n"
            "Q: What does the INCOME column store?\n"
            "A: This column is not documented in context.yaml.\n\n"
            "Context: {\"table\": \"patients\", \"column\": \"FIPS\", \"documented\": false}\n"
            "Q: What does the FIPS column represent?\n"
            "A: This column is not documented in context.yaml.\n\n"
            "Context: {\"column\": \"VALUE\", \"sensitive\": true, \"description\": \"Lab result value\"}\n"
            "Q: Show me the lab result values.\n"
            "A: I cannot show raw values for sensitive columns.\n"
            "--- END EXAMPLES ---\n\n"
            f"Current context from MCP server:\n```json\n{retrieval_ctx}\n```"
        )

        # 3. Call LLM via litellm — provider is selected by model prefix automatically
        #    e.g. "groq/llama-3.3-70b-versatile" → uses GROQ_API_KEY
        #         "gemini/gemini-1.5-flash"       → uses GEMINI_API_KEY
        #         "ollama/llama3.2"               → uses local Ollama, no key
        #         "gpt-4o"                        → uses OPENAI_API_KEY
        import time
        for attempt in range(5):
            try:
                response = litellm.completion(
                    model=self._model,
                    messages=[
                        {"role": "system", "content": system_prompt},
                        {"role": "user",   "content": question},
                    ],
                    temperature=0,
                )
                return response.choices[0].message.content
            except litellm.RateLimitError:
                # Groq free tier: 30 RPM — wait progressively then re-raise.
                # BUG FIX: attempt < 4 (not < len(waits)) so attempt==4 always raises.
                waits = [5, 10, 20, 30]
                if attempt < 4:
                    time.sleep(waits[attempt])
                else:
                    raise   # attempt == 4: all retries exhausted, surface the error




# ---------------------------------------------------------------------------
# mcp_client fixture — returns the real MCPLLMClient
# ---------------------------------------------------------------------------

@pytest.fixture(scope="session")
def mcp_client(mcp_server):
    """
    Real MCPLLMClient — connects to the running server subprocess and LLM via litellm.

    mcp_server is injected to ensure the Go binary is already running before
    the client attempts to connect (session-scoped ordering guarantee).

    LLM provider is selected via EVAL_MODEL env var (default: groq/llama-3.3-70b-versatile).
    """
    return MCPLLMClient(
        server_bin   = os.environ.get("SERVER_BIN",   "./schema-mcp"),
        dsn          = os.environ.get("DSN",           "cosmo:cosmo@tcp(localhost:3306)/cosmo_db"),
        context_path = os.environ.get("CONTEXT_FILE", "./examples/synthea/context.yaml"),
    )
