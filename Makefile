# Load environment variables from .env
include .env
export

# Database connection URL
DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Migration directory
MIGRATIONS_DIR=./internal/db/migrations

# Number of migrations for down/steps
STEPS?=1

.PHONY: migrate-up migrate-down migrate-force migrate-version migrate-create

# Run all up migrations
migrate-up:
	migrate.exe -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

# Rollback one migration (or specify STEPS=n)
migrate-down:
	migrate.exe -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down $(STEPS)

# Show current migration version
migrate-version:
	migrate.exe -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

# Force migration version (usage: make migrate-force VERSION=1)
migrate-force:
	migrate.exe -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)

# Create a new migration
# usage:
# make migrate-create NAME=create_users_table
migrate-create:
	migrate.exe create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)