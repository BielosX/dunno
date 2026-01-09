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

resource "aws_sqs_queue" "queue" {
  name_prefix                = "pc-scaling-queue-"
  visibility_timeout_seconds = 120
}

data "aws_iam_policy_document" "role_policy" {
  statement {
    effect = "Allow"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:ChangeMessageVisibility"
    ]
    resources = [aws_sqs_queue.queue.arn]
  }
}

resource "aws_iam_role" "role" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "basic_execution_attachment" {
  role       = aws_iam_role.role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "role_policy" {
  policy = data.aws_iam_policy_document.role_policy.json
  role   = aws_iam_role.role.id
}

resource "aws_lambda_function" "function" {
  function_name    = "pc_scaling"
  role             = aws_iam_role.role.arn
  runtime          = "provided.al2023"
  handler          = "dummy"
  timeout          = 20
  memory_size      = 512
  publish          = true
  filename         = var.bundle_path
  architectures    = ["arm64"]
  source_code_hash = filebase64sha256(var.bundle_path)
}

resource "aws_lambda_alias" "latest" {
  function_name    = aws_lambda_function.function.function_name
  function_version = aws_lambda_function.function.version
  name             = "latest"
}

resource "aws_lambda_event_source_mapping" "mapping" {
  event_source_arn = aws_sqs_queue.queue.arn
  function_name    = aws_lambda_alias.latest.arn
  batch_size       = 10
  scaling_config {
    maximum_concurrency = 10
  }
}

resource "aws_lambda_provisioned_concurrency_config" "concurrency" {
  function_name                     = aws_lambda_alias.latest.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_alias.latest.name
  lifecycle {
    ignore_changes = [provisioned_concurrent_executions]
  }
}

resource "aws_appautoscaling_target" "lambda_pc" {
  depends_on         = [aws_lambda_provisioned_concurrency_config.concurrency]
  service_namespace  = "lambda"
  scalable_dimension = "lambda:function:ProvisionedConcurrency"
  resource_id        = "function:${aws_lambda_alias.latest.function_name}:${aws_lambda_alias.latest.name}"
  min_capacity       = 0
  max_capacity       = 10
}

resource "aws_appautoscaling_policy" "lambda_pc_target_tracking" {
  name               = "lambda-pc-utilization"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = "lambda"
  scalable_dimension = "lambda:function:ProvisionedConcurrency"
  resource_id        = aws_appautoscaling_target.lambda_pc.resource_id

  target_tracking_scaling_policy_configuration {
    target_value = 0.5

    predefined_metric_specification {
      predefined_metric_type = "LambdaProvisionedConcurrencyUtilization"
    }

    scale_in_cooldown  = 60
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_scheduled_action" "lambda_pc_schedule_scale_up" {
  name               = "lambda-pc-schedule"
  service_namespace  = aws_appautoscaling_target.lambda_pc.service_namespace
  resource_id        = aws_appautoscaling_target.lambda_pc.resource_id
  scalable_dimension = aws_appautoscaling_target.lambda_pc.scalable_dimension
  schedule           = "cron(0 10 * * ? *)"

  scalable_target_action {
    min_capacity = 2
    max_capacity = 10
  }
}

resource "aws_appautoscaling_scheduled_action" "lambda_pc_schedule_scale_down" {
  name               = "lambda-pc-scale-down"
  service_namespace  = aws_appautoscaling_target.lambda_pc.service_namespace
  resource_id        = aws_appautoscaling_target.lambda_pc.resource_id
  scalable_dimension = aws_appautoscaling_target.lambda_pc.scalable_dimension
  schedule           = "cron(0 18 * * ? *)"

  scalable_target_action {
    min_capacity = 0
    max_capacity = 0
  }
}