# CEP API Challenge - Multithreading

Este projeto implementa um desafio de multithreading em Go que busca informações de CEP em duas APIs diferentes e retorna a resposta mais rápida.

## APIs utilizadas

1. **BrasilAPI**: `https://brasilapi.com.br/api/cep/v1/{cep}`
2. **ViaCEP**: `http://viacep.com.br/ws/{cep}/json/`

## Funcionalidades

- ✅ Requisições simultâneas para ambas as APIs usando goroutines
- ✅ Retorna a resposta da API mais rápida
- ✅ Timeout de 1 segundo
- ✅ Exibe os dados do endereço e qual API respondeu primeiro
- ✅ Tratamento de erros e validações

## Como usar

### Opção 1: Executar diretamente
```bash
# Executar o programa
go run main.go <CEP>

# Exemplo com CEP válido
go run main.go 01153000

# Exemplo com outro CEP
go run main.go 20040020
```

### Opção 2: Compilar e executar
```bash
# Usando Go build diretamente
go build -o cep-challenge.exe main.go
.\cep-challenge.exe 01153000

# Usando Makefile (se make estiver disponível)
make build
.\cep-challenge.exe 01153000
```

### Opção 3: Executar testes
```bash
# Executar testes com múltiplos CEPs
make test

# Executar apenas um CEP específico
make run
```

## Exemplo de saída

```
=== RESULTADO MAIS RÁPIDO ===
API: BrasilAPI
CEP: 01153-000
Logradouro: Rua Vitorino Carmilo
Bairro: Campos Elíseos
Cidade: São Paulo
Estado: SP
```

## Estrutura do código

- **main.go**: Arquivo principal com toda a lógica
- **Structs**: `BrasilAPIResponse`, `ViaCEPResponse` e `CEPResult`
- **Funções**: `fetchBrasilAPI`, `fetchViaCEP` e `main`

## Tecnologias

- Go 1.22.5
- Goroutines para concorrência
- Channels para comunicação entre goroutines
- HTTP client nativo do Go
- JSON encoding/decoding

## Tratamento de erros

- Timeout após 1 segundo
- Validação de argumentos de linha de comando
- Tratamento de erros HTTP
- Tratamento de erros JSON
