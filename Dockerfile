# ===================================================================================
# === Stage 1: Build the Go backend service =========================================
# ===================================================================================
FROM golang:1.24 AS builder
WORKDIR /build

ARG VERSION="1.0.0"

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd/ ./cmd
COPY api/ ./api
COPY store/ ./store
COPY pkg/ ./pkg

RUN go build -o server ./cmd/api-server

# ================================================================================================
# === Stage 2: Get backend binary into a lightweight container ===================================
# ================================================================================================
FROM debian:latest AS dev
RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app

EXPOSE 8080

ARG CONFIG_PATH
ARG DATA_SOURCE_NAME
ARG AUTH0_CLIENT_SECRET
ARG HMAC_SECRET

ENV CONFIG_PATH=${CONFIG_PATH}
ENV DATA_SOURCE_NAME=${DATA_SOURCE_NAME}
ENV LOG_LEVEL=${LOG_LEVEL}
ENV AUTH0_CLIENT_SECRET=${AUTH0_CLIENT_SECRET}
ENV HMAC_SECRET=${HMAC_SECRET}

COPY --from=builder /build/server . 
COPY config/ ./config
CMD sh -c "./server \
  --configPath=${CONFIG_PATH} \
  --dsn=${DATA_SOURCE_NAME} \
  --log.level=${LOG_LEVEL} \
  --auth0ClientSecret=${AUTH0_CLIENT_SECRET} \
  --hmacSecret=${HMAC_SECRET}"
