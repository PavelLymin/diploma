include .env
export

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

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a migration name using the 'name' variable."; \
		exit 1; \
	fi; \
	docker compose run --rm postgres-migration \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(name)"

migrate-up:
	@$(MAKE) migrate-action action=up

migrate-down:
	@$(MAKE) migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Error: Please provide an action using the 'action' variable."; \
		exit 1; \
	fi; \
	docker compose run --rm postgres-migration \
		-path /migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/$(POSTGRES_DB)?sslmode=disable" \
		$(action)