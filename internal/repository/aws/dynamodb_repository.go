package awsrepository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rldb-br/rldb-api-universal/internal/models/aws"
)

// DynamoDBRepository - Repositorio para operacoes com DynamoDB
type DynamoDBRepository struct {
	client    *dynamodb.Client
	tableName string
	region    string
}

// NewDynamoDBRepository - Cria novo repositorio DynamoDB
def NewDynamoDBRepository(tableName, region string) (*DynamoDBRepository, error) {
	// Carregar configuracao AWS
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDBRepository{
		client:    client,
		tableName: tableName,
		region:    region,
	}, nil
}

// PutItem - Insere um item no DynamoDB
def (r *DynamoDBRepository) PutItem(ctx context.Context, item interface{}, pk, sk string) error {
	// Converter item para mapa de attribute values
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %v", err)
	}

	// Criar key
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}

	// Adicionar GSI1
	if gsi1pk, ok := av["GSI1PK"]; ok {
		key["GSI1PK"] = gsi1pk
	}
	if gsi1sk, ok := av["GSI1SK"]; ok {
		key["GSI1SK"] = gsi1sk
	}

	// Criar input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	// Executar PutItem
	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item: %v", err)
	}

	return nil
}

// GetItem - Obtem um item do DynamoDB
def (r *DynamoDBRepository) GetItem(ctx context.Context, pk, sk string, result interface{}) error {
	// Criar key
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}

	// Criar input
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	}

	// Executar GetItem
	output, err := r.client.GetItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get item: %v", err)
	}

	if output.Item == nil {
		return fmt.Errorf("item not found")
	}

	// Converter para struct
	err = attributevalue.UnmarshalMap(output.Item, result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal item: %v", err)
	}

	return nil
}

// UpdateItem - Atualiza um item no DynamoDB
def (r *DynamoDBRepository) UpdateItem(ctx context.Context, pk, sk string, updateExpr string, exprAttrValues map[string]types.AttributeValue) error {
	// Criar key
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}

	// Criar input
	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.tableName),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprAttrValues,
		ReturnValues:              types.ReturnValueAllNew,
	}

	// Executar UpdateItem
	_, err := r.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update item: %v", err)
	}

	return nil
}

// DeleteItem - Remove um item do DynamoDB
def (r *DynamoDBRepository) DeleteItem(ctx context.Context, pk, sk string) error {
	// Criar key
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}

	// Criar input
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	}

	// Executar DeleteItem
	_, err := r.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete item: %v", err)
	}

	return nil
}

// QueryItems - Consulta itens no DynamoDB
def (r *DynamoDBRepository) QueryItems(ctx context.Context, pk string, result interface{}) ([]interface{}, error) {
	// Criar key condition
	keyCondition := "PK = :pk"
	exprAttrValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: pk},
	}

	// Criar input
	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: exprAttrValues,
		Limit:                     aws.Int32(1000),
	}

	// Executar Query
	output, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %v", err)
	}

	// Converter resultados
	var results []interface{}
	for _, item := range output.Items {
		var resultItem map[string]interface{}
		err = attributevalue.UnmarshalMap(item, &resultItem)
		if err != nil {
			log.Printf("Warning: failed to unmarshal item: %v", err)
			continue
		}
		results = append(results, resultItem)
	}

	return results, nil
}

// TransactionDynamoDBRepository - Repositorio especializado para transacoes
type TransactionDynamoDBRepository struct {
	dynamoRepo *DynamoDBRepository
}

// NewTransactionDynamoDBRepository - Cria novo repositorio para transacoes
def NewTransactionDynamoDBRepository(dynamoRepo *DynamoDBRepository) *TransactionDynamoDBRepository {
	return &TransactionDynamoDBRepository{
		dynamoRepo: dynamoRepo,
	}
}

// SaveTransaction - Salva uma transacao no DynamoDB
def (r *TransactionDynamoDBRepository) SaveTransaction(ctx context.Context, transaction *awsmodels.TransactionStreamEvent) error {
	// Criar PK e SK
	pk := fmt.Sprintf("TRANSACTION#%s", transaction.ShipmentID)
	sk := fmt.Sprintf("TIMESTAMP#%d", transaction.Timestamp.UnixNano())

	// Adicionar metadados para GSI
	transaction.GSI1PK = aws.String(fmt.Sprintf("OPERATOR#%s", transaction.Operator))
	transaction.GSI1SK = aws.String(fmt.Sprintf("STATUS#%s#%d", transaction.Status, transaction.Timestamp.UnixNano()))

	return r.dynamoRepo.PutItem(ctx, transaction, pk, sk)
}

