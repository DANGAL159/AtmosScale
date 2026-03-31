# ==========================================
# Etapa 1: Compilación (Builder)
# ==========================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copiamos solo los archivos de dependencias primero (aprovecha la caché de Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el código fuente
COPY main.go ./

# Compilamos el binario de forma estática
RUN CGO_ENABLED=0 GOOS=linux go build -o kafka-consumer-bin main.go

# ==========================================
# Etapa 2: Imagen Final (Runtime)
# ==========================================
FROM alpine:latest

# Certificados de seguridad
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiamos SOLO el binario compilado
COPY --from=builder /app/kafka-consumer-bin .

# IMPORTANTE: Un consumidor no necesita exponer puertos (EXPOSE) 
# porque nadie se conecta a él; él es quien se conecta a Kafka y a Valkey.

# Comando de ejecución
CMD ["./kafka-consumer-bin"]