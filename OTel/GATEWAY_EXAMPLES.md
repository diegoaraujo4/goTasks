# OTEL - Exemplos de API

## Serviço A - Gateway (otel-gateway) - Porta 8080

### POST /cep - Processar CEP
```bash
# CEP válido
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "29902555"}'

# Resposta esperada (200):
{
  "city": "Linhares",
  "temp_C": 25.5,
  "temp_F": 77.9,
  "temp_K": 298.5
}
```

```bash
# CEP inválido - muito curto
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'

# Resposta esperada (422):
{
  "message": "invalid zipcode"
}
```

```bash
# CEP inválido - com letras
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "abc12345"}'

# Resposta esperada (422):
{
  "message": "invalid zipcode"
}
```

## Serviço B - Orchestration (otel-orchestration) - Porta 8081

### GET /weather/{cep} - Consultar temperatura por CEP
```bash
# CEP válido (8 dígitos)
curl http://localhost:8081/weather/29902555

# Resposta esperada (200):
{
  "city": "Linhares",
  "temp_C": 25.5,
  "temp_F": 77.9,
  "temp_K": 298.5
}
```

```bash
# CEP inválido (formato incorreto)
curl http://localhost:8081/weather/123

# Resposta esperada (422):
{
  "message": "invalid zipcode"
}
```

```bash
# CEP não encontrado (formato correto mas inexistente)
curl http://localhost:8081/weather/99999999

# Resposta esperada (404):
{
  "message": "can not find zipcode"
}
```

## Requisitos Atendidos

### ✅ Serviço A (Gateway):
- **Input:** Recebe POST com `{"cep": "29902555"}`
- **Validação:** 8 dígitos, apenas números, formato string
- **Encaminhamento:** Para Serviço B via HTTP quando válido
- **Erro 422:** "invalid zipcode" quando inválido

### ✅ Serviço B (Orchestration):
- **Input:** CEP válido de 8 dígitos
- **Processamento:** Busca localização + temperatura
- **Resposta 200:** `{"city": "São Paulo", "temp_C": 28.5, "temp_F": 28.5, "temp_K": 28.5}`
- **Erro 422:** "invalid zipcode" (formato incorreto)
- **Erro 404:** "can not find zipcode" (CEP não encontrado)

## Estrutura do Projeto

### 📂 **Localização dos Códigos:**
- **Gateway (Serviço A):** `cmd/gateway/` → Porta 8080
- **Orchestrator (Serviço B):** `cmd/orchestrator/` → Porta 8081

## Testes Completos

```bash
# CEPs válidos para teste
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": "01310100"}' # São Paulo - SP
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": "20040020"}' # Rio de Janeiro - RJ
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": "30112000"}' # Belo Horizonte - MG
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": "29902555"}' # Linhares - ES

# CEPs inválidos para teste
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": ""}'        # Vazio → 422
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": "123"}'     # Muito curto → 422
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": "123456789"}' # Muito longo → 422
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": "abc12345"}' # Com letras → 422

# Teste direto do Serviço B (Orchestrator)
curl http://localhost:8081/weather/29902555  # CEP válido → 200
curl http://localhost:8081/weather/123       # CEP inválido → 422  
curl http://localhost:8081/weather/99999999  # CEP não encontrado → 404
```
