resource "aws_apigatewayv2_api" "api" {
  name          = "http-single-binary"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "prod" {
  api_id      = aws_apigatewayv2_api.api.id
  auto_deploy = true
  name        = "$default"
}

resource "aws_apigatewayv2_integration" "integration" {
  for_each           = local.handlers
  api_id             = aws_apigatewayv2_api.api.id
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.function[each.key].invoke_arn
}

resource "aws_apigatewayv2_route" "get-movie-by-id" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "GET /movies/{movieId}"
  target    = "integrations/${aws_apigatewayv2_integration.integration["GetMovieById"].id}"
}

resource "aws_apigatewayv2_route" "save-movie" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "POST /movies"
  target    = "integrations/${aws_apigatewayv2_integration.integration["SaveMovie"].id}"
}

resource "aws_apigatewayv2_route" "list-movies" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "GET /movies"
  target    = "integrations/${aws_apigatewayv2_integration.integration["ListMovies"].id}"
}
