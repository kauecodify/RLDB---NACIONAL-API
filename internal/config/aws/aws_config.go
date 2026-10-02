package awsconfig

import (
	"time"

	"github.com/joho/godotenv"
)

// AWSConfig - Configuracao completa da AWS
type AWSConfig struct {
	// Kinesis
	KinesisEnabled         bool          `json:"kinesis_enabled" yaml:"kinesis_enabled"`
	KinesisStreamName      string        `json:"kinesis_stream_name" yaml:"kinesis_stream_name"`
	KinesisShardCount      int           `json:"kinesis_shard_count" yaml:"kinesis_shard_count"`
	KinesisRetentionPeriod int           `json:"kinesis_retention_period" yaml:"kinesis_retention_period"`
	KinesisRegion          string        `json:"kinesis_region" yaml:"kinesis_region"`

	// DynamoDB
	DynamoDBEnabled      bool    `json:"dynamodb_enabled" yaml:"dynamodb_enabled"`
	DynamoDBTableName    string  `json:"dynamodb_table_name" yaml:"dynamodb_table_name"`
	DynamoDBRegion       string  `json:"dynamodb_region" yaml:"dynamodb_region"`
	DynamoDBReadCapacity int64   `json:"dynamodb_read_capacity" yaml:"dynamodb_read_capacity"`
	DynamoDBWriteCapacity int64  `json:"dynamodb_write_capacity" yaml:"dynamodb_write_capacity"`

	// S3
	S3Enabled   bool   `json:"s3_enabled" yaml:"s3_enabled"`
	S3Bucket    string `json:"s3_bucket" yaml:"s3_bucket"`
	S3Region    string `json:"s3_region" yaml:"s3_region"`
	S3Prefix    string `json:"s3_prefix" yaml:"s3_prefix"`

	// Lambda
	LambdaEnabled     bool          `json:"lambda_enabled" yaml:"lambda_enabled"`
	LambdaRegion      string        `json:"lambda_region" yaml:"lambda_region"`
	LambdaMemorySize  int           `json:"lambda_memory_size" yaml:"lambda_memory_size"`
	LambdaTimeout    time.Duration `json:"lambda_timeout" yaml:"lambda_timeout"`
	LambdaConcurrency int           `json:"lambda_concurrency" yaml:"lambda_concurrency"`

	// SageMaker
	SageMakerEnabled  bool   `json:"sagemaker_enabled" yaml:"sagemaker_enabled"`
	SageMakerRegion   string `json:"sagemaker_region" yaml:"sagemaker_region"`
	SageMakerEndpoint string `json:"sagemaker_endpoint" yaml:"sagemaker_endpoint"`
	SageMakerModel    string `json:"sagemaker_model" yaml:"sagemaker_model"`

	// SNS
	SNSEnabled        bool   `json:"sns_enabled" yaml:"sns_enabled"`
	SNSRegion         string `json:"sns_region" yaml:"sns_region"`
	SNSFraudAlertArn  string `json:"sns_fraud_alert_arn" yaml:"sns_fraud_alert_arn"`
	SNSNotificationArn string `json:"sns_notification_arn" yaml:"sns_notification_arn"`

	// CloudWatch
	CloudWatchEnabled bool   `json:"cloudwatch_enabled" yaml:"cloudwatch_enabled"`
	CloudWatchRegion  string `json:"cloudwatch_region" yaml:"cloudwatch_region"`
	CloudWatchLogGroup string `json:"cloudwatch_log_group" yaml:"cloudwatch_log_group"`

	// SQS
	SQSEnabled          bool   `json:"sqs_enabled" yaml:"sqs_enabled"`
	SQSRegion           string `json:"sqs_region" yaml:"sqs_region"`
	SQSQueueName        string `json:"sqs_queue_name" yaml:"sqs_queue_name"`
	SQSMaxReceiveCount  int    `json:"sqs_max_receive_count" yaml:"sqs_max_receive_count"`
	SQSVisibilityTimeout int    `json:"sqs_visibility_timeout" yaml:"sqs_visibility_timeout"`

	// IA/ML Configuration
	AIFraudDetectionEnabled  bool    `json:"ai_fraud_detection_enabled" yaml:"ai_fraud_detection_enabled"`
	AIDemandForecastEnabled   bool    `json:"ai_demand_forecast_enabled" yaml:"ai_demand_forecast_enabled"`
	AIRouteOptimizationEnabled bool  `json:"ai_route_optimization_enabled" yaml:"ai_route_optimization_enabled"`
	AIFraudModelThreshold     float64 `json:"ai_fraud_model_threshold" yaml:"ai_fraud_model_threshold"`
	AIDemandModelThreshold    float64 `json:"ai_demand_model_threshold" yaml:"ai_demand_model_threshold"`

	// Real-time Processing
	RealTimeProcessingEnabled bool   `json:"real_time_processing_enabled" yaml:"real_time_processing_enabled"`
	BatchProcessingEnabled     bool   `json:"batch_processing_enabled" yaml:"batch_processing_enabled"`
	BatchSize                  int    `json:"batch_size" yaml:"batch_size"`
	BatchInterval              string `json:"batch_interval" yaml:"batch_interval"` // e.g., "5m", "1h"
}

