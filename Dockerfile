# ==========================================
# ETAPA 1: Construcción (Builder)
# ==========================================
FROM golang:1.23-alpine AS builder

# Establecemos el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiamos los archivos de dependencias primero (Optimización de caché de Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código fuente
COPY . .

# MAGIA DEVSECOPS: Compilamos el binario
# CGO_ENABLED=0 -> Le dice a Go que NO dependa de librerías en C del sistema operativo.
# GOOS=linux -> Asegura que el binario es para Linux.
# -ldflags="-w -s" -> Elimina información de debug para que el binario pese menos y sea más difícil de aplicar ingeniería inversa.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o audit-api ./cmd/api/main.go

# ==========================================
# ETAPA 2: Imagen Final Ultraligera y Segura
# ==========================================
# Usamos 'scratch': una imagen vacía sin shell, sin bash, sin vulnerabilidades de SO.
FROM scratch

# Copiamos ÚNICAMENTE el binario compilado desde la etapa anterior
COPY --from=builder /app/audit-api /audit-api

# Exponemos el puerto de nuestro servicio
EXPOSE 8081

# Ejecutamos el binario directamente
ENTRYPOINT ["/audit-api"]