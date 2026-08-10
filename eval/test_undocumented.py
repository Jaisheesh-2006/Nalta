"""
Tests: LLM must correctly handle completely nonexistent columns and tables.

These tests verify the LLM propagates 'not found' signals from the MCP server
rather than inventing schema information from its training knowledge.

Assertion: Direct string match on the required phrase from the system prompt:
  'This column does not exist in the database.'

Note: AnswerRelevancyMetric and HallucinationMetric are NOT used here — both score
refusals poorly because refusals are intentionally not-relevant and may not perfectly
mirror the retrieval_context wording (triggering false hallucination flags).
"""

NOT_EXIST_PHRASE = "does not exist in the database"


def test_nonexistent_column(mcp_client):
    """LLM must state that patients.favorite_color does not exist — not guess a meaning."""
    question = "What does the favorite_color column in the patients table represent?"
    response = mcp_client.ask(question, table="patients", column="favorite_color")

    assert NOT_EXIST_PHRASE.lower() in response.lower(), (
        f"Expected 'does not exist in the database' response.\nGot: {response[:300]}"
    )


def test_nonexistent_table(mcp_client):
    """LLM must state the appointments table does not exist — not invent its schema."""
    question = "What does the scheduled_at column in the appointments table mean?"
    response = mcp_client.ask(question, table="appointments", column="scheduled_at")

    assert NOT_EXIST_PHRASE.lower() in response.lower(), (
        f"Expected 'does not exist in the database' response.\nGot: {response[:300]}"
    )
