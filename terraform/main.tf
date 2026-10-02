terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "rldb-terraform-state"
    key            = "rldb-api-universal/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "rldb-terraform-lock"
  }
}

# Configure the AWS Provider
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "RLDB-API-UNIVERSAL"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# Create VPC
resource "aws_vpc" "rldb_vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "rldb-vpc-${var.environment}"
  }
}

# Create Public Subnets
resource "aws_subnet" "public_subnets" {
  count                   = length(var.public_subnet_cidrs)
  vpc_id                  = aws_vpc.rldb_vpc.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "rldb-public-subnet-${var.availability_zones[count.index]}-${var.environment}"
    Type = "public"
  }
}

# Create Private Subnets
resource "aws_subnet" "private_subnets" {
  count             = length(var.private_subnet_cidrs)
  vpc_id            = aws_vpc.rldb_vpc.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name = "rldb-private-subnet-${var.availability_zones[count.index]}-${var.environment}"
    Type = "private"
  }
}

# Create Internet Gateway
resource "aws_internet_gateway" "rldb_igw" {
  vpc_id = aws_vpc.rldb_vpc.id

  tags = {
    Name = "rldb-igw-${var.environment}"
  }
}

# Create NAT Gateway (for private subnets)
resource "aws_eip" "nat_eip" {
  domain = "vpc"

  tags = {
    Name = "rldb-nat-eip-${var.environment}"
  }
}

resource "aws_nat_gateway" "rldb_nat" {
  allocation_id = aws_eip.nat_eip.id
  subnet_id     = aws_subnet.public_subnets[0].id

  tags = {
    Name = "rldb-nat-${var.environment}"
  }

  depends_on = [aws_internet_gateway.rldb_igw]
}

# Create Route Tables
resource "aws_route_table" "public_rt" {
  vpc_id = aws_vpc.rldb_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.rldb_igw.id
  }

  tags = {
    Name = "rldb-public-rt-${var.environment}"
  }
}

resource "aws_route_table" "private_rt" {
  vpc_id = aws_vpc.rldb_vpc.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.rldb_nat.id
  }

  tags = {
    Name = "rldb-private-rt-${var.environment}"
  }
}

# Associate Route Tables with Subnets
resource "aws_route_table_association" "public_associations" {
  count          = length(aws_subnet.public_subnets)
  subnet_id      = aws_subnet.public_subnets[count.index].id
  route_table_id = aws_route_table.public_rt.id
}

resource "aws_route_table_association" "private_associations" {
  count          = length(aws_subnet.private_subnets)
  subnet_id      = aws_subnet.private_subnets[count.index].id
  route_table_id = aws_route_table.private_rt.id
}

# Create Security Groups
resource "aws_security_group" "api_sg" {
  name        = "rldb-api-sg-${var.environment}"
  description = "Security group for RLDB API"
  vpc_id      = aws_vpc.rldb_vpc.id

  # HTTP access
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS access
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # API port
  ingress {
    from_port   = var.api_port
    to_port     = var.api_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH access (restricted)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.ssh_allowed_cidrs
  }

  # Outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "rldb-api-sg-${var.environment}"
  }
}

resource "aws_security_group" "database_sg" {
  name        = "rldb-database-sg-${var.environment}"
  description = "Security group for RLDB Database"
  vpc_id      = aws_vpc.rldb_vpc.id

  # PostgreSQL access from API
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.api_sg.id]
  }

  # Redis access from API
  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.api_sg.id]
  }

  # Outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "rldb-database-sg-${var.environment}"
  }
}

# Create RDS PostgreSQL Database
resource "aws_db_subnet_group" "rldb_db_subnet_group" {
  name       = "rldb-db-subnet-group-${var.environment}"
  subnet_ids = aws_subnet.private_subnets[*].id

  tags = {
    Name = "rldb-db-subnet-group-${var.environment}"
  }
}

resource "aws_db_instance" "rldb_postgres" {
  identifier             = "rldb-postgres-${var.environment}"
  instance_class         = var.db_instance_class
  allocated_storage      = var.db_allocated_storage
  engine                 = "postgres"
  engine_version         = var.db_engine_version
  username               = var.db_username
  password               = var.db_password
  db_name                = var.db_name
  parameter_group_name   = "default.postgres15"
  skip_final_snapshot    = true
  publicly_accessible    = false
  vpc_security_group_ids = [aws_security_group.database_sg.id]
  db_subnet_group_name   = aws_db_subnet_group.rldb_db_subnet_group.name
  backup_retention_period = var.db_backup_retention
  multi_az               = var.db_multi_az

  tags = {
    Name = "rldb-postgres-${var.environment}"
  }
}

