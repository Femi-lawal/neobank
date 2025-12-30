-- ============================================================================
-- NeoBank Enhanced Seed Data Script
-- Run: Get-Content seed.sql | docker exec -i new_bank-postgres-1 psql -U user -d newbank_core
-- ============================================================================

-- Clear existing data
DELETE FROM postings;
DELETE FROM journal_entries;
DELETE FROM cards;
DELETE FROM accounts;
DELETE FROM products;
DELETE FROM users WHERE email IN ('demo@neobank.com', 'user@example.com', 'jane@neobank.com');

-- ============================================================================
-- USERS (password: "password123" for all - bcrypt hash below)
-- ============================================================================
INSERT INTO users (id, email, password_hash, first_name, last_name, role, kyc_status, created_at, updated_at) VALUES 
('11111111-1111-1111-1111-111111111111', 'demo@neobank.com', '$2a$12$KqTOwPMvEUhkF0wkYXKtzO3EM3IUtv8EWtpkWrIzNDkI1hGxs.CCm', 'Alex', 'Johnson', 'customer', 'VERIFIED', NOW() - INTERVAL '1 year', NOW()),
('22222222-2222-2222-2222-222222222222', 'user@example.com', '$2a$12$KqTOwPMvEUhkF0wkYXKtzO3EM3IUtv8EWtpkWrIzNDkI1hGxs.CCm', 'John', 'Doe', 'customer', 'VERIFIED', NOW() - INTERVAL '6 months', NOW()),
('33333333-3333-3333-3333-333333333333', 'jane@neobank.com', '$2a$12$KqTOwPMvEUhkF0wkYXKtzO3EM3IUtv8EWtpkWrIzNDkI1hGxs.CCm', 'Jane', 'Smith', 'admin', 'VERIFIED', NOW() - INTERVAL '2 years', NOW());

-- ============================================================================
-- PRODUCTS (15 Banking Products)
-- ============================================================================
INSERT INTO products (id, code, name, type, interest_rate, currency_code, metadata, created_at, updated_at) VALUES
('a0000001-0001-0001-0001-000000000001', 'SAVINGS-STD', 'Standard Savings', 'SAVINGS', 0.0250, 'USD', '{"min_balance": 0, "features": ["No fees", "Free transfers"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000002', 'SAVINGS-HY', 'High-Yield Savings', 'SAVINGS', 0.0450, 'USD', '{"min_balance": 1000, "features": ["4.5% APY", "FDIC insured"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000003', 'SAVINGS-KIDS', 'Kids Savings', 'SAVINGS', 0.0300, 'USD', '{"age_limit": 17, "features": ["Parent controls"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000004', 'CHECKING-STD', 'Everyday Checking', 'CHECKING', 0.0010, 'USD', '{"features": ["Free debit card", "Mobile deposit"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000005', 'CHECKING-PREMIUM', 'Premium Checking', 'CHECKING', 0.0100, 'USD', '{"min_balance": 5000, "features": ["Priority support", "No foreign fees"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000006', 'CHECKING-BUSINESS', 'Business Checking', 'CHECKING', 0.0050, 'USD', '{"features": ["500 free transactions", "QuickBooks sync"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000007', 'LOAN-PERSONAL', 'Personal Loan', 'LOAN', 0.0899, 'USD', '{"max_amount": 50000, "terms": [12, 24, 36, 48, 60]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000008', 'LOAN-AUTO', 'Auto Loan', 'LOAN', 0.0599, 'USD', '{"max_amount": 100000, "terms": [36, 48, 60, 72, 84]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000009', 'LOAN-MORTGAGE', 'Home Mortgage', 'LOAN', 0.0689, 'USD', '{"max_amount": 1000000, "terms": [180, 360]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000010', 'SAVINGS-EUR', 'Euro Savings', 'SAVINGS', 0.0350, 'EUR', '{"features": ["Multi-currency", "Competitive FX"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000011', 'SAVINGS-GBP', 'Sterling Savings', 'SAVINGS', 0.0400, 'GBP', '{"features": ["UK banking", "ISA eligible"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000012', 'CHECKING-CAD', 'Canadian Checking', 'CHECKING', 0.0015, 'CAD', '{"features": ["Cross-border", "No CAD fees"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000013', 'SAVINGS-JPY', 'Yen Savings', 'SAVINGS', 0.0010, 'JPY', '{"features": ["Japan banking", "Wire transfers"]}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000014', 'LOAN-STUDENT', 'Student Loan', 'LOAN', 0.0499, 'USD', '{"max_amount": 150000, "deferment": true}', NOW(), NOW()),
('a0000001-0001-0001-0001-000000000015', 'SAVINGS-RETIREMENT', 'Retirement Savings', 'SAVINGS', 0.0500, 'USD', '{"features": ["IRA eligible", "Tax advantages"]}', NOW(), NOW());

