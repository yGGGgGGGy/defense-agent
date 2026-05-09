"""Tests for LLM client in mock mode."""

import sys
sys.path.insert(0, ".")

from core.llm_client import LLMClient


def test_mock_mode():
    client = LLMClient()  # no API key = mock mode
    assert client.mock_mode


def test_mock_perceiver():
    client = LLMClient()
    response = client._mock_response("system prompt for perceiver agent")
    assert "mock" in response.model


def test_mock_default():
    client = LLMClient()
    response = client._mock_response("general instructions")
    assert response.content != ""
