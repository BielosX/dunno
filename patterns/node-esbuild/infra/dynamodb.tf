resource "aws_dynamodb_table" "users" {
  name           = "users"
  hash_key       = "userId"
  billing_mode   = "PROVISIONED"
  write_capacity = 2
  read_capacity  = 2

  attribute {
    name = "userId"
    type = "S"
  }
}