variable "lambda_count" {
  type    = number
  default = 1
}

variable "language" {
  type = string
}

variable "bucket" {
  type = string
}

variable "bucket_key" {
  type = string
}

variable "handler" {
  type    = string
  default = "dummy"
}

variable "snap_start" {
  type    = bool
  default = false
}