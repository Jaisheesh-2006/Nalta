# Evaluation Harness & CI pipelines (Developer 3)

This directory contains the automated evaluation harness for the `schema-context-mcp` project. It uses [DeepEval](https://github.com/confident-ai/deepeval) to rigorously test the LLM's behavior when interacting with our MCP Server.

## What is this?
To ensure our LLMs don't hallucinate missing context or leak PII, we treat the LLM as a client of our Go MCP Server. The tests spin up the server subprocess, connect to it over stdio via the `mcp` Python SDK, and use an LLM-as-a-Judge to evaluate the responses.

### Test Categories
1. **Grounding (`test_grounding.py`)**: Tests whether the LLM correctly uses the schema injected via `schema://full` or `explain_column`.
2. **Hallucination (`test_hallucination.py`)**: Tests that the LLM does not fabricate meaning for undocumented or missing columns.
3. **Sensitive Data (`test_sensitive.py`)**: Tests that the LLM appropriately refuses or restricts access to columns marked `sensitive: true`.
4. **Undocumented Columns (`test_undocumented.py`)**: Tests behavior when users ask for columns that exist in the DB but are not in `context.yaml`.

All test scenarios and expected criteria are defined in `testcases.yaml`.

## How to run locally

### 1. Prerequisites
You must have Python 3.11+ installed. You also need the `schema-mcp` binary built in the project root, as the tests invoke it.

```bash
# In the project root, build the server:
go build -o schema-mcp ./server

# Navigate to eval and set up Python
cd eval
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 2. API Keys
The evaluations require a fast/cheap judge model. By default, it uses **Groq** (`llama-3.3-70b-versatile`) as the primary judge, falling back to **Gemini** if rate limited.

Create a `.env` file in the project root or export these directly:
```bash
export GROQ_API_KEY="your-groq-key"
export GEMINI_API_KEY="your-gemini-key"
```
*(Note: OpenAI is not required unless you override the `EVAL_MODEL` env var to use it).*

### 3. Run the Tests
You can run the entire suite using `pytest`:

```bash
pytest -v
```

## GitHub Actions (CI)
This repository contains two CI workflows located in `.github/workflows/`:

1. **`eval.yml`**: Runs the evaluation harness (this directory) on PRs. It requires the `GROQ_API_KEY` and `GEMINI_API_KEY` GitHub Secrets to run the DeepEval judges.
2. **`schema-guard.yml`**: Runs Developer 2's `schema-guard` Go Action. It spins up an ephemeral MySQL container, runs `golang-migrate`, and posts a PR comment diffing the database schema against `context.yaml`. It requires the `GITHUB_TOKEN` secret.
