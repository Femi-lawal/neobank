# NeoBank Cookbook - Default Attributes

# Application
default['neobank']['app_name'] = 'neobank'
default['neobank']['user'] = 'neobank'
default['neobank']['group'] = 'neobank'
default['neobank']['install_dir'] = '/opt/neobank'
default['neobank']['log_dir'] = '/var/log/neobank'

# Docker Registry
default['neobank']['docker_registry'] = 'registry.neobank.com'
default['neobank']['docker_tag'] = 'latest'

# Database
default['neobank']['database']['host'] = 'localhost'
default['neobank']['database']['port'] = 5432
default['neobank']['database']['name'] = 'newbank_core'
default['neobank']['database']['user'] = 'neobank'
# Password should come from encrypted data bag

# Redis
default['neobank']['redis']['host'] = 'localhost'
default['neobank']['redis']['port'] = 6379

# Kafka
default['neobank']['kafka']['brokers'] = ['localhost:9092']

# Services configuration
default['neobank']['services'] = {
  'identity-service' => { 'port' => 8081, 'replicas' => 2 },
  'ledger-service' => { 'port' => 8082, 'replicas' => 2 },
  'payment-service' => { 'port' => 8083, 'replicas' => 2 },
  'product-service' => { 'port' => 8084, 'replicas' => 2 },
  'card-service' => { 'port' => 8085, 'replicas' => 2 },
  'frontend' => { 'port' => 3000, 'replicas' => 2 }
}

# Health checks
default['neobank']['health_check']['path'] = '/health'
default['neobank']['health_check']['interval'] = 30
default['neobank']['health_check']['timeout'] = 10
