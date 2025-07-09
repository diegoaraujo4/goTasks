build:
	go build -o cep-challenge.exe main.go

run:
	go run main.go 01153000

test:
	@echo Testing multiple CEPs...
	@echo Testing CEP 01153000 (São Paulo):
	go run main.go 01153000
	@echo.
	@echo Testing CEP 20040020 (Rio de Janeiro):
	go run main.go 20040020
	@echo.
	@echo Testing CEP 04038001 (São Paulo - Vila Clementino):
	go run main.go 04038001
	@echo.
	@echo Testing CEP 30112000 (Belo Horizonte):
	go run main.go 30112000
	@echo.
	@echo Testing invalid CEP format:
	go run main.go 123 || echo "Expected error for invalid CEP format"
	@echo.
	@echo All tests completed!

clean:
	del /f cep-challenge.exe 2>nul || echo "No files to clean"

fmt:
	go fmt ./...

vet:
	go vet ./...

help:
	@echo Available targets:
	@echo   build  - Build the executable
	@echo   run    - Run with test CEP
	@echo   test   - Run tests with multiple CEPs
	@echo   clean  - Clean build artifacts
	@echo   fmt    - Format code
	@echo   vet    - Check for issues
	@echo   help   - Show this help

.PHONY: build run test clean fmt vet help
