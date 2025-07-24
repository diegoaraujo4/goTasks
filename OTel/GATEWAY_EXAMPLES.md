# OTEL - Exemplos de API

## Serviço A - Gateway (otel-gateway) - Porta 8081

### POST /cep - Processar CEP
```bash
# CEP válido
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "29902555"}'

# Resposta esperada (200):
{
  "location": "Linhares - ES",
  "temperature": {
    "celsius": 25.5,
    "fahrenheit": 77.9,
    "kelvin": 298.65
  }
}
```

```bash
# CEP inválido - muito curto
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'

# Resposta esperada (422):
{
  "message": "invalid zipcode"
}
```

```bash
# CEP inválido - com letras
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "abc12345"}'

# Resposta esperada (422):
{
  "message": "invalid zipcode"
}
```

```bash
# JSON inválido
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d 'invalid json'

# Resposta esperada (400):
{
  "message": "invalid request body"
}
```

### GET /health - Health Check do Gateway
```bash
curl http://localhost:8081/health

# Resposta esperada (200):
{
  "status": "healthy",
  "service": "otel-gateway"
}
```

## Serviço B - Orchestration (otel-orchestration) - Porta 8080

### GET /weather/{cep} - Consultar temperatura por CEP
```bash
# CEP válido
curl http://localhost:8080/weather/29902-555

# Resposta esperada (200):
{
  "location": "Linhares - ES",
  "temperature": {
    "celsius": 25.5,
    "fahrenheit": 77.9,
    "kelvin": 298.65
  }
}
```

```bash
# CEP inválido
curl http://localhost:8080/weather/123

# Resposta esperada (400):
{
  "error": "CEP inválido"
}
```

### GET /health - Health Check do Orchestration
```bash
curl http://localhost:8080/health

# Resposta esperada (200):
{
  "status": "OK"
}
```

## Fluxo de Integração

### Fluxo Normal (CEP Válido)
1. **Cliente** → POST /cep {"cep": "29902555"} → **Gateway (8081)**
2. **Gateway** → Valida CEP (8 dígitos, apenas números)
3. **Gateway** → GET /weather/29902-555 → **Orchestration (8080)**
4. **Orchestration** → Consulta ViaCEP para localização
5. **Orchestration** → Consulta WeatherAPI para temperatura
6. **Orchestration** → Converte temperaturas (C, F, K)
7. **Orchestration** → Resposta → **Gateway**
8. **Gateway** → Resposta → **Cliente**

### Fluxo de Erro (CEP Inválido)
1. **Cliente** → POST /cep {"cep": "123"} → **Gateway (8081)**
2. **Gateway** → Valida CEP (falha na validação)
3. **Gateway** → Resposta 422 "invalid zipcode" → **Cliente**

## Testes com diferentes CEPs

```bash
# CEPs válidos para teste
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "01310100"}' # São Paulo - SP
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "20040020"}' # Rio de Janeiro - RJ
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "30112000"}' # Belo Horizonte - MG
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "80010000"}' # Curitiba - PR
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "29902555"}' # Linhares - ES

# CEPs inválidos para teste
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": ""}'        # Vazio
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "123"}'     # Muito curto
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "123456789"}' # Muito longo
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "abc12345"}' # Com letras
curl -X POST http://localhost:8081/cep -H "Content-Type: application/json" -d '{"cep": "12345-67"}' # Com hífen
```
