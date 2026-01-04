resource "aws_apigatewayv2_api" "api" {
  name          = "kotlin-connection-pool"
  protocol_type = "HTTP"
}

data "aws_iam_policy_document" "api_gw_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["apigateway.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "api_gw_role" {
  assume_role_policy = data.aws_iam_policy_document.api_gw_assume_role.json
}

resource "aws_iam_role_policy_attachment" "api_gw_role_policy" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"
  role       = aws_iam_role.api_gw_role.id
}

resource "aws_api_gateway_account" "account" {
  cloudwatch_role_arn = aws_iam_role.api_gw_role.arn
}

resource "aws_cloudwatch_log_group" "api_log_group" {
  name = "/api-gateway/${aws_apigatewayv2_api.api.name}-access"
}

resource "aws_apigatewayv2_stage" "prod" {
  api_id      = aws_apigatewayv2_api.api.id
  auto_deploy = true
  name        = "$default"
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_log_group.arn
    format          = replace(file("${path.module}/apigw_log_format.json"), "\n", "")
  }
}

resource "aws_apigatewayv2_integration" "books" {
  api_id             = aws_apigatewayv2_api.api.id
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  integration_uri    = aws_lambda_alias.latest.invoke_arn
}

resource "aws_apigatewayv2_route" "route" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.books.id}"
}