-- ============================================================================
-- ACCOUNTS (12 Demo User Accounts)
-- ============================================================================
INSERT INTO accounts (id, user_id, account_number, name, type, currency_code, status, balance_version, cached_balance, metadata, created_at, updated_at) VALUES
('b0000001-0001-0001-0001-000000000001', '11111111-1111-1111-1111-111111111111', 'ACC-1001', 'Primary Checking', 'ASSET', 'USD', 'ACTIVE', 150, 18458.75, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '11 months', NOW()),
('b0000001-0001-0001-0001-000000000002', '11111111-1111-1111-1111-111111111111', 'ACC-1002', 'High-Yield Savings', 'ASSET', 'USD', 'ACTIVE', 80, 52230.00, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '10 months', NOW()),
('b0000001-0001-0001-0001-000000000003', '11111111-1111-1111-1111-111111111111', 'ACC-1003', 'Emergency Fund', 'ASSET', 'USD', 'ACTIVE', 40, 25000.00, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '9 months', NOW()),
('b0000001-0001-0001-0001-000000000004', '11111111-1111-1111-1111-111111111111', 'ACC-1004', 'Vacation Fund', 'ASSET', 'USD', 'ACTIVE', 25, 4850.50, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '6 months', NOW()),
('b0000001-0001-0001-0001-000000000005', '11111111-1111-1111-1111-111111111111', 'ACC-1005', 'Euro Account', 'ASSET', 'EUR', 'ACTIVE', 15, 12500.00, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '5 months', NOW()),
('b0000001-0001-0001-0001-000000000006', '11111111-1111-1111-1111-111111111111', 'ACC-1006', 'British Pounds', 'ASSET', 'GBP', 'ACTIVE', 12, 6200.00, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '4 months', NOW()),
('b0000001-0001-0001-0001-000000000007', '11111111-1111-1111-1111-111111111111', 'ACC-1007', 'Investment Reserve', 'ASSET', 'USD', 'ACTIVE', 30, 35000.00, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '8 months', NOW()),
('b0000001-0001-0001-0001-000000000008', '11111111-1111-1111-1111-111111111111', 'ACC-1008', 'Business Account', 'ASSET', 'USD', 'ACTIVE', 60, 98450.25, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '7 months', NOW()),
('b0000001-0001-0001-0001-000000000009', '11111111-1111-1111-1111-111111111111', 'ACC-1009', 'Retirement Fund', 'ASSET', 'USD', 'ACTIVE', 20, 125000.00, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '1 year', NOW()),
('b0000001-0001-0001-0001-000000000010', '11111111-1111-1111-1111-111111111111', 'ACC-1010', 'Kids College Fund', 'ASSET', 'USD', 'ACTIVE', 15, 45000.00, '{"owner_id": "11111111-1111-1111-1111-111111111111"}', NOW() - INTERVAL '2 years', NOW()),
('b0000002-0002-0002-0002-000000000001', '22222222-2222-2222-2222-222222222222', 'ACC-2001', 'Main Checking', 'ASSET', 'USD', 'ACTIVE', 50, 8680.30, '{"owner_id": "22222222-2222-2222-2222-222222222222"}', NOW() - INTERVAL '5 months', NOW()),
('b0000002-0002-0002-0002-000000000002', '22222222-2222-2222-2222-222222222222', 'ACC-2002', 'Savings', 'ASSET', 'USD', 'ACTIVE', 25, 18500.00, '{"owner_id": "22222222-2222-2222-2222-222222222222"}', NOW() - INTERVAL '5 months', NOW());

