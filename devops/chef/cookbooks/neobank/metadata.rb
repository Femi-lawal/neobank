name 'neobank'
maintainer 'NeoBank Team'
maintainer_email 'devops@neobank.com'
license 'MIT'
description 'Deploys NeoBank banking application'
version '1.0.0'
chef_version '>= 16.0'

depends 'docker', '~> 7.0'
depends 'postgresql', '~> 11.0'

supports 'ubuntu', '>= 20.04'
supports 'centos', '>= 8.0'

issues_url 'https://github.com/neobank/chef-neobank/issues'
source_url 'https://github.com/neobank/chef-neobank'
