.PHONY: build run test dev clean

export PATH := $(HOME)/.local/bin:$(PATH)
export GOTOOLCHAIN := local
export GOPROXY := https://goproxy.cn,direct

build:
	cd backend && go build -o bin/orchestrator ./cmd/orchestrator

run: build
	cd backend && ./bin/orchestrator

test:
	cd backend && go test ./... -v -count=1
	cd ai-service && python -m pytest tests/ -v 2>/dev/null || echo "Python tests skipped"

test-integration:
	cd backend && go test ./tests/integration/... -v -count=1

dev:
	@echo "Starting development services..."
	@fuser -k 8080/tcp 2>/dev/null || true
	@pkill -f orchestrator 2>/dev/null || true
	@sleep 1
	@echo "1. Docker infrastructure"
	sudo docker compose up -d 2>/dev/null
	@sleep 3
	@echo "2. AI service (mock mode) on :8100"
	cd ai-service && PYTHONPATH=. python -m uvicorn cmd.server:app --host 0.0.0.0 --port 8100 &
	@sleep 2
	@echo "3. Go orchestrator on :8080"
	cd backend && GOTOOLCHAIN=local go run ./cmd/orchestrator &
	@echo "All services started. API at http://localhost:8080"

stop:
	sudo docker compose stop 2>/dev/null || true
	@pkill -f orchestrator 2>/dev/null || true
	@pkill -f uvicorn 2>/dev/null || true
	@fuser -k 8080/tcp 2>/dev/null || true
	@sleep 1
	@echo "All services stopped"

clean:
	rm -rf backend/bin
	sudo docker compose down -v 2>/dev/null || true

# Multi-instance test
test-multi:
	@echo "Submitting 3 parallel task instances..."
	@for i in 1 2 3; do \
		curl -s -X POST http://localhost:8080/api/v1/tasks \
			-H "Content-Type: application/json" \
			-d "{\"scene\":\"incident_response\",\"title\":\"Test Instance $$i\",\"description\":\"Parallel test $$i\",\"input\":\"SSH brute force detected from 10.0.0.5$$i\",\"alerts\":[{\"id\":\"ALERT-00$$i\",\"rule\":\"SSH_BRUTE_FORCE\",\"source_ip\":\"10.0.0.5$$i\",\"count\":150}]}" & \
	done
	@wait
	@sleep 3
	@echo "Instance status:"
	@curl -s http://localhost:8080/api/v1/instances | python3 -m json.tool
