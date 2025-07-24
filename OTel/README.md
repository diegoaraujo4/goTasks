# OTEL - Open Telemetry Weather Services

Sistema distribuído para consulta de temperatura por CEP brasileiro, baseado em arquitetura de microserviços.

## Arquitetura

O sistema é composto por dois serviços principais:

### Serviço A - Gateway (otel-gateway)
**Responsabilidade:** Input e Validação
- **Porta:** 8080
- **Função:** Recebe e valida inputs de CEP, encaminhando para o serviço de orquestração

### Serviço B - Orchestration (otel-orchestration)  
**Responsabilidade:** Orquestração e Processamento
- **Porta:** 8081
- **Função:** Processa CEPs, consulta APIs externas e retorna dados de temperatura

## API do Gateway (Serviço A)

### POST /cep
Recebe um CEP para validação e processamento.

**Request Body:**
```json
{
  "cep": "29902555"
}
```

**Validações:**
- CEP deve ser uma string
- CEP deve conter exatamente 8 dígitos
- CEP deve conter apenas números

**Responses:**

**Sucesso (200):**
```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

**CEP Inválido (422):**
```json
{
  "message": "invalid zipcode"
}
```

**Request Inválido (400):**
```json
{
  "message": "invalid request body"
}
```

### GET /health
Health check do gateway.

**Response (200):**
```json
{
  "status": "healthy",
  "service": "otel-gateway"
}
```

## API do Orchestration (Serviço B)

### GET /weather/{cep}
Consulta temperatura por CEP (8 dígitos).

**Response (200):**
```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

**CEP Inválido (422):**
```json
{
  "message": "invalid zipcode"
}
```

**CEP Não Encontrado (404):**
```json
{
  "message": "can not find zipcode"
}
```

### GET /health
Health check do serviço de orquestração.

## Como Executar

### Docker Compose (Recomendado)
```bash
docker-compose up --build
```

### Executar Individualmente

#### Gateway (Serviço A)
```bash
# Terminal 1
export ORCHESTRATION_SERVICE_URL=http://localhost:8081
export PORT=8080
go run cmd/gateway/main.go
```

#### Orchestration (Serviço B)
```bash
# Terminal 2
export WEATHER_API_KEY=your_api_key
export PORT=8081
go run cmd/orchestrator/main.go
```

## Testes

### Executar todos os testes
```bash
go test ./...
```

### Executar testes do Gateway
```bash
go test ./internal/gateway/...
```

### Executar testes do Orchestration
```bash
go test ./internal/...
```

## Exemplos de Uso

### Testando o Gateway
```bash
# CEP válido
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "29902555"}'

# CEP inválido
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'
```

### Testando o Orchestration diretamente
```bash
curl http://localhost:8081/weather/29902-555
```

## Estrutura do Projeto

```
.
├── cmd/
│   ├── orchestrator/  # Serviço B (Orchestration)
│   └── gateway/       # Serviço A (Gateway)
├── internal/
│   ├── gateway/       # Lógica do Gateway
│   ├── handler/       # Handlers do Orchestration
│   ├── repository/    # Repositórios (ViaCEP, WeatherAPI)
│   └── service/       # Serviços de negócio
├── pkg/
│   ├── temperature/   # Conversor de temperatura
│   └── validator/     # Validador de CEP
├── config/            # Configurações
├── docs/              # Documentação Swagger
├── docker-compose.yml       # Orquestração dos serviços
├── Dockerfile.orchestration # Serviço B (Orchestration)
└── Dockerfile.gateway       # Serviço A (Gateway)
```

## Variáveis de Ambiente

### Gateway (Serviço A)
- `PORT`: Porta do serviço (padrão: 8080)
- `ORCHESTRATION_SERVICE_URL`: URL do serviço de orquestração (padrão: http://localhost:8081)

### Orchestration (Serviço B)
- `PORT`: Porta do serviço (padrão: 8081)
- `WEATHER_API_KEY`: Chave da API Weather (obrigatória)

## Health Checks

Ambos os serviços expõem endpoints de health check:
- Gateway: http://localhost:8080/health
- Orchestration: http://localhost:8081/health
