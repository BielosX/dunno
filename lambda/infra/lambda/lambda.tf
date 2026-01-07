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

resource "aws_iam_role_policy_attachment" "attachment" {
  role       = aws_iam_role.role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

locals {
  runtimes = {
    "golang" = "provided.al2023"
    "rust"   = "provided.al2023"
    "python" = "python3.14"
    "java"   = "java25"
    "ruby"   = "ruby3.4"
    "js"     = "nodejs24.x"
    "dotnet" = "dotnet8" // dotnet9 container only
  }
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