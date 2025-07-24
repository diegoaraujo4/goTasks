# OTEL - Open Telemetry Weather Services

Sistema distribuído para consulta de temperatura por CEP brasileiro, baseado em arquitetura de microserviços com **OpenTelemetry** e **Zipkin** para observabilidade e tracing distribuído.

## Arquitetura

O sistema é composto por dois serviços principais com tracing distribuído:

### Serviço A - Gateway (otel-gateway)
**Responsabilidade:** Input e Validação
- **Porta:** 8080
- **Função:** Recebe e valida inputs de CEP, encaminhando para o serviço de orquestração
- **Tracing:** Instrumentado com OpenTelemetry para rastreamento de requests

### Serviço B - Orchestration (otel-orchestration)  
**Responsabilidade:** Orquestração e Processamento
- **Porta:** 8081
- **Função:** Processa CEPs, consulta APIs externas e retorna dados de temperatura
- **Tracing:** Instrumentado com spans detalhados para cada operação

### Zipkin - Distributed Tracing
**Responsabilidade:** Coleta e Visualização de Traces
- **Porta:** 9411
- **Função:** Interface web para visualização de traces distribuídos
- **UI:** http://localhost:9411

## OpenTelemetry Features

### Instrumentação Automática
- **HTTP Requests:** Todos os requests HTTP são automaticamente instrumentados
- **Gorilla Mux:** Middleware automático para rotas
- **Client HTTP:** Instrumentação de chamadas para APIs externas

### Spans Customizados
- **CEP Validation:** Medição do tempo de validação
- **Location Lookup:** Tracing de consultas ao ViaCEP
- **Weather Fetch:** Tracing de consultas à WeatherAPI
- **Temperature Conversion:** Medição de conversões de temperatura

### Métricas e Atributos
- **Request Duration:** Tempo total de processamento
- **CEP Input/Output:** Rastreamento de entrada e saída
- **API Response Times:** Tempo de resposta das APIs externas
- **Error Tracking:** Rastreamento de erros com stack traces

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

Este comando iniciará:
- **Gateway Service** em http://localhost:8080
- **Orchestration Service** em http://localhost:8081  
- **Zipkin** em http://localhost:9411

### Executar Individualmente

#### Gateway (Serviço A)
```bash
# Terminal 1
export ORCHESTRATION_SERVICE_URL=http://localhost:8081
export ZIPKIN_URL=http://localhost:9411/api/v2/spans
export PORT=8080
go run cmd/gateway/main.go
```

#### Orchestration (Serviço B)
```bash
# Terminal 2
export WEATHER_API_KEY=your_api_key
export ZIPKIN_URL=http://localhost:9411/api/v2/spans
export PORT=8081
go run cmd/orchestrator/main.go
```

#### Zipkin (Tracing)
```bash
# Terminal 3 - Usando Docker
docker run -d -p 9411:9411 openzipkin/zipkin
```

## Observabilidade e Tracing

### Zipkin Dashboard
Acesse http://localhost:9411 para visualizar:
- **Traces distribuídos** entre Gateway e Orchestration
- **Spans detalhados** de cada operação
- **Dependências** entre serviços
- **Performance metrics** e latência

### Spans Implementados

#### Gateway Service
- `gateway.process_cep` - Processamento completo da requisição
- `gateway.validate_cep` - Validação do formato do CEP
- `gateway.call_orchestration_service` - Chamada para o serviço de orquestração

#### Orchestration Service  
- `orchestration.get_weather_by_cep` - Processamento completo
- `weather_service.get_weather_by_cep` - Lógica de negócio
- `weather_service.validate_cep` - Validação do CEP
- `weather_service.get_location_by_cep` - Consulta ao ViaCEP
- `weather_service.get_weather_by_location` - Consulta à WeatherAPI
- `weather_service.convert_temperatures` - Conversões de temperatura

### Trace Context Propagation
O contexto de trace é propagado automaticamente entre:
- **Gateway → Orchestration:** Via HTTP headers
- **Orchestration → External APIs:** Via instrumented HTTP client
- **Internal Operations:** Via context propagation

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

## 🚀 Quick Start

### 1. **Startup Completo**
```bash
docker-compose up --build
```

