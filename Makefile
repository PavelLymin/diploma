include .env
export

# Postgres commands
postgres-up:
	@docker compose up -d postgres

postgres-stop:
	@docker compose stop postgres

postgres-cleanup:
	@read -p "Are you sure you want to remove the database? (y/n): " answer; \
	if [ "$$answer" = "y" ]; then \
		docker compose down -v; \
		echo "Database removed."; \
	else \
		echo "Operation canceled."; \
	fi

port-forward:
	@docker compose up -d port-forwarder

port-forward-down:
	@docker compose down port-forwarder

# migrations commands
service ?= auth
db ?= auth_db
steps ?= 1

migration_dir := ./services/$(service)/migrations

database_url := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/$(db)?sslmode=disable

.PHONY: postgres-up migrate-create migrate-up migrate-down

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create service=auth name=create_auth_otps_table"; \
		exit 1; \
	fi
	@test -d "$(migration_dir)" || \
		(echo "Migration directory not found: $(migration_dir)"; exit 1)
	@docker compose run --rm --no-deps \
		-v "$(migration_dir):/migrations" \
		postgres-migration \
		create -ext sql -dir /migrations -seq "$(name)"

migrate-up:
	@test -d "$(migration_dir)" || \
		(echo "Migration directory not found: $(migration_dir)"; exit 1)
	@docker compose run --rm \
		-v "$(migration_dir):/migrations" \
		postgres-migration \
		-path /migrations \
		-database "$(database_url)" \
		up

migrate-down:
	@test -d "$(migration_dir)" || \
		(echo "Migration directory not found: $(migration_dir)"; exit 1)
	@docker compose run --rm \
		-v "$(migration_dir):/migrations" \
		postgres-migration \
		-path /migrations \
		-database "$(database_url)" \
		down $(steps)