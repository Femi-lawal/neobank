-- Migration: Create cards table
-- Version: 000004
-- Description: Cards table for card service
-- NOTE: CVV is NEVER stored per PCI DSS 3.2 - only used for single-transaction validation

CREATE TABLE IF NOT EXISTS cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    -- Store encrypted card number (AES-256-GCM in production via app-level encryption)
    encrypted_card_number TEXT NOT NULL,
    -- Store only last 4 digits for display purposes
    masked_card_number VARCHAR(19) NOT NULL, -- Format: **** **** **** 1234
    expiration_date VARCHAR(5) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'BLOCKED', 'INACTIVE', 'EXPIRED')),
    -- Card token for payment processing (replaces actual card number)
    card_token UUID DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_cards_user_id ON cards(user_id);
CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards(account_id);
CREATE INDEX IF NOT EXISTS idx_cards_status ON cards(status);
CREATE INDEX IF NOT EXISTS idx_cards_token ON cards(card_token);

-- Add trigger for updated_at
CREATE TRIGGER update_cards_updated_at
    BEFORE UPDATE ON cards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

