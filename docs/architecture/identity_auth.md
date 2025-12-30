# Identity & Authentication Architecture

This document details the authentication and identity management flows within NeoBank.

## Overview
The **Identity Service** is responsible for:
*   User Registration (KYC)
*   Authentication (Login with Email/Password)
*   Token Management (JWT Access & Refresh Tokens)
*   Security Controls (Account Lockout, Rate Limiting)

## Login Flow
The detailed flow for a user logging in.

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant IdentityService
    participant DB as PostgreSQL
    participant Redis

    User->>Frontend: Enter Email & Password
    Frontend->>IdentityService: POST /auth/login
    
    IdentityService->>Redis: Check Account Lockout (IsLocked?)
    alt Account Locked
        Redis-->>IdentityService: Locked
        IdentityService-->>Frontend: 401 Unauthorized (Locked)
        Frontend-->>User: "Account Locked" error
    else Account Not Locked
        IdentityService->>DB: Fetch User & Password Hash
        alt User Not Found (or Bad Password)
            IdentityService->>Redis: Increment Failed Attempts
            IdentityService-->>Frontend: 401 Unauthorized
        else Password Valid
            IdentityService->>Redis: Clear Failed Attempts
            IdentityService->>IdentityService: Generate JWT (Access + Refresh)
            IdentityService-->>Frontend: 200 OK { access_token, refresh_token }
            Frontend->>User: Redirect to Dashboard
        end
    end
```

## Security Mechanisms
1.  **Bcrypt**: Passwords are hashed using `bcrypt` (Cost 12).
2.  **JWT**:
    *   **Access Token**: Short-lived (15 mins). Contains `sub` (user_id), `role`.
    *   **Refresh Token**: Long-lived (7 days).
3.  **Account Lockout**:
    *   Implemented via `AccountLockout` (In-Memory/Redis).
    *   After `N` failed attempts, account is locked for `T` duration.
