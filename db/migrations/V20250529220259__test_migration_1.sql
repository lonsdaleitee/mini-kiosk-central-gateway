-- Description: Add email column to user table and create audit function
-- V20250529220259__test_migration_1.sql

-- Add email column to existing user table
ALTER TABLE "user" ADD COLUMN email VARCHAR(255);

-- Create index on email
CREATE INDEX idx_user_email ON "user"(email);

-- Create audit function for logging user changes
CREATE OR REPLACE FUNCTION log_user_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        RAISE NOTICE 'User % was updated at %', NEW.username, NOW();
    ELSIF TG_OP = 'INSERT' THEN
        RAISE NOTICE 'User % was created at %', NEW.username, NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for user changes
CREATE TRIGGER user_audit_trigger
    AFTER INSERT OR UPDATE ON "user"
    FOR EACH ROW EXECUTE FUNCTION log_user_changes();