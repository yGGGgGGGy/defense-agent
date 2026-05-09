# Defense Agent System

自主决策通用防御智能体 — an autonomous multi-agent system for cybersecurity defense operations.

Built with Go + React + Python, powered by 11 specialized AI agents collaborating through a DAG-based orchestration engine.

## Overview

Defense Agent is a self-hosted, AI-driven security operations platform that enables autonomous threat perception, analysis, response, and maintenance. It transitions cybersecurity defense from **automation to intelligence** — each decision carries a complete audit trail with SHA256 tamper-evident chain.

## Architecture

```
┌────────────────────────────────────────┐
│         React + TypeScript UI          │
│    Dashboard │ DAG View │ Audit Chain  │
└─────────────────┬──────────────────────┘
                  │ REST / SSE
┌─────────────────▼──────────────────────┐
│        Go Orchestrator Engine          │
│  ┌──────────┐ ┌──────┐ ┌───────────┐  │
│  │ DAG Sched│ │Planner│ │Checkpoint │  │
│  └──────────┘ └──────┘ └───────────┘  │
│  ┌──────────────────────────────────┐  │
│  │        11 AI Agents              │  │
│  │  Perceiver  Analyst  Responder   │  │
│  │  Operator   Researcher Developer │  │
│  │  Executor   Adviser   Reflector  │  │
│  │  Auditor    Memorist             │  │
│  └──────────────────────────────────┘  │
└─────────────────┬──────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    ▼             ▼             ▼
┌────────┐ ┌──────────┐ ┌──────────┐
│Postgres│ │  Neo4j   │ │  NATS    │
│+pgvect │ │ Knowledge │ │ Message  │
│ Memory │ │  Graph   │ │  Queue   │
└────────┘ └──────────┘ └──────────┘
```

## Features

### Multi-Agent Collaboration
11 specialized agents work together through a DAG (Directed Acyclic Graph) engine. Kahn's topological sort determines execution order, with parallel-safe groups running concurrently.

| Agent | Role | Risk Level |
|-------|------|------------|
| **Perceiver** | Threat perception, log collection, asset discovery | General (100 calls) |
| **Analyst** | Alert correlation, ATT&CK mapping, root cause analysis | General |
| **Responder** | IP blocking, host isolation, service recovery | General |
| **Operator** | Health check, compliance, backup, certificate management | General |
| **Researcher** | CVE lookup, IOC correlation, threat intelligence | Limited (20 calls) |
| **Developer** | Attack planning, tool selection, strategy generation | Limited |
| **Executor** | Tool dispatch, sandbox orchestration, result collection | General |
| **Adviser** | Execution monitoring, loop detection, mentor guidance | Limited |
| **Reflector** | Failure analysis, error recovery, graceful termination | Limited |
| **Auditor** | Decision review, evidence verification, compliance check | Limited |
| **Memorist** | Long-term memory store/retrieve, pattern extraction | Limited |

### 5 Operational Scenes

| Scene | DAG Flow |
|-------|----------|
| **Incident Response** | Perceiver → Analyst → (Responder ∥ Operator) |
| **Penetration Test** | Researcher → Developer → Executor → Auditor |
| **Vulnerability Research** | Researcher → Developer → Executor → Memorist |
| **Reverse Engineering** | (Static Perceiver ∥ Dynamic Perceiver) → Analyst → Executor → Memorist |
| **Ops Maintenance** | Operator (health, compliance, backup, logs) |

### Audit Chain
Every agent action is recorded in a SHA256 tamper-evident hash chain with a **4-tier security gate**:

| Risk Level | Approval | Examples |
|------------|----------|---------|
| **Low** | Auto-approved | Log queries, port scans, CVE search |
| **Medium** | Auditor auto-review | IP blocking, config changes |
| **High** | **Human approval required** | Host isolation, traffic blocking |
| **Dangerous** | **Mandatory human approval** | Database drops, kernel changes |

### Real-time SSE Streaming
Frontend receives live updates via Server-Sent Events: node state changes, audit record appends, instance status transitions.

### Knowledge Graph
Neo4j stores attack paths, asset relationships, and decision chains. All agent actions are automatically persisted as graph nodes for semantic querying.

