variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "aws_account_id" {
  description = "AWS Account ID"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

# VPC Configuration
variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "Public subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "private_subnet_cidrs" {
  description = "Private subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.11.0/24", "10.0.12.0/24"]
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

# API Configuration
variable "api_port" {
  description = "API port"
  type        = number
  default     = 8080
}

variable "api_domain_name" {
  description = "API domain name"
  type        = string
  default     = ""
}

variable "ssl_certificate_arn" {
  description = "SSL certificate ARN"
  type        = string
  default     = ""
}

variable "create_route53_record" {
  description = "Create Route53 record"
  type        = bool
  default     = false
}

variable "route53_zone_id" {
  description = "Route53 zone ID"
  type        = string
  default     = ""
}

# Database Configuration
variable "db_instance_class" {
  description = "Database instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "db_allocated_storage" {
  description = "Database allocated storage"
  type        = number
  default     = 20
}

variable "db_engine_version" {
  description = "Database engine version"
  type        = string
  default     = "15.4"
}

variable "db_username" {
  description = "Database username"
  type        = string
  default     = "rldb_user"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "rldb_api"
}

variable "db_backup_retention" {
  description = "Database backup retention period"
  type        = number
  default     = 7
}

variable "db_multi_az" {
  description = "Enable multi-AZ deployment"
  type        = bool
  default     = false
}

# Redis Configuration
variable "redis_node_type" {
  description = "Redis node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "redis_num_nodes" {
  description = "Number of Redis nodes"
  type        = number
  default     = 1
}

variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.0"
}

# Kinesis Configuration
variable "kinesis_shard_count" {
  description = "Kinesis shard count"
  type        = number
  default     = 4
}

variable "kinesis_retention_period" {
  description = "Kinesis retention period in hours"
  type        = number
  default     = 24
}

variable "kinesis_logs_shard_count" {
  description = "Kinesis logs shard count"
  type        = number
  default     = 2
}

# DynamoDB Configuration
variable "dynamodb_read_capacity" {
  description = "DynamoDB read capacity"
  type        = number
  default     = 100
}

variable "dynamodb_write_capacity" {
  description = "DynamoDB write capacity"
  type        = number
  default     = 100
}

variable "dynamodb_gsi_read_capacity" {
  description = "DynamoDB GSI read capacity"
  type        = number
  default     = 50
}

variable "dynamodb_gsi_write_capacity" {
  description = "DynamoDB GSI write capacity"
  type        = number
  default     = 50
}

# S3 Configuration
variable "s3_bucket_name" {
  description = "S3 bucket name"
  type        = string
  default     = ""
}

# CloudWatch Configuration
variable "log_retention_days" {
  description = "Log retention in days"
  type        = number
  default     = 30
}

# Lambda Configuration
variable "lambda_memory_size" {
  description = "Lambda memory size"
  type        = number
  default     = 512
}

variable "lambda_timeout" {
  description = "Lambda timeout"
  type        = number
  default     = 30
}

variable "lambda_batch_size" {
  description = "Lambda batch size"
  type        = number
  default     = 100
}

# ECS Configuration
variable "ecs_cpu" {
  description = "ECS CPU"
  type        = number
  default     = 1024
}

variable "ecs_memory" {
  description = "ECS memory"
  type        = number
  default     = 2048
}

variable "ecs_desired_count" {
  description = "ECS desired count"
  type        = number
  default     = 2
}

# Security Configuration
variable "ssh_allowed_cidrs" {
  description = "Allowed CIDR blocks for SSH"
  type        = list(string)
  default     = ["10.0.0.0/8"]
}

variable "enable_deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false
}

# AI/ML Configuration
variable "enable_fraud_detection" {
  description = "Enable fraud detection"
  type        = bool
  default     = true
}

variable "enable_demand_forecast" {
  description = "Enable demand forecast"
  type        = bool
  default     = true
}

variable "enable_route_optimization" {
  description = "Enable route optimization"
  type        = bool
  default     = true
}

variable "fraud_model_threshold" {
  description = "Fraud detection model threshold"
  type        = number
  default     = 0.7
}

variable "demand_model_threshold" {
  description = "Demand forecast model threshold"
  type        = number
  default     = 0.8
}
