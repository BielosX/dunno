locals {
  prefix = "dunno"
}

data "aws_iam_policy_document" "api_lambda_policy" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:BatchGetItem",
    ]
    resources = [aws_dynamodb_table.books.arn]
  }
  statement {
    effect = "Allow"
    actions = [
      "es:ESHttp*",
    ]
    resources = ["${aws_opensearch_domain.opensearch.arn}/*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData"
    ]
    resources = ["*"]
  }
}

module "api_lambda" {
  source      = "./go_lambda"
  bundle_path = var.bundle_path
  handler     = "apiGatewayHandler"
  name        = "${local.prefix}-api"
  env_vars = {
    BOOKS_TABLE_ARN     = aws_dynamodb_table.books.arn
    OPEN_SEARCH_ADDRESS = "https://${aws_opensearch_domain.opensearch.endpoint}"
    LOG_LEVEL           = "info"
    METRICS_NAMESPACE   = local.prefix
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

module "streams_lambda" {
  source      = "./go_lambda"
  bundle_path = var.bundle_path
  handler     = "dynamoDbStreamsHandler"
  name        = "${local.prefix}-dynamodb-streams"
  env_vars = {
    LOG_LEVEL           = "info"
    OPEN_SEARCH_ADDRESS = "https://${aws_opensearch_domain.opensearch.endpoint}"
  }
}

data "aws_iam_policy_document" "streams_lambda_policy" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:DescribeStream",
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
      "dynamodb:ListStreams"
    ]
    resources = ["${aws_dynamodb_table.books.arn}/stream/*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "es:ESHttp*"
    ]
    resources = ["${aws_opensearch_domain.opensearch.arn}/*"]
  }
}

resource "aws_iam_role_policy" "streams_lambda_policy" {
  policy = data.aws_iam_policy_document.streams_lambda_policy.json
  role   = module.streams_lambda.role_name
}

resource "aws_lambda_event_source_mapping" "streams_lambda_mapping" {
  function_name     = module.streams_lambda.function_name
  event_source_arn  = aws_dynamodb_table.books.stream_arn
  starting_position = "TRIM_HORIZON"
  batch_size        = 10
}
