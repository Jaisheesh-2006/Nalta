"""Tests: LLM descriptions should be grounded in context.yaml."""
import pytest


@pytest.mark.skip(reason="MCP client not yet wired — implement in Phase 2")
def test_grounded_description(mcp_client):
    """LLM should describe toxicity_class using context.yaml's definition."""
    question = "What does the toxicity_class column in ingredients mean?"
    response = mcp_client.ask(question)

    from deepeval import assert_test
    from deepeval.test_case import LLMTestCase
    from deepeval.metrics import FaithfulnessMetric

    context = [
        "Internal safety tier: 'safe', 'restricted', or 'banned'. "
        "Drives formulation guardrails."
    ]

    test_case = LLMTestCase(
        input=question,
        actual_output=response,
        retrieval_context=context,
    )
    assert_test(test_case, [FaithfulnessMetric(threshold=0.7)])
