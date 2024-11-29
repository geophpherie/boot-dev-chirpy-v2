# Makefile for managing database migrations and code generation

.PHONY: all
all:
	@echo "Run a specific target, e.g., 'make migrate-up'"

# Targets
.PHONY: migrate-up migrate-down generate

# Migrate up: apply the migrations
migrate-up:
	cd ./sql/schema/ && goose postgres postgres://postgres:postgres@localhost:5432/chirpy up


# Migrate down: revert the migrations
migrate-down:
	cd ./sql/schema/ && goose postgres postgres://postgres:postgres@localhost:5432/chirpy down

# Generate: run the code generation tool (e.g., to generate models, services, etc.)
generate:
	sqlc generate
