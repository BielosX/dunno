locals {
  cluster_id              = "mi-cluster"
  master_password_version = 1
  cluster_port            = 5432
}

ephemeral "random_password" "password" {
  length  = 32
  special = false
  upper   = true
}

resource "aws_ssm_parameter" "master_password" {
  name             = "/rds/${local.cluster_id}/master-password"
  type             = "SecureString"
  value_wo         = ephemeral.random_password.password.result
  value_wo_version = local.master_password_version
}

resource "aws_db_subnet_group" "group" {
  name       = "main"
  subnet_ids = local.private_subnet_ids
}

resource "aws_security_group" "cluster_sg" {
  vpc_id = local.vpc_id
  ingress {
    cidr_blocks = [local.vpc_cidr]
    protocol    = "tcp"
    from_port   = local.cluster_port
    to_port     = local.cluster_port
  }
}

resource "aws_rds_cluster" "cluster" {
  cluster_identifier                  = local.cluster_id
  engine                              = "aurora-postgresql"
  database_name                       = "postgres"
  engine_version                      = "17.7"
  master_username                     = "master"
  engine_mode                         = "provisioned"
  port                                = local.cluster_port
  master_password_wo                  = ephemeral.random_password.password.result
  master_password_wo_version          = local.master_password_version
  db_subnet_group_name                = aws_db_subnet_group.group.name
  vpc_security_group_ids              = [aws_security_group.cluster_sg.id]
  iam_database_authentication_enabled = true
  skip_final_snapshot                 = true
}

resource "aws_rds_cluster_instance" "instance" {
  cluster_identifier   = aws_rds_cluster.cluster.cluster_identifier
  engine               = aws_rds_cluster.cluster.engine
  engine_version       = aws_rds_cluster.cluster.engine_version
  instance_class       = "db.t4g.medium"
  db_subnet_group_name = aws_db_subnet_group.group.name
}