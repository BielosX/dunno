data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_security_group" "group" {
  vpc_id = aws_vpc.vpc.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
  }
}

resource "aws_iam_role" "role" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "policy_attachment" {
  policy_arn = "arn:aws:iam::aws:policy/AWSLambdaManagedEC2ResourceOperator"
  role       = aws_iam_role.role.name
}

locals {
  providers = {
    "private-subnets" = aws_subnet.private[*].id
    "public-subnets"  = aws_subnet.public[*].id
  }
}

resource "aws_lambda_capacity_provider" "providers" {
  for_each = local.providers
  name     = each.key
  capacity_provider_scaling_config {
    scaling_mode   = "Manual"
    max_vcpu_count = 20
    scaling_policies = [{
      predefined_metric_type = "LambdaCapacityProviderAverageCPUUtilization"
      target_value           = 70.0
    }]
  }
  instance_requirements {
    architectures          = ["arm64"]
    allowed_instance_types = ["m7g.large"]
  }
  vpc_config {
    security_group_ids = [aws_security_group.group.id]
    subnet_ids         = each.value
  }
  permissions_config {
    capacity_provider_operator_role_arn = aws_iam_role.role.arn
  }
}

locals {
  arm_name_prefix = "/dunno/CapacityProviders/arm64"
}

resource "aws_ssm_parameter" "private_provider" {
  name  = "${local.arm_name_prefix}/LambdaPrivateSubnetsCapacityProviderArn"
  type  = "String"
  value = aws_lambda_capacity_provider.providers["private-subnets"].arn
}

resource "aws_ssm_parameter" "public_provider" {
  name  = "${local.arm_name_prefix}/LambdaPublicSubnetsCapacityProviderArn"
  type  = "String"
  value = aws_lambda_capacity_provider.providers["public-subnets"].arn
}
