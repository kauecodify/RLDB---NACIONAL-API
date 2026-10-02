module github.com/rldb-br/rldb-api-universal

go 1.21

require (
	github.com/aws/aws-sdk-go-v2 v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.18.47
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.10.47
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.28.7
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.24.7
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.23.10
	github.com/aws/aws-sdk-go-v2/service/lambda v1.41.3
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.32.4
	github.com/aws/aws-sdk-go-v2/service/sns v1.19.11
	github.com/aws/aws-sdk-go-v2/service/sqs v1.27.14
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.0.5
	github.com/stretchr/testify v1.8.4
	golang.org/x/crypto v0.14.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

// AWS SDK dependencies
require (
	github.com/aws/aws-sdk-go-v2/credentials v1.13.47
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.145.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.30.0
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.25.3
	github.com/aws/aws-sdk-go-v2/service/iam v1.26.1
	github.com/aws/smithy-go v1.17.2
)