### 2. **Verificar Serviços**
```bash
# Gateway Health
curl http://localhost:8080/health

# Orchestration Health  
curl http://localhost:8081/health

# Zipkin UI
# Abrir http://localhost:9411 no navegador
```

### 3. **Acessar Swagger Documentation**
- **Gateway Service**: http://localhost:8080/swagger/index.html
- **Orchestration Service**: http://localhost:8081/swagger/index.html

### 4. **Testar via Swagger UI**
1. Abrir qualquer URL do Swagger
2. Expandir o endpoint desejado
3. Clicar em "Try it out"
4. Inserir os parâmetros
5. Clicar em "Execute"
6. Visualizar a resposta

### 5. **Testar via cURL**
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
- `ZIPKIN_URL`: URL do Zipkin para envio de traces (padrão: http://localhost:9411/api/v2/spans)

### Orchestration (Serviço B)
- `PORT`: Porta do serviço (padrão: 8081)
- `WEATHER_API_KEY`: Chave da API Weather (obrigatória)
- `ZIPKIN_URL`: URL do Zipkin para envio de traces (padrão: http://localhost:9411/api/v2/spans)

### Zipkin
- `STORAGE_TYPE`: Tipo de armazenamento (padrão: mem para desenvolvimento)

## Health Checks

Ambos os serviços expõem endpoints de health check:
- Gateway: http://localhost:8080/health
- Orchestration: http://localhost:8081/health
- Zipkin: http://localhost:9411/health

## Exemplos de Uso com Tracing

### Testando o Sistema Completo
```bash
# 1. Iniciar os serviços
docker-compose up --build

# 2. Fazer uma requisição
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "29902555"}'

# 3. Visualizar o trace no Zipkin
# Acesse: http://localhost:9411
# Clique em "Run Query" para ver os traces
```

### Analisando Performance
1. **Acesse Zipkin UI:** http://localhost:9411
2. **Execute várias requisições** para coletar dados
3. **Visualize métricas:**
   - Tempo total de processamento
   - Tempo de cada span individual
   - Dependências entre serviços
   - Identificação de gargalos

### Debugging com Traces
- **Request Completo:** Veja o fluxo completo Gateway → Orchestration
- **API Calls:** Monitore chamadas para ViaCEP e WeatherAPI
- **Error Tracking:** Identifique onde ocorrem falhas
- **Performance:** Identifique operações lentas

## Arquitetura de Observabilidade

```
[Client] → [Gateway:8080] → [Orchestration:8081] → [External APIs]
              ↓                     ↓                      ↓
         [Zipkin Trace]       [Zipkin Trace]        [Zipkin Trace]
              ↓                     ↓                      ↓
                           [Zipkin Collector:9411]
                                    ↓
                            [Zipkin UI Dashboard]
```

## Dependencies OpenTelemetry

O projeto utiliza as seguintes bibliotecas OpenTelemetry:

- `go.opentelemetry.io/otel` - Core OpenTelemetry
- `go.opentelemetry.io/otel/exporters/zipkin` - Zipkin exporter
- `go.opentelemetry.io/otel/sdk` - OpenTelemetry SDK
- `go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux` - Gorilla Mux instrumentation
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` - HTTP client instrumentation

## 📖 Swagger API Documentation

Both services have complete Swagger documentation available:

### Gateway Service (Port 8080)
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **API Endpoints**:
  - `POST /cep` - Process CEP input with validation
  - `GET /health` - Gateway health check

### Orchestration Service (Port 8081)  
- **Swagger UI**: http://localhost:8081/swagger/index.html
- **API Endpoints**:
  - `GET /weather/{cep}` - Get weather by CEP
  - `GET /health` - Service health check

### Swagger Features
- **Interactive API Testing** - Test endpoints directly from the UI
- **Request/Response Examples** - Complete examples for all endpoints
- **Schema Documentation** - Detailed request and response models
- **Error Response Documentation** - All possible error scenarios

## 🔍 Access URLs

- **Zipkin UI**: http://localhost:9411
- **Gateway Swagger**: http://localhost:8080/swagger/index.html
- **Gateway Health**: http://localhost:8080/health  
- **Orchestration Swagger**: http://localhost:8081/swagger/index.html
- **Orchestration Health**: http://localhost:8081/health
