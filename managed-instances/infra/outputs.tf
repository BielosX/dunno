output "private_subnet_ids" {
  value = aws_subnet.private[*].id
}

output "public_subnet_ids" {
  value = aws_subnet.public[*].id
}

output "private_capacity_provider_arn" {
  value = aws_lambda_capacity_provider.providers["private_subnets"].arn
}

output "public_capacity_provider_arn" {
  value = aws_lambda_capacity_provider.providers["public_subnets"].arn
}

output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "vpc_cidr" {
  value = aws_vpc.vpc.cidr_block
}