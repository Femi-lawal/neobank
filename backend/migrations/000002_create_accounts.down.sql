-- Rollback: Drop accounts table
-- Version: 000002

DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP INDEX IF EXISTS idx_accounts_account_type;
DROP INDEX IF EXISTS idx_accounts_user_id;
DROP TABLE IF EXISTS accounts;
