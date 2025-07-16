-- Description: Add first_name and last_name columns to users table
-- V2__add_user_names.sql

-- Add first_name and last_name columns to users table
ALTER TABLE users 
ADD COLUMN first_name VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN last_name VARCHAR(255) NOT NULL DEFAULT '';

-- Remove default constraints after adding columns
ALTER TABLE users 
ALTER COLUMN first_name DROP DEFAULT,
ALTER COLUMN last_name DROP DEFAULT;
