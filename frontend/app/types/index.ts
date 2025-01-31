export interface Account {
    ID: string;
    Name: string;
    AccountNumber: string;
    CurrencyCode: string;
    CachedBalance: string; // Decimal comes as string from JSON often
    Status: string;
    Type?: string; // Account type: ASSET, SAVINGS, CHECKING, INVESTMENT
}


export interface Product {
    Code: string;
    Name: string;
    Type: string;
    InterestRate: string;
}

export interface Card {
    ID: string;
    CardNumber: string;
    ExpirationDate: string;
    Status: string;
}
