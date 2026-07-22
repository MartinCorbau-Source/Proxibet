# syntax=docker/dockerfile:1

# -------------------------------------------------------
# Étape 1 : compilation
# -------------------------------------------------------
FROM golang:1.26-alpine AS builder

WORKDIR /app




# On copie d'abord uniquement les fichiers de dépendances
# pour profiter du cache Docker.
COPY go.mod go.sum ./

RUN go mod download

# Puis on copie le reste du projet.
COPY . .

# Construction d'un binaire Linux statique.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/bin/proxibet-api \
    ./cmd/api

# -------------------------------------------------------
# Étape 2 : image d'exécution
# -------------------------------------------------------
FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates \
    && addgroup -S proxibet \
    && adduser -S proxibet -G proxibet

WORKDIR /app

COPY --from=builder /app/bin/proxibet-api /app/proxibet-api

USER proxibet

EXPOSE 8080

ENTRYPOINT ["/app/proxibet-api"]