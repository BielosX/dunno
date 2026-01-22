terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.28.0"
    }
  }
  required_version = ">=1.11.2"
}

provider "aws" {}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}


locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.region
}