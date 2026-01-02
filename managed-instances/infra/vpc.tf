locals {
  num_subnets  = var.num_az * 2
  new_bits     = ceil(log(local.num_subnets, 2))
  subnet_cidrs = [for i in range(local.num_subnets) : cidrsubnet(var.cidr, local.new_bits, i)]
}

resource "aws_vpc" "vpc" {
  enable_dns_hostnames = true
  enable_dns_support   = true
  cidr_block           = var.cidr
}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_subnet" "public" {
  count                   = var.num_az
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = local.subnet_cidrs[count.index]
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "public-${count.index}"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "public"
  }
}

resource "aws_route_table_association" "public" {
  count          = var.num_az
  route_table_id = aws_route_table.public.id
  subnet_id      = aws_subnet.public[count.index].id
}

resource "aws_eip" "nat_eip" {
  domain = "vpc"
}

resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat_eip.id
  subnet_id     = aws_subnet.public[0].id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.nat.id
  }

  tags = {
    Name = "private"
  }
}

resource "aws_subnet" "private" {
  count             = var.num_az
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = local.subnet_cidrs[count.index + var.num_az]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "private-${count.index}"
  }
}

resource "aws_route_table_association" "private" {
  count          = var.num_az
  route_table_id = aws_route_table.private.id
  subnet_id      = aws_subnet.private[count.index].id
}
