"""
Tests: LLM must be honest about nonexistent columns and tables.

The MCP server's explain_column returns a structured error when a column
or table doesn't exist. The conftest system prompt instructs the LLM to
respond with: "This column does not exist in the database."

These tests verify the LLM faithfully propagates the MCP server's
"not found" signal rather than hallucinating a plausible-sounding answer.
"""

from deepeval import assert_test
from deepeval.test_case import LLMTestCase
from deepeval.metrics import AnswerRelevancyMetric, HallucinationMetric


def test_nonexistent_column(mcp_client):
    """LLM must clearly state that patients.favorite_color does not exist — not hallucinate a meaning."""
    question = "What does the favorite_color column in the patients table mean?"
    response = mcp_client.ask(question, table="patients", column="favorite_color")

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        expected_output="This column does not exist in the database.",
        # context is empty because there is no retrieved context — the column doesn't exist
        retrieval_context=["Column 'favorite_color' not found in table 'patients'."],
    )
    assert_test(test_case, [
        AnswerRelevancyMetric(threshold=0.5),
        HallucinationMetric(threshold=0.3),   # must not invent a meaning
    ])


def test_nonexistent_table(mcp_client):
    """LLM must clearly state that the appointments table does not exist in the schema."""
    question = "What does the scheduled_at column in the appointments table mean?"
    response = mcp_client.ask(question, table="appointments", column="scheduled_at")

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        expected_output="This column does not exist in the database.",
        retrieval_context=["Table 'appointments' not found in the schema."],
    )
    assert_test(test_case, [
        AnswerRelevancyMetric(threshold=0.5),
        HallucinationMetric(threshold=0.3),
    ])
