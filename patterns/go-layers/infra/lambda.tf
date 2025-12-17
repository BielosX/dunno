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

data "aws_iam_policy_document" "s3" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*"
    ]
  }
}

resource "aws_iam_role_policy" "s3" {
  policy = data.aws_iam_policy_document.s3.json
  role   = aws_iam_role.role.id
}

resource "aws_iam_role_policy_attachment" "attachment" {
  role       = aws_iam_role.role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_layer_version" "plugin" {
  layer_name       = "plugin"
  filename         = var.plugin_zip
  source_code_hash = filebase64sha256(var.plugin_zip)
}

resource "aws_lambda_function" "lambda" {
  function_name    = "go-layers"
  role             = aws_iam_role.role.arn
  handler          = "handler"
  runtime          = "provided.al2023"
  timeout          = 30
  memory_size      = 512
  package_type     = "Zip"
  architectures    = ["arm64"]
  filename         = var.lambda_zip
  source_code_hash = filebase64sha256(var.lambda_zip)
  layers           = [aws_lambda_layer_version.plugin.arn]
  environment {
    variables = {
      PLUGIN_PATH = "/opt/plugin-bin"
      BUCKET_NAME = aws_s3_bucket.bucket.id
    }
  }
}

resource "aws_lambda_permission" "apigw_permission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*"
}