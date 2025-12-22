data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.region
}

module "source_bucket" {
  source      = "./private_bucket"
  bucket_name = "source-${local.account_id}-${local.region}"
}

module "target_bucket" {
  source      = "./private_bucket"
  bucket_name = "target-${local.account_id}-${local.region}"
}