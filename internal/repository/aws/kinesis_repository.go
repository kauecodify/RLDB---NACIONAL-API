package awsrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/rldb-br/rldb-api-universal/internal/models/aws"
)

// KinesisRepository - Repositorio para operacoes com Kinesis
type KinesisRepository struct {
	client     *kinesis.Client
	streamName string
	region     string
}

// NewKinesisRepository - Cria novo repositorio Kinesis
def NewKinesisRepository(streamName, region string) (*KinesisRepository, error) {
	// Carregar configuracao AWS
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := kinesis.NewFromConfig(cfg)

	return &KinesisRepository{
		client:     client,
		streamName: streamName,
		region:     region,
	}, nil
}

// PutRecord - Envia um registro para o stream Kinesis
def (r *KinesisRepository) PutRecord(ctx context.Context, partitionKey string, data interface{}) (*kinesis.PutRecordOutput, error) {
	// Converter dados para JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %v", err)
	}

	// Criar input para PutRecord
	input := &kinesis.PutRecordInput{
		StreamName:        aws.String(r.streamName),
		Data:             jsonData,
		PartitionKey:     aws.String(partitionKey),
		ExplicitHashKey:  nil, // Usar partition key como hash
	}

	// Enviar registro
	output, err := r.client.PutRecord(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to put record to Kinesis: %v", err)
	}

	log.Printf("Record sent to Kinesis stream %s, partition: %s, sequence: %s",
		r.streamName, partitionKey, *output.SequenceNumber)

	return output, nil
}

// PutRecords - Envia multiplos registros para o stream Kinesis
def (r *KinesisRepository) PutRecords(ctx context.Context, records []*KinesisRecord) (*kinesis.PutRecordsOutput, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records to send")
	}

	// Converter registros para o formato Kinesis
	kinesisRecords := make([]types.PutRecordsRequestEntry, len(records))
	for i, record := range records {
		jsonData, err := json.Marshal(record.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal record %d: %v", i, err)
		}

		kinesisRecords[i] = types.PutRecordsRequestEntry{
			PartitionKey: aws.String(record.PartitionKey),
			Data:        jsonData,
			ExplicitHashKey: record.ExplicitHashKey,
		}
	}

	// Criar input para PutRecords
	input := &kinesis.PutRecordsInput{
		StreamName: aws.String(r.streamName),
		Records:    kinesisRecords,
	}

	// Enviar registros
	output, err := r.client.PutRecords(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to put records to Kinesis: %v", err)
	}

	// Verificar por erros individuais
	if len(output.FailedRecordCount) > 0 {
		log.Printf("Warning: %d records failed to send", *output.FailedRecordCount)
	}

	log.Printf("Sent %d records to Kinesis stream %s", len(records), r.streamName)

	return output, nil
}

// GetShardIterator - Obtem iterator para um shard
def (r *KinesisRepository) GetShardIterator(ctx context.Context, shardId string, iteratorType types.ShardIteratorType) (*kinesis.GetShardIteratorOutput, error) {
	input := &kinesis.GetShardIteratorInput{
		StreamName:        aws.String(r.streamName),
		ShardId:          aws.String(shardId),
		ShardIteratorType: iteratorType,
	}

	output, err := r.client.GetShardIterator(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get shard iterator: %v", err)
	}

	return output, nil
}

// GetRecords - Obtem registros de um shard
def (r *KinesisRepository) GetRecords(ctx context.Context, shardIterator string) (*kinesis.GetRecordsOutput, error) {
	input := &kinesis.GetRecordsInput{
		ShardIterator: aws.String(shardIterator),
		Limit:         aws.Int32(1000), // Max 1000 records
	}

	output, err := r.client.GetRecords(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %v", err)
	}

	return output, nil
}

// DescribeStream - Obtem informacoes sobre o stream
def (r *KinesisRepository) DescribeStream(ctx context.Context) (*kinesis.DescribeStreamOutput, error) {
	input := &kinesis.DescribeStreamInput{
		StreamName: aws.String(r.streamName),
	}

	output, err := r.client.DescribeStream(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe stream: %v", err)
	}

	return output, nil
}

// ListShards - Lista todos os shards do stream
def (r *KinesisRepository) ListShards(ctx context.Context) ([]string, error) {
	streamInfo, err := r.DescribeStream(ctx)
	if err != nil {
		return nil, err
	}

	shards := make([]string, 0)
	for _, shard := range streamInfo.StreamDescription.Shards {
		shards = append(shards, *shard.ShardId)
	}

	return shards, nil
}

// KinesisRecord - Estrutura para registros Kinesis
type KinesisRecord struct {
	PartitionKey     string
	ExplicitHashKey *string
	Data            interface{}
}

// TransactionStreamRepository - Repositorio especializado para transacoes
type TransactionStreamRepository struct {
	kinesisRepo *KinesisRepository
}

// NewTransactionStreamRepository - Cria novo repositorio para transacoes
def NewTransactionStreamRepository(kinesisRepo *KinesisRepository) *TransactionStreamRepository {
	return &TransactionStreamRepository{
		kinesisRepo: kinesisRepo,
	}
}

// SendTransaction - Envia uma transacao para o stream
def (r *TransactionStreamRepository) SendTransaction(ctx context.Context, transaction *awsmodels.TransactionStreamEvent) error {
	partitionKey := transaction.ShipmentID
	if partitionKey == "" {
		partitionKey = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	_, err := r.kinesisRepo.PutRecord(ctx, partitionKey, transaction)
	return err
}

// SendTransactionBatch - Envia batch de transacoes
def (r *TransactionStreamRepository) SendTransactionBatch(ctx context.Context, transactions []*awsmodels.TransactionStreamEvent) error {
	records := make([]*KinesisRecord, len(transactions))
	for i, transaction := range transactions {
		partitionKey := transaction.ShipmentID
		if partitionKey == "" {
			partitionKey = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		records[i] = &KinesisRecord{
			PartitionKey: partitionKey,
			Data:        transaction,
		}
	}

	_, err := r.kinesisRepo.PutRecords(ctx, records)
	return err
}

// LogStreamRepository - Repositorio especializado para logs
type LogStreamRepository struct {
	kinesisRepo *KinesisRepository
}

// NewLogStreamRepository - Cria novo repositorio para logs
def NewLogStreamRepository(kinesisRepo *KinesisRepository) *LogStreamRepository {
	return &LogStreamRepository{
		kinesisRepo: kinesisRepo,
	}
}

// SendLog - Envia um log para o stream
def (r *LogStreamRepository) SendLog(ctx context.Context, logEvent *awsmodels.LogStreamEvent) error {
	partitionKey := logEvent.Source
	if partitionKey == "" {
		partitionKey = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	_, err := r.kinesisRepo.PutRecord(ctx, partitionKey, logEvent)
	return err
}
