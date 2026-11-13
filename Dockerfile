# syntax=docker/dockerfile:1

FROM golang:1.27-bookworm AS build-stage
WORKDIR /queue
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /queue/queuebot ./cmd/main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build-stage /queue/queuebot /usr/local/bin/queuebot
RUN mkdir -p /app/data && chmod 777 /app/data
CMD ["/usr/local/bin/queuebot"]
