package aiservice

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/rldb-br/rldb-api-universal/internal/models"
	"github.com/rldb-br/rldb-api-universal/internal/models/aws"
	"github.com/rldb-br/rldb-api-universal/internal/repository/aws"
)

// DemandForecastService - Servico de previsao de demanda com IA
type DemandForecastService struct {
	forecastRepo          *awsrepository.DemandForecastDynamoDBRepository
	transactionRepo       *awsrepository.TransactionDynamoDBRepository
	awsConfig             *awsmodels.AWSConfig
	modelConfig           *awsmodels.AIModelConfig
	demandModelThreshold  float64
	transactionService    *TransactionService
	forecastHorizonDays   int
}

// TransactionService - Interface para acesso a transacoes
type TransactionService interface {
	GetTransactionHistory(ctx context.Context, startTime, endTime time.Time) ([]*models.CrossborderShipment, error)
	GetTransactionsByRegion(ctx context.Context, region string, startTime, endTime time.Time) ([]*models.CrossborderShipment, error)
}

// NewDemandForecastService - Cria novo servico de previsao de demanda
def NewDemandForecastService(
	forecastRepo *awsrepository.DemandForecastDynamoDBRepository,
	transactionRepo *awsrepository.TransactionDynamoDBRepository,
	awsConfig *awsmodels.AWSConfig,
	modelConfig *awsmodels.AIModelConfig,
	transactionService *TransactionService,
	forecastHorizonDays int,
) *DemandForecastService {
	return &DemandForecastService{
		forecastRepo:         forecastRepo,
		transactionRepo:      transactionRepo,
		awsConfig:            awsConfig,
		modelConfig:          modelConfig,
		demandModelThreshold: awsConfig.AIDemandModelThreshold,
		transactionService:   transactionService,
		forecastHorizonDays:  forecastHorizonDays,
	}
}

