package awsmodels

import (
	"time"
)

// AWSStreamConfig - Configuracao de streams AWS
type AWSStreamConfig struct {
	KinesisStreamName      string        `json:"kinesis_stream_name" yaml:"kinesis_stream_name"`
	KinesisShardCount      int           `json:"kinesis_shard_count" yaml:"kinesis_shard_count"`
	KinesisRetentionPeriod int           `json:"kinesis_retention_period" yaml:"kinesis_retention_period"`
	DynamoDBTableName     string        `json:"dynamodb_table_name" yaml:"dynamodb_table_name"`
	DynamoDBReadCapacity  int64         `json:"dynamodb_read_capacity" yaml:"dynamodb_read_capacity"`
	DynamoDBWriteCapacity int64         `json:"dynamodb_write_capacity" yaml:"dynamodb_write_capacity"`
	S3BucketName          string        `json:"s3_bucket_name" yaml:"s3_bucket_name"`
	S3Region              string        `json:"s3_region" yaml:"s3_region"`
	LambdaFunctionName    string        `json:"lambda_function_name" yaml:"lambda_function_name"`
	LambdaMemorySize      int           `json:"lambda_memory_size" yaml:"lambda_memory_size"`
	LambdaTimeout        time.Duration `json:"lambda_timeout" yaml:"lambda_timeout"`
	SageMakerEndpoint     string        `json:"sagemaker_endpoint" yaml:"sagemaker_endpoint"`
	SageMakerModelName    string        `json:"sagemaker_model_name" yaml:"sagemaker_model_name"`
	SNSFraudAlertTopic    string        `json:"sns_fraud_alert_topic" yaml:"sns_fraud_alert_topic"`
	CloudWatchLogGroup    string        `json:"cloudwatch_log_group" yaml:"cloudwatch_log_group"`
}

// TransactionStreamEvent - Evento de transacao no stream
type TransactionStreamEvent struct {
	EventID        string                 `json:"event_id"`
	EventType      string                 `json:"event_type"`
	Timestamp      time.Time              `json:"timestamp"`
	ShipmentID     string                 `json:"shipment_id"`
	TrackingCode   string                 `json:"tracking_code"`
	Operator       string                 `json:"operator"`
	Status         string                 `json:"status"`
	Location       *LocationData          `json:"location,omitempty"`
	Value          float64                `json:"value,omitempty"`
	WeightKG       float64                `json:"weight_kg,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Processed      bool                   `json:"processed"`
	ProcessedAt    time.Time              `json:"processed_at,omitempty"`
}

// LocationData - Dados de localizacao
type LocationData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city"`
	State     string  `json:"state"`
	Country   string  `json:"country"`
	CEP       string  `json:"cep,omitempty"`
}

// LogStreamEvent - Evento de log no stream
type LogStreamEvent struct {
	LogID        string                 `json:"log_id"`
	LogLevel     string                 `json:"log_level"`
	Timestamp    time.Time              `json:"timestamp"`
	Source       string                 `json:"source"`
	Message      string                 `json:"message"`
	ShipmentID   string                 `json:"shipment_id,omitempty"`
	Operator     string                 `json:"operator,omitempty"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Processed    bool                   `json:"processed"`
	ProcessedAt  time.Time              `json:"processed_at,omitempty"`
}

// FraudDetectionResult - Resultado de deteccao de fraude
type FraudDetectionResult struct {
	DetectionID    string                 `json:"detection_id"`
	ShipmentID     string                 `json:"shipment_id"`
	TrackingCode   string                 `json:"tracking_code"`
	Timestamp      time.Time              `json:"timestamp"`
	FraudScore     float64                `json:"fraud_score"`
	FraudType      string                 `json:"fraud_type"`
	Confidence     float64                `json:"confidence"`
	RiskLevel      string                 `json:"risk_level"` // low, medium, high, critical
	AlertSent      bool                   `json:"alert_sent"`
	ActionTaken    string                 `json:"action_taken,omitempty"` // block, flag, monitor, none
	Details        map[string]interface{} `json:"details,omitempty"`
}

// DemandForecast - Previsao de demanda
type DemandForecast struct {
	ForecastID     string    `json:"forecast_id"`
	Region         string    `json:"region"`
	Operator       string    `json:"operator"`
	Timestamp      time.Time `json:"timestamp"`
	ForecastPeriod string    `json:"forecast_period"` // daily, weekly, monthly
	PredictedValue  float64   `json:"predicted_value"`
	Confidence     float64   `json:"confidence"`
	Trend          string    `json:"trend"` // increasing, decreasing, stable
	Seasonality    string    `json:"seasonality,omitempty"`
	ModelVersion   string    `json:"model_version"`
}

// RouteOptimization - Otimizacao de rota
type RouteOptimization struct {
	OptimizationID string                 `json:"optimization_id"`
	ShipmentID     string                 `json:"shipment_id"`
	TrackingCode   string                 `json:"tracking_code"`
	Timestamp      time.Time              `json:"timestamp"`
	CurrentRoute   []LocationData         `json:"current_route"`
	OptimizedRoute []LocationData         `json:"optimized_route"`
	DistanceSaved  float64                `json:"distance_saved_km"`
	TimeSaved      float64                `json:"time_saved_hours"`
	CostSaved      float64                `json:"cost_saved_brl"`
	CO2Saved       float64                `json:"co2_saved_kg"`
	Constraints    []string               `json:"constraints,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// RealTimeMetrics - Metricas em tempo real
type RealTimeMetrics struct {
	MetricID       string    `json:"metric_id"`
	MetricType     string    `json:"metric_type"` // transactions, logs, fraud, demand, routes
	Value          float64   `json:"value"`
	Timestamp      time.Time `json:"timestamp"`
	Aggregation    string    `json:"aggregation"` // count, sum, avg, max, min
	Region         string    `json:"region,omitempty"`
	Operator       string    `json:"operator,omitempty"`
	Service        string    `json:"service,omitempty"`
}

// StreamProcessingResult - Resultado do processamento de stream
type StreamProcessingResult struct {
	ProcessID      string                 `json:"process_id"`
	StreamType     string                 `json:"stream_type"` // transactions, logs
	RecordsProcessed int                    `json:"records_processed"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	SuccessCount   int                    `json:"success_count"`
	ErrorCount     int                    `json:"error_count"`
	Errors         []StreamError          `json:"errors,omitempty"`
	Metrics        map[string]interface{} `json:"metrics,omitempty"`
}

// StreamError - Erro no processamento de stream
type StreamError struct {
	ErrorID    string    `json:"error_id"`
	RecordID   string    `json:"record_id"`
	ErrorType  string    `json:"error_type"`
	ErrorMsg   string    `json:"error_message"`
	Timestamp  time.Time `json:"timestamp"`
}

// AIModelConfig - Configuracao de modelos de IA
type AIModelConfig struct {
	ModelName        string                 `json:"model_name"`
	ModelVersion     string                 `json:"model_version"`
	ModelType        string                 `json:"model_type"` // fraud, demand, routing
	EndpointURL      string                 `json:"endpoint_url"`
	InferenceTimeout time.Duration          `json:"inference_timeout"`
	InputSchema      map[string]interface{} `json:"input_schema"`
	OutputSchema     map[string]interface{} `json:"output_schema"`
	Thresholds       map[string]float64    `json:"thresholds,omitempty"`
}