### Dynamic Replanning & Checkpoint
- Failed nodes auto-retry up to `max_retries`
- Substitute agents deployed when retries exhausted
- Automatic checkpoint snapshots at every DAG group completion
- Full rollback to last checkpoint on critical failure

## Quick Start

### Prerequisites
- Go 1.23+, Docker, Python 3.11+, Node 24+

### Start Everything

```bash
make dev
# Starts: Docker infra → AI service (:8100) → Orchestrator (:8080)
```

### Create a Task

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "scene": "incident_response",
    "title": "SSH Brute Force Alert",
    "description": "150 failed SSH logins from 10.0.0.50",
    "input": "SSH brute force detected",
    "alerts": [{"id": "A1", "rule": "SSH_BRUTE_FORCE", "source_ip": "10.0.0.50", "count": 150}]
  }'
```

### Multi-Instance Parallel Test

```bash
make test-multi
# Submits 3 tasks simultaneously and verifies audit chains
```

### Run Tests

```bash
make test                # Unit tests
make test-integration    # Integration tests
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/health` | Health check |
| `POST` | `/api/v1/tasks` | Submit new task |
| `GET` | `/api/v1/tasks/:id` | Task status + DAG state |
| `POST` | `/api/v1/tasks/:id/cancel` | Cancel running task |
| `GET` | `/api/v1/tasks/:id/audit` | Full audit trail |
| `GET` | `/api/v1/agents` | List 11 agents |
| `GET` | `/api/v1/instances` | All instances |
| `GET` | `/api/v1/events` | SSE stream |
| `GET` | `/metrics` | Prometheus metrics |

## Services

| Service | Port | Credentials |
|---------|------|-------------|
| Go Orchestrator | 8080 | — |
| React Frontend | 5173 | — |
| Python AI Service | 8100 | — |
| PostgreSQL + pgvector | 5432 | defense/defense123 |
| Neo4j | 7474/7687 | neo4j/defense123 |
| NATS | 4222/8222 | — |
| Prometheus | 9090 | — |
| Grafana | 3000 | admin/defense123 |

## Configuring LLM

By default, agents run in **mock mode** (no API key required). To use a real LLM:

```bash
# Set via the AI service configuration endpoint
curl -X POST http://localhost:8100/v1/config \
  -H "Content-Type: application/json" \
  -d '{"api_key": "sk-...", "base_url": "https://api.openai.com/v1", "model": "gpt-4o"}'

# Or set environment variables
export LLM_API_KEY="sk-..."
export LLM_BASE_URL="https://api.deepseek.com/v1"
export LLM_MODEL="deepseek-chat"
```

## Project Structure

```
├── backend/
│   ├── cmd/orchestrator/main.go     # Entry point
│   ├── internal/
│   │   ├── agent/      # 11 agent implementations
│   │   ├── dag/        # DAG scheduler & executor
│   │   ├── orchestrator/ # Main engine + scene router
│   │   ├── audit/      # SHA256 chain + security gate
│   │   ├── memory/     # pgvector memory store
│   │   ├── comm/       # NATS message bus
│   │   ├── graphiti/   # Neo4j knowledge graph
│   │   ├── sandbox/    # Docker sandbox
│   │   ├── tools/      # Agent tools (search, terminal)
│   │   ├── sse/        # Server-Sent Events broker
│   │   ├── api/        # HTTP handlers
│   │   └── types/      # Shared type definitions
│   └── tests/integration/
├── ai-service/          # Python FastAPI AI microservice
├── frontend/            # React + TypeScript + Vite
├── observability/       # Prometheus + Grafana configs
├── docker-compose.yml   # Core infrastructure
└── docker-compose-observability.yml  # Monitoring stack
```

## Key Design Decisions

- **DAG over linear Flow** — enables parallel execution and dynamic replanning
- **Per-instance audit chains** — prevents cross-instance hash contamination
- **Mock-first agent design** — all agents work without LLM API keys
- **SSE over WebSocket** — simpler, HTTP-native, auto-reconnect
- **Go channels + NATS fallback** — works in-memory for single process, NATS for distributed
- **Allowlist over full sandbox** — safe commands execute directly, dangerous ones require Docker
