# Environment-Based Configuration

The gateway application now supports running with different environment configurations. This allows you to easily switch between development, staging, and production environments without modifying code.

## Available Configurations

The following environment configurations are available:
- `local` - Local development environment (default)
- `dev` - Development environment
- `sit` - System Integration Testing environment
- `staging` - Staging environment
- `prod` - Production environment

Each configuration corresponds to a YAML file in the `configs/` directory:
- `configs/local.config.yaml`
- `configs/dev.config.yaml`
- `configs/sit.config.yaml`
- `configs/staging.config.yaml`
- `configs/prod.config.yaml`

## Usage

### Method 1: Using the generic environment command

```bash
# Run with a specific environment
make gateway-run-env ENV=dev
make gateway-run-env ENV=sit
make gateway-run-env ENV=staging
make gateway-run-env ENV=prod
```

### Method 2: Using convenience shortcuts

```bash
# Development environment
make gateway-dev

# SIT environment
make gateway-sit

# Staging environment
make gateway-staging

# Production environment
make gateway-prod

# Local environment (default behavior)
make gateway-run
```

## How it Works

1. The Makefile sets the `GATEWAY_CONFIG_ENV` environment variable based on your choice
2. The application's configuration loader reads this environment variable
3. It loads the corresponding configuration file from the `configs/` directory
4. If no environment is specified, it defaults to `local.config.yaml`

## Error Handling

- If you don't specify an environment for `gateway-run-env`, you'll get an error message with usage instructions
- If the specified configuration file doesn't exist, you'll get an error listing available configurations
- The application will log which configuration file it's using when it starts

## Example Output

```bash
$ make gateway-run-env ENV=dev
Running gateway with dev environment configuration...
2024/07/21 10:30:45 Using config file: /path/to/configs/dev.config.yaml (environment: dev)
```

## Migration from Previous Version

The default behavior remains unchanged - running `make gateway-run` will still use the local configuration. Existing scripts and workflows will continue to work without modification.