// GetTransaction - Obtem uma transacao do DynamoDB
def (r *TransactionDynamoDBRepository) GetTransaction(ctx context.Context, shipmentID string, timestamp time.Time) (*awsmodels.TransactionStreamEvent, error) {
	pk := fmt.Sprintf("TRANSACTION#%s", shipmentID)
	sk := fmt.Sprintf("TIMESTAMP#%d", timestamp.UnixNano())

	var result awsmodels.TransactionStreamEvent
	err := r.dynamoRepo.GetItem(ctx, pk, sk, &result)
	return &result, err
}

// GetTransactionsByShipment - Obtem todas as transacoes de um shipment
def (r *TransactionDynamoDBRepository) GetTransactionsByShipment(ctx context.Context, shipmentID string) ([]*awsmodels.TransactionStreamEvent, error) {
	pk := fmt.Sprintf("TRANSACTION#%s", shipmentID)

	items, err := r.dynamoRepo.QueryItems(ctx, pk, &awsmodels.TransactionStreamEvent{})
	if err != nil {
		return nil, err
	}

	// Converter para slice de TransactionStreamEvent
	var transactions []*awsmodels.TransactionStreamEvent
	for _, item := range items {
		if tx, ok := item.(*awsmodels.TransactionStreamEvent); ok {
			transactions = append(transactions, tx)
		}
	}

	return transactions, nil
}

// UpdateTransaction - Atualiza uma transacao
def (r *TransactionDynamoDBRepository) UpdateTransaction(ctx context.Context, transaction *awsmodels.TransactionStreamEvent) error {
	pk := fmt.Sprintf("TRANSACTION#%s", transaction.ShipmentID)
	sk := fmt.Sprintf("TIMESTAMP#%d", transaction.Timestamp.UnixNano())

	// Criar expressao de update
	updateExpr := "SET #processed = :processed, #processed_at = :processed_at"
	exprAttrNames := map[string]string{
		"#processed":      "processed",
		"#processed_at":  "processed_at",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":processed":     &types.AttributeValueMemberBOOL{Value: transaction.Processed},
		":processed_at": &types.AttributeValueMemberS{Value: transaction.ProcessedAt.Format(time.RFC3339)},
	}

	return r.dynamoRepo.UpdateItem(ctx, pk, sk, updateExpr, exprAttrValues)
}

// FraudDetectionDynamoDBRepository - Repositorio especializado para deteccao de fraude
type FraudDetectionDynamoDBRepository struct {
	dynamoRepo *DynamoDBRepository
}

// NewFraudDetectionDynamoDBRepository - Cria novo repositorio para deteccao de fraude
def NewFraudDetectionDynamoDBRepository(dynamoRepo *DynamoDBRepository) *FraudDetectionDynamoDBRepository {
	return &FraudDetectionDynamoDBRepository{
		dynamoRepo: dynamoRepo,
	}
}

// SaveFraudDetection - Salva resultado de deteccao de fraude
def (r *FraudDetectionDynamoDBRepository) SaveFraudDetection(ctx context.Context, detection *awsmodels.FraudDetectionResult) error {
	pk := fmt.Sprintf("FRAUD#%s", detection.DetectionID)
	sk := fmt.Sprintf("SHIPMENT#%s", detection.ShipmentID)

	// Adicionar metadados para GSI
	detection.GSI1PK = aws.String(fmt.Sprintf("RISK_LEVEL#%s", detection.RiskLevel))
	detection.GSI1SK = aws.String(fmt.Sprintf("TIMESTAMP#%d", detection.Timestamp.UnixNano()))

	return r.dynamoRepo.PutItem(ctx, detection, pk, sk)
}

// GetFraudDetectionsByRiskLevel - Obtem deteccoes por nivel de risco
def (r *FraudDetectionDynamoDBRepository) GetFraudDetectionsByRiskLevel(ctx context.Context, riskLevel string) ([]*awsmodels.FraudDetectionResult, error) {
	pk := fmt.Sprintf("RISK_LEVEL#%s", riskLevel)

	items, err := r.dynamoRepo.QueryItems(ctx, pk, &awsmodels.FraudDetectionResult{})
	if err != nil {
		return nil, err
	}

	var detections []*awsmodels.FraudDetectionResult
	for _, item := range items {
		if detection, ok := item.(*awsmodels.FraudDetectionResult); ok {
			detections = append(detections, detection)
		}
	}

	return detections, nil
}

