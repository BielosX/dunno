resource "aws_dynamodb_table" "table" {
  name           = "books"
  hash_key       = "id"
  billing_mode   = "PROVISIONED"
  write_capacity = 2
  read_capacity  = 2

  attribute {
    name = "id"
    type = "S"
  }
}