# Create ElastiCache Redis
resource "aws_elasticache_subnet_group" "rldb_redis_subnet_group" {
  name       = "rldb-redis-subnet-group-${var.environment}"
  subnet_ids = aws_subnet.private_subnets[*].id

  tags = {
    Name = "rldb-redis-subnet-group-${var.environment}"
  }
}

resource "aws_elasticache_cluster" "rldb_redis" {
  cluster_id           = "rldb-redis-${var.environment}"
  engine              = "redis"
  node_type           = var.redis_node_type
  num_cache_nodes      = var.redis_num_nodes
  parameter_group_name = "default.redis7"
  engine_version       = var.redis_engine_version
  port                 = 6379
  security_group_ids   = [aws_security_group.database_sg.id]
  subnet_group_name    = aws_elasticache_subnet_group.rldb_redis_subnet_group.name

  tags = {
    Name = "rldb-redis-${var.environment}"
  }
}

# Create Kinesis Stream for real-time processing
resource "aws_kinesis_stream" "rldb_transactions_stream" {
  name             = "rldb-transactions-stream-${var.environment}"
  shard_count      = var.kinesis_shard_count
  retention_period = var.kinesis_retention_period

  shard_level_metrics = ["IncomingBytes", "IncomingRecords", "OutgoingBytes", "OutgoingRecords"]

  tags = {
    Name = "rldb-transactions-stream-${var.environment}"
  }
}

# Create Kinesis Stream for logs
resource "aws_kinesis_stream" "rldb_logs_stream" {
  name             = "rldb-logs-stream-${var.environment}"
  shard_count      = var.kinesis_logs_shard_count
  retention_period = var.kinesis_retention_period

  shard_level_metrics = ["IncomingBytes", "IncomingRecords"]

  tags = {
    Name = "rldb-logs-stream-${var.environment}"
  }
}

# Create DynamoDB Table for real-time data
resource "aws_dynamodb_table" "rldb_transactions_table" {
  name           = "rldb-transactions-table-${var.environment}"
  billing_mode   = "PROVISIONED"
  read_capacity  = var.dynamodb_read_capacity
  write_capacity = var.dynamodb_write_capacity
  hash_key       = "PK"
  range_key      = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  # Global Secondary Index 1
  global_secondary_index {
    name            = "GSI1"
    hash_key        = "GSI1PK"
    range_key       = "GSI1SK"
    read_capacity   = var.dynamodb_gsi_read_capacity
    write_capacity  = var.dynamodb_gsi_write_capacity
    projection_type = "ALL"
  }

  # Global Secondary Index 2
  global_secondary_index {
    name            = "GSI2"
    hash_key        = "GSI2PK"
    range_key       = "GSI2SK"
    read_capacity   = var.dynamodb_gsi_read_capacity
    write_capacity  = var.dynamodb_gsi_write_capacity
    projection_type = "ALL"
  }

  # Enable streams
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  # Enable TTL
  ttl {
    attribute_name = "expires_at"
    enabled        = true
  }

  tags = {
    Name = "rldb-transactions-table-${var.environment}"
  }
}

# Create S3 Bucket for data storage
resource "aws_s3_bucket" "rldb_data_bucket" {
  bucket = "rldb-data-bucket-${var.environment}-${var.aws_account_id}"

  tags = {
    Name = "rldb-data-bucket-${var.environment}"
  }
}

resource "aws_s3_bucket_versioning" "rldb_data_bucket_versioning" {
  bucket = aws_s3_bucket.rldb_data_bucket.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "rldb_data_bucket_encryption" {
  bucket = aws_s3_bucket.rldb_data_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Create SNS Topic for alerts
resource "aws_sns_topic" "rldb_fraud_alerts" {
  name = "rldb-fraud-alerts-${var.environment}"

  tags = {
    Name = "rldb-fraud-alerts-${var.environment}"
  }
}

resource "aws_sns_topic" "rldb_notifications" {
  name = "rldb-notifications-${var.environment}"

  tags = {
    Name = "rldb-notifications-${var.environment}"
  }
}

# Create CloudWatch Log Group
resource "aws_cloudwatch_log_group" "rldb_api_logs" {
  name              = "/rldb/api/${var.environment}"
  retention_in_days = var.log_retention_days

  tags = {
    Name = "rldb-api-logs-${var.environment}"
  }
}

resource "aws_cloudwatch_log_group" "rldb_transactions_logs" {
  name              = "/rldb/transactions/${var.environment}"
  retention_in_days = var.log_retention_days

  tags = {
    Name = "rldb-transactions-logs-${var.environment}"
  }
}

# Create IAM Role for Lambda
resource "aws_iam_role" "rldb_lambda_role" {
  name = "rldb-lambda-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "rldb-lambda-role-${var.environment}"
  }
}

# Create IAM Policy for Lambda
resource "aws_iam_policy" "rldb_lambda_policy" {
  name        = "rldb-lambda-policy-${var.environment}"
  description = "Policy for RLDB Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [aws_dynamodb_table.rldb_transactions_table.arn]
      },
      {
        Effect = "Allow"
        Action = [
          "kinesis:GetRecords",
          "kinesis:GetShardIterator",
          "kinesis:DescribeStream"
        ]
        Resource = [
          aws_kinesis_stream.rldb_transactions_stream.arn,
          aws_kinesis_stream.rldb_logs_stream.arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject"
        ]
        Resource = ["${aws_s3_bucket.rldb_data_bucket.arn}/*"]
      },
      {
        Effect = "Allow"
        Action = [
          "sns:Publish"
        ]
        Resource = [
          aws_sns_topic.rldb_fraud_alerts.arn,
          aws_sns_topic.rldb_notifications.arn
        ]
      }
    ]
  })

  tags = {
    Name = "rldb-lambda-policy-${var.environment}"
  }
}

