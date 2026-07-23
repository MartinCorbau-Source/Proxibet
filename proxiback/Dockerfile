# syntax=docker/dockerfile:1

FROM node:24-alpine AS frontend-builder

WORKDIR /src/proxifront

COPY proxifront/package*.json ./

RUN npm ci

COPY proxifront/ ./

RUN npm run build -- --output-path=/out/proxifront \
    && test -f /out/proxifront/browser/index.html

FROM golang:1.26-alpine AS backend-builder

WORKDIR /src/proxiback

COPY proxiback/go.mod proxiback/go.sum ./

RUN go mod download

COPY proxiback/ ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/proxibet-api \
    ./cmd/api

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates \
    && addgroup -S proxibet \
    && adduser -S proxibet -G proxibet

WORKDIR /app

ENV STATIC_DIR=/app/public

COPY --from=backend-builder /out/proxibet-api /app/proxibet-api
COPY --from=frontend-builder /out/proxifront/browser /app/public

USER proxibet

EXPOSE 8080

ENTRYPOINT ["/app/proxibet-api"]