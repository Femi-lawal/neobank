-- Rollback: Drop cards table
-- Version: 000004

DROP TRIGGER IF EXISTS update_cards_updated_at ON cards;
DROP INDEX IF EXISTS idx_cards_status;
DROP INDEX IF EXISTS idx_cards_account_id;
DROP INDEX IF EXISTS idx_cards_user_id;
DROP TABLE IF EXISTS cards;
