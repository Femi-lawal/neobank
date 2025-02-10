# NeoBank Chef Cookbook

This cookbook deploys and manages the NeoBank banking application.

## Requirements

- Chef 16+
- Docker
- Target OS: Ubuntu 20.04/22.04 or CentOS 8+

## Recipes

| Recipe | Description |
|--------|-------------|
| `neobank::default` | Full deployment |
| `neobank::database` | PostgreSQL setup |
| `neobank::redis` | Redis caching layer |
| `neobank::backend` | All backend microservices |
| `neobank::frontend` | Next.js frontend |

## Usage

```ruby
# In your node's run_list:
include_recipe 'neobank::default'

# Or individual components:
include_recipe 'neobank::database'
include_recipe 'neobank::backend'
```

## Attributes

See `attributes/default.rb` for configurable options.

## Data Bags

Sensitive data is stored in encrypted data bags:

```bash
knife data bag create neobank secrets --secret-file ~/.chef/encrypted_data_bag_secret
```
