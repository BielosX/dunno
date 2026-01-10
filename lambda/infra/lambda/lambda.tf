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

resource "aws_iam_role" "role" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

locals {
  managed_policies = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
    "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  ]
}

resource "aws_iam_role_policy_attachment" "attachment" {
  for_each   = toset(local.managed_policies)
  role       = aws_iam_role.role.id
  policy_arn = each.value
}

locals {
  runtimes = {
    "quarkus" = "provided.al2023"
    "golang"  = "provided.al2023"
    "rust"    = "provided.al2023"
    "python"  = "python3.14"
    "java"    = "java25"
    "ruby"    = "ruby3.4"
    "js"      = "nodejs24.x"
    "dotnet"  = "dotnet8" // dotnet9 container only
  }
}

module "vpc" {
  count  = var.vpc ? 1 : 0
  source = "./vpc"
}

resource "aws_lambda_function" "lambda" {
  count         = var.lambda_count
  function_name = "${var.language}-${count.index}"
  role          = aws_iam_role.role.arn
  handler       = var.handler
  package_type  = "Zip"
  memory_size   = 1769 // 1vCPU
  timeout       = 30
  runtime       = local.runtimes[var.language]
  s3_bucket     = var.bucket
  s3_key        = var.bucket_key
  architectures = [var.architecture]
  publish       = true # Required for SnapStart
  dynamic "vpc_config" {
    for_each = var.vpc ? [1] : []
    content {
      security_group_ids = [module.vpc[0].lambda_sg_id]
      subnet_ids         = [module.vpc[0].private_subnet_id]
    }
  }
  tags = {
    language = var.language
  }
  dynamic "snap_start" {
    for_each = var.snap_start ? [1] : []
    content {
      apply_on = "PublishedVersions"
    }
  }
}

resource "aws_lambda_alias" "alias" {
  count            = var.snap_start ? var.lambda_count : 0
  function_name    = aws_lambda_function.lambda[count.index].function_name
  function_version = aws_lambda_function.lambda[count.index].version
  name             = "latest"
}