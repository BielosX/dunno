locals {
  handlers = {
    "GetMovieById" : 30
    "SaveMovie" : 30
    "ListMovies" : 30
  }
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

data "aws_iam_policy_document" "dynamodb" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:Scan"
    ]
    resources = [aws_dynamodb_table.movies.arn]
  }
}

resource "aws_iam_role_policy" "dynamodb" {
  policy = data.aws_iam_policy_document.dynamodb.json
  role   = aws_iam_role.role.id
}

resource "aws_lambda_function" "function" {
  for_each         = local.handlers
  function_name    = "movies-${each.key}"
  role             = aws_iam_role.role.arn
  runtime          = "provided.al2023"
  timeout          = each.value
  memory_size      = 512
  handler          = each.key
  architectures    = ["arm64"]
  filename         = var.zip_path
  source_code_hash = filesha256(var.zip_path)
  environment {
    variables = {
      MOVIES_TABLE_ARN : aws_dynamodb_table.movies.arn
    }
  }
}

resource "aws_lambda_permission" "apigw_permission" {
  for_each      = local.handlers
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.function[each.key].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*"
}