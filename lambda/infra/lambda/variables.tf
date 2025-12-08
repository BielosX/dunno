variable "lambda_count" {
  type    = number
  default = 1
}

variable "language" {
  type    = string
  default = "golang"
}

variable "bucket" {
  type    = string
  default = "dummy"
}

variable "bucket_key" {
  type    = string
  default = "dummy.zip"
}

variable "handler" {
  type    = string
  default = "dummy"
}

variable "snap_start" {
  type    = bool
  default = false
}

variable "architecture" {
  type    = string
  default = "arm64"
}