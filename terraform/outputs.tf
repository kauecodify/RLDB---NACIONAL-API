# Outputs for RLDB API UNIVERSAL Infrastructure

# VPC Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.rldb_vpc.id
}

output "vpc_cidr" {
  description = "VPC CIDR block"
  value       = aws_vpc.rldb_vpc.cidr_block
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = aws_subnet.public_subnets[*].id
}

output "public_subnet_cidrs" {
  description = "Public subnet CIDR blocks"
  value       = aws_subnet.public_subnets[*].cidr_block
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = aws_subnet.private_subnets[*].id
}

output "private_subnet_cidrs" {
  description = "Private subnet CIDR blocks"
  value       = aws_subnet.private_subnets[*].cidr_block
}

# Network Outputs
output "internet_gateway_id" {
  description = "Internet Gateway ID"
  value       = aws_internet_gateway.rldb_igw.id
}

output "nat_gateway_id" {
  description = "NAT Gateway ID"
  value       = aws_nat_gateway.rldb_nat.id
}

output "nat_gateway_public_ip" {
  description = "NAT Gateway public IP"
  value       = aws_eip.nat_eip.public_ip
}

# Security Outputs
output "api_security_group_id" {
  description = "API security group ID"
  value       = aws_security_group.api_sg.id
}

output "database_security_group_id" {
  description = "Database security group ID"
  value       = aws_security_group.database_sg.id
}

# Database Outputs
output "postgres_endpoint" {
  description = "PostgreSQL endpoint"
  value       = aws_db_instance.rldb_postgres.endpoint
}

output "postgres_address" {
  description = "PostgreSQL address"
  value       = aws_db_instance.rldb_postgres.address
}

output "postgres_port" {
  description = "PostgreSQL port"
  value       = aws_db_instance.rldb_postgres.port
}

output "postgres_username" {
  description = "PostgreSQL username"
  value       = aws_db_instance.rldb_postgres.username
  sensitive   = true
}

output "postgres_db_name" {
  description = "PostgreSQL database name"
  value       = aws_db_instance.rldb_postgres.db_name
}

# Redis Outputs
output "redis_endpoint" {
  description = "Redis endpoint"
  value       = aws_elasticache_cluster.rldb_redis.cache_nodes[0].address
}

output "redis_port" {
  description = "Redis port"
  value       = aws_elasticache_cluster.rldb_redis.cache_nodes[0].port
}

# Kinesis Outputs
output "kinesis_transactions_stream_name" {
  description = "Kinesis transactions stream name"
  value       = aws_kinesis_stream.rldb_transactions_stream.name
}

output "kinesis_transactions_stream_arn" {
  description = "Kinesis transactions stream ARN"
  value       = aws_kinesis_stream.rldb_transactions_stream.arn
}

output "kinesis_logs_stream_name" {
  description = "Kinesis logs stream name"
  value       = aws_kinesis_stream.rldb_logs_stream.name
}

output "kinesis_logs_stream_arn" {
  description = "Kinesis logs stream ARN"
  value       = aws_kinesis_stream.rldb_logs_stream.arn
}

# DynamoDB Outputs
output "dynamodb_table_name" {
  description = "DynamoDB table name"
  value       = aws_dynamodb_table.rldb_transactions_table.name
}

output "dynamodb_table_arn" {
  description = "DynamoDB table ARN"
  value       = aws_dynamodb_table.rldb_transactions_table.arn
}

output "dynamodb_table_stream_arn" {
  description = "DynamoDB table stream ARN"
  value       = aws_dynamodb_table.rldb_transactions_table.stream_arn
}

# S3 Outputs
output "s3_bucket_name" {
  description = "S3 bucket name"
  value       = aws_s3_bucket.rldb_data_bucket.bucket
}

output "s3_bucket_arn" {
  description = "S3 bucket ARN"
  value       = aws_s3_bucket.rldb_data_bucket.arn
}

output "s3_bucket_domain" {
  description = "S3 bucket domain"
  value       = aws_s3_bucket.rldb_data_bucket.bucket_domain_name
}

# SNS Outputs
output "sns_fraud_alerts_topic_name" {
  description = "SNS fraud alerts topic name"
  value       = aws_sns_topic.rldb_fraud_alerts.name
}

output "sns_fraud_alerts_topic_arn" {
  description = "SNS fraud alerts topic ARN"
  value       = aws_sns_topic.rldb_fraud_alerts.arn
}

output "sns_notifications_topic_name" {
  description = "SNS notifications topic name"
  value       = aws_sns_topic.rldb_notifications.name
}

output "sns_notifications_topic_arn" {
  description = "SNS notifications topic ARN"
  value       = aws_sns_topic.rldb_notifications.arn
}

# CloudWatch Outputs
output "cloudwatch_api_log_group" {
  description = "CloudWatch API log group"
  value       = aws_cloudwatch_log_group.rldb_api_logs.name
}

output "cloudwatch_transactions_log_group" {
  description = "CloudWatch transactions log group"
  value       = aws_cloudwatch_log_group.rldb_transactions_logs.name
}

