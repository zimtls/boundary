# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# For this scenario to work, add the following line to /etc/hosts
# 127.0.0.1 localhost boundary

scenario "minio" {
  terraform_cli = terraform_cli.default
  terraform     = terraform.default
  providers = [
    provider.aws.default,
    provider.enos.default
  ]

  step "create_docker_network" {
    module = module.docker_network
  }

  step "create_minio" {
    depends_on = [
      step.create_docker_network
    ]
    variables {
      network_name = step.create_docker_network.network_name
    }
    module = module.docker_minio
  }

  output "test_results" {
    value = "1"
  }
}
