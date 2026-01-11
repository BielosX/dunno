resource "aws_dynamodb_table" "table" {
  name           = "books"
  hash_key       = "Id"
  billing_mode   = "PROVISIONED"
  write_capacity = 2
  read_capacity  = 2

  attribute {
    name = "Id"
    type = "S"
  }
}