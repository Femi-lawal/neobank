export interface Account {
    id: string;
    user_id: string;
    account_number: string;
    name: string;
    type: string; // Account type: ASSET, SAVINGS, CHECKING, INVESTMENT
    currency_code: string;
    status: string;
    balance: string; // Decimal comes as string from JSON
    created_at: string;
    updated_at: string;
}


export interface Product {
    ID: string;
    Code: string;
    Name: string;
    Type: string;
    InterestRate: string;
    CurrencyCode: string;
}

export interface Card {
    id: string;
    user_id: string;
    account_id: string;
    card_number: string;    // Masked card number: **** **** **** 1234
    expiration_date: string; // MM/YY
    status: string;
    card_token: string;
    daily_limit: string;
    created_at: string;
    updated_at: string;
}