# Attach Policy to Role
resource "aws_iam_role_policy_attachment" "rldb_lambda_attachment" {
  role       = aws_iam_role.rldb_lambda_role.name
  policy_arn = aws_iam_policy.rldb_lambda_policy.arn
}

# Create Lambda Function for stream processing
resource "aws_lambda_function" "rldb_stream_processor" {
  function_name = "rldb-stream-processor-${var.environment}"
  role          = aws_iam_role.rldb_lambda_role.arn
  handler       = "main.processStream"
  runtime       = "go1.x"
  memory_size   = var.lambda_memory_size
  timeout        = var.lambda_timeout

  filename         = "../lambda/rldb-stream-processor.zip"
  source_code_hash = filebase64sha256("../lambda/rldb-stream-processor.zip")

  environment {
    variables = {
      ENVIRONMENT       = var.environment
      DYNAMODB_TABLE    = aws_dynamodb_table.rldb_transactions_table.name
      SNS_FRAUD_TOPIC   = aws_sns_topic.rldb_fraud_alerts.arn
      SNS_NOTIFY_TOPIC  = aws_sns_topic.rldb_notifications.arn
      LOG_GROUP         = aws_cloudwatch_log_group.rldb_transactions_logs.name
    }
  }

  tags = {
    Name = "rldb-stream-processor-${var.environment}"
  }
}

# Create Lambda Event Source Mapping for Kinesis
resource "aws_lambda_event_source_mapping" "rldb_transactions_stream_mapping" {
  event_source_arn  = aws_kinesis_stream.rldb_transactions_stream.arn
  function_name    = aws_lambda_function.rldb_stream_processor.arn
  starting_position = "LATEST"
  batch_size        = var.lambda_batch_size

  depends_on = [aws_lambda_function.rldb_stream_processor]
}

resource "aws_lambda_event_source_mapping" "rldb_logs_stream_mapping" {
  event_source_arn  = aws_kinesis_stream.rldb_logs_stream.arn
  function_name    = aws_lambda_function.rldb_stream_processor.arn
  starting_position = "LATEST"
  batch_size        = var.lambda_batch_size

  depends_on = [aws_lambda_function.rldb_stream_processor]
}

# Create ECR Repository for containers
resource "aws_ecr_repository" "rldb_api_ecr" {
  name                 = "rldb-api-universal/${var.environment}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "rldb-api-ecr-${var.environment}"
  }
}

# Create ECS Cluster
resource "aws_ecs_cluster" "rldb_api_cluster" {
  name = "rldb-api-cluster-${var.environment}"

  tags = {
    Name = "rldb-api-cluster-${var.environment}"
  }
}

# Create ECS Task Definition
resource "aws_ecs_task_definition" "rldb_api_task" {
  family                   = "rldb-api-task-${var.environment}"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.ecs_cpu
  memory                   = var.ecs_memory
  execution_role_arn       = aws_iam_role.rldb_ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name      = "rldb-api-universal"
      image     = "${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/${aws_ecr_repository.rldb_api_ecr.name}:latest"
      essential = true
      portMappings = [
        {
          containerPort = var.api_port
          hostPort      = var.api_port
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "ENVIRONMENT"
          value = var.environment
        },
        {
          name  = "DB_HOST"
          value = aws_db_instance.rldb_postgres.address
        },
        {
          name  = "DB_PORT"
          value = "5432"
        },
        {
          name  = "DB_NAME"
          value = var.db_name
        },
        {
          name  = "REDIS_HOST"
          value = aws_elasticache_cluster.rldb_redis.cache_nodes[0].address
        },
        {
          name  = "REDIS_PORT"
          value = "6379"
        },
        {
          name  = "AWS_KINESIS_ENABLED"
          value = "true"
        },
        {
          name  = "AWS_KINESIS_STREAM_NAME"
          value = aws_kinesis_stream.rldb_transactions_stream.name
        },
        {
          name  = "AWS_DYNAMODB_ENABLED"
          value = "true"
        },
        {
          name  = "AWS_DYNAMODB_TABLE_NAME"
          value = aws_dynamodb_table.rldb_transactions_table.name
        }
      ]
      secrets = [
        {
          name      = "DB_USERNAME"
          valueFrom = aws_db_instance.rldb_postgres.username
        },
        {
          name      = "DB_PASSWORD"
          valueFrom = aws_db_instance.rldb_postgres.password
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.rldb_api_logs.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "rldb-api"
        }
      }
    }
  ])

  tags = {
    Name = "rldb-api-task-${var.environment}"
  }
}

