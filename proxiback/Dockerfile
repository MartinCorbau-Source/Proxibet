# syntax=docker/dockerfile:1

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

COPY --from=backend-builder /out/proxibet-api /app/proxibet-api

USER proxibet

EXPOSE 8080

ENTRYPOINT ["/app/proxibet-api"]