// LoadAWSConfig - Carrega configuracao da AWS
def LoadAWSConfig() (*AWSConfig, error) {
	// Carregar variaveis de ambiente
	if err := godotenv.Load(); err != nil {
		// Ignorar erro se .env nao existir
		_ = err
	}

	return &AWSConfig{
		// Kinesis
		KinesisEnabled:         getBoolEnv("AWS_KINESIS_ENABLED", true),
		KinesisStreamName:      getStringEnv("AWS_KINESIS_STREAM_NAME", "rldb-transactions-stream"),
		KinesisShardCount:      getIntEnv("AWS_KINESIS_SHARD_COUNT", 4),
		KinesisRetentionPeriod: getIntEnv("AWS_KINESIS_RETENTION_PERIOD", 24),
		KinesisRegion:          getStringEnv("AWS_KINESIS_REGION", "us-east-1"),

		// DynamoDB
		DynamoDBEnabled:      getBoolEnv("AWS_DYNAMODB_ENABLED", true),
		DynamoDBTableName:    getStringEnv("AWS_DYNAMODB_TABLE_NAME", "rldb-transactions-table"),
		DynamoDBRegion:       getStringEnv("AWS_DYNAMODB_REGION", "us-east-1"),
		DynamoDBReadCapacity:  getInt64Env("AWS_DYNAMODB_READ_CAPACITY", 100),
		DynamoDBWriteCapacity: getInt64Env("AWS_DYNAMODB_WRITE_CAPACITY", 100),

		// S3
		S3Enabled: getBoolEnv("AWS_S3_ENABLED", true),
		S3Bucket:  getStringEnv("AWS_S3_BUCKET", "rldb-data-bucket"),
		S3Region:  getStringEnv("AWS_S3_REGION", "us-east-1"),
		S3Prefix:  getStringEnv("AWS_S3_PREFIX", "transactions"),

		// Lambda
		LambdaEnabled:     getBoolEnv("AWS_LAMBDA_ENABLED", true),
		LambdaRegion:      getStringEnv("AWS_LAMBDA_REGION", "us-east-1"),
		LambdaMemorySize:  getIntEnv("AWS_LAMBDA_MEMORY_SIZE", 512),
		LambdaTimeout:    getDurationEnv("AWS_LAMBDA_TIMEOUT", 30*time.Second),
		LambdaConcurrency: getIntEnv("AWS_LAMBDA_CONCURRENCY", 100),

		// SageMaker
		SageMakerEnabled:  getBoolEnv("AWS_SAGEMAKER_ENABLED", true),
		SageMakerRegion:   getStringEnv("AWS_SAGEMAKER_REGION", "us-east-1"),
		SageMakerEndpoint: getStringEnv("AWS_SAGEMAKER_ENDPOINT", ""),
		SageMakerModel:    getStringEnv("AWS_SAGEMAKER_MODEL", "rldb-fraud-detection-model"),

		// SNS
		SNSEnabled:        getBoolEnv("AWS_SNS_ENABLED", true),
		SNSRegion:         getStringEnv("AWS_SNS_REGION", "us-east-1"),
		SNSFraudAlertArn:  getStringEnv("AWS_SNS_FRAUD_ALERT_ARN", ""),
		SNSNotificationArn: getStringEnv("AWS_SNS_NOTIFICATION_ARN", ""),

		// CloudWatch
		CloudWatchEnabled: getBoolEnv("AWS_CLOUDWATCH_ENABLED", true),
		CloudWatchRegion:  getStringEnv("AWS_CLOUDWATCH_REGION", "us-east-1"),
		CloudWatchLogGroup: getStringEnv("AWS_CLOUDWATCH_LOG_GROUP", "/rldb/transactions"),

		// SQS
		SQSEnabled:          getBoolEnv("AWS_SQS_ENABLED", true),
		SQSRegion:           getStringEnv("AWS_SQS_REGION", "us-east-1"),
		SQSQueueName:        getStringEnv("AWS_SQS_QUEUE_NAME", "rldb-processing-queue"),
		SQSMaxReceiveCount:  getIntEnv("AWS_SQS_MAX_RECEIVE_COUNT", 3),
		SQSVisibilityTimeout: getIntEnv("AWS_SQS_VISIBILITY_TIMEOUT", 30),

		// IA/ML Configuration
		AIFraudDetectionEnabled:  getBoolEnv("AI_FRAUD_DETECTION_ENABLED", true),
		AIDemandForecastEnabled:   getBoolEnv("AI_DEMAND_FORECAST_ENABLED", true),
		AIRouteOptimizationEnabled: getBoolEnv("AI_ROUTE_OPTIMIZATION_ENABLED", true),
		AIFraudModelThreshold:     getFloat64Env("AI_FRAUD_MODEL_THRESHOLD", 0.7),
		AIDemandModelThreshold:    getFloat64Env("AI_DEMAND_MODEL_THRESHOLD", 0.8),

		// Real-time Processing
		RealTimeProcessingEnabled: getBoolEnv("REAL_TIME_PROCESSING_ENABLED", true),
		BatchProcessingEnabled:     getBoolEnv("BATCH_PROCESSING_ENABLED", false),
		BatchSize:                  getIntEnv("BATCH_SIZE", 100),
		BatchInterval:              getStringEnv("BATCH_INTERVAL", "5m"),
	}, nil
}

