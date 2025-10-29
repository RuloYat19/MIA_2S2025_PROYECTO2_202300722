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
RUN apk --no-cache add ca-certificates bash

# Crear un usuario no-root para mayor seguridad
RUN adduser -D -g '' appuser

# Establecer el directorio de trabajo
WORKDIR /root/

# Copiar el binario compilado desde la etapa builder
COPY --from=builder /app/main .
COPY --from=builder /app/Comandos ./Comandos
COPY --from=builder /app/Globales ./Globales
COPY --from=builder /app/Structs ./Structs

# Copiar el script shell y dar permisos de ejecución
COPY b.sh .
RUN chmod +x b.sh

# Crear la estructura de directorios que necesita el script
RUN mkdir -p /home/rauly/USAC/Archivos/MIA_2S2025_P2_202300722/Calificacion_MIA/Discos && \
    mkdir -p /home/rauly/USAC/Archivos/MIA_2S2025_P2_202300722/Calificacion_MIA/Reportes && \
    mkdir -p /home/rauly/USAC/Archivos/MIA_2S2025_P2_202300722/proyecto

# Cambiar propietario de los archivos al usuario appuser
RUN chown -R appuser:appuser /root/ /home/rauly/

# Cambiar al usuario no-root
USER appuser

# Exponer el puerto que usa la aplicación
EXPOSE 8080

# Ejecutar el script de limpieza y luego la aplicación
CMD ["sh", "-c", "./b.sh && ./main"]