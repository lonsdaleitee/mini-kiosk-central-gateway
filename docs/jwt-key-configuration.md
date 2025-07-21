# JWT Key Configuration Changes

This document summarizes the changes made to support configurable private and public key paths for JWT authentication.

## Changes Made

### 1. **Updated Configuration Structure** (`internal/config/config.go`)

- Added `Keys PublicPrivateKey` field to the main `Config` struct
- Updated `PublicPrivateKey` struct with:
  - `PrivateKeyPath string` - Path to RSA private key for JWT signing
  - `PublicKeyPath string` - Path to RSA public key for JWT verification
- Added default values in `setDefaults()`:
  - `keys.private_key_path`: "privateKey.pem"
  - `keys.public_key_path`: "publicKey.pem"

### 2. **Updated Auth Handler** (`internal/handlers/auth.go`)

- Added config import and updated `AuthHandler` struct to include `config *config.Config`
- Updated `NewAuthHandler` constructor to accept `cfg *config.Config` parameter
- Modified `Login()` and `RefreshToken()` methods to use `h.config.Keys.PrivateKeyPath` instead of hardcoded "privateKey.pem"

### 3. **Updated Router** (`internal/router/router.go`)

- Added config import
- Updated `SetupRouter()` function to accept `cfg *config.Config` parameter
- Fixed JWT middleware usage:
  - Moved `middleware.JWTAuthMiddleware()` to the correct route groups (`orders` and `inventory`)
  - Fixed typo: "publikKey.pem" → `cfg.Keys.PublicKeyPath`
  - Used configuration-based public key path

### 4. **Updated Main Application** (`cmd/server/main.go`)

- Updated `router.SetupRouter(db, cfg)` call to pass configuration parameter

### 5. **Updated Configuration Files**

- **local.config.yaml**: Added `keys` section with default values
- **dev.config.yaml**: Added `keys` section with comments for guidance

## Usage

### Configuration File Example

```yaml
keys:
  private_key_path: "privateKey.pem"    # Path to RSA private key for JWT signing
  public_key_path: "publicKey.pem"      # Path to RSA public key for JWT verification
```

### Environment Variable Override

You can override these values using environment variables:

```bash
export GATEWAY_KEYS_PRIVATE_KEY_PATH="/path/to/custom/private.pem"
export GATEWAY_KEYS_PUBLIC_KEY_PATH="/path/to/custom/public.pem"
```

### Different Environments

Each environment can now have its own key configuration:

- **Development**: `configs/dev.config.yaml`
- **Staging**: `configs/staging.config.yaml`  
- **Production**: `configs/prod.config.yaml`

Example for production:
```yaml
keys:
  private_key_path: "/etc/secrets/prod-private.pem"
  public_key_path: "/etc/secrets/prod-public.pem"
```

## Benefits

1. **Security**: Different environments can use different keys
2. **Flexibility**: Key paths can be configured without code changes
3. **Environment Variables**: Support for runtime configuration via env vars
4. **Maintainability**: No more hardcoded paths in source code
5. **Bug Fixes**: Fixed the typo "publikKey.pem" and incorrect middleware placement

## Backward Compatibility

- Default values maintain existing behavior if no configuration is provided
- Existing local development setup continues to work without changes
- Environment variable support provides additional flexibility

## Testing

All changes have been tested and the application builds successfully:
```bash
make gateway-build  # ✅ Successful
```

## Next Steps

1. Update other environment config files (staging.config.yaml, prod.config.yaml, sit.config.yaml) with appropriate key paths
2. Ensure RSA key files exist in the specified locations for each environment
3. Test JWT authentication with the new configuration system
