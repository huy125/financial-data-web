
# ===================================================================================
# === Database Migration ============================================================
# ===================================================================================
MIGRATION_DIR=migrations
MIGRATE_IMAGE=migrate/migrate
DATA_SOURCE_NAME=postgres://root:.U6dfWsJcLB48uc-QUdAXHpPQ9T_HTUC@localhost:5432/db?sslmode=disable

# Network settings - simplified for cross-platform use
NETWORK_NAME=host

# Check if we're on Windows
ifeq ($(OS),Windows_NT)
	# Windows-specific settings
	WINDOWS=1
	# Use Docker volume mount syntax for Windows
	VOLUME_MOUNT=-v /$(subst :,,$(subst \,/,$(CURDIR)))/$(MIGRATION_DIR):/migrations
else
	# Unix/Linux/MacOS settings
	VOLUME_MOUNT=-v $(CURDIR)/$(MIGRATION_DIR):/migrations
endif

version:
	@echo "Current database version:"
	@docker run --rm --network $(NETWORK_NAME) \
		$(VOLUME_MOUNT) \
		$(MIGRATE_IMAGE) \
		-source file:///migrations \
		-database $(DATA_SOURCE_NAME) version
.PHONY: version

up-schema: 
	@echo "Upgrade database schema"
	@docker run --rm --network $(NETWORK_NAME) $(VOLUME_MOUNT) \
		$(MIGRATE_IMAGE) \
		-source file:///migrations \
		-database $(DATA_SOURCE_NAME) up
.PHONY: up-schema

down-schema:
	@echo "Downgrade last migration"
	@docker run --rm --network $(NETWORK_NAME) $(VOLUME_MOUNT) \
		$(MIGRATE_IMAGE) \
		-source file:///migrations \
		-database $(DATA_SOURCE_NAME) down 1
.PHONY: down-schema

force-schema:
	@echo "Force to the specified version $(VERSION)..."
	@docker run --rm --network $(NETWORK_NAME) $(VOLUME_MOUNT) \
		$(MIGRATE_IMAGE) \
		-source file:///migrations \
		-database $(DATA_SOURCE_NAME) force $(VERSION)
.PHONY: force-schema

# ===================================================================================
# === Supabase Schema Synchronization ===============================================
# ===================================================================================

build-migrate:
	@echo "Building Supabase migration tool..."
	@go build -o bin/migrate ./cmd/migrate
.PHONY: build-migrate

sync-supabase: build-migrate
	@echo "Synchronizing schema with Supabase..."
	@./bin/migrate
.PHONY: sync-supabase

build-seed:
	@echo "Building Supabase seed data tool..."
	@go build -o bin/seed ./cmd/seed
.PHONY: build-seed

seed-supabase: build-seed
	@echo "Seeding Supabase database with dummy data..."
	@./bin/seed
.PHONY: seed-supabase

# Reset database - drop all tables and objects for a clean migration
reset-db:
	@echo "Resetting database (dropping all tables)..."
	@docker exec -it financial-db psql -U postgres -d db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "Database reset complete. Run 'make up-schema' to apply migrations."
.PHONY: reset-db

# ===================================================================================
# === Linter ========================================================================
# ===================================================================================
lint:
	@golangci-lint run
.PHONY: lint

lint-verbose:
	@golangci-lint run ./... \
		--verbose \
		--config ./.golangci.yml \
		--issues-exit-code=1 \
		--print-resources-usage
.PHONY: lint-verbose

sort-imports:
	@gci write --skip-generated -s standard -s default .
.PHONY: sort-imports
