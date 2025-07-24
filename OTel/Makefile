.PHONY: build-gateway build-orchestration run-gateway run-orchestration test test-gateway test-orchestration docker-build docker-run docker-stop run-with-tracing

# Variáveis
GATEWAY_NAME=otel-gateway
ORCHESTRATION_NAME=otel-orchestration
GATEWAY_IMAGE=otel-gateway:latest
ORCHESTRATION_IMAGE=otel-orchestration:latest
ZIPKIN_URL=http://localhost:9411/api/v2/spans

# Build local
build-gateway:
	go build -o bin/gateway ./cmd/gateway

build-orchestration:
	go build -o bin/orchestration ./cmd/orchestrator

# Executar localmente sem tracing
run-gateway:
	PORT=8080 ORCHESTRATION_SERVICE_URL=http://localhost:8081 go run ./cmd/gateway

run-orchestration:
	PORT=8081 WEATHER_API_KEY=34d03a56db334a6caca234735252207 go run ./cmd/orchestrator

# Executar localmente com tracing (requer Zipkin rodando)
run-gateway-with-tracing:
	PORT=8080 ORCHESTRATION_SERVICE_URL=http://localhost:8081 ZIPKIN_URL=$(ZIPKIN_URL) go run ./cmd/gateway

run-orchestration-with-tracing:
	PORT=8081 WEATHER_API_KEY=34d03a56db334a6caca234735252207 ZIPKIN_URL=$(ZIPKIN_URL) go run ./cmd/orchestrator

# Executar sistema completo com tracing
run-with-tracing:
	@echo "Starting Zipkin..."
	docker run -d --name zipkin-temp -p 9411:9411 openzipkin/zipkin || true
	@echo "Waiting for Zipkin to be ready..."
	@sleep 5
	@echo "Starting services with tracing..."
	@echo "Open http://localhost:9411 for Zipkin UI"
	@echo "Services will be available at:"
	@echo "  Gateway: http://localhost:8080"
	@echo "  Orchestration: http://localhost:8081"

# Parar Zipkin temporário
stop-zipkin:
	docker stop zipkin-temp || true
	docker rm zipkin-temp || true

# Testes
test:
	go test ./...

test-gateway:
	go test ./internal/gateway/...

test-orchestration:
	go test ./internal/handler/... ./internal/service/... ./internal/repository/...

# Docker - Build das imagens
docker-build:
	docker build -t $(ORCHESTRATION_IMAGE) .
	docker build -t $(GATEWAY_IMAGE) -f Dockerfile.gateway .

# Executar com Docker Compose
docker-run:
	docker-compose up --build

# Parar containers
docker-stop:
	docker-compose down

# Executar testes em container
docker-test:
	docker build -f Dockerfile.test -t $(APP_NAME)-test .
	docker run --rm $(APP_NAME)-test
