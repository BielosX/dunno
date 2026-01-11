variable "name" {
  type = string
}

variable "lambda_invoke_arn" {
  type = string
}

resource "aws_apigatewayv2_api" "api" {
  name          = var.name
  protocol_type = "HTTP"
}

output "execution_arn" {
  value = aws_apigatewayv2_api.api.execution_arn
}

resource "aws_apigatewayv2_stage" "prod" {
  api_id      = aws_apigatewayv2_api.api.id
  auto_deploy = true
  name        = "$default"
}

resource "aws_apigatewayv2_integration" "books" {
  api_id             = aws_apigatewayv2_api.api.id
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  integration_uri    = var.lambda_invoke_arn
}

resource "aws_apigatewayv2_route" "books" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "ANY /books/{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.books.id}"
}

resource "aws_apigatewayv2_route" "books_root" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "ANY /books"
  target    = "integrations/${aws_apigatewayv2_integration.books.id}"
}