# Create IAM Role for ECS Task Execution
resource "aws_iam_role" "rldb_ecs_task_execution_role" {
  name = "rldb-ecs-task-execution-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "rldb-ecs-task-execution-role-${var.environment}"
  }
}

resource "aws_iam_role_policy_attachment" "rldb_ecs_task_execution_role_policy" {
  role       = aws_iam_role.rldb_ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Create ECS Service
resource "aws_ecs_service" "rldb_api_service" {
  name            = "rldb-api-service-${var.environment}"
  cluster         = aws_ecs_cluster.rldb_api_cluster.id
  task_definition = aws_ecs_task_definition.rldb_api_task.arn
  desired_count   = var.ecs_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.public_subnets[*].id
    security_groups  = [aws_security_group.api_sg.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.rldb_api_target_group.arn
    container_name   = "rldb-api-universal"
    container_port   = var.api_port
  }

  depends_on = [aws_lb_listener.rldb_api_listener]

  tags = {
    Name = "rldb-api-service-${var.environment}"
  }
}

# Create Application Load Balancer
resource "aws_lb" "rldb_api_alb" {
  name               = "rldb-api-alb-${var.environment}"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.api_sg.id]
  subnets            = aws_subnet.public_subnets[*].id

  enable_deletion_protection = var.enable_deletion_protection

  tags = {
    Name = "rldb-api-alb-${var.environment}"
  }
}

resource "aws_lb_target_group" "rldb_api_target_group" {
  name        = "rldb-api-tg-${var.environment}"
  port        = var.api_port
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.rldb_vpc.id

  health_check {
    path                = "/api/health"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = "rldb-api-tg-${var.environment}"
  }
}

resource "aws_lb_listener" "rldb_api_listener" {
  load_balancer_arn = aws_lb.rldb_api_alb.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = var.ssl_certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.rldb_api_target_group.arn
  }

  tags = {
    Name = "rldb-api-listener-${var.environment}"
  }
}

resource "aws_lb_listener" "rldb_api_listener_http" {
  load_balancer_arn = aws_lb.rldb_api_alb.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  tags = {
    Name = "rldb-api-listener-http-${var.environment}"
  }
}

# Create Route 53 Record
resource "aws_route53_record" "rldb_api_record" {
  count   = var.create_route53_record ? 1 : 0
  zone_id = var.route53_zone_id
  name    = var.api_domain_name
  type    = "A"

  alias {
    name                   = aws_lb.rldb_api_alb.dns_name
    zone_id                = aws_lb.rldb_api_alb.zone_id
    evaluate_target_health = true
  }
}

# Outputs
output "vpc_id" {
  value = aws_vpc.rldb_vpc.id
}

output "public_subnet_ids" {
  value = aws_subnet.public_subnets[*].id
}

output "private_subnet_ids" {
  value = aws_subnet.private_subnets[*].id
}

output "database_endpoint" {
  value = aws_db_instance.rldb_postgres.endpoint
}

output "redis_endpoint" {
  value = aws_elasticache_cluster.rldb_redis.cache_nodes[0].address
}

output "kinesis_stream_name" {
  value = aws_kinesis_stream.rldb_transactions_stream.name
}

output "kinesis_logs_stream_name" {
  value = aws_kinesis_stream.rldb_logs_stream.name
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.rldb_transactions_table.name
}

output "s3_bucket_name" {
  value = aws_s3_bucket.rldb_data_bucket.bucket
}

output "sns_fraud_topic_arn" {
  value = aws_sns_topic.rldb_fraud_alerts.arn
}

output "sns_notification_topic_arn" {
  value = aws_sns_topic.rldb_notifications.arn
}

output "alb_dns_name" {
  value = aws_lb.rldb_api_alb.dns_name
}

output "ecs_service_name" {
  value = aws_ecs_service.rldb_api_service.name
}

output "lambda_function_name" {
  value = aws_lambda_function.rldb_stream_processor.function_name
}
