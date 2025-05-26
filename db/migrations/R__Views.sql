-- Description: Repeatable migrations for views
-- R__Views.sql

-- Create view for active users
CREATE OR REPLACE VIEW active_users AS
SELECT id, username, email, created_at
FROM users
WHERE created_at > now() - interval '90 days';
