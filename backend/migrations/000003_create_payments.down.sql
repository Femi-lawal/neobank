-- Rollback: Drop payments table
-- Version: 000003

DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_to_account;
DROP INDEX IF EXISTS idx_payments_from_account;
DROP TABLE IF EXISTS payments;