-- ============================================================================
-- CARDS (8 Cards)
-- ============================================================================
INSERT INTO cards (id, account_id, user_id, masked_card_number, encrypted_card_number, expiration_date, status, daily_limit, created_at, updated_at) VALUES
('e0000001-0001-0001-0001-000000000001', 'b0000001-0001-0001-0001-000000000001', '11111111-1111-1111-1111-111111111111', '**** **** **** 0366', 'encrypted_placeholder', '12/27', 'ACTIVE', 5000.00, NOW() - INTERVAL '11 months', NOW()),
('e0000001-0001-0001-0001-000000000002', 'b0000001-0001-0001-0001-000000000002', '11111111-1111-1111-1111-111111111111', '**** **** **** 9903', 'encrypted_placeholder', '03/28', 'ACTIVE', 10000.00, NOW() - INTERVAL '6 months', NOW()),
('e0000001-0001-0001-0001-000000000003', 'b0000001-0001-0001-0001-000000000008', '11111111-1111-1111-1111-111111111111', '**** **** **** 0126', 'encrypted_placeholder', '08/26', 'ACTIVE', 25000.00, NOW() - INTERVAL '7 months', NOW()),
('e0000001-0001-0001-0001-000000000004', 'b0000001-0001-0001-0001-000000000001', '11111111-1111-1111-1111-111111111111', '**** **** **** 2832', 'encrypted_placeholder', '06/25', 'ACTIVE', 1000.00, NOW() - INTERVAL '2 months', NOW()),
('e0000001-0001-0001-0001-000000000005', 'b0000001-0001-0001-0001-000000000001', '11111111-1111-1111-1111-111111111111', '**** **** **** 4305', 'encrypted_placeholder', '09/24', 'BLOCKED', 2000.00, NOW() - INTERVAL '1 year', NOW()),
('e0000001-0001-0001-0001-000000000006', 'b0000001-0001-0001-0001-000000000005', '11111111-1111-1111-1111-111111111111', '**** **** **** 5556', 'encrypted_placeholder', '05/27', 'ACTIVE', 3000.00, NOW() - INTERVAL '4 months', NOW()),
('e0000002-0001-0001-0001-000000000001', 'b0000002-0002-0002-0002-000000000001', '22222222-2222-2222-2222-222222222222', '**** **** **** 5100', 'encrypted_placeholder', '11/26', 'ACTIVE', 3000.00, NOW() - INTERVAL '5 months', NOW()),
('e0000002-0001-0001-0001-000000000002', 'b0000002-0002-0002-0002-000000000002', '22222222-2222-2222-2222-222222222222', '**** **** **** 1111', 'encrypted_placeholder', '02/28', 'ACTIVE', 5000.00, NOW() - INTERVAL '3 months', NOW());

-- ============================================================================
-- JOURNAL ENTRIES & POSTINGS (100+ transactions over 12 months)
-- ============================================================================

-- Monthly Salary (12 months)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c1' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 month' * (12 - n)),
    'Salary Deposit - ' || to_char(NOW() - (INTERVAL '1 month' * (12 - n)), 'Month YYYY'),
    'PAY-' || to_char(NOW() - (INTERVAL '1 month' * (12 - n)), 'YYYY-MM'),
    'POSTED',
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d1' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c1' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    8500 + (n * 100),
    1,
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Recurring subscriptions (12 months each)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c2' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 month' * (12 - n)),
    'Netflix Subscription',
    'NFLX-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d2' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c2' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    15.99,
    -1,
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Spotify (12 months)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c3' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 month' * (12 - n)),
    'Spotify Premium',
    'SPOT-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d3' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c3' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    9.99,
    -1,
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Gym membership (12 months)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c4' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 month' * (12 - n)),
    'Gym Membership - Planet Fitness',
    'GYM-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d4' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c4' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    49.99,
    -1,
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Electric bills (12 months, varying amounts)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c5' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 month' * (12 - n)),
    'Electric Bill - ConEdison',
    'ELEC-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d5' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c5' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    80 + (random() * 100)::numeric(10,2),
    -1,
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Rent payments (12 months)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c6' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 month' * (12 - n)),
    'Rent Payment - Apartment',
    'RENT-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d6' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c6' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    2500.00,
    -1,
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Grocery shopping (weekly for 52 weeks)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c7' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 week' * (52 - n)),
    CASE n % 4 
        WHEN 0 THEN 'Whole Foods Market'
        WHEN 1 THEN 'Trader Joes'
        WHEN 2 THEN 'Costco'
        ELSE 'Target'
    END,
    'GRC-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 week' * (52 - n))
FROM generate_series(1, 52) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d7' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c7' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    50 + (random() * 150)::numeric(10,2),
    -1,
    NOW() - (INTERVAL '1 week' * (52 - n))
FROM generate_series(1, 52) n;

-- Dining out (2x per week for 52 weeks = 104)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c8' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '3 days' * (104 - n)),
    CASE n % 6 
        WHEN 0 THEN 'Chipotle'
        WHEN 1 THEN 'Starbucks'
        WHEN 2 THEN 'Local Restaurant'
        WHEN 3 THEN 'Uber Eats'
        WHEN 4 THEN 'DoorDash'
        ELSE 'McDonalds'
    END,
    'DINE-' || n,
    'POSTED',
    NOW() - (INTERVAL '3 days' * (104 - n))
FROM generate_series(1, 104) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d8' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c8' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    8 + (random() * 50)::numeric(10,2),
    -1,
    NOW() - (INTERVAL '3 days' * (104 - n))
