# Etapa de compilación
FROM golang:1.23-alpine AS builder

# Establecer el directorio de trabajo en el contenedor
WORKDIR /app

# Copiar archivos go.mod y go.sum primero (para aprovechar la caché de capas de Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente completo
COPY . .

# Compilar la aplicación para producción
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Etapa de ejecución
FROM alpine:latest

# Instalar dependencias básicas
RUN apk --no-cache add ca-certificates tzdata

# Establecer zona horaria por defecto
ENV TZ=America/Guatemala

# Crear directorios para discos con permisos adecuados
RUN mkdir -p /discos
RUN mkdir -p /app/jorgis/Calificacion_MIA/Discos

# Crear directorio entrada explícitamente y luego el archivo con contenido "Jorge"
RUN mkdir -p /discos
RUN echo "Jorge" > /discos/NAME.txt

# Crear un usuario no-root para ejecutar la aplicación
RUN adduser -D -g '' appuser

WORKDIR /app

# Copiar el ejecutable compilado desde la etapa de compilación
COPY --from=builder /app/main .
# Copiar archivos necesarios para el funcionamiento
COPY --from=builder /app/disk_registry.json ./

# Dar permisos para crear archivos en los directorios de discos
RUN chown -R appuser:appuser /app/jorgis
RUN chown -R appuser:appuser /discos

# Usar el usuario no-root para ejecutar la aplicación
USER appuser

# Exponer el puerto que usa la aplicación
EXPOSE 8080

# Volumen para persistir los discos
VOLUME ["/discos", "/app/jorgis"]

# Comando para ejecutar la aplicación
CMD ["./main"]