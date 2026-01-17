data "aws_iam_policy_document" "lambda_assume_role" {
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
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

locals {
  common_policies = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
    "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
  ]
}

resource "aws_iam_role_policy_attachment" "common" {
  for_each   = toset(local.common_policies)
  policy_arn = each.value
  role       = aws_iam_role.role.name
}

resource "aws_lambda_function" "lambda" {
  function_name    = var.name
  handler          = var.handler
  role             = aws_iam_role.role.arn
  runtime          = "provided.al2023"
  timeout          = 60
  memory_size      = 1769 // 1vCPU
  architectures    = ["arm64"]
  filename         = var.bundle_path
  source_code_hash = filebase64sha256(var.bundle_path)
  tracing_config {
    mode = "Active"
  }
  environment {
    variables = var.env_vars
  }
}