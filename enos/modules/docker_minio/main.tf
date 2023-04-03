# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.1"
    }

    enos = {
      source = "app.terraform.io/hashicorp-qti/enos"
    }
  }
}

variable "image_name_server" {
  description = "Name of Docker Image for minio server"
  type        = string
  default     = "docker.mirror.hashicorp.services/minio/minio:latest"
}
variable "image_name_client" {
  description = "Name of Docker Image for minio client"
  type        = string
  default     = "docker.mirror.hashicorp.services/minio/mc:latest"
}
variable "network_name" {
  description = "Name of Docker Network"
  type        = string
}
variable "container_name" {
  description = "Name of Docker Container"
  type        = string
  default     = "minio"
}
variable "region" {
  description = "AWS Region"
  type        = string
  default     = "us-east-1"
}
variable "bucket_name" {
  description = "Name of storage bucket"
  type        = string
  default     = "testbucket" # this needs to match the bucket in policy.json
}
variable "root_user" {
  description = "Username for minio root user"
  type        = string
  default     = "minio"
}
variable "root_password" {
  description = "Password for minio root user"
  type        = string
  default     = "minioadmin"
}
variable "access_key_id" {
  description = "Username/Access Key Id for user that can access bucket"
  type        = string
  default     = "testuser"
}
variable "secret_access_key" {
  description = "Password/Secret Access Key for user that can access bucket"
  type        = string
  default     = "password"
}

resource "docker_image" "minio_server" {
  name         = var.image_name_server
  keep_locally = true
}

resource "docker_container" "minio_server" {
  depends_on = [
    docker_image.minio_server
  ]
  image   = docker_image.minio_server.image_id
  name    = var.container_name
  command = ["minio", "server", "/data", "--console-address", ":9090"]
  env = [
    "MINIO_ROOT_USER=minio",
    "MINIO_ROOT_PASSWORD=minioadmin",
    "MINIO_REGION=${var.region}"
  ]
  ports {
    internal = 9000
    external = 9000
  }
  ports {
    internal = 9090
    external = 9090
  }
  networks_advanced {
    name = var.network_name
  }
}

# does this need to be in a script?
resource "docker_image" "minio_client" {
  name         = var.image_name_client
  keep_locally = true
}

resource "enos_local_exec" "init_minio" {
  depends_on = [
    docker_image.minio_client
  ]
  environment = {
    MINIO_SERVER_CONTAINER_NAME = var.container_name,
    MINIO_CLIENT_IMAGE          = var.image_name_client,
    MINIO_BUCKET_NAME           = var.bucket_name,
    MINIO_ROOT_USER             = var.root_user,
    MINIO_ROOT_PASSWORD         = var.root_password,
    MINIO_ACCESS_KEY_ID         = var.access_key_id,
    MINIO_SECRET_ACCESS_KEY     = var.secret_access_key,
    TEST_NETWORK_NAME           = var.network_name,

  }
  inline = ["bash ./${path.module}/init.sh \"${var.image_name_client}\""]
}

output "bucket_name" {
  value = var.bucket_name
}

output "access_key_id" {
  value = var.access_key_id
}

output "secret_access_key" {
  value = var.secret_access_key
}
