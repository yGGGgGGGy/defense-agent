"""FastAPI AI inference microservice for the defense agent system."""

import os
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

from core.llm_client import LLMClient

app = FastAPI(title="Defense Agent AI Service", version="0.2.0")
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])

llm = LLMClient(
    api_key=os.getenv("LLM_API_KEY", ""),
    base_url=os.getenv("LLM_BASE_URL", "https://api.openai.com/v1"),
    model=os.getenv("LLM_MODEL", "gpt-4o"),
)


class ChatPayload(BaseModel):
    messages: list[dict] = []
    system_prompt: str = ""
    temperature: float = 0.0
    tools: list[dict] | None = None


class ConfigPayload(BaseModel):
    api_key: str | None = None
    base_url: str | None = None
    model: str | None = None


@app.get("/v1/health")
async def health():
    return {"status": "ok", "mode": "mock" if llm.mock_mode else f"live ({llm.provider})", "model": llm.model}


@app.post("/v1/chat")
async def chat(payload: ChatPayload):
    response = await llm.chat(
        messages=payload.messages,
        system_prompt=payload.system_prompt,
        temperature=payload.temperature,
        tools=payload.tools,
    )
    return {
        "content": response.content,
        "tool_calls": response.tool_calls,
        "model": response.model,
        "usage": response.usage,
    }


@app.post("/v1/config")
async def configure(payload: ConfigPayload):
    if payload.api_key:
        llm.set_api_key(payload.api_key)
    if payload.base_url:
        llm.base_url = payload.base_url
        llm.provider = llm._detect_provider()
    if payload.model:
        llm.set_model(payload.model)
    return {"status": "configured", "mode": "mock" if llm.mock_mode else f"live ({llm.provider})", "model": llm.model}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("AI_SERVICE_PORT", "8100"))
    uvicorn.run("cmd.server:app", host="0.0.0.0", port=port, reload=True)
