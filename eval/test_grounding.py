"""
Tests: LLM descriptions must be grounded in context.yaml (no hallucination).

Metrics used:
  FaithfulnessMetric (≥0.7)  — LLM answer must be grounded in the MCP-retrieved context
  HallucinationMetric (≤0.3) — LLM must not fabricate column meanings beyond the context

Columns chosen are:
  - Documented in context.yaml with a meaningful description
  - NOT marked sensitive or pii — so the refusal rules don't fire
  - Diverse across different tables (patients, conditions, encounters)

This combination ensures we test "does the LLM explain grounded in context?" without
the sensitive/PII refusal rules interfering.
"""

from deepeval import assert_test
from deepeval.test_case import LLMTestCase
from deepeval.metrics import FaithfulnessMetric, HallucinationMetric


def test_grounded_birthdate_description(mcp_client):
    """LLM must explain patients.BIRTHDATE using the context.yaml definition."""
    question = "What does the BIRTHDATE column in the patients table represent?"
    response = mcp_client.ask(question, table="patients", column="BIRTHDATE")

    context = [
        "Date of birth in ISO-8601 format (YYYY-MM-DD). "
        "Used to derive patient age and for cohort filtering."
    ]
    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        retrieval_context=context,
        context=context,          # HallucinationMetric requires context= as well
    )
    assert_test(test_case, [
        FaithfulnessMetric(threshold=0.7),
        HallucinationMetric(threshold=0.3),
    ])


def test_grounded_conditions_code(mcp_client):
    """LLM must describe conditions.CODE as a SNOMED CT code using the context."""
    question = "What does the CODE column in the conditions table contain?"
    response = mcp_client.ask(question, table="conditions", column="CODE")

    context = [
        "SNOMED CT concept code that uniquely identifies the condition "
        "(e.g., 44054006 for Type 2 diabetes mellitus). "
        "Look up the human-readable name in the DESCRIPTION column."
    ]
    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        retrieval_context=context,
        context=context,
    )
    assert_test(test_case, [
        FaithfulnessMetric(threshold=0.7),
        HallucinationMetric(threshold=0.3),
    ])


def test_grounded_encounterclass(mcp_client):
    """LLM must describe encounters.ENCOUNTERCLASS as an encounter type field."""
    question = "What is stored in the ENCOUNTERCLASS column of the encounters table?"
    response = mcp_client.ask(question, table="encounters", column="ENCOUNTERCLASS")

    context = [
        "Type of clinical encounter: 'outpatient', 'ambulatory', 'inpatient', "
        "'emergency', 'urgentcare', or 'wellness'. "
        "Useful for filtering encounter modality in analytical queries."
    ]
    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        retrieval_context=context,
        context=context,
    )
    assert_test(test_case, [
        FaithfulnessMetric(threshold=0.7),
        HallucinationMetric(threshold=0.3),
    ])
