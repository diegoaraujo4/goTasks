# Serviço de Cotação USD/BRL

Este projeto implementa uma arquitetura cliente-servidor para buscar e armazenar cotações de câmbio USD/BRL usando Go, com containerização Docker.

## Arquitetura

- **Servidor** (`server.go`): Servidor HTTP que busca cotações USD/BRL de API externa e armazena em banco SQLite
- **Cliente** (`client.go`): Cliente HTTP que solicita cotação do servidor e salva em arquivo

## Funcionalidades

- Timeouts baseados em contexto para todas as operações:
  - Chamada à API externa: timeout de 200ms
  - Operações de banco de dados: timeout de 10ms  
  - Comunicação cliente-servidor: timeout de 300ms
- Banco de dados SQLite para armazenar histórico de cotações
- Saída em arquivo com cotação atual
- Containerização Docker com persistência de volumes

## Requisitos Atendidos

✅ Servidor consome API https://economia.awesomeapi.com.br/json/last/USD-BRL  
✅ Servidor retorna JSON com campo "bid" para o cliente  
✅ Servidor armazena cada cotação no banco SQLite  
✅ Timeouts de contexto: 200ms (API), 10ms (BD), 300ms (cliente)  
✅ Log de erros para cenários de timeout  
✅ Cliente salva cotação no arquivo "cotacao.txt"  
✅ Servidor roda na porta 8080 com endpoint `/cotacao`  

## Início Rápido

### Pré-requisitos

1. **Docker Desktop**: Certifique-se de que o Docker Desktop está instalado e rodando
   - Download em: https://www.docker.com/products/docker-desktop
   - Inicie o Docker Desktop antes de executar os comandos abaixo

### Usando Docker Compose (Recomendado)

1. **Construir e executar todos os serviços:**
   ```bash
   docker-compose up --build
   ```

2. **Executar em modo detached:**
   ```bash
   docker-compose up --build -d
   ```

3. **Executar apenas o cliente (servidor deve estar rodando):**
   ```bash
   docker-compose up client
   ```

4. **Parar todos os serviços:**
   ```bash
   docker-compose down
   ```

5. **Ver logs:**
   ```bash
   docker-compose logs -f server
   docker-compose logs -f client
   ```

### Build Manual do Docker

1. **Construir imagem do servidor:**
   ```bash
   docker build -f Dockerfile.server -t go-quotation-server .
   ```

2. **Construir imagem do cliente:**
   ```bash
   docker build -f Dockerfile.client -t go-quotation-client .
   ```

3. **Executar servidor:**
   ```bash
   docker run -p 8080:8080 -v quotes_data:/data go-quotation-server
   ```

4. **Executar cliente:**
   ```bash
   docker run -v client_data:/data --link server go-quotation-client
   ```

### Desenvolvimento Local

1. **Instalar dependências:**
   ```bash
   go mod tidy
   ```

2. **Executar servidor:**
   ```bash
   cd cmd/server
   go run server.go
   ```

3. **Executar cliente (em outro terminal):**
   ```bash
   cd cmd/client
   go run client.go
   ```

## Uso da API

### Obter Cotação Atual
```bash
curl http://localhost:8080/cotacao
```

Resposta:
```json
{
  "bid": "5.1234"
}
```

## Persistência de Dados

- **Banco SQLite**: Armazenado em `/data/quotes.db` (Docker) ou `./quotes.db` (local)
- **Saída do Cliente**: Salvo em `/data/cotacao.txt` (Docker) ou `./cotacao.txt` (local)

## Monitoramento

O servidor inclui um endpoint de health check acessível em `/cotacao`. A configuração Docker inclui monitoramento de saúde que reiniciará o serviço se ficar sem resposta.

## Logs

Todos os cenários de timeout e erros são logados no stdout. Veja os logs com:
```bash
docker-compose logs -f server
docker-compose logs -f client
```

## Variáveis de Ambiente

A aplicação usa valores padrão sensatos mas pode ser customizada:

| Variável | Padrão | Descrição |
|----------|---------|-------------|
| PORT | 8080 | Porta de escuta do servidor |
| DB_PATH | /data/quotes.db | Caminho do arquivo do banco SQLite |
| OUTPUT_PATH | /data/cotacao.txt | Caminho do arquivo de saída do cliente |

## Solução de Problemas

### Problemas Comuns

1. **Erros de timeout**: Verifique conectividade de rede e disponibilidade da API externa
2. **Erros de banco de dados**: Certifique-se das permissões de escrita no volume de dados
3. **Erros de conexão do cliente**: Verifique se o servidor está rodando e acessível

### Visualizando Dados

Verificar banco SQLite:
```bash
docker-compose exec server sqlite3 /data/quotes.db "SELECT * FROM quotes ORDER BY timestamp DESC LIMIT 10;"
```

Verificar saída do cliente:
```bash
docker-compose exec client cat /data/cotacao.txt
```
