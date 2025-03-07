# Contributing to NeoBank

Thank you for your interest in contributing to NeoBank! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Pull Request Process](#pull-request-process)
- [Style Guidelines](#style-guidelines)

## Code of Conduct

Please be respectful and inclusive. We welcome contributions from everyone.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/neobank.git`
3. Add upstream remote: `git remote add upstream https://github.com/Femi-lawal/neobank.git`

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### Backend Setup

```bash
cd backend
go work sync
make build
```

### Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

### Running with Docker

```bash
docker-compose up -d
```

## Making Changes

1. Create a feature branch: `git checkout -b feature/your-feature`
2. Make your changes
3. Write/update tests
4. Run tests: `make test`
5. Commit with conventional commits:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation
   - `test:` for tests
   - `chore:` for maintenance

## Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all tests pass
4. Update the README if applicable
5. Submit PR against `main` branch

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How was this tested?

## Checklist
- [ ] My code follows the style guidelines
- [ ] I have added tests
- [ ] All tests pass
- [ ] I have updated documentation
```

## Style Guidelines

### Go

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing

### TypeScript/React

- Use functional components with hooks
- Follow ESLint configuration
- Use TypeScript strict mode

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

## Questions?

Open an issue or reach out to the maintainers.
