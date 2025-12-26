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
    effect = "Allow"
    actions = [
      "kinesis:GetRecords",
      "kinesis:GetShardIterator",
      "kinesis:DescribeStream",
      "kinesis:DescribeStreamSummary",
      "kinesis:ListShards"
    ]
    resources = [aws_kinesis_stream.stream.arn]
  }
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:PutItem",
      "dynamodb:DeleteItem"
    ]
    resources = [aws_dynamodb_table.users.arn]
  }
  statement {
    effect    = "Allow"
    actions   = ["cloudwatch:PutMetricData"]
    resources = ["*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["sqs:SendMessage"]
    resources = [aws_sqs_queue.dlq.arn]
  }
}

resource "aws_iam_role_policy" "policy" {
  policy = data.aws_iam_policy_document.policy.json
  role   = aws_iam_role.role.id
}

resource "aws_lambda_function" "function" {
  function_name    = "node-esbuild"
  role             = aws_iam_role.role.arn
  runtime          = "nodejs24.x"
  timeout          = 60
  memory_size      = 512
  architectures    = ["arm64"]
  handler          = "index.handler"
  filename         = var.bundle_path
  source_code_hash = filebase64sha256(var.bundle_path)
  environment {
    variables = {
      CONCURRENCY_LIMIT    = 40
      FAILED_RECORDS_QUEUE = aws_sqs_queue.dlq.url
      USERS_TABLE : aws_dynamodb_table.users.arn
      LOG_LEVEL : "info"
    }
  }
}

resource "aws_lambda_event_source_mapping" "source_mapping" {
  event_source_arn                   = aws_kinesis_stream.stream.arn
  function_name                      = aws_lambda_function.function.arn
  starting_position                  = "LATEST"
  batch_size                         = 100
  enabled                            = true
  maximum_batching_window_in_seconds = 10
  parallelization_factor             = 1
  bisect_batch_on_function_error     = true
}