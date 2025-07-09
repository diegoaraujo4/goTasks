# Go USD/BRL Exchange Rate Service
.PHONY: all build server client clean run-server run-client docker help

all: build

build:
	@cd cmd/server && go build -o ../../server
	@cd cmd/client && go build -o ../../client

server:
	@cd cmd/server && go build -o ../../server

client:
	@cd cmd/client && go build -o ../../client

run-server: server
	@./server

run-client: client
	@./client

docker:
	@docker-compose up --build

clean:
	@rm -f server client server.exe client.exe quotes.db cotacao.txt

help:
	@echo "Go USD/BRL Exchange Rate Service - Available targets:"
	@echo ""
	@echo "  build        Build server and client binaries"
	@echo "  server       Build server binary only"
	@echo "  client       Build client binary only"
	@echo "  run-server   Build and start the server"
	@echo "  run-client   Build and run the client"
	@echo "  docker       Start services with Docker Compose"
	@echo "  clean        Remove built binaries and data files"
	@echo "  help         Show this help message"
