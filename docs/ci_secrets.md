# CI Secrets Configuration Guide

This document lists all GitHub Actions secrets used by this repository and explains how to obtain them.

Go to **Repository → Settings → Secrets and variables → Actions → New repository secret** to add each one.

---

## Required Secrets

### `GROQ_API_KEY` — used by `eval.yml`

**Purpose**: Authenticates LLM calls to Groq's inference API (used for eval metric scoring and subject LLM responses).

**Free tier**: Groq offers a generous free tier — no credit card required for sign-up.

**How to get it**:
1. Go to [https://console.groq.com](https://console.groq.com)
2. Sign up / log in
3. Navigate to **API Keys** → **Create API Key**
4. Copy the key (starts with `gsk_...`) and add it as `GROQ_API_KEY` in GitHub secrets

**Default model**: `groq/llama-3.3-70b-versatile` (set via `EVAL_MODEL` in `eval.yml`)

---

### `GEMINI_API_KEY` — used by `eval.yml` (optional alternative)

**Purpose**: Authenticates LLM calls to Google Gemini (free-tier alternative to Groq).

**Free tier**: Available via Google AI Studio, no billing required.

**How to get it**:
1. Go to [https://aistudio.google.com/apikey](https://aistudio.google.com/apikey)
2. Click **Create API Key**
3. Copy the key and add it as `GEMINI_API_KEY` in GitHub secrets
4. Change `EVAL_MODEL` in `eval.yml` to `gemini/gemini-1.5-flash`

---

### `OPENAI_API_KEY` — used by `eval.yml` (optional, paid)

**Purpose**: Authenticates LLM calls to OpenAI (gpt-4o). DeepEval also uses this internally as its judge model if set.

**Note**: This is a **paid** service. Only needed if you explicitly set `EVAL_MODEL: gpt-4o` in `eval.yml`.

---

## Auto-Injected Secrets (no setup needed)

### `GITHUB_TOKEN` — used by `schema-guard.yml`

**Purpose**: Allows the schema-guard action to post PR comments when `--dry-run` is removed.

**Note**: GitHub automatically injects this token for all workflows. No manual setup required.
The token has read/write access scoped to the current repository only.

---

## Optional Overrides

### `MYSQL_ROOT_PASSWORD` — used by `eval.yml`

Falls back to `ci-root-only` if not set. Only needed if you want a custom root password in CI MySQL.

### `MYSQL_PASSWORD` — used by `eval.yml`

Falls back to `ci-cosmo-only` if not set. Only needed if you want a custom `cosmo` user password in CI MySQL.

---

## Summary Table

| Secret | Workflow | Required? | Free? | Where to get |
|---|---|---|---|---|
| `GROQ_API_KEY` | `eval.yml` | ✅ Yes (default) | ✅ Yes | [console.groq.com](https://console.groq.com) |
| `GEMINI_API_KEY` | `eval.yml` | ❌ Alternative | ✅ Yes | [aistudio.google.com](https://aistudio.google.com/apikey) |
| `OPENAI_API_KEY` | `eval.yml` | ❌ Alternative | ❌ Paid | [platform.openai.com](https://platform.openai.com/api-keys) |
| `GITHUB_TOKEN` | `schema-guard.yml` | ✅ Yes | ✅ Auto-injected | Automatic |
| `MYSQL_ROOT_PASSWORD` | `eval.yml` | ❌ Optional | — | Set any value |
| `MYSQL_PASSWORD` | `eval.yml` | ❌ Optional | — | Set any value |