# Lambda Outputs
output "lambda_stream_processor_name" {
  description = "Lambda stream processor name"
  value       = aws_lambda_function.rldb_stream_processor.function_name
}

output "lambda_stream_processor_arn" {
  description = "Lambda stream processor ARN"
  value       = aws_lambda_function.rldb_stream_processor.arn
}

output "lambda_execution_role_arn" {
  description = "Lambda execution role ARN"
  value       = aws_iam_role.rldb_lambda_role.arn
}

# ECR Outputs
output "ecr_repository_name" {
  description = "ECR repository name"
  value       = aws_ecr_repository.rldb_api_ecr.name
}

output "ecr_repository_url" {
  description = "ECR repository URL"
  value       = aws_ecr_repository.rldb_api_ecr.repository_url
}

# ECS Outputs
output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.rldb_api_cluster.name
}

output "ecs_cluster_arn" {
  description = "ECS cluster ARN"
  value       = aws_ecs_cluster.rldb_api_cluster.arn
}

output "ecs_task_definition_arn" {
  description = "ECS task definition ARN"
  value       = aws_ecs_task_definition.rldb_api_task.arn
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = aws_ecs_service.rldb_api_service.name
}

output "ecs_service_arn" {
  description = "ECS service ARN"
  value       = aws_ecs_service.rldb_api_service.arn
}

output "ecs_task_execution_role_arn" {
  description = "ECS task execution role ARN"
  value       = aws_iam_role.rldb_ecs_task_execution_role.arn
}

# Load Balancer Outputs
output "alb_name" {
  description = "ALB name"
  value       = aws_lb.rldb_api_alb.name
}

output "alb_arn" {
  description = "ALB ARN"
  value       = aws_lb.rldb_api_alb.arn
}

output "alb_dns_name" {
  description = "ALB DNS name"
  value       = aws_lb.rldb_api_alb.dns_name
}

output "alb_zone_id" {
  description = "ALB zone ID"
  value       = aws_lb.rldb_api_alb.zone_id
}

output "target_group_name" {
  description = "Target group name"
  value       = aws_lb_target_group.rldb_api_target_group.name
}

output "target_group_arn" {
  description = "Target group ARN"
  value       = aws_lb_target_group.rldb_api_target_group.arn
}

output "listener_arn" {
  description = "HTTPS listener ARN"
  value       = aws_lb_listener.rldb_api_listener.arn
}

# Route 53 Outputs
output "route53_record_name" {
  description = "Route53 record name"
  value       = var.api_domain_name != "" ? var.api_domain_name : ""
}

output "route53_record_fqdn" {
  description = "Route53 record FQDN"
  value       = var.api_domain_name != "" ? "${var.api_domain_name}." : ""
}

# Connection Strings
output "postgres_connection_string" {
  description = "PostgreSQL connection string"
  value       = "postgresql://${aws_db_instance.rldb_postgres.username}:${aws_db_instance.rldb_postgres.password}@${aws_db_instance.rldb_postgres.endpoint}:${aws_db_instance.rldb_postgres.port}/${aws_db_instance.rldb_postgres.db_name}?sslmode=require"
  sensitive   = true
}

output "redis_connection_string" {
  description = "Redis connection string"
  value       = "redis://${aws_elasticache_cluster.rldb_redis.cache_nodes[0].address}:${aws_elasticache_cluster.rldb_redis.cache_nodes[0].port}"
}

# API Endpoints
output "api_endpoint" {
  description = "API endpoint"
  value       = var.api_domain_name != "" ? "https://${var.api_domain_name}" : "http://${aws_lb.rldb_api_alb.dns_name}"
}

output "api_health_endpoint" {
  description = "API health check endpoint"
  value       = var.api_domain_name != "" ? "https://${var.api_domain_name}/api/health" : "http://${aws_lb.rldb_api_alb.dns_name}/api/health"
}

# Environment Configuration
output "environment" {
  description = "Environment name"
  value       = var.environment
}

output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
}

# Summary Output
output "infrastructure_summary" {
  description = "Infrastructure summary"
  value = {
    vpc_id                     = aws_vpc.rldb_vpc.id
    public_subnet_count        = length(aws_subnet.public_subnets)
    private_subnet_count       = length(aws_subnet.private_subnets)
    database_endpoint          = aws_db_instance.rldb_postgres.endpoint
    redis_endpoint             = aws_elasticache_cluster.rldb_redis.cache_nodes[0].address
    kinesis_transactions_stream = aws_kinesis_stream.rldb_transactions_stream.name
    kinesis_logs_stream        = aws_kinesis_stream.rldb_logs_stream.name
    dynamodb_table             = aws_dynamodb_table.rldb_transactions_table.name
    s3_bucket                  = aws_s3_bucket.rldb_data_bucket.bucket
    alb_dns_name               = aws_lb.rldb_api_alb.dns_name
    ecs_service_name           = aws_ecs_service.rldb_api_service.name
    lambda_function_name       = aws_lambda_function.rldb_stream_processor.function_name
    environment                = var.environment
    api_endpoint               = var.api_domain_name != "" ? "https://${var.api_domain_name}" : "http://${aws_lb.rldb_api_alb.dns_name}"
  }
}
