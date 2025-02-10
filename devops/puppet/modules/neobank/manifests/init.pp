# NeoBank Puppet Module - Main Manifest
#
# Deploys and manages the NeoBank banking application
#
# @param environment Environment name (dev, staging, prod)
# @param docker_registry Docker registry URL
# @param docker_tag Docker image tag
# @param db_host Database host
# @param db_port Database port
# @param db_name Database name
# @param redis_host Redis host
#
class neobank (
  String $environment     = 'dev',
  String $docker_registry = 'registry.neobank.com',
  String $docker_tag      = 'latest',
  String $db_host         = 'localhost',
  Integer $db_port        = 5432,
  String $db_name         = 'newbank_core',
  String $redis_host      = 'localhost',
  Integer $redis_port     = 6379,
  Hash $services          = {},
) {

  # Ensure Docker is installed
  include docker
  
  # Create application user
  user { 'neobank':
    ensure     => present,
    system     => true,
    managehome => true,
    home       => '/opt/neobank',
  }
  
  # Create necessary directories
  $directories = [
    '/opt/neobank',
    '/opt/neobank/config',
    '/opt/neobank/logs',
    '/var/log/neobank',
  ]
  
  file { $directories:
    ensure  => directory,
    owner   => 'neobank',
    group   => 'neobank',
    mode    => '0755',
    require => User['neobank'],
  }
  
  # Create Docker network
  docker_network { 'neobank_network':
    ensure => present,
  }
  
  # Include component classes based on roles
  if $facts['role'] == 'neobank-db' {
    include neobank::database
  }
  
  if $facts['role'] == 'neobank-redis' {
    include neobank::redis
  }
  
  if $facts['role'] == 'neobank-backend' {
    include neobank::backend
  }
  
  if $facts['role'] == 'neobank-frontend' {
    include neobank::frontend
  }
}
