terraform {
  required_providers {
    opensearch = {
      source  = "opensearch-project/opensearch"
      version = "2.3.2"
    }
  }
  required_version = ">=1.11.2"
}

data "terraform_remote_state" "lambda" {
  backend = "local"
  config = {
    path = "${path.module}/../lambda/terraform.tfstate"
  }
}

locals {
  endpoint = data.terraform_remote_state.lambda.outputs.opensearch_endpoint
}

provider "opensearch" {
  url = "https://${local.endpoint}"
}

resource "opensearch_index" "books" {
  name          = "books"
  force_destroy = true
  mappings = jsonencode({
    properties = {
      id      = { type = "keyword" }
      title   = { type = "text" }
      isbn    = { type = "text" }
      authors = { type = "text" }
    }
  })
}