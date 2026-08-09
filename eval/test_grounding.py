"""
Tests: LLM descriptions must be grounded in context.yaml (no hallucination).

Metrics used:
  FaithfulnessMetric (≥0.7)  — LLM answer must be grounded in the MCP-retrieved context
  HallucinationMetric (≤0.3) — LLM must not fabricate column meanings beyond the context

All questions use targeted explain_column retrieval (table + column specified)
so DeepEval has exact retrieval_context to score against.
"""

from deepeval import assert_test
from deepeval.test_case import LLMTestCase
from deepeval.metrics import FaithfulnessMetric, HallucinationMetric


def test_grounded_ssn_description(mcp_client):
    """LLM must describe patients.SSN using context.yaml definition — not guess or extrapolate."""
    question = "What does the SSN column in the patients table represent?"
    response = mcp_client.ask(question, table="patients", column="SSN")

    context = [
        "Social Security Number. Synthetic but structurally valid. "
        "Marked pii: true — never expose raw values."
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


def test_grounded_conditions_description(mcp_client):
    """LLM must describe conditions.DESCRIPTION as a clinical diagnosis field."""
    question = "What does the DESCRIPTION column in the conditions table mean?"
    response = mcp_client.ask(question, table="conditions", column="DESCRIPTION")

    context = [
        "Human-readable name of the condition (e.g., Hypertension, "
        "Type 2 diabetes mellitus). This is a clinical diagnosis. "
        "Marked sensitive: true — reveals clinical history."
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


def test_grounded_observation_value(mcp_client):
    """LLM must describe observations.VALUE as a raw clinical measurement."""
    question = "What is stored in the VALUE column of the observations table?"
    response = mcp_client.ask(question, table="observations", column="VALUE")

    context = [
        "The actual measured or recorded result (e.g., 193.3, Negative, Positive). "
        "Raw clinical result — marked sensitive: true."
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
