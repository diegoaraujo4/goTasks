.PHONY: build run test clean docker-build docker-run docker-test swagger-install swagger-gen swagger-serve

# Variáveis
APP_NAME=weather-api
DOCKER_IMAGE=weather-api:latest

# Build da aplicação
build:
	go build -o $(APP_NAME) ./cmd/api

# Executar a aplicação localmente
run:
	go run ./cmd/api

# Executar testes
test:
	go test -v ./...

# Executar testes com coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Instalar ferramenta swag
swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

# Gerar documentação Swagger
swagger-gen:
	swag init -g cmd/api/main.go

# Executar aplicação e abrir Swagger UI
swagger-serve: swagger-gen
	@echo "Starting server with Swagger documentation..."
	@echo "Swagger UI available at: http://localhost:8080/swagger/index.html"
	go run ./cmd/api

# Limpar arquivos gerados
clean:
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	rm -rf docs/

# Build da imagem Docker
docker-build:
	docker build -f deployments/Dockerfile -t $(DOCKER_IMAGE) .

# Executar com Docker Compose
docker-run:
	cd deployments && docker-compose up --build

# Parar containers
docker-stop:
	docker-compose down

# Executar testes em container
docker-test:
	docker build -f Dockerfile.test -t $(APP_NAME)-test .
	docker run --rm $(APP_NAME)-test

# Baixar dependências
deps:
	go mod download
	go mod tidy

# Verificar código
lint:
	go vet ./...
	go fmt ./...

# Deploy no Google Cloud Run (requer gcloud configurado)
deploy:
	gcloud run deploy weather-api \
		--source . \
		--platform managed \
		--region us-central1 \
		--allow-unauthenticated \
		--set-env-vars WEATHER_API_KEY=$(WEATHER_API_KEY)
