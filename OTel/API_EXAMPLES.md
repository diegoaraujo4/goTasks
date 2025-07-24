# Exemplos de Uso da Weather API

Este arquivo contém exemplos práticos de como usar a Weather API.

## 1. Acessando a Documentação Swagger

### Swagger UI (Interface Interativa)
```
http://localhost:8080/swagger/index.html
```

### JSON da Documentação
```
http://localhost:8080/swagger/doc.json
```

## 2. Endpoints da API

### Health Check
```bash
curl http://localhost:8080/health
```

**Resposta:**
```
OK
```

### Consulta de Temperatura por CEP

#### CEP Válido - São Paulo (01310-100)
```bash
curl http://localhost:8080/weather/01310100
```

**Resposta de Sucesso (200):**
```json
{
  "temp_C": 23.5,
  "temp_F": 74.3,
  "temp_K": 296.5
}
```

#### CEP com Hífen
```bash
curl http://localhost:8080/weather/01310-100
```

#### CEP Válido - Rio de Janeiro (20040-020)
```bash
curl http://localhost:8080/weather/20040020
```

#### CEP Válido - Belo Horizonte (30112-000)
```bash
curl http://localhost:8080/weather/30112000
```

## 3. Cenários de Erro

### CEP com Formato Inválido (422)
```bash
curl -v http://localhost:8080/weather/123
```

**Resposta:**
```json
{
  "message": "invalid zipcode"
}
```

### CEP Não Encontrado (404)
```bash
curl -v http://localhost:8080/weather/99999999
```

**Resposta:**
```json
{
  "message": "can not find zipcode"
}
```

## 4. Testando com PowerShell

### Health Check
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
```

### Consulta de Temperatura
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8080/weather/01310100" -Method Get
$response | ConvertTo-Json -Depth 3
```

### Tratamento de Erros
```powershell
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/weather/123" -Method Get
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Status Code: $($_.Exception.Response.StatusCode)"
    $errorStream = $_.Exception.Response.GetResponseStream()
    $reader = New-Object System.IO.StreamReader($errorStream)
    $errorBody = $reader.ReadToEnd()
    Write-Host "Response Body: $errorBody"
}
```

## 5. Testando com JavaScript/Fetch

### HTML + JavaScript Simples
```html
<!DOCTYPE html>
<html>
<head>
    <title>Weather API Test</title>
</head>
<body>
    <h1>Weather API Test</h1>
    <input type="text" id="cep" placeholder="Digite o CEP (ex: 01310100)" />
    <button onclick="getWeather()">Consultar Clima</button>
    <div id="result"></div>

    <script>
        async function getWeather() {
            const cep = document.getElementById('cep').value;
            const resultDiv = document.getElementById('result');
            
            try {
                const response = await fetch(`http://localhost:8080/weather/${cep}`);
                const data = await response.json();
                
                if (response.ok) {
                    resultDiv.innerHTML = `
                        <h3>Temperatura:</h3>
                        <p>Celsius: ${data.temp_C}°C</p>
                        <p>Fahrenheit: ${data.temp_F}°F</p>
                        <p>Kelvin: ${data.temp_K}K</p>
                    `;
                } else {
                    resultDiv.innerHTML = `<p>Erro: ${data.message}</p>`;
                }
            } catch (error) {
                resultDiv.innerHTML = `<p>Erro de conexão: ${error.message}</p>`;
            }
        }
    </script>
</body>
</html>
```

## 6. Testando com Python

```python
import requests
import json

# Configuração
BASE_URL = "http://localhost:8080"

def test_health():
    """Testa o health check"""
    response = requests.get(f"{BASE_URL}/health")
    print(f"Health Check: {response.status_code} - {response.text}")

def test_weather(cep):
    """Testa consulta de temperatura por CEP"""
    response = requests.get(f"{BASE_URL}/weather/{cep}")
    
    if response.status_code == 200:
        data = response.json()
        print(f"CEP {cep}:")
        print(f"  Celsius: {data['temp_C']}°C")
        print(f"  Fahrenheit: {data['temp_F']}°F")
        print(f"  Kelvin: {data['temp_K']}K")
    else:
        error = response.json()
        print(f"Erro {response.status_code}: {error['message']}")

# Exemplos de uso
if __name__ == "__main__":
    test_health()
    print()
    
    # Testes com CEPs válidos
    test_weather("01310100")  # São Paulo
    test_weather("20040020")  # Rio de Janeiro
    test_weather("30112000")  # Belo Horizonte
    
    print()
    
    # Testes com erros
    test_weather("123")       # CEP inválido
    test_weather("99999999")  # CEP não encontrado
```

## 7. Fórmulas de Conversão

A API usa as seguintes fórmulas para conversão de temperatura:

### Celsius para Fahrenheit
```
F = C × 1.8 + 32
```

### Celsius para Kelvin
```
K = C + 273
```

## 8. Status Codes HTTP

| Código | Descrição | Exemplo |
|--------|-----------|---------|
| 200 | Sucesso | CEP encontrado e temperatura retornada |
| 404 | CEP não encontrado | CEP válido mas não existe na base |
| 422 | CEP inválido | Formato incorreto (não tem 8 dígitos) |
| 500 | Erro interno | Erro na API de clima ou conexão |

## 9. Headers HTTP Recomendados

```bash
curl -H "Content-Type: application/json" \
     -H "Accept: application/json" \
     http://localhost:8080/weather/01310100
```

## 10. Monitoramento

### Logs da Aplicação
A aplicação gera logs úteis para monitoramento:
- Inicialização do servidor
- Requisições recebidas
- Erros de API externa
- Status de health checks
