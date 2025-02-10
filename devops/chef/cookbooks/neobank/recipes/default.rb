#
# Cookbook:: neobank
# Recipe:: default
#
# Deploys the complete NeoBank application stack

include_recipe 'docker'

# Get secrets from encrypted data bag
secrets = data_bag_item('neobank', 'secrets')

# Create application user
user node['neobank']['user'] do
  group node['neobank']['group']
  system true
  shell '/bin/bash'
  home node['neobank']['install_dir']
end

# Create necessary directories
[
  node['neobank']['install_dir'],
  "#{node['neobank']['install_dir']}/config",
  node['neobank']['log_dir']
].each do |dir|
  directory dir do
    owner node['neobank']['user']
    group node['neobank']['group']
    mode '0755'
    recursive true
  end
end

# Create Docker network
docker_network 'neobank_network' do
  action :create
end

# Include component recipes
include_recipe 'neobank::database' if node['roles'].include?('neobank-db')
include_recipe 'neobank::redis' if node['roles'].include?('neobank-redis')
include_recipe 'neobank::backend' if node['roles'].include?('neobank-backend')
include_recipe 'neobank::frontend' if node['roles'].include?('neobank-frontend')
