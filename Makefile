.PHONY: setup migrate clean info validate new-migration new-repeatable build run test gateway-build gateway-run gateway-test

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

gateway-test:
	go test ./...

gateway-dev: migrate gateway-run

gateway-clean:
	rm -rf bin/
