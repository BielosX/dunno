resource "aws_dynamodb_table" "books" {
  name             = "books"
  hash_key         = "id"
  billing_mode     = "PROVISIONED"
  write_capacity   = 2
  read_capacity    = 2
  stream_enabled   = true
  stream_view_type = "NEW_IMAGE"

  attribute {
    name = "id"
    type = "S"
  }
}