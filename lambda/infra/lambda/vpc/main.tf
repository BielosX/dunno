resource "aws_vpc" "vpc" {
  enable_dns_hostnames = true
  enable_dns_support   = true
  cidr_block           = "10.0.0.0/16"
}

resource "aws_subnet" "private" {
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = "10.0.0.0/20"
  map_public_ip_on_launch = false
}

resource "aws_security_group" "log_endpoint" {
  vpc_id = aws_vpc.vpc.id
  ingress {
    protocol    = "tcp"
    from_port   = 443
    to_port     = 443
    cidr_blocks = [aws_vpc.vpc.cidr_block]
  }
  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "cloudwatch_logs" {
  vpc_id            = aws_vpc.vpc.id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.logs"
  vpc_endpoint_type = "Interface"

  subnet_ids         = [aws_subnet.private.id]
  security_group_ids = [aws_security_group.log_endpoint.id]

  private_dns_enabled = true
}

resource "aws_security_group" "lambda" {
  vpc_id = aws_vpc.vpc.id
  egress {
    security_groups = [aws_security_group.log_endpoint.id]
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
  }
}

output "lambda_sg_id" {
  value = aws_security_group.lambda.id
}

output "private_subnet_id" {
  value = aws_subnet.private.id
}