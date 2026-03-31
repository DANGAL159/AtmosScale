# ==========================================
# Etapa 1: Compilación (Builder)
# ==========================================
FROM rust:slim AS builder

# reqwest usa OpenSSL por defecto, así que necesitamos estas librerías para compilar
RUN apt-get update && apt-get install -y pkg-config libssl-dev

# Establecemos el directorio de trabajo
WORKDIR /usr/src/api-rust

# Copiamos los archivos de dependencias y el código fuente
COPY Cargo.toml ./
COPY src ./src

# Compilamos en modo "release" (optimizado al máximo para producción)
RUN cargo build --release

# ==========================================
# Etapa 2: Imagen Final (Runtime)
# ==========================================
FROM debian:bookworm-slim

# Instalamos ca-certificates para peticiones web y libssl para que el binario funcione
RUN apt-get update && apt-get install -y ca-certificates libssl3 && rm -rf /var/lib/apt/lists/*

# (Opcional pero recomendado) Por seguridad en Kubernetes, creamos un usuario que no sea root
RUN useradd -ms /bin/bash appuser
USER appuser

WORKDIR /app

# Copiamos SOLO el binario compilado desde la etapa 1
COPY --from=builder /usr/src/api-rust/target/release/api-rust .

# Exponemos el puerto que configuramos en Actix-web
EXPOSE 8000

# Ejecutamos el binario
CMD ["./api-rust"]