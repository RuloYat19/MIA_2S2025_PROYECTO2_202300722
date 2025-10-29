# Usar la imagen oficial de Go para construir la aplicación
FROM golang:1.22.2-alpine AS builder

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar los archivos de módulos primero (para aprovechar la cache de Docker)
COPY go.mod ./
COPY go.sum ./

# Descargar las dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN go build -o main .

# Etapa final: imagen mínima para ejecutar
FROM alpine:latest

# Instalar dependencias necesarias para runtime
RUN apk --no-cache add ca-certificates

# Crear un usuario no-root para mayor seguridad
RUN adduser -D -g '' appuser

# Establecer el directorio de trabajo
WORKDIR /root/

# Copiar el binario compilado desde la etapa builder
COPY --from=builder /app/main .
COPY --from=builder /app/Comandos ./Comandos
COPY --from=builder /app/Globales ./Globales
COPY --from=builder /app/Structs ./Structs

# Crear la estructura de directorios que necesita la aplicación
RUN mkdir -p /home/ubuntu/MIA_2S2025_PROYECTO2_202300722/Calificacion_MIA/Discos && \
    mkdir -p /home/ubuntu/MIA_2S2025_PROYECTO2_202300722/Calificacion_MIA/Reportes && \
    mkdir -p /home/ubuntu/MIA_2S2025_PROYECTO2_202300722/proyecto

# Cambiar propietario de los archivos al usuario appuser
RUN chown -R appuser:appuser /root/ /home/ubuntu/

# Cambiar al usuario no-root
USER appuser

# Exponer el puerto que usa la aplicación
EXPOSE 8080

# Comando para ejecutar la aplicación directamente
CMD ["./main"]