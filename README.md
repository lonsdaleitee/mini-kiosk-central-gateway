# mini-kiosk-central-gateway
Central Gateway part of microservices architecture for mini kiosk app

## Overview

The **Mini Kiosk Central Gateway** is a secure API gateway built with Go and Gin framework that serves as the central entry point for a microservices-based kiosk application. It provides JWT-based authentication, request routing, and acts as a security layer between clients and downstream services.

## Features

### 🔐 Authentication & Security
- **JWT Authentication** with RSA-256 signing
- **Refresh Token** mechanism for secure token renewal
- **User Registration** with password hashing (bcrypt)
- **User Login** with credential validation
- **Token-based Authorization** middleware
- **Stateless Authentication** for horizontal scalability

### 🚪 Gateway Functionality
- **Request Routing** to downstream microservices
- **User Context Forwarding** via HTTP headers
- **CORS Support** for web applications
- **Health Check Endpoints** for monitoring
- **Request/Response Logging** with unique request IDs

### 🗄️ Database Integration
- **PostgreSQL** database with UUID support
- **Flyway Migrations** for schema management
- **User Management** with profile storage
- **Refresh Token Storage** with expiration handling

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
- `POST /api/v1/auth/register` - User registration
  ```json
  {
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "password": "secure123"
  }
  ```
- `POST /api/v1/auth/login` - User login
  ```json
  {
    "username": "johndoe",
    "password": "secure123"
  }
  ```
- `POST /api/v1/auth/refresh` - Refresh access token
  ```json
  {
    "refresh_token": "uuid-refresh-token"
  }
  ```
- `POST /api/v1/auth/logout` - User logout (revoke refresh token)
  ```json
  {
    "refresh_token": "uuid-refresh-token"
  }
  ```

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
1. `configs/local.config.yaml` file (for local development)
2. Environment variables (higher priority)

#### Configuration Structure:
```yaml
server:
  port: 8080
  host: "localhost"
  read_timeout: 30
  write_timeout: 30

database:
  host: "localhost"
  port: 5432
  user: "harrywijaya"
  password: ""
  dbname: "central_gateway_mini_kiosk"
  sslmode: "disable"

services:
  auth_service:
    base_url: "http://localhost:8081"
    timeout: 30
  order_service:
    base_url: "http://localhost:8082"
    timeout: 30
  inventory_service:
    base_url: "http://localhost:8083"
    timeout: 30
  payment_service:
    base_url: "http://localhost:8084"
    timeout: 30

gin:
  mode: "debug"  # debug, release, test

flyway:
  url: "jdbc:postgresql://localhost:5432/central_gateway_mini_kiosk"
  user: "harrywijaya"
  password: ""
  locations: "filesystem:db/migrations"
  connectRetries: 3
  outOfOrder: false
  validateMigrationNaming: true
  cleanDisabled: true

authentication:
  privateKeyLocation: "./"
```

## JWT Authentication

### RSA Key Generation

The gateway uses RSA-256 for JWT signing. Generate your keys:

```bash
# Generate private key
openssl genrsa -out privateKey.pem 2048

# Generate public key
openssl rsa -in privateKey.pem -pubout -out publicKey.pem
```

### Authentication Flow

1. **Registration**: User creates account with username, email, and password
2. **Login**: User receives JWT access token (15 min) + refresh token (7 days)
3. **API Requests**: Include `Authorization: Bearer <access_token>` header
4. **Token Refresh**: Use refresh token to get new access token when expired
5. **Logout**: Revoke refresh token to prevent further token generation

### JWT Claims Structure
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "full_name": "John Doe",
  "exp": 1234567890,
  "iat": 1234567000
}
```

## Database Schema

### Core Tables

**users table**
```sql
CREATE TABLE "user" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true
);
```

**refresh_tokens table**
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    token VARCHAR(512) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE
);
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestAuthHandler_Login_Success ./internal/handlers

# Run tests in verbose mode
go test -v ./internal/handlers
```

### Test Categories

- **Unit Tests**: Individual function testing
- **Integration Tests**: Handler and middleware testing
- **Authentication Tests**: JWT token validation and generation
- **Database Tests**: User registration and login flows

### Test Structure

```
internal/handlers/
├── auth.go              # Authentication handlers
├── auth_test.go         # Authentication tests
└── health.go           # Health check handlers
```

## Development

### Project Structure

```
mini-kiosk-central-gateway/
├── cmd/
│   └── server/
│       ├── main.go          # Application entry point
│       └── main_test.go     # Main function tests
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── database/
│   │   ├── database.go      # Database connection
│   │   └── migration.go     # Flyway migration runner
│   ├── handlers/
│   │   ├── auth.go          # Authentication handlers
│   │   ├── auth_test.go     # Authentication tests
│   │   └── health.go        # Health check handlers
│   ├── middleware/
│   │   ├── middleware.go    # General middleware
│   │   └── jwt.go          # JWT authentication middleware
│   ├── proxy/
│   │   └── proxy.go        # Service proxy functionality
│   ├── router/
│   │   └── router.go       # Route definitions
│   └── server/
│       └── server.go       # HTTP server setup
├── db/
│   └── migrations/         # Flyway SQL migrations
├── configs/
│   └── local.config.yaml  # Local configuration
├── docker/
│   └── docker-compose.yml # Docker setup
├── docs/                   # Documentation
├── Makefile               # Build and run commands
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── privateKey.pem         # RSA private key (generated)
├── publicKey.pem          # RSA public key (generated)
└── README.md             # This file
```

### API Testing Examples

Using curl:

```bash
# Register new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com", 
    "first_name": "John",
    "last_name": "Doe",
    "password": "secure123"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "secure123"
  }'

# Access protected endpoint
curl -X GET http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <your-jwt-token>"

# Refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<your-refresh-token>"
  }'
```

### Environment Variables

Override configuration with environment variables:

- `SERVER_PORT`: Gateway server port (default: 8080)
- `DB_HOST`: Database host (default: localhost)
- `DB_USER`: Database user (default: harrywijaya)
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name (default: central_gateway_mini_kiosk)
- `GIN_MODE`: Gin mode (debug, release, test)
- `PRIVATE_KEY_PATH`: Path to RSA private key
- `PUBLIC_KEY_PATH`: Path to RSA public key

### Next Steps

1. **✅ JWT Authentication**: RSA-based JWT with refresh tokens
2. **✅ User Management**: Registration, login, logout
3. **✅ Database Integration**: PostgreSQL with Flyway migrations
4. **✅ Testing Suite**: Comprehensive unit and integration tests
5. **🔄 Service Discovery**: Dynamic service registry
6. **🔄 Rate Limiting**: Request throttling and abuse prevention
7. **🔄 Circuit Breaker**: Resilience patterns for downstream services
8. **🔄 Monitoring**: Metrics collection and distributed tracing
9. **🔄 API Documentation**: OpenAPI/Swagger specification
10. **🔄 Docker Deployment**: Container orchestration setup
