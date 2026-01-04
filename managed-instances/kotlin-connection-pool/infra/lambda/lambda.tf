data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

locals {
  region        = data.aws_region.current.region
  account_id    = data.aws_caller_identity.current.account_id
  rds_user_name = "lambda_user"
}

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

resource "aws_iam_role_policy_attachment" "basic_execution_attachment" {
  role       = aws_iam_role.role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_ssm_parameter" "config" {
  name = "KotlinConnectionPoolConfig"
  type = "String"
  value = jsonencode({
    db = {
      username = local.rds_user_name
      host     = local.cluster_endpoint
      port     = local.db_port
      name     = local.db_name
    }
  })
}

data "aws_iam_policy_document" "lambda_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["rds-db:connect"]
    resources = [
      "arn:aws:rds-db:${local.region}:${local.account_id}:dbuser:${local.cluster_resource_id}/${local.rds_user_name}"
    ]
  }
  statement {
    effect    = "Allow"
    actions   = ["ssm:GetParameter"]
    resources = [aws_ssm_parameter.config.arn]
  }
}

resource "aws_iam_role_policy" "lambda_role_policy" {
  policy = data.aws_iam_policy_document.lambda_role_policy.json
  role   = aws_iam_role.role.id
}

resource "aws_lambda_function" "lambda" {
  depends_on       = [aws_ssm_parameter.config]
  function_name    = "kotlin-connection-pool"
  role             = aws_iam_role.role.arn
  runtime          = "java25"
  timeout          = 60
  handler          = "org.dunno.Handler::handleRequest"
  filename         = var.bundle_path
  memory_size      = 2048 // 2048 minimum for Managed Instances
  publish          = true // Published version required for Managed Instances
  source_code_hash = filebase64sha256(var.bundle_path)
  architectures    = ["arm64"]

  capacity_provider_config {
    lambda_managed_instances_capacity_provider_config {
      capacity_provider_arn = local.private_capacity_provider_arn
    }
  }
}

resource "aws_lambda_alias" "latest" {
  function_name    = aws_lambda_function.lambda.function_name
  function_version = aws_lambda_function.lambda.version
  name             = "latest"
}

resource "aws_lambda_permission" "apigw_permission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*"
  qualifier     = aws_lambda_alias.latest.name
}