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

resource "aws_iam_role" "api_lambda_role" {
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "api_lambda_policy" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem"
    ]
    resources = [aws_dynamodb_table.books.arn]
  }
}

resource "aws_iam_role_policy" "api_lambda_role_policy" {
  policy = data.aws_iam_policy_document.api_lambda_policy.json
  role   = aws_iam_role.api_lambda_role.id
}

locals {
  name_prefix = "dunno-"
  common_policies = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
    "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
  ]
  functions = {
    api = {
      handler = "apiGatewayHandler"
      role    = aws_iam_role.api_lambda_role.arn
      env_vars = {
        BOOKS_TABLE_ARN = aws_dynamodb_table.books.arn
        LOG_LEVEL       = "info"
      }
    }
  }
  roles_common_policies = [
    for pair in setproduct(local.common_policies, values(local.functions)) : [split("/", pair[1].role)[1], pair[0]]
  ]
}

resource "aws_iam_role_policy_attachment" "common" {
  count      = length(local.common_policies) * length((keys(local.functions)))
  policy_arn = local.roles_common_policies[count.index][1]
  role       = local.roles_common_policies[count.index][0]
}

resource "aws_lambda_function" "lambda" {
  for_each         = local.functions
  function_name    = "${local.name_prefix}${each.key}"
  handler          = each.value["handler"]
  role             = each.value["role"]
  runtime          = "provided.al2023"
  timeout          = 60
  architectures    = ["arm64"]
  filename         = var.bundle_path
  source_code_hash = filebase64sha256(var.bundle_path)
  tracing_config {
    mode = "Active"
  }
  environment {
    variables = lookup(each.value, "env_vars", {})
  }
}

resource "aws_lambda_permission" "apigw_permission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda["api"].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*"
}