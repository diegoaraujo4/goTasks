FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY configs/ configs/
COPY .env .env

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/ordersystem/main.go cmd/ordersystem/wire_gen.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary and config from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8000 8080 50051

CMD ["./main"]
