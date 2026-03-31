# ==========================================
# Etapa 1: Compilación (Builder)
# ==========================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

# Descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY main.go ./

# Compilar
RUN CGO_ENABLED=0 GOOS=linux go build -o rabbit-consumer-bin main.go

# ==========================================
# Etapa 2: Imagen Final (Runtime)
# ==========================================
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/rabbit-consumer-bin .

# Sin EXPOSE porque es un consumidor

CMD ["./rabbit-consumer-bin"]