// Validate - Valida configuracao
def (c *AWSConfig) Validate() error {
	// Validacoes basicas
	if c.KinesisEnabled && c.KinesisStreamName == "" {
		return &ConfigError{Field: "KinesisStreamName", Message: "Kinesis stream name cannot be empty when enabled"}
	}
	if c.DynamoDBEnabled && c.DynamoDBTableName == "" {
		return &ConfigError{Field: "DynamoDBTableName", Message: "DynamoDB table name cannot be empty when enabled"}
	}
	if c.S3Enabled && c.S3Bucket == "" {
		return &ConfigError{Field: "S3Bucket", Message: "S3 bucket cannot be empty when enabled"}
	}
	return nil
}

// ConfigError - Erro de configuracao
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}

// Helper functions para ler variaveis de ambiente
func getStringEnv(key, defaultValue string) string {
	if value, exists := getEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := getEnv(key); exists {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value, exists := getEnv(key); exists {
		var result int64
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := getEnv(key); exists {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getFloat64Env(key string, defaultValue float64) float64 {
	if value, exists := getEnv(key); exists {
		var result float64
		if _, err := fmt.Sscanf(value, "%f", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := getEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnv(key string) (string, bool) {
	// Implementacao padrao
	// Em producao, usar os.Getenv ou similar
	return "", false
}

// Inicializar funcoes helper
import (
	"fmt"
	"os"
)

func getEnv(key string) (string, bool) {
	value, exists := os.LookupEnv(key)
	return value, exists
}
