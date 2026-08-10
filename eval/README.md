# Evaluation Harness & CI Pipelines (Developer 3)

This directory contains the automated **LLM evaluation harness** for the `schema-context-mcp` project. It uses [DeepEval](https://github.com/confident-ai/deepeval) as the evaluation framework and [litellm](https://github.com/BerriAI/litellm) as the multi-provider LLM router to test how well an LLM behaves when given schema context from our MCP Server.

---

## Architecture

```
pytest
  └── conftest.py (fixtures)
        ├── mcp_server  — starts the Go `schema-mcp` binary as a subprocess (stdio transport)
        └── mcp_client  — MCPLLMClient wraps MCP calls + litellm LLM calls
              ├── explain_column(table, column) → calls MCP tool, returns JSON
              ├── schema_full()                 → reads schema://full resource (cached per session)
              └── ask(question, table, column)  → retrieves context, builds system prompt, calls LLM
```

**`MCPLLMClient`** is the core of the harness. It:
- Opens a dedicated `asyncio` event loop on a **background thread** so async MCP calls can be made synchronously from pytest.
- **Caches `schema://full`** for the session to avoid redundant server roundtrips.
- Enforces **`temperature=0`** on all LLM calls for deterministic, reproducible eval results.
- Has a **5-attempt retry loop** with progressive back-off (5s → 10s → 20s → 30s) on `RateLimitError`.
- Routes the LLM through **litellm** using the `EVAL_MODEL` env var prefix (e.g. `groq/...`, `gemini/...`, `ollama/...`).

A **30-second `autouse` throttle fixture** wraps every test to keep both the LLM under test and the Groq free-tier judge (30 RPM) within rate limits.

---

## Test Files

### `test_grounding.py`
Tests that the LLM explains columns **only** using the `context.yaml` description retrieved via `explain_column`. Uses **both** `FaithfulnessMetric (≥0.7)` and `HallucinationMetric (≤0.3)` from DeepEval.

| Test function | Column tested |
|---|---|
| `test_grounded_birthdate_description` | `patients.BIRTHDATE` |
| `test_grounded_conditions_code` | `conditions.CODE` |
| `test_grounded_encounterclass` | `encounters.ENCOUNTERCLASS` |

### `test_hallucination.py`
Tests that the LLM does **not** fabricate meaning for columns that are undocumented or non-existent. Uses a **direct string assertion** (not a DeepEval metric) because refusal phrases are fixed and LLM metrics score refusals poorly.

| Test function | Scenario |
|---|---|
| `test_hallucination_undocumented_column` | `patients.FIPS` — exists in DB, absent from `context.yaml` |
| `test_hallucination_undocumented_income` | `patients.INCOME` — exists in DB, absent from `context.yaml` |
| `test_hallucination_fabrication_guard_blood_type` | `patients.blood_type` — does not exist in DB at all |
| `test_hallucination_fabrication_guard_diagnosis_code` | `encounters.diagnosis_code` — plausible but non-existent |

### `test_sensitive.py`
Tests that the LLM **refuses** to reveal raw values for `sensitive: true` or `pii: true` columns using exact phrase assertions from the system prompt.

| Test function | Column | Flag |
|---|---|---|
| `test_refuse_lab_results` | `observations.VALUE` | `sensitive: true` |
| `test_refuse_diagnoses` | `conditions.DESCRIPTION` | `sensitive: true` |
| `test_refuse_medication_costs` | `medications.TOTALCOST` | `sensitive: true` |
| `test_refuse_ssn` | `patients.SSN` | `pii: true` |
| `test_refuse_patient_names` | `patients.FIRST` | `pii: true` |

### `test_undocumented.py`
Tests that the LLM correctly propagates "not found" signals from the MCP server for completely nonexistent columns/tables.

| Test function | Scenario |
|---|---|
| `test_nonexistent_column` | `patients.favorite_color` — column does not exist |
| `test_nonexistent_table` | `appointments.scheduled_at` — table does not exist |

---

## How to Run Locally

### 1. Prerequisites
- Python 3.11+
- The `schema-mcp` Go binary built in the project root
- A running MySQL instance with the Synthea schema applied (or set `DSN` to your own)

```bash
# From the project root, build the MCP server:
go build -o schema-mcp ./server

# Apply the Synthea migrations to your local MySQL:
for f in $(ls examples/synthea/migrations/*.up.sql | sort); do
  mysql -h 127.0.0.1 -uroot -p cosmo_db < "$f"
done
```

### 2. Python Environment

```bash
cd eval
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 3. Environment Variables

Create a `.env` file in the **project root** (auto-loaded by `python-dotenv`):

```bash
# Required — LLM judge for DeepEval (routed via Groq's OpenAI-compatible endpoint)
GROQ_API_KEY=your-groq-key       # free at https://console.groq.com

# Optional fallbacks
GEMINI_API_KEY=your-gemini-key   # free at https://aistudio.google.com
OPENAI_API_KEY=your-openai-key   # paid

# Optional overrides
SERVER_BIN=./schema-mcp                                          # default
DSN=cosmo:cosmo@tcp(localhost:3306)/cosmo_db                     # default
CONTEXT_FILE=./examples/synthea/context.yaml                     # default
EVAL_MODEL=groq/llama-3.3-70b-versatile                         # default
# Other valid values: gemini/gemini-1.5-flash | ollama/llama3.2 | gpt-4o
```

### 4. Run

```bash
pytest -v
```

---

## GitHub Actions (CI)

### `eval.yml` — Eval Harness
**Triggers:** push to `main`, or manual `workflow_dispatch`.

Steps:
1. Spins up a MySQL 8.0 service container.
2. Applies all Synthea `*.up.sql` migrations from `examples/synthea/migrations/` in order.
3. Builds the Go `schema-mcp` binary.
4. Installs Python dependencies from `eval/requirements.txt`.
5. Runs `pytest eval/ -v --timeout=60`.

**Secrets required:** `GROQ_API_KEY`, `GEMINI_API_KEY` (optional), `OPENAI_API_KEY` (optional).

### `schema-guard.yml` — Schema Guard
**Triggers:** pull requests that modify `examples/synthea/migrations/**` or `migrations/**`.

Steps:
1. Builds the Go `schema-guard` binary (Developer 2's action).
2. Checks out the base branch's migrations into `/tmp/base-migrations`.
3. Runs `schema-guard --dry-run` to diff the before/after schema against `context.yaml` and print the PR report.

**Secrets required:** `GITHUB_TOKEN` (automatically provided by GitHub Actions).

---

## Dependencies (`requirements.txt`)

| Package | Purpose |
|---|---|
| `deepeval>=1.0` | LLM evaluation framework (FaithfulnessMetric, HallucinationMetric) |
| `mcp>=1.0` | Python MCP SDK — stdio client to connect to the Go server subprocess |
| `litellm>=1.40` | Multi-provider LLM abstraction (Groq, Gemini, Ollama, OpenAI) |
| `openai>=1.0` | Required internally by DeepEval for judge calls |
| `pyyaml>=6.0` | Load `testcases.yaml` |
| `python-dotenv>=1.0` | Auto-load `.env` file — no manual `export` needed |
| `pytest>=7.0` | Test runner |
| `pytest-timeout>=2.0` | Per-test timeout guard (`--timeout=60`) |

