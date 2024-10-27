# ===================================================================================
# === Stage 1: Build the Go backend service =========================================
# ===================================================================================
FROM golang:1.23 as builder
WORKDIR /build

ARG VERSION="1.0.0"

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd/ ./cmd
COPY api/ ./api

RUN go build -o server ./cmd/api-server

# ================================================================================================
# === Stage 2: Get backend binary into a lightweight container ===================================
# ================================================================================================
FROM debian:latest as dev
RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app

EXPOSE 8080

ARG ALPHA_VANTAGE_API_KEY
ENV ALPHA_VANTAGE_API_KEY=${ALPHA_VANTAGE_API_KEY}

ARG HOST_IP="0.0.0.0"

COPY --from=builder /build/server . 
CMD sh -c "./server --apiKey=${ALPHA_VANTAGE_API_KEY} --host=${HOST_IP}"
