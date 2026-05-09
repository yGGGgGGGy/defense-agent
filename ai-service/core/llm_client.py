"""LLM provider abstraction with mock mode and real API support."""

import os
from dataclasses import dataclass, field
from typing import Any


@dataclass
class ChatMessage:
    role: str
    content: str


@dataclass
class ChatResponse:
    content: str
    tool_calls: list[dict] = field(default_factory=list)
    model: str = "mock"
    usage: dict = field(default_factory=dict)


MOCK_RESPONSES = {
    "perceiver": ChatResponse(
        content="Perceiver: Security event detected. Multiple indicators of compromise found. Initial asset discovery and log collection initiated.",
        model="mock-perceiver",
    ),
    "analyst": ChatResponse(
        content="Analyst: Threat assessed and mapped to ATT&CK T1110 (Brute Force). Severity: CRITICAL. Correlated 3 alerts from single source IP. Recommend immediate containment.",
        model="mock-analyst",
    ),
    "responder": ChatResponse(
        content="Responder: Containment actions executed. Blocked source IP via iptables. Isolated affected host. Verified service integrity post-action.",
        model="mock-responder",
    ),
    "operator": ChatResponse(
        content="Operator: System health check completed. Logs rotated and archived. Configuration snapshot created. All services running normally.",
        model="mock-operator",
    ),
    "researcher": ChatResponse(
        content="Researcher: CVE database search complete. Found 3 relevant CVEs. IOC correlation positive. Threat intelligence gathered.",
        model="mock-researcher",
    ),
    "developer": ChatResponse(
        content="Developer: Generated response plan with 3 prioritized phases. Tool selection: nmap, hydra, metasploit. Risk assessment included.",
        model="mock-developer",
    ),
    "executor": ChatResponse(
        content="Executor: All planned tools executed successfully in sandbox. Results collected and normalized. No errors encountered.",
        model="mock-executor",
    ),
    "adviser": ChatResponse(
        content="Adviser: Agent progressing normally toward objectives. No loops detected. Execution efficiency: optimal.",
        model="mock-adviser",
    ),
    "reflector": ChatResponse(
        content="Reflector: Attempt 1/3. Error analyzed. Suggested: verify tool arguments, try alternative approach.",
        model="mock-reflector",
    ),
    "auditor": ChatResponse(
        content="Auditor: All decisions reviewed. 8/10 low risk (auto-approved), 2/10 medium risk (approved with notes). Compliance: OK.",
        model="mock-auditor",
    ),
    "memorist": ChatResponse(
        content="Memorist: Stored execution results to long-term memory. Found 3 similar past cases for pattern comparison.",
        model="mock-memorist",
    ),
    "default": ChatResponse(
        content="Task analyzed and completed successfully. All steps executed as planned.",
        model="mock-default",
    ),
}


class LLMClient:
    """LLM client supporting OpenAI, Anthropic, DeepSeek, and mock mode."""

    def __init__(self, api_key: str = "", base_url: str = "", model: str = ""):
        self.api_key = api_key
        self.base_url = base_url
        self.model = model
        self.mock_mode = not api_key
        self.provider = self._detect_provider()

    def _detect_provider(self) -> str:
        if not self.base_url:
            return "mock"
        if "openai.com" in self.base_url:
            return "openai"
        if "anthropic.com" in self.base_url:
            return "anthropic"
        if "deepseek.com" in self.base_url:
            return "deepseek"
        return "custom"

    async def chat(
        self,
        messages: list[dict[str, str]],
        system_prompt: str = "",
        temperature: float = 0.0,
        tools: list[dict[str, Any]] | None = None,
    ) -> ChatResponse:
        if self.mock_mode:
            return self._mock_response(system_prompt)

        try:
            import httpx
            async with httpx.AsyncClient(timeout=120.0) as client:
                payload = self._build_payload(messages, system_prompt, temperature, tools)
                headers = self._build_headers()
                resp = await client.post(
                    f"{self.base_url}/chat/completions",
                    json=payload,
                    headers=headers,
                )
                resp.raise_for_status()
                return self._parse_response(resp.json())
        except Exception as e:
            print(f"LLM call failed: {e}, falling back to mock")
            return self._mock_response(system_prompt)

    def _build_payload(self, messages, system_prompt, temperature, tools):
        full_messages = []
        if system_prompt:
            full_messages.append({"role": "system", "content": system_prompt})
        full_messages.extend(messages)

        payload = {
            "model": self.model or "gpt-4o",
            "messages": full_messages,
            "temperature": temperature,
        }
        if tools:
            payload["tools"] = tools
            payload["tool_choice"] = "auto"
        return payload

    def _build_headers(self):
        return {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
        }

    def _parse_response(self, data: dict) -> ChatResponse:
        choice = data["choices"][0]["message"]
        return ChatResponse(
            content=choice.get("content", ""),
            tool_calls=choice.get("tool_calls", []),
            model=data.get("model", ""),
            usage=data.get("usage", {}),
        )

    def _mock_response(self, system_prompt: str) -> ChatResponse:
        for agent_type, resp in MOCK_RESPONSES.items():
            if agent_type in system_prompt.lower():
                return resp
        return MOCK_RESPONSES["default"]

    def set_api_key(self, key: str):
        self.api_key = key
        self.mock_mode = not key

    def set_model(self, model: str):
        self.model = model
