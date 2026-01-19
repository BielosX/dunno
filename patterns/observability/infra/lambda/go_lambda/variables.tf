variable "bundle_path" {
  type = string
}

variable "name" {
  type = string
}

variable "handler" {
  type = string
}

variable "env_vars" {
  type    = map(string)
  default = {}
}

variable "role_policies" {
  type    = list(string)
  default = []
}