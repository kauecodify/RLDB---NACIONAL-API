package repository

import (
	"context"
	"fmt"

	"github.com/rldb-br/rldb-api-universal/internal/models"
	"gorm.io/gorm"
)

// CrossborderTR8Repository handles database operations for TR8 crossborder
type CrossborderTR8Repository struct {
	db *gorm.DB
}

// NewCrossborderTR8Repository creates a new TR8 repository
func NewCrossborderTR8Repository(db *gorm.DB) *CrossborderTR8Repository {
	return &CrossborderTR8Repository{db: db}
}

// CreateShipment creates a new TR8 shipment
func (r *CrossborderTR8Repository) CreateShipment(ctx context.Context, shipment *models.CrossborderShipment) error {
	return r.db.WithContext(ctx).Create(shipment).Error
}

// GetShipmentByID retrieves a shipment by ID
func (r *CrossborderTR8Repository) GetShipmentByID(ctx context.Context, id string) (*models.CrossborderShipment, error) {
	var shipment models.CrossborderShipment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&shipment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("shipment not found")
		}
		return nil, err
	}
	return &shipment, nil
}

// GetShipmentByTrackingCode retrieves a shipment by tracking code
func (r *CrossborderTR8Repository) GetShipmentByTrackingCode(ctx context.Context, trackingCode string) (*models.CrossborderShipment, error) {
	var shipment models.CrossborderShipment
	if err := r.db.WithContext(ctx).Where("tracking_code = ?", trackingCode).First(&shipment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("shipment not found")
		}
		return nil, err
	}
	return &shipment, nil
}

