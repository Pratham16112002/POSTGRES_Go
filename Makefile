include .envrc
MIGRATIONS_PATH = ./cmd/migrate/migrations

migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

migrate-revert:
	@migrate -path=$(MIGRATIONS_PATH)  -database=$(DB_ADDR)  force 10

.PHONY: migration  migrate  migrate-down
