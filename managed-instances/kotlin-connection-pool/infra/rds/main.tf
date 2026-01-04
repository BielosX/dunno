terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.27.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
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

locals {
  vpc_id             = data.terraform_remote_state.main.outputs.vpc_id
  vpc_cidr           = data.terraform_remote_state.main.outputs.vpc_cidr
  public_subnet_ids  = data.terraform_remote_state.main.outputs.public_subnet_ids
  private_subnet_ids = data.terraform_remote_state.main.outputs.private_subnet_ids
}