# ==========================================
# Etapa 1: Compilación (Builder)
# ==========================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copiamos proto para el replace
COPY proto/ ./proto/

# Copiamos la carpeta de este microservicio
COPY go-writers/go-writer-rabbitmq/ ./go-writers/go-writer-rabbitmq/

# Nos movemos a la carpeta correcta
WORKDIR /app/go-writers/go-writer-rabbitmq

RUN go mod download

# Compilamos
RUN CGO_ENABLED=0 GOOS=linux go build -o rabbit-writer-bin main.go

# ==========================================
# Etapa 2: Imagen Final (Runtime)
# ==========================================
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copiamos el binario
COPY --from=builder /app/go-writers/go-writer-rabbitmq/rabbit-writer-bin .

# Exponemos el puerto de RabbitMQ Writer
EXPOSE 50052

CMD ["./rabbit-writer-bin"]