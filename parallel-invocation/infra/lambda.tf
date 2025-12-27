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

resource "aws_lambda_function" "lambda" {
  function_name                  = "parallel-invocation"
  role                           = aws_iam_role.role.arn
  reserved_concurrent_executions = 5
  runtime                        = "nodejs24.x"
  handler                        = "index.handler"
  timeout                        = 60
  memory_size                    = 128
  filename                       = var.bundle_path
  source_code_hash               = filebase64sha256(var.bundle_path)
}