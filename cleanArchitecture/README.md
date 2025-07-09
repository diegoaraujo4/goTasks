# Clean Architecture - Order System

Sistema de pedidos implementado com Clean Architecture em Go, oferecendo APIs REST, gRPC e GraphQL.

## Funcionalidades

- Criar pedidos
- Listar pedidos
- APIs disponíveis: REST, gRPC e GraphQL
- Persistência em MySQL
- Mensageria com RabbitMQ
- Containerização com Docker

## Portas dos Serviços

- **REST API**: 8000
- **gRPC**: 50051
- **GraphQL**: 8080
- **MySQL**: 3306
- **RabbitMQ**: 5672 (Management: 15672)

## Executando o Projeto

### Com Docker (Recomendado)

1. Clone o repositório
2. Execute o comando:

```bash
docker-compose up --build
```

Isso irá:
- Inicializar o banco MySQL com as tabelas necessárias
- Configurar o RabbitMQ
- Compilar e executar a aplicação

### Executando Localmente

1. Certifique-se de ter Go 1.19+ instalado
2. Configure o banco MySQL e RabbitMQ
3. Copie o arquivo `.env` e ajuste as configurações se necessário
4. Execute:

```bash
go mod tidy
go run cmd/ordersystem/main.go
```

## Testando as APIs

### REST API

**Criar Pedido:**
```bash
POST http://localhost:8000/order
Content-Type: application/json

{
    "id": "order-1",
    "price": 100.5,
    "tax": 0.5
}
```

**Listar Pedidos:**
```bash
GET http://localhost:8000/order
```

### GraphQL

Acesse o playground em: http://localhost:8080

**Criar Pedido:**
```graphql
mutation {
    createOrder(input: {
        id: "order-gql-1",
        Price: 100.0,
        Tax: 10.0
    }) {
        id
        Price
        Tax
        FinalPrice
    }
}
```

**Listar Pedidos:**
```graphql
query {
    orders {
        id
        Price
        Tax
        FinalPrice
    }
}
```

### gRPC

Use um cliente gRPC como Evans ou Postman para testar:

**Criar Pedido:**
```proto
rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse)
```

**Listar Pedidos:**
```proto
rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse)
```

## Estrutura do Projeto

```
cleanArchitecture/
├── cmd/ordersystem/          # Aplicação principal
├── internal/
│   ├── entity/              # Entidades de domínio
│   ├── usecase/             # Casos de uso
│   ├── infra/
│   │   ├── database/        # Repositórios
│   │   ├── web/             # Handlers REST
│   │   ├── grpc/            # Serviços gRPC
│   │   └── graph/           # Resolvers GraphQL
│   └── event/               # Eventos de domínio
├── sql/migrations/          # Scripts de migração
├── api/                     # Arquivos de teste da API
├── docker-compose.yaml      # Configuração Docker
└── .env                     # Variáveis de ambiente
```

## Banco de Dados

A tabela `orders` é criada automaticamente via migração no Docker:

```sql
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(255) PRIMARY KEY,
    price DECIMAL(10,2) NOT NULL,
    tax DECIMAL(10,2) NOT NULL,
    final_price DECIMAL(10,2) NOT NULL
);
```

## Arquivos de Teste

Utilize o arquivo `api/create_order.http` para testar as APIs usando extensões como REST Client no VS Code.
