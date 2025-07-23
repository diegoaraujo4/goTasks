# Use a imagem oficial do Go como base
FROM golang:1.24.5-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git ca-certificates

# Definir diretório de trabalho
WORKDIR /app

# Copiar go.mod e go.sum
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar o código fonte
COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Usar uma imagem mínima para produção
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates

# Criar diretório de trabalho
WORKDIR /root/

# Copiar o binário da etapa de build
COPY --from=builder /app/main .

# Expor a porta
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]
