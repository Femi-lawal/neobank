# NeoBank Backend Class
#
# Deploys backend microservices using Docker
#
class neobank::backend (
  String $docker_registry = $neobank::docker_registry,
  String $docker_tag      = $neobank::docker_tag,
  Hash $services          = {},
) inherits neobank {

  # Get secrets from Hiera
  $db_password = lookup('neobank::secrets::db_password', String)
  $jwt_secret  = lookup('neobank::secrets::jwt_secret', String)
  
  # Common environment variables
  $common_env = [
    "DB_HOST=${neobank::db_host}",
    "DB_PORT=${neobank::db_port}",
    "DB_NAME=${neobank::db_name}",
    "DB_USER=neobank",
    "DB_PASSWORD=${db_password}",
    "JWT_SECRET=${jwt_secret}",
    "REDIS_ADDR=${neobank::redis_host}:${neobank::redis_port}",
  ]
  
  # Default services if not specified
  $default_services = {
    'identity-service' => { 'port' => 8081 },
    'ledger-service'   => { 'port' => 8082 },
    'payment-service'  => { 'port' => 8083 },
    'product-service'  => { 'port' => 8084 },
    'card-service'     => { 'port' => 8085 },
  }
  
  $managed_services = empty($services) ? {
    true    => $default_services,
    default => $services,
  }
  
  # Deploy each service
  $managed_services.each |String $service_name, Hash $config| {
    $port = $config['port']
    
    # Create environment file
    file { "/opt/neobank/config/${service_name}.env":
      ensure  => file,
      owner   => 'neobank',
      group   => 'neobank',
      mode    => '0600',
      content => template('neobank/service.env.erb'),
      notify  => Docker::Run[$service_name],
    }
    
    # Pull and run Docker container
    docker::run { $service_name:
      image            => "${docker_registry}/${service_name}:${docker_tag}",
      ports            => ["${port}:${port}"],
      env_file         => "/opt/neobank/config/${service_name}.env",
      net              => 'neobank_network',
      restart_service  => true,
      health_check_cmd => "wget --spider -q http://localhost:${port}/health || exit 1",
      require          => [
        File['/opt/neobank/config'],
        Docker_network['neobank_network'],
      ],
    }
  }
  
  # Health check exec
  exec { 'verify_backend_services':
    command     => '/opt/neobank/scripts/health_check.sh',
    refreshonly => true,
    subscribe   => Docker::Run[$managed_services.keys],
  }
}