// DemandForecastDynamoDBRepository - Repositorio especializado para previsao de demanda
type DemandForecastDynamoDBRepository struct {
	dynamoRepo *DynamoDBRepository
}

// NewDemandForecastDynamoDBRepository - Cria novo repositorio para previsao de demanda
def NewDemandForecastDynamoDBRepository(dynamoRepo *DynamoDBRepository) *DemandForecastDynamoDBRepository {
	return &DemandForecastDynamoDBRepository{
		dynamoRepo: dynamoRepo,
	}
}

// SaveDemandForecast - Salva previsao de demanda
def (r *DemandForecastDynamoDBRepository) SaveDemandForecast(ctx context.Context, forecast *awsmodels.DemandForecast) error {
	pk := fmt.Sprintf("FORECAST#%s", forecast.ForecastID)
	sk := fmt.Sprintf("REGION#%s#%s", forecast.Region, forecast.Operator)

	// Adicionar metadados para GSI
	forecast.GSI1PK = aws.String(fmt.Sprintf("REGION#%s", forecast.Region))
	forecast.GSI1SK = aws.String(fmt.Sprintf("PERIOD#%s#%d", forecast.ForecastPeriod, forecast.Timestamp.UnixNano()))

	return r.dynamoRepo.PutItem(ctx, forecast, pk, sk)
}

// GetDemandForecastsByRegion - Obtem previsoes por regiao
def (r *DemandForecastDynamoDBRepository) GetDemandForecastsByRegion(ctx context.Context, region string) ([]*awsmodels.DemandForecast, error) {
	pk := fmt.Sprintf("REGION#%s", region)

	items, err := r.dynamoRepo.QueryItems(ctx, pk, &awsmodels.DemandForecast{})
	if err != nil {
		return nil, err
	}

	var forecasts []*awsmodels.DemandForecast
	for _, item := range items {
		if forecast, ok := item.(*awsmodels.DemandForecast); ok {
			forecasts = append(forecasts, forecast)
		}
	}

	return forecasts, nil
}

// RouteOptimizationDynamoDBRepository - Repositorio especializado para otimizacao de rotas
type RouteOptimizationDynamoDBRepository struct {
	dynamoRepo *DynamoDBRepository
}

// NewRouteOptimizationDynamoDBRepository - Cria novo repositorio para otimizacao de rotas
def NewRouteOptimizationDynamoDBRepository(dynamoRepo *DynamoDBRepository) *RouteOptimizationDynamoDBRepository {
	return &RouteOptimizationDynamoDBRepository{
		dynamoRepo: dynamoRepo,
	}
}

// SaveRouteOptimization - Salva otimizacao de rota
def (r *RouteOptimizationDynamoDBRepository) SaveRouteOptimization(ctx context.Context, optimization *awsmodels.RouteOptimization) error {
	pk := fmt.Sprintf("ROUTE#%s", optimization.OptimizationID)
	sk := fmt.Sprintf("SHIPMENT#%s", optimization.ShipmentID)

	// Adicionar metadados para GSI
	optimization.GSI1PK = aws.String(fmt.Sprintf("SHIPMENT#%s", optimization.ShipmentID))
	optimization.GSI1SK = aws.String(fmt.Sprintf("TIMESTAMP#%d", optimization.Timestamp.UnixNano()))

	return r.dynamoRepo.PutItem(ctx, optimization, pk, sk)
}

// GetRouteOptimizationsByShipment - Obtem otimizacoes por shipment
def (r *RouteOptimizationDynamoDBRepository) GetRouteOptimizationsByShipment(ctx context.Context, shipmentID string) ([]*awsmodels.RouteOptimization, error) {
	pk := fmt.Sprintf("SHIPMENT#%s", shipmentID)

	items, err := r.dynamoRepo.QueryItems(ctx, pk, &awsmodels.RouteOptimization{})
	if err != nil {
		return nil, err
	}

	var optimizations []*awsmodels.RouteOptimization
	for _, item := range items {
		if optimization, ok := item.(*awsmodels.RouteOptimization); ok {
			optimizations = append(optimizations, optimization)
		}
	}

	return optimizations, nil
}
