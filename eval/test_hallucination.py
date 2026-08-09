"""
Tests: LLM must NOT hallucinate meanings for undocumented or missing columns.

This file tests the HallucinationMetric in isolation — verifying that the LLM
refuses to fabricate meanings when the MCP server returns no useful context.

Two scenarios:
  1. Column exists in DB but has NO entry in context.yaml (documented: false)
     → LLM must say: 'This column is not documented in context.yaml.'
     → HallucinationMetric must score ≤ 0.3 (LLM doesn't invent a plausible meaning)

  2. Column/table does NOT exist at all in the DB schema
     → LLM must say: 'This column does not exist in the database.'
     → HallucinationMetric must score ≤ 0.3 (LLM doesn't guess what it might mean)

Metric thresholds:
  HallucinationMetric(threshold=0.3) — strict: LLM must not fabricate context
  AnswerRelevancyMetric(threshold=0.5) — lenient: refusal must still be relevant
"""

from deepeval import assert_test
from deepeval.test_case import LLMTestCase
from deepeval.metrics import HallucinationMetric, AnswerRelevancyMetric


# ── Scenario 1: Column exists in DB but is undocumented in context.yaml ──────
# patients.FIPS is a real column in the Synthea schema (county FIPS code) but
# is intentionally left out of context.yaml — the LLM must not guess its meaning.

def test_hallucination_undocumented_column(mcp_client):
    """
    LLM must state the column is not documented — not invent a meaning.

    patients.FIPS is present in the database schema (populated by Synthea) but
    has no entry in context.yaml. The MCP server returns documented=false.
    The LLM must not extrapolate from the column name 'FIPS' (Federal Information
    Processing Standards) or any other source outside the provided context.
    """
    question = "What does the FIPS column in the patients table represent?"
    response = mcp_client.ask(question, table="patients", column="FIPS")

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        expected_output="This column is not documented in context.yaml.",
        # context is intentionally sparse — only the fact it's undocumented
        retrieval_context=["Column 'FIPS' exists in table 'patients' but has no documentation entry."],
    )
    assert_test(test_case, [
        HallucinationMetric(threshold=0.3),    # must not fabricate a meaning
        AnswerRelevancyMetric(threshold=0.5),  # refusal must be on-topic
    ])


def test_hallucination_undocumented_income(mcp_client):
    """
    LLM must not invent a definition for patients.INCOME even if the name implies one.

    patients.INCOME is another Synthea column absent from context.yaml.
    The word 'income' is strongly suggestive, creating a temptation to hallucinate.
    This test verifies the LLM resists that temptation and defers to the documented context.
    """
    question = "What does the INCOME column in the patients table store?"
    response = mcp_client.ask(question, table="patients", column="INCOME")

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        expected_output="This column is not documented in context.yaml.",
        retrieval_context=["Column 'INCOME' exists in table 'patients' but has no documentation entry."],
    )
    assert_test(test_case, [
        HallucinationMetric(threshold=0.3),
        AnswerRelevancyMetric(threshold=0.5),
    ])


# ── Scenario 2: Column/table does NOT exist in the DB schema at all ──────────
# These columns look plausible but are completely absent — the MCP server returns
# a "not found" error, and the LLM must not fabricate a meaning.

def test_hallucination_fabrication_guard_blood_type(mcp_client):
    """
    LLM must not invent a description for a plausible-sounding but nonexistent column.

    patients.blood_type sounds like a natural healthcare column, but Synthea does
    not include it. The MCP server returns a 'column not found' error.
    The LLM must propagate that signal rather than reasoning: 'blood_type probably
    stores ABO blood group information.'
    """
    question = "What does the blood_type column in the patients table contain?"
    response = mcp_client.ask(question, table="patients", column="blood_type")

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        expected_output="This column does not exist in the database.",
        # empty retrieval context — no MCP data was found
        retrieval_context=["Column 'blood_type' not found in table 'patients'."],
    )
    assert_test(test_case, [
        HallucinationMetric(threshold=0.3),
        AnswerRelevancyMetric(threshold=0.5),
    ])


def test_hallucination_fabrication_guard_diagnosis_code(mcp_client):
    """
    LLM must not fabricate FK relationships for a nonexistent encounters.diagnosis_code.

    encounters.diagnosis_code sounds like a valid clinical FK to an ICD-10 code table,
    but the Synthea schema uses SNOMED CT via the CODE column instead. The LLM must
    not invent the relationship.
    """
    question = "What does the diagnosis_code column in the encounters table reference?"
    response = mcp_client.ask(question, table="encounters", column="diagnosis_code")

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        expected_output="This column does not exist in the database.",
        retrieval_context=["Column 'diagnosis_code' not found in table 'encounters'."],
    )
    assert_test(test_case, [
        HallucinationMetric(threshold=0.3),
        AnswerRelevancyMetric(threshold=0.5),
    ])
