# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security seriously at NeoBank. If you discover a security vulnerability, please follow these steps:

### Do NOT

- Open a public GitHub issue
- Disclose the vulnerability publicly before it's fixed
- Exploit the vulnerability beyond testing

### Do

1. **Email**: Send details to security@neobank.example.com
2. **Include**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

| Action | Timeline |
|--------|----------|
| Acknowledgment | 24 hours |
| Initial Assessment | 72 hours |
| Resolution Target | 7-30 days (based on severity) |
| Public Disclosure | After fix is deployed |

## Security Measures

### Authentication & Authorization

- JWT tokens with short expiration (15 minutes)
- Refresh token rotation
- Password hashing using bcrypt (cost factor 12)
- Account lockout after 5 failed attempts
- Multi-factor authentication support

### Data Protection

- All data encrypted in transit (TLS 1.3)
- Sensitive data encrypted at rest (AES-256)
- PCI DSS compliant card data handling
- No sensitive data in logs or error messages
- Automatic PII masking

### Input Validation

- All inputs validated and sanitized
- SQL injection prevention (parameterized queries)
- XSS prevention (content encoding)
- Request size limits enforced

### API Security

- Rate limiting per user and IP
- CORS properly configured
- Security headers (CSP, X-Frame-Options, etc.)
- API versioning
- Request/Response logging (sanitized)

### Infrastructure

- Secrets managed via Vault/K8s Secrets
- Container security scanning
- Dependency vulnerability scanning
- Network segmentation
- Principle of least privilege

## Security Headers

All API responses include:

```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Strict-Transport-Security: max-age=31536000; includeSubDomains
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

## Compliance

- PCI DSS Level 1 (for card data)
- GDPR (for EU users)
- SOC 2 Type II (in progress)

## Security Contacts

- Security Team: security@neobank.example.com
- Bug Bounty Program: Coming soon
