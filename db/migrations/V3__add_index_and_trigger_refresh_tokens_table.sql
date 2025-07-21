-- Description: Add more index and create trigger for last_used_at
-- V3__add_index_and_trigger_refresh_tokens_table.sql

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);

-- Add trigger for updating last_used_at
CREATE TRIGGER update_refresh_tokens_last_used
  BEFORE UPDATE ON refresh_tokens
  FOR EACH ROW
  EXECUTE FUNCTION update_modified_column();