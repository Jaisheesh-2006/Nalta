"""Tests: LLM should be honest about undocumented and nonexistent columns."""
import pytest


@pytest.mark.skip(reason="MCP client not yet wired — implement in Phase 2")
def test_undocumented_honesty(mcp_client):
    """LLM should admit when a column has no documentation."""
    question = "What does the created_at column in products mean?"
    response = mcp_client.ask(question)

    from deepeval import assert_test
    from deepeval.test_case import LLMTestCase
    from deepeval.metrics import AnswerRelevancyMetric

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        expected_output="This column is not documented in context.yaml.",
    )
    assert_test(test_case, [AnswerRelevancyMetric(threshold=0.5)])


@pytest.mark.skip(reason="MCP client not yet wired — implement in Phase 2")
def test_nonexistent_column(mcp_client):
    """LLM should clearly state the column doesn't exist."""
    question = "What does the favorite_color column in ingredients mean?"
    response = mcp_client.ask(question)

    from deepeval import assert_test
    from deepeval.test_case import LLMTestCase
    from deepeval.metrics import AnswerRelevancyMetric

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        expected_output="This column does not exist in the database.",
    )
    assert_test(test_case, [AnswerRelevancyMetric(threshold=0.5)])
