#!/usr/bin/env bash
set -euo pipefail
export PATH="$HOME/.local/bin:$PATH"

echo "=== Starting Defense Agent System ==="

# 1. Start Docker infrastructure
echo ">>> Starting Docker infrastructure..."
sudo docker compose up -d 2>&1 | tail -3
sleep 3

# 2. Start Python AI service in background
echo ">>> Starting AI service on :8100 (mock mode)..."
cd ai-service
PYTHONPATH=. python3 -m uvicorn cmd.server:app --host 0.0.0.0 --port 8100 &
AI_PID=$!
cd ..
sleep 2

# 3. Build and run Go orchestrator
echo ">>> Building orchestrator..."
cd backend
GOTOOLCHAIN=local go build -o bin/orchestrator ./cmd/orchestrator
echo ">>> Starting orchestrator on :8080..."
./bin/orchestrator &
ORCH_PID=$!
cd ..

echo ""
echo "=== All services started ==="
echo "  API:       http://localhost:8080"
echo "  AI:        http://localhost:8100"
echo "  Postgres:  localhost:5432"
echo "  NATS:      localhost:4222"
echo "  Neo4j:     localhost:7474"
echo ""
echo "PIDs: AI=$AI_PID  ORCH=$ORCH_PID"
echo "Stop with: kill $AI_PID $ORCH_PID && sudo docker compose stop"

wait
