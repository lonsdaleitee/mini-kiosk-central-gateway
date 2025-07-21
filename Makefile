.PHONY: setup migrate clean info validate new-migration new-repeatable build run test gateway-build gateway-run gateway-run-env gateway-dev gateway-sit gateway-staging gateway-prod gateway-test

# Database migration commands
migrate:
	flyway -configFiles=flyway.conf migrate

clean:
	flyway -configFiles=flyway.conf clean

info:
	flyway -configFiles=flyway.conf info

validate:
	flyway -configFiles=flyway.conf validate

repair:
	flyway -configFiles=flyway.conf repair

# Create a new migration file
new-migration:
	@read -p "Enter migration name: " name; \
	version=$$(date +%Y%m%d%H%M%S); \
	touch db/migrations/V$$version"__"$$name.sql; \
	echo "Created migration file: db/migrations/V$$version"__"$$name.sql"

# Create a repeatable migration
new-repeatable:
	@read -p "Enter repeatable migration name: " name; \
	touch db/migrations/R"__"$$name.sql; \
	echo "Created repeatable migration: db/migrations/R"__"$$name.sql"

# Build application
build:
	go build -o bin/server cmd/server/main.go

# Run application
run: build
	./bin/server

# Run tests
test:
	go test -v ./...

# Setup project dependencies
setup:
	go mod tidy
	go get github.com/lib/pq

# Gateway application commands
gateway-build:
	go build -o bin/gateway ./cmd/server

gateway-run: gateway-build
	./bin/gateway

# Run gateway application with specified environment
# Usage: make gateway-run-env ENV=dev|sit|staging|prod|local
# Example: make gateway-run-env ENV=dev
gateway-run-env: gateway-build
	@if [ -z "$(ENV)" ]; then \
		echo "Error: ENV parameter is required. Usage: make gateway-run-env ENV=dev|sit|staging|prod|local"; \
		exit 1; \
	fi; \
	if [ ! -f "configs/$(ENV).config.yaml" ]; then \
		echo "Error: Configuration file configs/$(ENV).config.yaml not found"; \
		echo "Available configurations: dev, local, prod, sit, staging"; \
		exit 1; \
	fi; \
	echo "Running gateway with $(ENV) environment configuration..."; \
	GATEWAY_CONFIG_ENV=$(ENV) ./bin/gateway

# Convenience shortcuts for common environments
gateway-dev: gateway-build
	@echo "Running gateway with dev environment configuration..."
	GATEWAY_CONFIG_ENV=dev ./bin/gateway

gateway-sit: gateway-build
	@echo "Running gateway with SIT environment configuration..."
	GATEWAY_CONFIG_ENV=sit ./bin/gateway

gateway-staging: gateway-build
	@echo "Running gateway with staging environment configuration..."
	GATEWAY_CONFIG_ENV=staging ./bin/gateway

gateway-prod: gateway-build
	@echo "Running gateway with production environment configuration..."
	GATEWAY_CONFIG_ENV=prod ./bin/gateway

gateway-test:
	go test ./...

gateway-dev: migrate gateway-run

gateway-clean:
	rm -rf bin/
