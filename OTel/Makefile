.PHONY: build-gateway build-orchestration run-gateway run-orchestration test test-gateway test-orchestration docker-build docker-run docker-stop

# Variáveis
GATEWAY_NAME=otel-gateway
ORCHESTRATION_NAME=otel-orchestration
GATEWAY_IMAGE=otel-gateway:latest
ORCHESTRATION_IMAGE=otel-orchestration:latest

# Build local
build-gateway:
	go build -o bin/gateway ./cmd/gateway

build-orchestration:
	go build -o bin/orchestration ./cmd/api

# Executar localmente
run-gateway:
	PORT=8081 ORCHESTRATION_SERVICE_URL=http://localhost:8080 go run ./cmd/gateway

run-orchestration:
	PORT=8080 WEATHER_API_KEY=34d03a56db334a6caca234735252207 go run ./cmd/api

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