// GetShipmentByIdempotencyKey retrieves a shipment by idempotency key
func (r *CrossborderTR8Repository) GetShipmentByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.CrossborderShipment, error) {
	var shipment models.CrossborderShipment
	if err := r.db.WithContext(ctx).Where("metadata->>$.idempotency_key = ?", idempotencyKey).First(&shipment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &shipment, nil
}

// UpdateShipment updates a shipment
func (r *CrossborderTR8Repository) UpdateShipment(ctx context.Context, shipment *models.CrossborderShipment) error {
	return r.db.WithContext(ctx).Save(shipment).Error
}

// CreateTrackingEvent creates a tracking event
func (r *CrossborderTR8Repository) CreateTrackingEvent(ctx context.Context, event *models.CrossborderTrackingEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetTrackingEvents retrieves all tracking events for a shipment
func (r *CrossborderTR8Repository) GetTrackingEvents(ctx context.Context, shipmentID string) ([]models.CrossborderTrackingEvent, error) {
	var events []models.CrossborderTrackingEvent
	if err := r.db.WithContext(ctx).Where("shipment_id = ?", shipmentID).Order("timestamp asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// CreateCustomsDocument creates a customs document
func (r *CrossborderTR8Repository) CreateCustomsDocument(ctx context.Context, doc *models.CrossborderCustomsDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// GetCustomsDocuments retrieves all customs documents for a shipment
func (r *CrossborderTR8Repository) GetCustomsDocuments(ctx context.Context, shipmentID string) ([]models.CrossborderCustomsDocument, error) {
	var docs []models.CrossborderCustomsDocument
	if err := r.db.WithContext(ctx).Where("shipment_id = ?", shipmentID).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// GetCustomsDocumentByID retrieves a customs document by ID
func (r *CrossborderTR8Repository) GetCustomsDocumentByID(ctx context.Context, id string) (*models.CrossborderCustomsDocument, error) {
	var doc models.CrossborderCustomsDocument
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document not found")
		}
		return nil, err
	}
	return &doc, nil
}

// UpdateCustomsDocument updates a customs document
func (r *CrossborderTR8Repository) UpdateCustomsDocument(ctx context.Context, doc *models.CrossborderCustomsDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

// GetTR8Metrics retrieves TR8 metrics for a period
func (r *CrossborderTR8Repository) GetTR8Metrics(ctx context.Context, startTime, endTime time.Time) (*models.TR8MetricsResponse, error) {
	// Count shipments
	var totalShipments int64
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&totalShipments).Error; err != nil {
		return nil, err
	}

	// Sum declared values
	var totalValueBRL float64
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("COALESCE(SUM(declared_value_brl), 0)").
		Scan(&totalValueBRL).Error; err != nil {
		return nil, err
	}

	// Average delivery days
	var avgDeliveryDays float64
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ? AND actual_delivery IS NOT NULL", startTime, endTime).
		Select("COALESCE(AVG(EXTRACT(DAY FROM (actual_delivery - created_at))), 0)").
		Scan(&avgDeliveryDays).Error; err != nil {
		return nil, err
	}

	// On-time delivery rate
	var onTimeCount, totalDelivered int64
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ? AND actual_delivery IS NOT NULL", startTime, endTime).
		Count(&totalDelivered).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ? AND actual_delivery IS NOT NULL AND actual_delivery <= estimated_delivery", startTime, endTime).
		Count(&onTimeCount).Error; err != nil {
		return nil, err
	}

	onTimeRate := float64(0)
	if totalDelivered > 0 {
		onTimeRate = float64(onTimeCount) / float64(totalDelivered)
	}

	// Average cost per kg
	var avgCostPerKg float64
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("COALESCE(AVG(declared_value_brl / NULLIF(weight_kg, 0)), 0)").
		Scan(&avgCostPerKg).Error; err != nil {
		return nil, err
	}

	// Top destinations
	var topDestinations []models.CountryMetric
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("destination_country as country, COUNT(*) as shipment_count, SUM(declared_value_brl) as value_brl, AVG(EXTRACT(DAY FROM (actual_delivery - created_at))) as avg_delivery_days").
		Group("destination_country").
		Order("shipment_count DESC").
		Limit(10).
		Scan(&topDestinations).Error; err != nil {
		return nil, err
	}

	// Top operators
	var topOperators []models.OperatorMetric
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("operator as operator, COUNT(*) as shipment_count, SUM(declared_value_brl) as value_brl, AVG(EXTRACT(DAY FROM (actual_delivery - created_at))) as avg_delivery_days").
		Group("operator").
		Order("shipment_count DESC").
		Limit(10).
		Scan(&topOperators).Error; err != nil {
		return nil, err
	}

	// Calculate CO2 per package
	var totalCO2, totalPackages float64
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderQuote{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("COALESCE(SUM(co2_kg), 0)").
		Scan(&totalCO2).Error; err != nil {
		return nil, err
	}
	totalPackages = float64(totalShipments)
	co2PerPackage := float64(0)
	if totalPackages > 0 {
		co2PerPackage = totalCO2 / totalPackages
	}

	// Customs clearance rate
	var customsCleared, totalWithCustoms int64
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ? AND metadata->>$.customs_status = 'verified'", startTime, endTime).
		Count(&customsCleared).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ? AND metadata->>$.customs_status IS NOT NULL", startTime, endTime).
		Count(&totalWithCustoms).Error; err != nil {
		return nil, err
	}

	customsClearanceRate := float64(0)
	if totalWithCustoms > 0 {
		customsClearanceRate = float64(customsCleared) / float64(totalWithCustoms)
	}

	// Reverse logistics rate
	var reverseCount int64
	if err := r.db.WithContext(ctx).
		Model(&models.CrossborderShipment{}).
		Where("created_at BETWEEN ? AND ? AND reverse_logistics_enabled = true", startTime, endTime).
		Count(&reverseCount).Error; err != nil {
		return nil, err
	}

	reverseLogisticsRate := float64(0)
	if totalShipments > 0 {
		reverseLogisticsRate = float64(reverseCount) / float64(totalShipments)
	}

	return &models.TR8MetricsResponse{
		Period:                fmt.Sprintf("%s/%s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		TotalShipments:        totalShipments,
		TotalValueBRL:         totalValueBRL,
		TotalValueUSD:         totalValueBRL / 5.0, // Approximate conversion
		AverageDeliveryDays:   avgDeliveryDays,
		OnTimeDeliveryRate:    onTimeRate,
		AverageCostPerKgBRL:   avgCostPerKg,
		AverageCostPerKgUSD:   avgCostPerKg / 5.0,
		CO2PerPackageKg:      co2PerPackage,
		CustomsClearanceRate:  customsClearanceRate,
		ReverseLogisticsRate: reverseLogisticsRate,
		TopDestinations:       topDestinations,
		TopOperators:          topOperators,
	}, nil
}

// GetTR8ComplianceReport generates a compliance report
func (r *CrossborderTR8Repository) GetTR8ComplianceReport(ctx context.Context, startTime, endTime time.Time) (*models.TR8ComplianceReport, error) {
	// Count total requests (from audit logs)
	var totalRequests int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Count(&totalRequests).Error; err != nil {
		return nil, err
	}

	// Count blocked requests
	var blockedRequests int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("timestamp BETWEEN ? AND ? AND status_code >= 400", startTime, endTime).
		Count(&blockedRequests).Error; err != nil {
		return nil, err
	}

	// Get operators
	var operators []models.OperatorCompliance
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Select("operator_id as operator, COUNT(*) as requests_made, SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as requests_blocked").
		Group("operator_id").
		Scan(&operators).Error; err != nil {
		return nil, err
	}

	// Calculate compliance scores
	for i := range operators {
		total := operators[i].RequestsMade
		blocked := operators[i].RequestsBlocked
		if total > 0 {
			operators[i].ComplianceScore = 1.0 - (float64(blocked) / float64(total))
		} else {
			operators[i].ComplianceScore = 1.0
		}
	}

	return &models.TR8ComplianceReport{
		Period:          fmt.Sprintf("%s/%s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		GeneratedAt:     time.Now().UTC(),
		TotalRequests:   totalRequests,
		BlockedRequests: blockedRequests,
		Operators:       operators,
	}, nil
}

// CreateQuote creates a quote record
func (r *CrossborderTR8Repository) CreateQuote(ctx context.Context, quote *models.CrossborderQuote) error {
	return r.db.WithContext(ctx).Create(quote).Error
}

// CreateWebhook creates a webhook configuration
func (r *CrossborderTR8Repository) CreateWebhook(ctx context.Context, webhook *models.CrossborderWebhook) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

// GetWebhooks retrieves webhooks for a client
func (r *CrossborderTR8Repository) GetWebhooks(ctx context.Context, clientID string) ([]models.CrossborderWebhook, error) {
	var webhooks []models.CrossborderWebhook
	if err := r.db.WithContext(ctx).Where("client_id = ? AND active = true", clientID).Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}

// GetAuditLogs retrieves audit logs for a period
func (r *CrossborderTR8Repository) GetAuditLogs(ctx context.Context, startTime, endTime time.Time) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Order("timestamp desc").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// CreateAuditLog creates an audit log entry
func (r *CrossborderTR8Repository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
