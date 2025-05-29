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