FROM generate_series(1, 104) n;

-- Gas/Fuel (weekly for 52 weeks)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('c9' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 week' * (52 - n)),
    CASE n % 3 
        WHEN 0 THEN 'Shell Gas Station'
        WHEN 1 THEN 'Exxon Mobil'
        ELSE 'Chevron'
    END,
    'GAS-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 week' * (52 - n))
FROM generate_series(1, 52) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('d9' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('c9' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    35 + (random() * 35)::numeric(10,2),
    -1,
    NOW() - (INTERVAL '1 week' * (52 - n))
FROM generate_series(1, 52) n;

-- Savings transfers (monthly for 12 months)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('ca' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 month' * (12 - n)),
    'Monthly Savings Transfer',
    'SAV-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Debit from checking
INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('da' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('ca' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    2000.00,
    -1,
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Credit to savings
INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('db' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('ca' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000002',
    2000.00,
    1,
    NOW() - (INTERVAL '1 month' * (12 - n))
FROM generate_series(1, 12) n;

-- Amazon purchases (monthly)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at)
SELECT 
    ('cb' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    NOW() - (INTERVAL '1 month' * (12 - n)) + (INTERVAL '5 days'),
    'Amazon.com Purchase',
    'AMZ-' || n,
    'POSTED',
    NOW() - (INTERVAL '1 month' * (12 - n)) + (INTERVAL '5 days')
FROM generate_series(1, 12) n;

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at)
SELECT 
    ('dc' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    ('cb' || lpad(n::text, 6, '0') || '-0000-0000-0000-000000000000')::uuid,
    'b0000001-0001-0001-0001-000000000001',
    25 + (random() * 200)::numeric(10,2),
    -1,
    NOW() - (INTERVAL '1 month' * (12 - n)) + (INTERVAL '5 days')
FROM generate_series(1, 12) n;

-- Interest payments (quarterly)
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at) VALUES
('cc000001-0000-0000-0000-000000000000', NOW() - INTERVAL '9 months', 'Interest Payment Q1', 'INT-Q1', 'POSTED', NOW() - INTERVAL '9 months'),
('cc000002-0000-0000-0000-000000000000', NOW() - INTERVAL '6 months', 'Interest Payment Q2', 'INT-Q2', 'POSTED', NOW() - INTERVAL '6 months'),
('cc000003-0000-0000-0000-000000000000', NOW() - INTERVAL '3 months', 'Interest Payment Q3', 'INT-Q3', 'POSTED', NOW() - INTERVAL '3 months'),
('cc000004-0000-0000-0000-000000000000', NOW(), 'Interest Payment Q4', 'INT-Q4', 'POSTED', NOW());

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at) VALUES
('dd000001-0000-0000-0000-000000000000', 'cc000001-0000-0000-0000-000000000000', 'b0000001-0001-0001-0001-000000000002', 156.78, 1, NOW() - INTERVAL '9 months'),
('dd000002-0000-0000-0000-000000000000', 'cc000002-0000-0000-0000-000000000000', 'b0000001-0001-0001-0001-000000000002', 189.45, 1, NOW() - INTERVAL '6 months'),
('dd000003-0000-0000-0000-000000000000', 'cc000003-0000-0000-0000-000000000000', 'b0000001-0001-0001-0001-000000000002', 212.34, 1, NOW() - INTERVAL '3 months'),
('dd000004-0000-0000-0000-000000000000', 'cc000004-0000-0000-0000-000000000000', 'b0000001-0001-0001-0001-000000000002', 245.67, 1, NOW());

-- Tax refund
INSERT INTO journal_entries (id, transaction_date, description, reference_id, status, created_at) VALUES
('ce000001-0000-0000-0000-000000000000', NOW() - INTERVAL '8 months', 'IRS Tax Refund 2023', 'IRS-REFUND', 'POSTED', NOW() - INTERVAL '8 months');

INSERT INTO postings (id, journal_entry_id, account_id, amount, direction, created_at) VALUES
('de000001-0000-0000-0000-000000000000', 'ce000001-0000-0000-0000-000000000000', 'b0000001-0001-0001-0001-000000000001', 4523.00, 1, NOW() - INTERVAL '8 months');

-- ============================================================================
-- Summary
-- ============================================================================
SELECT 'Seed data loaded!' as status;
SELECT count(*) as users FROM users;
SELECT count(*) as products FROM products;
SELECT count(*) as accounts FROM accounts;
SELECT count(*) as cards FROM cards;
SELECT count(*) as transactions FROM journal_entries;
SELECT count(*) as postings FROM postings;
SELECT 'Login: demo@neobank.com / password123' as credentials;
