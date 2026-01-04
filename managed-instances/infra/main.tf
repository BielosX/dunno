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