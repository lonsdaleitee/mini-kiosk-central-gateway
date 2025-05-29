-- Description: Repeatable migrations for views
-- R__Views.sql

-- Create view for active users (using existing user table structure)
CREATE OR REPLACE VIEW active_users AS
SELECT id, username, first_name, last_name, created_date
FROM "user"
WHERE is_active = true 
  AND created_date > now() - interval '90 days';
