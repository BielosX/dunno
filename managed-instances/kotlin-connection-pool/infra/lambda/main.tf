terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.27.0"
    }
  }
  required_version = ">=1.11.2"
}

provider "aws" {}

data "terraform_remote_state" "main" {
  backend = "local"
  config = {
    path = "${path.module}/../../../infra/terraform.tfstate"
  }
}

data "terraform_remote_state" "rds" {
  backend = "local"
  config = {
    path = "${path.module}/../rds/terraform.tfstate"
  }
}

locals {
  cluster_resource_id           = data.terraform_remote_state.rds.outputs.cluster_resource_id
  db_name                       = data.terraform_remote_state.rds.outputs.db_name
  db_port                       = data.terraform_remote_state.rds.outputs.db_port
  cluster_endpoint              = data.terraform_remote_state.rds.outputs.cluster_endpoint
  private_capacity_provider_arn = data.terraform_remote_state.main.outputs.private_capacity_provider_arn
}