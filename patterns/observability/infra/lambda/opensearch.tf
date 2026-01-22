locals {
  opensearch_domain = "dunno"
}

data "aws_iam_policy_document" "opensearch_access_policy" {
  statement {
    effect  = "Allow"
    actions = ["es:*"]
    principals {
      identifiers = ["arn:aws:iam::${local.account_id}:root"]
      type        = "AWS"
    }
    resources = ["arn:aws:es:${local.region}:${local.account_id}:domain/${local.opensearch_domain}/*"]
  }
}

resource "aws_opensearch_domain" "opensearch" {
  domain_name    = local.opensearch_domain
  engine_version = "OpenSearch_3.3"
  cluster_config {
    instance_count = 1
    instance_type  = "t3.small.search"
  }
  domain_endpoint_options {
    enforce_https = true
  }
  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
  access_policies = data.aws_iam_policy_document.opensearch_access_policy.json
}
