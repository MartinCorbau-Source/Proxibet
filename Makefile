APP_NAME := proxibet-api
MAIN_PACKAGE := ./cmd/api
BINARY_DIRECTORY := ./bin
BINARY_PATH := $(BINARY_DIRECTORY)/$(APP_NAME)

GO := go
DOCKER_COMPOSE := docker compose

LOCAL_DATABASE_URL := postgres://proxibet:proxibet@localhost:5432/proxibet?sslmode=disable

.PHONY: help
help: ## Affiche les commandes disponibles
	@echo "Commandes disponibles :"
	@awk 'BEGIN {FS = ":.*## "; printf "\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

.PHONY: install
install: ## Télécharge les dépendances Go
	$(GO) mod download

.PHONY: tidy
tidy: ## Nettoie et synchronise go.mod et go.sum
	$(GO) mod tidy

.PHONY: format
format: ## Formate le code Go
	$(GO) fmt ./...

.PHONY: vet
vet: ## Lance l'analyse statique standard
	$(GO) vet ./...

.PHONY: test
test: ## Lance les tests
	$(GO) test ./...

.PHONY: test-race
test-race: ## Lance les tests avec détection des data races
	$(GO) test -race ./...

.PHONY: test-cover
test-cover: ## Génère un rapport de couverture
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html: test-cover ## Ouvre le rapport HTML de couverture
	$(GO) tool cover -html=coverage.out

.PHONY: check
check: format vet test ## Lance les vérifications principales

.PHONY: build
build: ## Compile l'API
	@mkdir -p $(BINARY_DIRECTORY)
	CGO_ENABLED=0 $(GO) build \
		-trimpath \
		-ldflags="-s -w" \
		-o $(BINARY_PATH) \
		$(MAIN_PACKAGE)

.PHONY: run
run: ## Lance l'API localement avec Go
	APP_ENV=development \
	HTTP_PORT=8080 \
	DATABASE_URL="$(LOCAL_DATABASE_URL)" \
	$(GO) run $(MAIN_PACKAGE)

.PHONY: clean
clean: ## Supprime les fichiers générés
	rm -rf $(BINARY_DIRECTORY)
	rm -f coverage.out

.PHONY: docker-build
docker-build: ## Construit l'image Docker de l'API
	$(DOCKER_COMPOSE) build api

.PHONY: docker-up
docker-up: ## Démarre l'API et PostgreSQL
	$(DOCKER_COMPOSE) up --build -d

.PHONY: docker-down
docker-down: ## Arrête les conteneurs
	$(DOCKER_COMPOSE) down

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Redémarre les conteneurs

.PHONY: docker-logs
docker-logs: ## Affiche les logs de tous les services
	$(DOCKER_COMPOSE) logs -f

.PHONY: api-logs
api-logs: ## Affiche les logs de l'API
	$(DOCKER_COMPOSE) logs -f api

.PHONY: database-logs
database-logs: ## Affiche les logs PostgreSQL
	$(DOCKER_COMPOSE) logs -f database

.PHONY: database-up
database-up: ## Démarre uniquement PostgreSQL
	$(DOCKER_COMPOSE) up -d database

.PHONY: database-shell
database-shell: ## Ouvre psql dans le conteneur PostgreSQL
	$(DOCKER_COMPOSE) exec database \
		psql \
		-U $${POSTGRES_USER:-proxibet} \
		-d $${POSTGRES_DB:-proxibet}

.PHONY: docker-clean
docker-clean: ## Supprime les conteneurs et les données PostgreSQL
	$(DOCKER_COMPOSE) down --volumes --remove-orphans