# Weather API - Sistema de Consulta de Clima por CEP

Sistema desenvolvido em Go que recebe um CEP brasileiro, identifica a cidade correspondente e retorna as informações de temperatura atual em Celsius, Fahrenheit e Kelvin.

## 🌐 Aplicação em Produção

A aplicação está disponível em produção no Google Cloud Run:
**https://cloud-run-k76zncoypa-rj.a.run.app**

### Exemplos de uso:
- **CEP válido**: https://cloud-run-k76zncoypa-rj.a.run.app/weather/68504-000
- **Health check**: https://cloud-run-k76zncoypa-rj.a.run.app/health
- **Documentação Swagger**: https://cloud-run-k76zncoypa-rj.a.run.app/swagger/index.html

## Funcionalidades

- ✅ Validação de CEP (8 dígitos)
- ✅ Consulta de localização via API ViaCEP
- ✅ Consulta de clima via WeatherAPI
- ✅ Conversão automática de temperaturas
- ✅ Tratamento de erros adequado
- ✅ Testes automatizados
- ✅ Containerização com Docker
- ✅ Documentação Swagger/OpenAPI
- ✅ Interface Swagger UI
- ✅ Deploy no Google Cloud Run

## Documentação da API

A API possui documentação completa no formato OpenAPI/Swagger. Após iniciar a aplicação, você pode acessar:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Swagger JSON**: http://localhost:8080/swagger/doc.json
- **Swagger YAML**: Arquivo `docs/swagger.yaml`

## API Endpoints

### GET /weather/{cep}

Retorna informações de temperatura para o CEP informado.

**Parâmetros:**
- `cep`: CEP brasileiro com 8 dígitos (com ou sem hífen)

**Respostas:**

**200 OK - Sucesso:**
```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

**422 Unprocessable Entity - CEP inválido:**
```json
{
  "message": "invalid zipcode"
}
```

**404 Not Found - CEP não encontrado:**
```json
{
  "message": "can not find zipcode"
}
```

### GET /health

Endpoint de health check.

**200 OK:**
```
OK
```

## ⚡ Quick Start

```bash
# 1. Navegar para o diretório
cd cloudRun

# 2. Configure a variável de ambiente
export WEATHER_API_KEY=your_api_key_here

# 3. Executar a aplicação
go run ./cmd/api

# 4. Testar a API
curl http://localhost:8080/weather/68504-000

# 5. Acessar documentação Swagger
# http://localhost:8080/swagger/index.html
```

## Configuração

### Variáveis de Ambiente

- `WEATHER_API_KEY`: Chave da API do WeatherAPI (obrigatória)
- `PORT`: Porta do servidor (padrão: 8080)

### Obter Chave da WeatherAPI

1. Acesse [https://www.weatherapi.com/](https://www.weatherapi.com/)
2. Crie uma conta gratuita
3. Obtenha sua API key
4. Configure a variável de ambiente `WEATHER_API_KEY`

## Executando Localmente

### Pré-requisitos
- Go 1.24.5+
- Docker (opcional)

### Com Go
```bash
# 1. Clone o repositório
git clone <repository-url>
cd cloudRun

# 2. Instale as dependências
go mod download

# 3. Configure a variável de ambiente
export WEATHER_API_KEY=your_api_key_here

# 4. Execute a aplicação
go run ./cmd/api
```

### Com Docker Compose
```bash
# 1. Configure o arquivo .env
cp .env.example .env
# Edite o arquivo .env com sua API key

# 2. Execute com Docker Compose
cd deployments && docker-compose up --build
```

### Com Make
```bash
# Build da imagem Docker
make docker-build

# Executar com Docker
make docker-run

# Executar testes em container Docker
make docker-test

# Parar containers
make docker-stop
```

## Documentação Swagger

### Gerando a Documentação

A documentação é gerada automaticamente a partir das anotações no código usando a ferramenta `swag`:

```bash
# Instalar ferramenta swag (primeira vez)
go install github.com/swaggo/swag/cmd/swag@latest

# Gerar documentação
swag init -g cmd/api/main.go

# Executar aplicação
go run ./cmd/api
```

### Acessando a Documentação

Após iniciar a aplicação:

1. **Swagger UI**: http://localhost:8080/swagger/index.html
2. **JSON**: http://localhost:8080/swagger/doc.json  
3. **Arquivos locais**: `docs/swagger.json` e `docs/swagger.yaml`

### Estrutura da Documentação

A documentação inclui:
- ✅ Descrição completa da API
- ✅ Exemplos de request/response
- ✅ Códigos de status HTTP
- ✅ Modelos de dados
- ✅ Interface interativa para testes

## Testes

O projeto inclui testes automatizados abrangentes:

```bash
# Executar todos os testes
go test -v ./...

# Executar testes com coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Executar testes em container Docker
make docker-test
```

### Casos de Teste Implementados

- ✅ Validação de formato de CEP
- ✅ Conversão de temperaturas (Celsius → Fahrenheit/Kelvin)
- ✅ Teste de endpoint com sucesso
- ✅ Teste de CEP inválido (422)
- ✅ Teste de CEP não encontrado (404)
- ✅ Teste de erro na API de clima (500)

## Deploy no Google Cloud Run

### Pré-requisitos
- Google Cloud SDK instalado e configurado
- Projeto do Google Cloud criado
- API do Cloud Run habilitada

### Deploy Manual
```bash
# 1. Configure sua chave da WeatherAPI
export WEATHER_API_KEY=your_api_key_here

