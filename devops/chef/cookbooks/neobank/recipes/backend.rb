#
# Cookbook:: neobank
# Recipe:: backend
#
# Deploys NeoBank backend microservices

# Get secrets from encrypted data bag
secrets = data_bag_item('neobank', 'secrets')

# Common environment variables
common_env = {
  'DB_HOST' => node['neobank']['database']['host'],
  'DB_PORT' => node['neobank']['database']['port'].to_s,
  'DB_USER' => node['neobank']['database']['user'],
  'DB_PASSWORD' => secrets['db_password'],
  'DB_NAME' => node['neobank']['database']['name'],
  'JWT_SECRET' => secrets['jwt_secret'],
  'REDIS_ADDR' => "#{node['neobank']['redis']['host']}:#{node['neobank']['redis']['port']}",
  'KAFKA_BROKERS' => node['neobank']['kafka']['brokers'].join(',')
}

# Deploy each backend service
node['neobank']['services'].each do |service_name, config|
  next if service_name == 'frontend'
  
  # Create environment file
  template "#{node['neobank']['install_dir']}/config/#{service_name}.env" do
    source 'service.env.erb'
    owner node['neobank']['user']
    group node['neobank']['group']
    mode '0600'
    variables(
      port: config['port'],
      env_vars: common_env
    )
    notifies :restart, "docker_container[#{service_name}]", :delayed
  end
  
  # Pull the image
  docker_image service_name do
    repo "#{node['neobank']['docker_registry']}/#{service_name}"
    tag node['neobank']['docker_tag']
    action :pull
    notifies :redeploy, "docker_container[#{service_name}]", :immediately
  end
  
  # Deploy the container
  docker_container service_name do
    repo "#{node['neobank']['docker_registry']}/#{service_name}"
    tag node['neobank']['docker_tag']
    port "#{config['port']}:#{config['port']}"
    env_file "#{node['neobank']['install_dir']}/config/#{service_name}.env"
    network_mode 'neobank_network'
    restart_policy 'unless-stopped'
    health_check(
      test: ['CMD', 'wget', '--spider', '-q', "http://localhost:#{config['port']}/health"],
      interval: 30,
      timeout: 10,
      retries: 3
    )
    action :run
  end
end

# Verify all services are healthy
ruby_block 'verify_backend_health' do
  block do
    require 'net/http'
    
    node['neobank']['services'].each do |service_name, config|
      next if service_name == 'frontend'
      
      uri = URI("http://localhost:#{config['port']}/health")
      retries = 10
      
      retries.times do |i|
        begin
          response = Net::HTTP.get_response(uri)
          break if response.code == '200'
        rescue => e
          Chef::Log.warn("#{service_name} not ready yet: #{e.message}")
        end
        sleep 5
      end
    end
  end
  action :run
end
