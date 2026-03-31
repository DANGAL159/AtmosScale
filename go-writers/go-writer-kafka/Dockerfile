# ==========================================
# Etapa 1: Compilación (Builder)
# ==========================================
FROM golang:1.24-alpine AS builder

# Instalamos git por si las dependencias lo requieren
RUN apk add --no-cache git

# Establecemos el directorio de trabajo principal en la raíz
WORKDIR /app

# 1. Copiamos la carpeta proto (vital para el replace)
COPY proto/ ./proto/

# 2. Copiamos específicamente la carpeta de este microservicio
COPY go-writers/go-writer-kafka/ ./go-writers/go-writer-kafka/

# 3. Nos movemos a la carpeta del microservicio para compilar
WORKDIR /app/go-writers/go-writer-kafka

# Descargamos las dependencias
RUN go mod download

# Compilamos el binario de forma estática
RUN CGO_ENABLED=0 GOOS=linux go build -o kafka-writer-bin main.go

# ==========================================
# Etapa 2: Imagen Final (Runtime)
# ==========================================
FROM alpine:latest

# Certificados de seguridad
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiamos SOLO el binario compilado desde la etapa 1
COPY --from=builder /app/go-writers/go-writer-kafka/kafka-writer-bin .

# Exponemos el puerto de gRPC (50051 es el que configuramos en el main.go)
EXPOSE 50051

# Comando de ejecución
CMD ["./kafka-writer-bin"]