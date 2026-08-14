GOOSE_VERSION := v3.27.3
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

MIGRATIONS_DIR := apps/api/migrations
ENV_FILE := .env

.PHONY: migrate-version migrate-status migrate-up migrate-down

migrate-version:
	@test -f $(ENV_FILE) || (echo "missing $(ENV_FILE); copy .env.example to .env"; exit 1)
	@set -a; . ./$(ENV_FILE); set +a; \
		$(GOOSE) postgres "$$DATABASE_URL" version

migrate-status:
	@test -f $(ENV_FILE) || (echo "missing $(ENV_FILE); copy .env.example to .env"; exit 1)
	@set -a; . ./$(ENV_FILE); set +a; \
		$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" status

migrate-up:
	@test -f $(ENV_FILE) || (echo "missing $(ENV_FILE); copy .env.example to .env"; exit 1)
	@set -a; . ./$(ENV_FILE); set +a; \
		$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" up

migrate-down:
	@test -f $(ENV_FILE) || (echo "missing $(ENV_FILE); copy .env.example to .env"; exit 1)
	@set -a; . ./$(ENV_FILE); set +a; \
		$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" down
