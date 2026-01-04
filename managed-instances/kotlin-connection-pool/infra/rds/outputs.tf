output "bastion_instance_id" {
  value = aws_instance.bastion.id
}

output "cluster_endpoint" {
  value = aws_rds_cluster.cluster.endpoint
}

output "cluster_id" {
  value = aws_rds_cluster.cluster.id
}

output "cluster_resource_id" {
  value = aws_rds_cluster.cluster.cluster_resource_id
}

output "db_port" {
  value = aws_rds_cluster.cluster.port
}

output "db_name" {
  value = aws_rds_cluster.cluster.database_name
}