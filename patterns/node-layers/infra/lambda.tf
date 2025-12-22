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

resource "aws_iam_role_policy_attachment" "attachment" {
  role       = aws_iam_role.role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect  = "Allow"
    actions = ["s3:GetObject", "s3:ListBucket"]
    resources = [
      module.source_bucket.bucket_arn,
      "${module.source_bucket.bucket_arn}/*"
    ]
  }
  statement {
    effect  = "Allow"
    actions = ["s3:GetObject", "s3:ListBucket", "s3:PutObject"]
    resources = [
      module.target_bucket.bucket_arn,
      "${module.target_bucket.bucket_arn}/*"
    ]
  }
  statement {
    effect    = "Allow"
    actions   = ["lambda:GetFunctionConfiguration"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "policy" {
  policy = data.aws_iam_policy_document.policy.json
  role   = aws_iam_role.role.id
}

locals {
  lambdas = {
    "archiver" : {
      "file_name" : var.archiver_zip
      "handler" : "index.archive"
      "prefix" : "files/"
    }
    "resizer" : {
      "file_name" : var.resizer_zip
      "handler" : "index.resize"
      "prefix" : "images/"
    }
  }
}

resource "aws_lambda_function" "lambdas" {
  for_each         = local.lambdas
  function_name    = each.key
  role             = aws_iam_role.role.arn
  runtime          = "nodejs24.x"
  timeout          = 30
  memory_size      = 512
  handler          = each.value["handler"]
  filename         = each.value["file_name"]
  source_code_hash = filebase64sha256(each.value["file_name"])
  layers           = [var.layer_arn]
  environment {
    variables = {
      TARGET_BUCKET : module.target_bucket.bucket_name
    }
  }
}

resource "aws_lambda_permission" "s3_permission" {
  for_each      = local.lambdas
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambdas[each.key].function_name
  principal     = "s3.amazonaws.com"
  source_arn    = module.source_bucket.bucket_arn
}

resource "aws_s3_bucket_notification" "notification" {
  bucket = module.source_bucket.bucket_name

  dynamic "lambda_function" {
    for_each = local.lambdas
    content {
      lambda_function_arn = aws_lambda_function.lambdas[lambda_function.key].arn
      events              = ["s3:ObjectCreated:*"]
      filter_prefix       = lambda_function.value.prefix
    }
  }
  depends_on = [aws_lambda_permission.s3_permission]
}