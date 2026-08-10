"""
Tests: LLM must NOT hallucinate meanings for undocumented or missing columns.

Two scenarios:
  1. Column EXISTS in DB but has NO entry in context.yaml (documented: false)
     → LLM must say: 'This column is not documented in context.yaml.'
     → Verified by: string assertion (most reliable for fixed-phrase tests)

  2. Column/table does NOT exist at all in the DB schema
     → LLM must say: 'This column does not exist in the database.'
     → Verified by: string assertion

Note: HallucinationMetric and AnswerRelevancyMetric are intentionally NOT used here
because they are designed for testing well-grounded positive answers, not refusals
or "not found" responses — which inherently score poorly on those metrics.
"""

from deepeval import assert_test
from deepeval.test_case import LLMTestCase
from deepeval.metrics import HallucinationMetric, AnswerRelevancyMetric

NOT_DOCUMENTED_PHRASE = "not documented in context.yaml"
NOT_EXIST_PHRASE      = "does not exist in the database"


# ── Scenario 1: Column exists in DB but is undocumented in context.yaml ──────
# patients.FIPS and patients.INCOME are real columns (added via ALTER TABLE) but
# intentionally absent from context.yaml — the LLM must not guess their meaning.

def test_hallucination_undocumented_column(mcp_client):
    """
    LLM must state the column is not documented — not invent a meaning.

    patients.FIPS is present in the database schema but has no entry in context.yaml.
    The MCP server returns documented=false. The LLM must not extrapolate from the
    column name 'FIPS' (Federal Information Processing Standards) or any other source.
    """
    question = "What does the FIPS column in the patients table represent?"
    response = mcp_client.ask(question, table="patients", column="FIPS")

    assert NOT_DOCUMENTED_PHRASE.lower() in response.lower(), (
        f"Expected 'not documented in context.yaml' response.\nGot: {response[:300]}"
    )


def test_hallucination_undocumented_income(mcp_client):
    """
    LLM must not invent a definition for patients.INCOME even if the name implies one.

    patients.INCOME is absent from context.yaml. The word 'income' is strongly suggestive,
    creating a temptation to hallucinate. This test verifies the LLM defers to the
    documented context and refuses to guess.
    """
    question = "What does the INCOME column in the patients table store?"
    response = mcp_client.ask(question, table="patients", column="INCOME")

    assert NOT_DOCUMENTED_PHRASE.lower() in response.lower(), (
        f"Expected 'not documented in context.yaml' response.\nGot: {response[:300]}"
    )


# ── Scenario 2: Column/table does NOT exist in the DB schema at all ──────────

def test_hallucination_fabrication_guard_blood_type(mcp_client):
    """
    LLM must not invent a description for a plausible-sounding but nonexistent column.

    patients.blood_type sounds like a natural healthcare column, but Synthea does
    not include it. The MCP server returns 'column not found'. The LLM must propagate
    that signal rather than reasoning from its training knowledge.
    """
    question = "What does the blood_type column in the patients table contain?"
    response = mcp_client.ask(question, table="patients", column="blood_type")

    assert NOT_EXIST_PHRASE.lower() in response.lower(), (
        f"Expected 'does not exist in the database' response.\nGot: {response[:300]}"
    )


def test_hallucination_fabrication_guard_diagnosis_code(mcp_client):
    """
    LLM must not fabricate FK relationships for a nonexistent encounters.diagnosis_code.

    encounters.diagnosis_code sounds like a valid clinical FK but the Synthea schema
    uses SNOMED CT via the CODE column. The LLM must not invent the relationship.
    """
    question = "What does the diagnosis_code column in the encounters table reference?"
    response = mcp_client.ask(question, table="encounters", column="diagnosis_code")

    assert NOT_EXIST_PHRASE.lower() in response.lower(), (
        f"Expected 'does not exist in the database' response.\nGot: {response[:300]}"
    )
