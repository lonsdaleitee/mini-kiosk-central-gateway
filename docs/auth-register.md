# Auth Handler - Register Endpoint

## Overview
The `/api/v1/auth/register` endpoint has been implemented to handle user registration.

## Request Format

**Endpoint:** `POST /api/v1/auth/register`

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john.doe@example.com", 
  "first_name": "John",
  "last_name": "Doe",
  "password": "password123"
}
```

## Validation Rules

- **username**: Required, cannot be empty
- **email**: Required, must be valid email format
- **first_name**: Required, cannot be empty  
- **last_name**: Required, cannot be empty
- **password**: Required, minimum 6 characters

## Response Format

### Success Response (201 Created)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "User registered successfully"
}
```

### Error Responses

#### 400 Bad Request - Invalid Input
```json
{
  "error": "Invalid request body",
  "details": "validation error details"
}
```

#### 400 Bad Request - Empty Fields
```json
{
  "error": "All fields are required and cannot be empty"
}
```

#### 409 Conflict - User Already Exists
```json
{
  "error": "User with this username or email already exists"
}
```

#### 500 Internal Server Error
```json
{
  "error": "Failed to create user"
}
```

## Features Implemented

1. **Input Validation**: All required fields are validated
2. **Duplicate Prevention**: Checks for existing usernames and emails
3. **Password Security**: Passwords are hashed using bcrypt
4. **Database Integration**: Creates user records in PostgreSQL database
5. **Error Handling**: Comprehensive error responses

## Database Schema Requirements

The handler requires the following database schema:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Testing the Endpoint

You can test the endpoint using curl:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john.doe@example.com",
    "first_name": "John", 
    "last_name": "Doe",
    "password": "password123"
  }'
```

## Security Considerations

- Passwords are hashed using bcrypt with default cost (10)
- Input sanitization removes leading/trailing whitespace
- Email format validation is enforced
- Duplicate username/email prevention
