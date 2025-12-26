resource "aws_sqs_queue" "dlq" {
  name                      = "node-esbuild-dlq"
  message_retention_seconds = 60 * 60 * 24 * 7 // 7 days
}