output "latest_alias_name" {
  value = aws_lambda_alias.latest.name
}

output "function_arn" {
  value = aws_lambda_function.lambda.arn
}