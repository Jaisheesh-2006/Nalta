# Evaluation & CI

Automated evaluation for `schema-context-mcp` using **DeepEval**.

The harness runs the Go MCP server, connects through the Python MCP SDK, and evaluates LLM responses with an LLM-as-a-Judge.

## Evaluation Architecture

The LLM is evaluated as an **MCP client**, rather than testing the MCP server in isolation.

```text
                 ┌──────────────────┐
                 │   Test Scenario  │
                 │  testcases.yaml  │
                 └────────┬─────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │ Python Evaluation│
                 │     Harness      │
                 └────────┬─────────┘
                          │ MCP / stdio
                          ▼
                 ┌──────────────────┐
                 │   Go MCP Server  │
                 │ schema-context-  │
                 │       mcp        │
                 └────────┬─────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │ Schema + Context │
                 │   context.yaml   │
                 └────────┬─────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │       LLM        │
                 └────────┬─────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │    DeepEval      │
                 │   LLM-as-Judge   │
                 └──────────────────┘
```

## Tests

* **Grounding** — Uses `schema://full` and `explain_column` correctly.
* **Hallucination** — Doesn't invent meaning for undocumented columns.
* **Sensitive Data** — Respects `sensitive: true`.
* **Undocumented Columns** — Handles DB columns missing from `context.yaml`.

Scenarios and criteria: `testcases.yaml`

## Run Locally

```bash
# From project root
go build -o nalta .

# Setup evaluation environment
cd eval
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

Set judge API keys:

```bash
export GROQ_API_KEY="your-key"
export GEMINI_API_KEY="your-key"
```

Run:

```bash
pytest -v
```

**Default judge:** Groq (`llama-3.3-70b-versatile`)
**Fallback:** Gemini

## CI

```text
Pull Request
     │
     ├── eval.yml
     │      └── DeepEval evaluation
     │
     └── schema-guard.yml
            ├── Start MySQL
            ├── Run migrations
            ├── Compare schema ↔ context.yaml
            └── Post diff to PR
```

### Required Secrets

**`eval.yml`**

* `GROQ_API_KEY`
* `GEMINI_API_KEY`

**`schema-guard.yml`**

* `GITHUB_TOKEN`

```
```
