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

data "aws_ssm_parameter" "capacity_provider" {
  name = "/dunno/CapacityProviders/arm64/LambdaPrivateSubnetsCapacityProviderArn"
}

resource "aws_lambda_function" "lambda" {
  function_name    = "kotlin-connection-pool"
  role             = aws_iam_role.role.arn
  runtime          = "java25"
  timeout          = 60
  handler          = "org.dunno.Handler::handleRequest"
  filename         = var.bundle_path
  memory_size      = 2048 // 2048 minimum for Managed Instances
  publish          = true // Published version required for Managed Instances
  source_code_hash = filebase64sha256(var.bundle_path)
  architectures    = ["arm64"]

  capacity_provider_config {
    lambda_managed_instances_capacity_provider_config {
      capacity_provider_arn = data.aws_ssm_parameter.capacity_provider.value
    }
  }
}

resource "aws_lambda_alias" "latest" {
  function_name    = aws_lambda_function.lambda.function_name
  function_version = aws_lambda_function.lambda.version
  name             = "latest"
}

resource "aws_lambda_permission" "apigw_permission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*"
  qualifier     = aws_lambda_alias.latest.name
}