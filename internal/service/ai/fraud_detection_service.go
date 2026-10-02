package aiservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/rldb-br/rldb-api-universal/internal/models"
	"github.com/rldb-br/rldb-api-universal/internal/models/aws"
	"github.com/rldb-br/rldb-api-universal/internal/repository/aws"
)

// FraudDetectionService - Servico de deteccao de fraude com IA
type FraudDetectionService struct {
	fraudRepo             *awsrepository.FraudDetectionDynamoDBRepository
	transactionRepo       *awsrepository.TransactionDynamoDBRepository
	awsConfig             *awsmodels.AWSConfig
	modelConfig           *awsmodels.AIModelConfig
	fraudModelThreshold   float64
	alertService          AlertService
	transactionService    *TransactionService
}

// AlertService - Interface para envio de alertas
type AlertService interface {
	SendFraudAlert(ctx context.Context, detection *awsmodels.FraudDetectionResult) error
}

// TransactionService - Interface para acesso a transacoes
type TransactionService interface {
	GetTransactionHistory(ctx context.Context, shipmentID string) ([]*models.CrossborderShipment, error)
}

// NewFraudDetectionService - Cria novo servico de deteccao de fraude
def NewFraudDetectionService(
	fraudRepo *awsrepository.FraudDetectionDynamoDBRepository,
	transactionRepo *awsrepository.TransactionDynamoDBRepository,
	awsConfig *awsmodels.AWSConfig,
	modelConfig *awsmodels.AIModelConfig,
	alertService AlertService,
	transactionService *TransactionService,
) *FraudDetectionService {
	return &FraudDetectionService{
		fraudRepo:           fraudRepo,
		transactionRepo:     transactionRepo,
		awsConfig:           awsConfig,
		modelConfig:         modelConfig,
		fraudModelThreshold: awsConfig.AIFraudModelThreshold,
		alertService:        alertService,
		transactionService:  transactionService,
	}
}

