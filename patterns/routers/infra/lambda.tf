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

data "aws_iam_policy_document" "dynamodb" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:Scan"
    ]
    resources = [aws_dynamodb_table.table.arn]
  }
}

resource "aws_iam_role_policy" "dynamodb" {
  policy = data.aws_iam_policy_document.dynamodb.json
  role   = aws_iam_role.role.id
}

resource "aws_iam_role_policy_attachment" "attachment" {
  role       = aws_iam_role.role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

locals {
  handlers = {
    "gorilla-mux" = {
      filename = var.gorilla_bundle_path
      runtime  = "provided.al2023"
      handler  = "dummy"
    }
  }
}

resource "aws_lambda_function" "lambda" {
  for_each         = local.handlers
  function_name    = each.key
  role             = aws_iam_role.role.arn
  handler          = each.value["handler"]
  runtime          = each.value["runtime"]
  timeout          = 30
  memory_size      = 512
  package_type     = "Zip"
  architectures    = ["arm64"]
  filename         = each.value["filename"]
  source_code_hash = filebase64sha256(each.value["filename"])
  environment {
    variables = {
      BOOKS_TABLE_ARN = aws_dynamodb_table.table.arn
    }
  }
}

module "apigw" {
  for_each          = local.handlers
  source            = "./apigw"
  lambda_invoke_arn = aws_lambda_function.lambda[each.key].invoke_arn
  name              = "http-${each.key}"
}

resource "aws_lambda_permission" "apigw_permission" {
  for_each      = local.handlers
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda[each.key].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${module.apigw[each.key].execution_arn}/*"
}