# 2. Deploy usando gcloud
gcloud run deploy weather-api \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=$WEATHER_API_KEY
```

### Deploy com Docker
```bash
# Build e teste local
make docker-build
make docker-run

# Para produção, use o Cloud Build ou faça push da imagem para Container Registry
```

## 🔧 Resolução de Problemas

### Erro "error fetching weather data"

Se você encontrar este erro, verifique:

1. **API Key válida**: Certifique-se de que `WEATHER_API_KEY` está configurada corretamente
2. **Conectividade HTTPS**: A aplicação usa HTTPS para conectar com `api.weatherapi.com`
3. **CEP válido**: Verifique se o CEP tem 8 dígitos e existe no Brasil

### Problemas de Rede

A aplicação faz chamadas para:
- `https://viacep.com.br/ws/{cep}/json/` - Para buscar informações do CEP
- `https://api.weatherapi.com/v1/current.json` - Para buscar dados meteorológicos

Certifique-se de que essas URLs estão acessíveis do seu ambiente.

## Estrutura do Projeto

```
cloudRun/
├── cmd/
│   └── api/
│       └── main.go          # Ponto de entrada da aplicação
├── internal/
│   ├── domain/
│   │   ├── weather.go       # Modelos de domínio
│   │   └── interfaces.go    # Interfaces de domínio
│   ├── handler/
│   │   ├── weather.go       # Handlers HTTP para weather
│   │   └── health.go        # Handler de health check
│   ├── service/
│   │   ├── weather.go       # Lógica de negócio
│   │   └── errors.go        # Erros de serviço
│   └── repository/
│       ├── viacep.go        # Integração com ViaCEP API
│       └── weather.go       # Integração com Weather API
├── pkg/
│   ├── temperature/
│   │   ├── converter.go     # Conversão de temperaturas
│   │   └── converter_test.go # Testes de conversão
│   └── validator/
│       ├── cep.go           # Validação de CEP
│       └── cep_test.go      # Testes de validação
├── config/
│   ├── config.go            # Configurações da aplicação
│   └── errors.go            # Erros de configuração
├── deployments/
│   ├── Dockerfile           # Container para produção
│   ├── Dockerfile.test      # Container para testes
│   ├── docker-compose.yml   # Orquestração local
│   └── service.yaml         # Configuração do Cloud Run
├── scripts/
│   ├── deploy.sh            # Script de deploy (Linux/Mac)
│   ├── deploy.ps1           # Script de deploy (Windows)
│   ├── test_api.sh          # Script de teste (Linux/Mac)
│   └── test_api.ps1         # Script de teste (Windows)
├── docs/                    # Documentação Swagger gerada
├── go.mod                   # Dependências do Go
├── go.sum                   # Checksums das dependências
├── Makefile                 # Comandos Docker simplificados
├── .env.example             # Exemplo de variáveis de ambiente
├── .env                     # Variáveis de ambiente (local)
├── docker-compose.yml       # Orquestração Docker local
├── Dockerfile               # Container para produção
└── README.md                # Esta documentação
```

## Tecnologias Utilizadas

- **Go 1.24.5**: Linguagem de programação
- **Gorilla Mux**: Router HTTP
- **ViaCEP API**: Consulta de informações por CEP (https://viacep.com.br)
- **WeatherAPI**: Consulta de informações meteorológicas (https://weatherapi.com)
- **Docker**: Containerização
- **Google Cloud Run**: Plataforma de deploy

## 📋 Changelog

### v1.1 - Correções de Conectividade
- ✅ **Fix**: Atualizado Weather API para usar HTTPS
- ✅ **Fix**: Adicionado URL encoding para cidades com caracteres especiais
- ✅ **Improvement**: Melhorada a tratativa de erros com logs mais detalhados
- ✅ **Update**: Removidas credenciais de API do README por segurança

### v1.0 - Versão Inicial
- ✅ Implementação completa da API de clima por CEP
- ✅ Integração com ViaCEP e WeatherAPI
- ✅ Documentação Swagger
- ✅ Deploy no Google Cloud Run

## Exemplos de Uso

### Exemplos de uso:
- **CEP válido**: https://cloud-run-k76zncoypa-rj.a.run.app/weather/01310100
- **Health check**: https://cloud-run-k76zncoypa-rj.a.run.app/health
- **Documentação Swagger**: https://cloud-run-k76zncoypa-rj.a.run.app/swagger/index.html

### CEP Válido (Local)
```bash
curl http://localhost:8080/weather/01310-100
```

### CEP Inválido (Local)
```bash
curl http://localhost:8080/weather/123
```

### CEP Não Encontrado (Local)
```bash
curl http://localhost:8080/weather/99999999
```

## Monitoramento

A aplicação inclui um endpoint de health check em `/health` que pode ser usado para monitoramento e load balancers.

## Contribuição

1. Fork o projeto
2. Crie sua feature branch (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para detalhes.
