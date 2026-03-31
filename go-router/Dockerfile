# ==========================================
# Etapa 1: Compilación (Builder)
# ==========================================
FROM golang:1.24-alpine AS builder

# Instalamos git por si alguna dependencia lo requiere
RUN apk add --no-cache git

# Establecemos un directorio de trabajo principal
WORKDIR /app

# ¡IMPORTANTE! Copiamos la carpeta proto para que el 'replace' de go.mod funcione
COPY proto/ ./proto/

# Copiamos el código del enrutador
COPY go-router/ ./go-router/

# Nos movemos a la carpeta del enrutador para compilar
WORKDIR /app/go-router

# Descargamos las dependencias
RUN go mod download

# Compilamos el binario de forma estática (sin dependencias externas de C)
RUN CGO_ENABLED=0 GOOS=linux go build -o go-router-bin main.go

# ==========================================
# Etapa 2: Imagen Final (Runtime)
# ==========================================
FROM alpine:latest

# Certificados de seguridad para peticiones web
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiamos SOLO el binario compilado
COPY --from=builder /app/go-router/go-router-bin .

# Exponemos el puerto de Fiber
EXPOSE 8080

# Comando de ejecución
CMD ["./go-router-bin"]