# ===================================================================================
# === Stage 1: Build the Go backend service =========================================
# ===================================================================================
FROM golang:1.23 AS builder
WORKDIR /build

ARG VERSION="1.0.0"

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd/ ./cmd
COPY api/ ./api
COPY store/ ./store

RUN go build -o server ./cmd/api-server

# ================================================================================================
# === Stage 2: Get backend binary into a lightweight container ===================================
# ================================================================================================
FROM debian:latest AS dev
RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app

EXPOSE 8080

ARG ALPHA_VANTAGE_API_KEY
ARG DATA_SOURCE_NAME
ARG ALGORITHM_FILE_NAME

ENV ALPHA_VANTAGE_API_KEY=${ALPHA_VANTAGE_API_KEY}
ENV DATA_SOURCE_NAME=${DATA_SOURCE_NAME}
ENV LOG_LEVEL=${LOG_LEVEL}
ENV ALGORITHM_FILE_NAME=${ALGORITHM_FILE_NAME}

ARG HOST_IP="0.0.0.0"

COPY --from=builder /build/server . 
COPY config/ ./config
CMD sh -c "./server --apiKey=${ALPHA_VANTAGE_API_KEY} --host=${HOST_IP} --dsn=${DATA_SOURCE_NAME} --log.level=${LOG_LEVEL} --algorithmPath=${ALGORITHM_FILE_NAME}"
