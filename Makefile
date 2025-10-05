# ---- Project Settings ----
APP_NAME := go-msg-dispatcher
COMPOSE  := docker compose
BUILD_DIR := build
API_URL := http://localhost:8080

# Default target
.PHONY: all
all: up

# ---- Docker lifecycle ----
.PHONY: up
up: ## Build and start all services (API, Postgres, Redis)
	$(COMPOSE) -f $(BUILD_DIR)/docker-compose.yml up --build -d

.PHONY: down
down: ## Stop and remove containers (keep volumes)
	$(COMPOSE) -f $(BUILD_DIR)/docker-compose.yml down

.PHONY: destroy
destroy: ## Stop containers and remove volumes (DB/Redis wiped)
	$(COMPOSE) -f $(BUILD_DIR)/docker-compose.yml down -v

.PHONY: logs
logs: ## Tail logs from the dispatcher API only
	$(COMPOSE) -f $(BUILD_DIR)/docker-compose.yml logs -f api
	
.PHONY: ps
ps: ## Show running containers
	$(COMPOSE) -f $(BUILD_DIR)/docker-compose.yml ps

.PHONY: restart
restart: ## Restart API only
	$(COMPOSE) -f $(BUILD_DIR)/docker-compose.yml restart api

# ---- Utilities ----
.PHONY: psql
psql: ## Open psql shell in Postgres
	docker exec -it msgsvc-postgres psql -U postgres -d msgsvc

.PHONY: redis
redis: ## Open redis-cli shell
	docker exec -it msgsvc-redis redis-cli

.PHONY: redis-dump
redis-dump: ## List cached sent metadata (messageId & sent_at)
	docker exec -it msgsvc-redis sh -lc '\
	for k in $$(redis-cli --raw --scan --pattern "msg:*:meta"); do \
	  v=$$(redis-cli --raw GET "$$k"); \
	  echo "$$k | $$v"; \
	done'

.PHONY: seed
seed: ## Insert 4 sample queued messages
	docker exec -i msgsvc-postgres psql -U postgres -d msgsvc -c "\
	  INSERT INTO messages (to_phone, content) VALUES \
	  ('+905551234567','Welcome to our platform! Your code is 4321.'), \
	  ('+905558888888','Reminder: Your appointment is tomorrow at 14:00.'), \
	  ('+905556666666','Security alert: A new login was detected on your account.'), \
	  ('+905557777777','Your package has been shipped and will arrive soon!');"

# ---- API helpers ----
.PHONY: start
start: ## Start the scheduler (POST /api/v1/scheduler/start)
	curl -s -X POST $(API_URL)/api/v1/scheduler/start | jq .

.PHONY: stop
stop: ## Stop the scheduler (POST /api/v1/scheduler/stop)
	curl -s -X POST $(API_URL)/api/v1/scheduler/stop | jq .

.PHONY: sent
sent: ## Show sent messages (GET /api/v1/messages/sent)
	curl -s "$(API_URL)/api/v1/messages/sent?limit=20" | jq .

.PHONY: create
create: ## Create a message (POST /api/v1/messages)
	curl -s -X POST $(API_URL)/api/v1/messages \
	  -H "Content-Type: application/json" \
	  -d '{"to_phone":"+905551234567","content":"Hello from Make!"}' | jq .

.PHONY: swagger
swagger: ## OpenAPI file location hint (update path if needed)
	@echo "OpenAPI spec at: internal/transport/http/swagger/openapi.yaml"

.PHONY: swagger-open
swagger-open: ## Open Swagger UI in your default browser
	@URL="$(API_URL)/swagger/"; \
	echo "Opening $$URL"; \
	( command -v xdg-open >/dev/null && xdg-open $$URL ) || \
	( command -v open >/dev/null && open $$URL ) || \
	( command -v start >/dev/null && start $$URL ) || true

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | sed 's/:.*##/: /' | column -t -s ':'
