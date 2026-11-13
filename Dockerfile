# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.21-bookworm AS build-stage
WORKDIR /queue

# Для go-sqlite3 нужен gcc и заголовки libc
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Включаем CGO — без этого go-sqlite3 станет "stub" и БД не заработает
ENV CGO_ENABLED=1
RUN go build -o /queue/queuebot ./cmd/main.go

# ---- Runtime stage ----
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Бинарник кладём в /usr/local/bin, как советует документация Bothost
COPY --from=build-stage /queue/queuebot /usr/local/bin/queuebot

# Папка для БД, чтобы /app/data сохранялся между перезапусками
RUN mkdir -p /app/data && chmod 777 /app/data

CMD ["/usr/local/bin/queuebot"]