# mini-kiosk-central-gateway
Central Gateway part of microservices architecture for mini kiosk app

## Database Migration with Flyway

This project uses Flyway for database migration to manage PostgreSQL schema, including tables, functions, triggers, and views.

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Flyway CLI (`brew install flyway`)

### Setup

1. Start the PostgreSQL database:
   ```bash
   docker-compose up -d
   ```
   
   (Or use your existing local PostgreSQL instance)

2. Run database migrations:
   ```bash
   make migrate
   ```

3. Check migration status:
   ```bash
   make info
   ```

### Creating Migrations

#### Version Migrations (V__*)
For schema changes that should only run once:

```bash
make new-migration
# You'll be prompted for a name
```

#### Repeatable Migrations (R__*)
For objects that can be replaced (views, functions, procedures):

```bash
make new-repeatable
# You'll be prompted for a name
```

### Other Commands

```bash
# Clean the database (drops all objects)
make clean

# Validate migrations
make validate

# Build and run the application
make run
```

### Flyway Configuration

The project is configured to use:
- Database: `central_gateway_mini_kiosk`
- User: `harrywijaya` (current system user)
- Migration files location: `db/migrations/`

### Migration Examples

The project includes examples of:
- Table creation with UUID support
- Custom functions and triggers
- Database views
- Indexes

All PostgreSQL objects (tables, functions, triggers, views, etc.) are fully supported through Flyway's SQL-based migrations.

## Central Gateway

The central gateway acts as the entry point for all client requests in the mini-kiosk microservices architecture. It handles:

- **Request Routing**: Routes incoming requests to appropriate downstream services
- **Health Checks**: Provides `/health` and `/ready` endpoints for monitoring
- **Request Logging**: Logs all incoming requests with unique request IDs
- **CORS Handling**: Manages cross-origin resource sharing
- **Configuration Management**: Centralized configuration for all services

### Gateway Endpoints

#### Health Endpoints
- `GET /health` - Returns gateway health status
- `GET /ready` - Returns gateway readiness status

#### API Endpoints (v1)
All API endpoints are prefixed with `/api/v1/`:

**Authentication Service**
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration  
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/profile` - Get user profile

**Order Service**
- `GET /api/v1/orders/` - List orders
- `POST /api/v1/orders/` - Create new order
- `GET /api/v1/orders/:id` - Get specific order
- `PUT /api/v1/orders/:id` - Update order
- `DELETE /api/v1/orders/:id` - Delete order

**Inventory Service**
- `GET /api/v1/inventory/` - List inventory items
- `GET /api/v1/inventory/:id` - Get specific item
- `PUT /api/v1/inventory/:id` - Update inventory item

**Payment Service**
- `POST /api/v1/payments/` - Create payment
- `GET /api/v1/payments/:id` - Get payment status
- `POST /api/v1/payments/:id/refund` - Process refund

### Running the Gateway

#### Using Make commands:
```bash
# Build the gateway
make gateway-build

# Run the gateway (includes build)
make gateway-run

# Run database migrations and then start gateway
make gateway-dev

# Run tests
make gateway-test

# Clean build artifacts
make gateway-clean
```

#### Manual commands:
```bash
# Install dependencies
go mod tidy

# Build the application
go build -o bin/gateway ./cmd/server

# Run the application
./bin/gateway
```

### Configuration

The gateway can be configured via:
1. `config.yaml` file in the project root
2. Environment variables (higher priority)

#### Environment Variables:
- `SERVER_PORT`: Gateway server port (default: 8080)
- `DB_HOST`: Database host (default: localhost)
- `DB_USER`: Database user (default: harrywijaya)
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name (default: central_gateway_mini_kiosk)

### Next Steps

1. **Implement Service Discovery**: Add service registry for dynamic service discovery
2. **Add Authentication Middleware**: Implement JWT token validation
3. **Rate Limiting**: Add rate limiting to prevent abuse
4. **Circuit Breaker**: Implement circuit breaker pattern for resilience
5. **Monitoring**: Add metrics collection and distributed tracing
6. **API Documentation**: Generate OpenAPI/Swagger documentation
