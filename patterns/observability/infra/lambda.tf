locals {
  prefix = "dunno"
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

module "api_lambda" {
  source      = "./go_lambda"
  bundle_path = var.bundle_path
  handler     = "apiGatewayHandler"
  name        = "${local.prefix}-api"
  env_vars = {
    BOOKS_TABLE_ARN = aws_dynamodb_table.books.arn
    LOG_LEVEL       = "info"
  }
}

resource "aws_iam_role_policy" "api_lambda_policy" {
  policy = data.aws_iam_policy_document.api_lambda_policy.json
  role   = module.api_lambda.role_name
}

resource "aws_lambda_permission" "apigw_permission" {
  action        = "lambda:InvokeFunction"
  function_name = module.api_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*"
}