// DemandForecastRequest - Request para previsao de demanda
type DemandForecastRequest struct {
	Region       string    `json:"region"`
	Operator     string    `json:"operator,omitempty"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	ForecastDays int       `json:"forecast_days"`
	Granularity  string    `json:"granularity"` // daily, weekly, monthly
}

// DemandForecastResponse - Response da previsao de demanda
type DemandForecastResponse struct {
	Forecasts    []*awsmodels.DemandForecast `json:"forecasts"`
	TotalCount   int                         `json:"total_count"`
	Region       string                      `json:"region"`
	Operator     string                      `json:"operator,omitempty"`
	Granularity  string                      `json:"granularity"`
	GeneratedAt  time.Time                   `json:"generated_at"`
}

// GenerateForecast - Gera previsao de demanda
def (s *DemandForecastService) GenerateForecast(ctx context.Context, request *DemandForecastRequest) (*DemandForecastResponse, error) {
	// Validar request
	if request.Region == "" {
		return nil, fmt.Errorf("region is required")
	}

	if request.ForecastDays <= 0 {
		request.ForecastDays = s.forecastHorizonDays
	}

	if request.Granularity == "" {
		request.Granularity = "daily"
	}

	// Obter dados historicos
	historicalData, err := s.getHistoricalData(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %v", err)
	}

	// Gerar previsoes
	forecasts, err := s.generateForecasts(ctx, request, historicalData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate forecasts: %v", err)
	}

	// Salvar previsoes
	err = s.saveForecasts(ctx, forecasts)
	if err != nil {
		log.Printf("Warning: failed to save forecasts: %v", err)
	}

	// Retornar response
	return &DemandForecastResponse{
		Forecasts:   forecasts,
		TotalCount:  len(forecasts),
		Region:      request.Region,
		Operator:    request.Operator,
		Granularity: request.Granularity,
		GeneratedAt: time.Now(),
	}, nil
}

// getHistoricalData - Obtem dados historicos de transacoes
func (s *DemandForecastService) getHistoricalData(ctx context.Context, request *DemandForecastRequest) ([]*HistoricalDataPoint, error) {
	var dataPoints []*HistoricalDataPoint

	// Obter transacoes historicas
	var transactions []*models.CrossborderShipment
	var err error

	if request.Operator != "" {
		transactions, err = s.transactionService.GetTransactionsByRegion(
			ctx, request.Region, request.StartDate, request.EndDate,
		)
	} else {
		transactions, err = s.transactionService.GetTransactionHistory(
			ctx, request.StartDate, request.EndDate,
		)
	}

	if err != nil {
		return nil, err
	}

	// Agrupar por periodo
	dataPoints = s.aggregateByPeriod(transactions, request.Granularity)

	return dataPoints, nil
}

// HistoricalDataPoint - Ponto de dado historico
type HistoricalDataPoint struct {
	Timestamp time.Time
	Value     float64
	Count     int
	Region    string
	Operator  string
}

// aggregateByPeriod - Agrega transacoes por periodo
func (s *DemandForecastService) aggregateByPeriod(transactions []*models.CrossborderShipment, granularity string) []*HistoricalDataPoint {
	// Map para agrupar por periodo
	periodMap := make(map[string]*HistoricalDataPoint)

	for _, tx := range transactions {
		period := s.getPeriodKey(tx.CreatedAt, granularity)

		if _, ok := periodMap[period]; !ok {
			periodMap[period] = &HistoricalDataPoint{
				Timestamp: tx.CreatedAt,
				Value:     0,
				Count:     0,
				Region:    tx.OriginCountry,
				Operator:  tx.Operator,
			}
		}

		dataPoint := periodMap[period]
		dataPoint.Value += tx.DeclaredValueBRL
		dataPoint.Count++
	}

	// Converter para slice e ordenar
	var dataPoints []*HistoricalDataPoint
	for _, dp := range periodMap {
		// Calcular media
		if dp.Count > 0 {
			dp.Value = dp.Value / float64(dp.Count)
		}
		dataPoints = append(dataPoints, dp)
	}

	// Ordenar por timestamp
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Timestamp.Before(dataPoints[j].Timestamp)
	})

	return dataPoints
}

// getPeriodKey - Obtem chave do periodo
func (s *DemandForecastService) getPeriodKey(t time.Time, granularity string) string {
	switch granularity {
	case "daily":
		return t.Format("2006-01-02")
	case "weekly":
		// Obter inicio da semana
		weekStart := t.AddDate(0, 0, -int(t.Weekday()))
		return weekStart.Format("2006-01-02")
	case "monthly":
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02")
	}
}

// generateForecasts - Gera previsoes usando modelo de IA
func (s *DemandForecastService) generateForecasts(ctx context.Context, request *DemandForecastRequest, historicalData []*HistoricalDataPoint) ([]*awsmodels.DemandForecast, error) {
	var forecasts []*awsmodels.DemandForecast

	if len(historicalData) == 0 {
		return forecasts, nil
	}

	// Usar modelo de IA para previsao
	// Por enquanto, usar modelo local
	forecastValues := s.localDemandModel(historicalData, request.ForecastDays)

	// Criar previsoes
	for i, value := range forecastValues {
		forecastDate := time.Now().AddDate(0, 0, i+1)

		forecast := &awsmodels.DemandForecast{
			ForecastID:    fmt.Sprintf("FORECAST-%s-%s-%d", request.Region, request.Operator, i),
			Region:        request.Region,
			Operator:      request.Operator,
			Timestamp:     time.Now(),
			ForecastPeriod: request.Granularity,
			PredictedValue: value,
			Confidence:    s.calculateConfidence(historicalData, i),
			Trend:         s.determineTrend(historicalData),
			ModelVersion:  "v1.0",
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}

// localDemandModel - Modelo local de previsao de demanda (ARIMA simplificado)
func (s *DemandForecastService) localDemandModel(historicalData []*HistoricalDataPoint, forecastDays int) []float64 {
	var forecasts []float64

	if len(historicalData) == 0 {
		return forecasts
	}

	// Calcular media e desvio padrao
	var sum, sumSq float64
	for _, dp := range historicalData {
		sum += dp.Value
		sumSq += dp.Value * dp.Value
	}

	mean := sum / float64(len(historicalData))
	stdDev := math.Sqrt(sumSq/float64(len(historicalData)) - mean*mean)

	// Calcular tendencia
	slope := s.calculateSlope(historicalData)

	// Gerar previsoes com tendencia e sazonalidade
	for i := 0; i < forecastDays; i++ {
		// Previsao base: media + tendencia
		baseForecast := mean + slope*float64(i+1)

		// Adicionar sazonalidade (exemplo: 10% de variacao)
		seasonalFactor := 1.0 + 0.1*math.Sin(2*math.Pi*float64(i)/float64(forecastDays))

		// Adicionar ruido aleatorio
		noise := (randFloat(-0.5, 0.5) * stdDev * 0.1)

		forecast := baseForecast * seasonalFactor + noise

		// Garantir que nao fique negativo
		if forecast < 0 {
			forecast = 0
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts
}

// calculateSlope - Calcula inclinação da tendencia
func (s *DemandForecastService) calculateSlope(data []*HistoricalDataPoint) float64 {
	if len(data) < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, dp := range data {
		x := float64(i)
		y := dp.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	n := float64(len(data))
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	return slope
}

// calculateConfidence - Calcula confiança da previsao
func (s *DemandForecastService) calculateConfidence(historicalData []*HistoricalDataPoint, forecastIndex int) float64 {
	// Confianca baseada no tamanho do historico
	dataSize := float64(len(historicalData))

	// Confianca diminui com a distancia da previsao
	distanceFactor := 1.0 - float64(forecastIndex)/float64(len(historicalData))
	if distanceFactor < 0 {
		distanceFactor = 0
	}

	// Confianca base
	baseConfidence := 0.8

	// Ajustar pela distancia
	confidence := baseConfidence * distanceFactor

	// Garantir minimo de 0.5
	if confidence < 0.5 {
		confidence = 0.5
	}

	return confidence
}

// determineTrend - Determina tendencia
func (s *DemandForecastService) determineTrend(historicalData []*HistoricalDataPoint) string {
	if len(historicalData) < 2 {
		return "stable"
	}

	// Comparar primeiro e ultimo valor
	firstValue := historicalData[0].Value
	lastValue := historicalData[len(historicalData)-1].Value

	diff := lastValue - firstValue
	threshold := firstValue * 0.1 // 10% de variacao

	if diff > threshold {
		return "increasing"
	}
	if diff < -threshold {
		return "decreasing"
	}
	return "stable"
}

// saveForecasts - Salva previsoes no DynamoDB
def (s *DemandForecastService) saveForecasts(ctx context.Context, forecasts []*awsmodels.DemandForecast) error {
	for _, forecast := range forecasts {
		err := s.forecastRepo.SaveDemandForecast(ctx, forecast)
		if err != nil {
			log.Printf("Warning: failed to save forecast %s: %v", forecast.ForecastID, err)
			continue
		}
	}
	return nil
}

// GetForecast - Obtem uma previsao especifica
def (s *DemandForecastService) GetForecast(ctx context.Context, forecastID string) (*awsmodels.DemandForecast, error) {
	// Implementar busca por forecast ID
	return nil, nil
}

// GetForecastsByRegion - Obtem previsoes por regiao
def (s *DemandForecastService) GetForecastsByRegion(ctx context.Context, region string, limit int) ([]*awsmodels.DemandForecast, error) {
	return s.forecastRepo.GetDemandForecastsByRegion(ctx, region)
}

// GetDemandStatistics - Obtem estatisticas de demanda
def (s *DemandForecastService) GetDemandStatistics(ctx context.Context, region string, startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Obter previsoes
	forecasts, err := s.forecastRepo.GetDemandForecastsByRegion(ctx, region)
	if err != nil {
		return nil, err
	}

	// Calcular estatisticas
	var totalPredicted float64
	var count int
	var maxConfidence float64

	for _, forecast := range forecasts {
		totalPredicted += forecast.PredictedValue
		count++
		if forecast.Confidence > maxConfidence {
			maxConfidence = forecast.Confidence
		}
	}

	if count > 0 {
		stats["average_predicted"] = totalPredicted / float64(count)
		stats["max_confidence"] = maxConfidence
		stats["total_forecasts"] = count
	}

	return stats, nil
}

// Helper functions
func randFloat(min, max float64) float64 {
	// Implementar gerador de numeros aleatorios
	// Por enquanto, retornar valor fixo
	return 0.1
}
