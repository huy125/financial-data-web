
# ===================================================================================
# === Database Migration ============================================================
# ===================================================================================
MIGRATION_DIR=migrations
MIGRATE_IMAGE=migrate/migrate
DATA_SOURCE_NAME=postgres://root:.U6dfWsJcLB48uc-QUdAXHpPQ9T_HTUC@db:5432/db?sslmode=disable
# Find internal_network
NETWORK_NAME := $(shell docker network ls | grep internal_network | awk '{print $$2}')
ifeq ($(strip $(NETWORK_NAME)),)
	$(error "Container network is not found!")
endif

version:
	@echo "Current database version:"
	@docker run --rm --network $(NETWORK_NAME) \
		-v $(PWD)/$(MIGRATION_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		-source file:///migrations \
		-database $(DATA_SOURCE_NAME) version
.PHONY: version

up-schema: 
	@echo "Upgrade database schema"
	@docker run --rm --network $(NETWORK_NAME) -v $(PWD)/$(MIGRATION_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		-source file:///migrations \
		-database $(DATA_SOURCE_NAME) up
.PHONY: up-schema

down-schema:
	@echo "Downgrade last migration"
	@docker run --rm --network $(NETWORK_NAME) -v $(PWD)/$(MIGRATION_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		-source file:///migrations \
		-database $(DATA_SOURCE_NAME) down 1
.PHONY: down-schema

force-schema:
	@echo "Force to the specified version"
	@docker run --rm --network $(NETWORK_NAME) -v $(PWD)/$(MIGRATION_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		-source file:///migrations \
		-database $(DATA_SOURCE_NAME) force $(VERSION)
.PHONY: force-schema

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
