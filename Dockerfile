
# ==================== Base Stage ====================
FROM golang:1.25-alpine AS base

WORKDIR /app

# Установка dev-зависимостей
RUN apk add --no-cache \
    git \
    bash \
    curl \
    make \
    sqlite-dev \
    gcc \
    musl-dev
# Копирование файлов зависимостей
COPY go.mod go.sum ./
RUN go mod download
# Копирование исходного кода
COPY . .
# Собираем с включённым CGO и флагами для SQLite
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# ==================== Development Stage ====================
FROM base AS dev

# Установка стабильной версии air, совместимой с Go 1.24
RUN go install github.com/air-verse/air@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest

# Создание точки для дебаггера
EXPOSE 40000 2345

# Запуск air для hot reload
CMD ["air", "-c", ".air.toml"]
