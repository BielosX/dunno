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

data "aws_iam_policy_document" "policy" {
  statement {
    effect    = "Allow"
    actions   = ["dynamodb:PutItem", "dynamodb:Scan"]
    resources = [aws_dynamodb_table.table.arn]
  }
}

resource "aws_iam_role_policy" "policy" {
  policy = data.aws_iam_policy_document.policy.json
  role   = aws_iam_role.role.id
}

resource "aws_lambda_function" "lambda" {
  function_name    = "jvm-spring-lambda"
  role             = aws_iam_role.role.arn
  runtime          = "java25"
  handler          = "org.dunno.Handler::handleRequest"
  timeout          = 60
  memory_size      = 512
  filename         = var.bundle_path
  source_code_hash = filebase64sha256(var.bundle_path)
  environment {
    variables = {
      AWS_DYNAMODB_TABLE_BOOKS = aws_dynamodb_table.table.name
    }
  }
}

resource "aws_lambda_permission" "apigw_permission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*"
}