// FraudDetectionRequest - Request para deteccao de fraude
type FraudDetectionRequest struct {
	ShipmentID     string                 `json:"shipment_id"`
	TrackingCode   string                 `json:"tracking_code"`
	Operator       string                 `json:"operator"`
	Value          float64                `json:"value"`
	WeightKG       float64                `json:"weight_kg"`
	OriginCountry  string                 `json:"origin_country"`
	DestinationCountry string            `json:"destination_country"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// FraudDetectionResponse - Response da deteccao de fraude
type FraudDetectionResponse struct {
	DetectionID    string                 `json:"detection_id"`
	ShipmentID     string                 `json:"shipment_id"`
	FraudScore     float64                `json:"fraud_score"`
	FraudType      string                 `json:"fraud_type"`
	Confidence     float64                `json:"confidence"`
	RiskLevel      string                 `json:"risk_level"`
	ActionTaken    string                 `json:"action_taken"`
	AlertSent      bool                   `json:"alert_sent"`
	Details        map[string]interface{} `json:"details,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// DetectFraud - Detecta fraude em uma transacao
def (s *FraudDetectionService) DetectFraud(ctx context.Context, request *FraudDetectionRequest) (*FraudDetectionResponse, error) {
	// Validar request
	if request.ShipmentID == "" {
		return nil, fmt.Errorf("shipment_id is required")
	}

	// Gerar detection ID
	detectionID := fmt.Sprintf("FRAUD-%s-%d", request.ShipmentID, time.Now().UnixNano())

	// Obter historico de transacoes
	var fraudScore float64
	var fraudType string
	var confidence float64

	// Calcular score de fraude usando modelo de ML
	fraudScore, fraudType, confidence = s.calculateFraudScore(ctx, request)

	// Determinar nivel de risco
	riskLevel := s.determineRiskLevel(fraudScore)

	// Determinar acao a ser tomada
	actionTaken := s.determineAction(riskLevel, fraudScore)

	// Criar resultado da deteccao
	detection := &awsmodels.FraudDetectionResult{
		DetectionID:  detectionID,
		ShipmentID:   request.ShipmentID,
		TrackingCode: request.TrackingCode,
		Timestamp:    time.Now(),
		FraudScore:   fraudScore,
		FraudType:    fraudType,
		Confidence:   confidence,
		RiskLevel:    riskLevel,
		ActionTaken:  actionTaken,
		AlertSent:    false,
		Details:      request.Metadata,
	}

	// Salvar resultado
	err := s.fraudRepo.SaveFraudDetection(ctx, detection)
	if err != nil {
		log.Printf("Warning: failed to save fraud detection: %v", err)
	}

	// Enviar alerta se necessario
	if actionTaken != "none" {
		detection.AlertSent = true
		err = s.alertService.SendFraudAlert(ctx, detection)
		if err != nil {
			log.Printf("Warning: failed to send fraud alert: %v", err)
		}
	}

	// Atualizar resultado
	detection.AlertSent = true

	// Retornar response
	return &FraudDetectionResponse{
		DetectionID:  detectionID,
		ShipmentID:   request.ShipmentID,
		FraudScore:   fraudScore,
		FraudType:    fraudType,
		Confidence:   confidence,
		RiskLevel:    riskLevel,
		ActionTaken:  actionTaken,
		AlertSent:    detection.AlertSent,
		Details:      detection.Details,
		Timestamp:    detection.Timestamp,
	}, nil
}

// calculateFraudScore - Calcula score de fraude usando modelo de ML
func (s *FraudDetectionService) calculateFraudScore(ctx context.Context, request *FraudDetectionRequest) (float64, string, float64) {
	// Obter features para o modelo
	features := s.extractFeatures(request)

	// Chamar modelo de ML (SageMaker ou local)
	// Por enquanto, usar modelo local de exemplo
	score, fraudType, confidence := s.localFraudModel(features)

	return score, fraudType, confidence
}

// extractFeatures - Extrai features para o modelo de fraude
func (s *FraudDetectionService) extractFeatures(request *FraudDetectionRequest) map[string]float64 {
	features := make(map[string]float64)

	// Features basicas
	features["value"] = request.Value
	features["weight"] = request.WeightKG

	// Features de localizacao (exemplo: risco por pais)
	features["origin_risk"] = s.getCountryRiskScore(request.OriginCountry)
	features["destination_risk"] = s.getCountryRiskScore(request.DestinationCountry)

	// Features temporais (exemplo: hora do dia)
	features["hour_of_day"] = float64(time.Now().Hour())

	// Features de operador
	features["operator_risk"] = s.getOperatorRiskScore(request.Operator)

	// Features adicionais do metadata
	if metadata := request.Metadata; metadata != nil {
		if paymentMethod, ok := metadata["payment_method"].(string); ok {
			features["payment_method_risk"] = s.getPaymentMethodRisk(paymentMethod)
		}
		if isNewCustomer, ok := metadata["is_new_customer"].(bool); ok {
			if isNewCustomer {
				features["new_customer"] = 1.0
			} else {
				features["new_customer"] = 0.0
			}
		}
	}

	return features
}

// localFraudModel - Modelo local de deteccao de fraude (exemplo)
func (s *FraudDetectionService) localFraudModel(features map[string]float64) (float64, string, float64) {
	// Pesos para cada feature (exemplo)
	weights := map[string]float64{
		"value":                0.15,
		"weight":               0.10,
		"origin_risk":          0.20,
		"destination_risk":     0.20,
		"hour_of_day":          0.05,
		"operator_risk":        0.15,
		"payment_method_risk":  0.10,
		"new_customer":         0.05,
	}

	// Calcular score ponderado
	var score float64
	for feature, weight := range weights {
		if value, ok := features[feature]; ok {
			score += value * weight
		}
	}

	// Normalizar score para 0-1
	score = 1.0 / (1.0 + math.Exp(-score)) // Sigmoid

	// Determinar tipo de fraude
	fraudType := s.determineFraudType(features, score)

	// Confianca baseada na distancia do threshold
	confidence := 1.0 - math.Abs(score-0.5)*2

	return score, fraudType, confidence
}

// determineFraudType - Determina tipo de fraude
func (s *FraudDetectionService) determineFraudType(features map[string]float64, score float64) string {
	// Logica simplificada
	if features["value"] > 10000 {
		return "HIGH_VALUE_FRAUD"
	}
	if features["origin_risk"] > 0.8 && features["destination_risk"] > 0.8 {
		return "INTERNATIONAL_FRAUD"
	}
	if features["new_customer"] == 1.0 && score > 0.7 {
		return "NEW_CUSTOMER_FRAUD"
	}
	if features["payment_method_risk"] > 0.7 {
		return "PAYMENT_FRAUD"
	}
	return "GENERIC_FRAUD"
}

// determineRiskLevel - Determina nivel de risco
func (s *FraudDetectionService) determineRiskLevel(score float64) string {
	if score >= 0.9 {
		return "critical"
	}
	if score >= 0.7 {
		return "high"
	}
	if score >= 0.5 {
		return "medium"
	}
	return "low"
}

// determineAction - Determina acao a ser tomada
func (s *FraudDetectionService) determineAction(riskLevel string, score float64) string {
	switch riskLevel {
	case "critical":
		return "block"
	case "high":
		if score >= s.fraudModelThreshold {
			return "flag"
		}
		return "monitor"
	case "medium":
		return "monitor"
	default:
		return "none"
	}
}

// getCountryRiskScore - Retorna score de risco por pais
func (s *FraudDetectionService) getCountryRiskScore(country string) float64 {
	// Scores de exemplo (0-1, onde 1 e alto risco)
	riskScores := map[string]float64{
		"BR": 0.3,
		"US": 0.2,
		"CN": 0.6,
		"RU": 0.9,
		"NG": 0.8,
		"IN": 0.5,
	}
	if score, ok := riskScores[country]; ok {
		return score
	}
	return 0.4 // Default
}

// getOperatorRiskScore - Retorna score de risco por operador
func (s *FraudDetectionService) getOperatorRiskScore(operator string) float64 {
	// Scores de exemplo
	riskScores := map[string]float64{
		"correios":       0.2,
		"fedex":          0.1,
		"dhl":            0.1,
		"unknown":        0.8,
		"new_operator":   0.6,
	}
	if score, ok := riskScores[operator]; ok {
		return score
	}
	return 0.3 // Default
}

// getPaymentMethodRisk - Retorna score de risco por metodo de pagamento
func (s *FraudDetectionService) getPaymentMethodRisk(method string) float64 {
	// Scores de exemplo
	riskScores := map[string]float64{
		"credit_card":     0.4,
		"debit_card":     0.2,
		"pix":            0.3,
		"boleto":         0.6,
		"cash":           0.8,
		"cryptocurrency": 0.9,
	}
	if score, ok := riskScores[method]; ok {
		return score
	}
	return 0.5 // Default
}

// BatchFraudDetection - Detecta fraude em batch de transacoes
def (s *FraudDetectionService) BatchFraudDetection(ctx context.Context, requests []*FraudDetectionRequest) ([]*FraudDetectionResponse, error) {
	var responses []*FraudDetectionResponse

	for _, request := range requests {
		response, err := s.DetectFraud(ctx, request)
		if err != nil {
			log.Printf("Warning: failed to detect fraud for shipment %s: %v", request.ShipmentID, err)
			continue
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetFraudDetectionHistory - Obtem historico de deteccoes de fraude
def (s *FraudDetectionService) GetFraudDetectionHistory(ctx context.Context, shipmentID string) ([]*awsmodels.FraudDetectionResult, error) {
	// Implementar busca por shipment
	return nil, nil
}

// GetFraudStatistics - Obtem estatisticas de fraude
def (s *FraudDetectionService) GetFraudStatistics(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	// Implementar estatisticas
	stats := make(map[string]interface{})

	// Total de deteccoes
	// Por nivel de risco
	// Por tipo de fraude
	// Media de score

	return stats, nil
}

// SNSAlertService - Implementacao de alerta via SNS
type SNSAlertService struct {
	snsClient     interface{} // AWS SNS client
	alertTopicArn string
}

// SendFraudAlert - Envia alerta via SNS
def (s *SNSAlertService) SendFraudAlert(ctx context.Context, detection *awsmodels.FraudDetectionResult) error {
	// Converter para JSON
	jsonData, err := json.Marshal(detection)
	if err != nil {
		return fmt.Errorf("failed to marshal detection: %v", err)
	}

	// Enviar para SNS (implementar)
	log.Printf("Sending fraud alert to SNS: %s", string(jsonData))

	return nil
}

// LocalAlertService - Implementacao local de alerta (para desenvolvimento)
type LocalAlertService struct {
	alerts chan *awsmodels.FraudDetectionResult
}

// SendFraudAlert - Envia alerta localmente
def (s *LocalAlertService) SendFraudAlert(ctx context.Context, detection *awsmodels.FraudDetectionResult) error {
	select {
	case s.alerts <- detection:
		log.Printf("Fraud alert sent: %s", detection.DetectionID)
		return nil
	default:
		return fmt.Errorf("alert channel full")
	}
}

// NewLocalAlertService - Cria novo servico de alerta local
def NewLocalAlertService(bufferSize int) *LocalAlertService {
	return &LocalAlertService{
		alerts: make(chan *awsmodels.FraudDetectionResult, bufferSize),
	}
}

// GetAlerts - Obtem canal de alertas
def (s *LocalAlertService) GetAlerts() <-chan *awsmodels.FraudDetectionResult {
	return s.alerts
}
