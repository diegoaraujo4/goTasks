.PHONY: docker-build docker-run docker-test docker-stop

# Variáveis
APP_NAME=weather-api
DOCKER_IMAGE=weather-api:latest

# Build da imagem Docker
docker-build:
	docker build -t $(DOCKER_IMAGE) .

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
