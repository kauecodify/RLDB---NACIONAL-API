package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rldb-br/rldb-api-universal/internal/models"
	"github.com/rldb-br/rldb-api-universal/internal/repository"
	"github.com/rldb-br/rldb-api-universal/pkg/governance"
)

// CrossborderTR8Service handles TR8 crossborder operations
type CrossborderTR8Service struct {
	repo           *repository.CrossborderTR8Repository
	governanceSvc *governance.DataGovernanceService
	cache          CacheService
	config         *models.TR8GovernanceConfig
}

// CacheService interface for caching
type CacheService interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
}

// NewCrossborderTR8Service creates a new TR8 service
func NewCrossborderTR8Service(
	repo *repository.CrossborderTR8Repository,
	governanceSvc *governance.DataGovernanceService,
	cache CacheService,
	config *models.TR8GovernanceConfig,
) *CrossborderTR8Service {
	return &CrossborderTR8Service{
		repo:           repo,
		governanceSvc: governanceSvc,
		cache:          cache,
		config:         config,
	}
}

// GetTR8Quotes returns shipping quotes for TR8 crossborder
func (s *CrossborderTR8Service) GetTR8Quotes(ctx context.Context, request models.TR8QuoteRequest, operatorID string) ([]models.CrossborderQuote, error) {
	// Apply data governance
	sanitizedRequest, err := s.governanceSvc.SanitizeRequest(ctx, request, operatorID, "GET /v1/crossborder/tr8/quotes")
	if err != nil {
		return nil, fmt.Errorf("governance sanitization failed: %w", err)
	}

	// Check rate limiting
	if err := s.checkRateLimit(ctx, operatorID, "quotes"); err != nil {
		return nil, err
	}

	// Generate cache key
	cacheKey := s.generateQuoteCacheKey(sanitizedRequest, operatorID)

	// Try to get from cache
	var cachedQuotes []models.CrossborderQuote
	if err := s.cache.Get(cacheKey, &cachedQuotes); err == nil {
		return cachedQuotes, nil
	}

	// Calculate quotes from operators
	quotes, err := s.calculateTR8Quotes(ctx, sanitizedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate quotes: %w", err)
	}

	// Cache for 5 minutes
	if err := s.cache.Set(cacheKey, quotes, 5*time.Minute); err != nil {
		// Log cache error but continue
		// log.Printf("failed to cache quotes: %v", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "GET /v1/crossborder/tr8/quotes", "success", sanitizedRequest)

	return quotes, nil
}

// CreateTR8Shipment creates a new TR8 crossborder shipment
func (s *CrossborderTR8Service) CreateTR8Shipment(ctx context.Context, request models.TR8ShipmentRequest, operatorID string) (*models.TR8ShipmentResponse, error) {
	// Apply data governance
	sanitizedRequest, err := s.governanceSvc.SanitizeRequest(ctx, request, operatorID, "POST /v1/crossborder/tr8/shipments")
	if err != nil {
		return nil, fmt.Errorf("governance sanitization failed: %w", err)
	}

	// Check idempotency
	if request.IdempotencyKey != "" {
		existing, err := s.repo.GetShipmentByIdempotencyKey(ctx, request.IdempotencyKey)
		if err == nil && existing != nil {
			return &models.TR8ShipmentResponse{
				TrackingCode:     existing.TrackingCode,
				LabelURL:         existing.LabelURL,
				EstimatedDelivery: existing.EstimatedDelivery,
				HubRoute:         existing.HubRoute,
				ShipmentID:       existing.ShipmentID,
				Operator:         existing.Operator,
				Service:          existing.Service,
				CreatedAt:        existing.CreatedAt,
			}, nil
		}
	}

	// Check rate limiting
	if err := s.checkRateLimit(ctx, operatorID, "shipments"); err != nil {
		return nil, err
	}

	// Validate request
	if err := s.validateTR8ShipmentRequest(sanitizedRequest); err != nil {
		return nil, err
	}

	// Generate tracking code
	trackingCode := s.generateTrackingCode()

	// Create shipment record
	shipment := &models.CrossborderShipment{
		ID:                uuid.New().String(),
		ShipmentID:        request.ShipmentID,
		Operator:          request.Operator,
		Service:           request.Service,
		OriginCountry:     request.OriginCountry,
		DestinationCountry: request.DestinationCountry,
		OriginZip:         request.OriginZip,
		DestinationZip:    request.DestinationZip,
		WeightKg:          request.WeightKg,
		DimensionsCm:      request.DimensionsCm,
		DeclaredValueBRL:  request.DeclaredValueBRL,
		DeclaredValueUSD:  request.DeclaredValueUSD,
		Currency:          "BRL",
		Incoterm:          request.Incoterm,
		HSCode:            request.HSCode,
		ProductCategory:   request.ProductCategory,
		ReverseLogistics:  request.ReverseLogistics,
		TrackingCode:      trackingCode,
		Status:            "pending",
		EstimatedDelivery: time.Now().AddDate(0, 0, 7), // Default 7 days
		Metadata: map[string]interface{}{
			"operator_id": operatorID,
			"request_id":   ctx.Value("request_id"),
		},
	}

	// Save shipment
	if err := s.repo.CreateShipment(ctx, shipment); err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	// Generate label
	labelURL := s.generateLabelURL(shipment.ID)
	shipment.LabelURL = labelURL

	// Update with label URL
	if err := s.repo.UpdateShipment(ctx, shipment); err != nil {
		// Log error but continue
		// log.Printf("failed to update shipment with label URL: %v", err)
	}

	// Create tracking events
	initialEvent := &models.CrossborderTrackingEvent{
		ID:           uuid.New().String(),
		ShipmentID:   shipment.ID,
		TrackingCode: trackingCode,
		Timestamp:    time.Now().UTC(),
		Location:     request.OriginZip,
		Country:      request.OriginCountry,
		Status:       "posted",
		Operator:     request.Operator,
		Description:  "Shipment created",
	}

	if err := s.repo.CreateTrackingEvent(ctx, initialEvent); err != nil {
		// Log error but continue
		// log.Printf("failed to create initial tracking event: %v", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "POST /v1/crossborder/tr8/shipments", "success", sanitizedRequest)

	// Return response
	response := &models.TR8ShipmentResponse{
		TrackingCode:     trackingCode,
		LabelURL:         labelURL,
		EstimatedDelivery: shipment.EstimatedDelivery,
		HubRoute:         []string{"Origin Hub", "Destination Hub"},
		ShipmentID:       shipment.ShipmentID,
		Operator:         shipment.Operator,
		Service:          shipment.Service,
		CustomsStatus:    "pending",
		CreatedAt:        shipment.CreatedAt,
	}

	return response, nil
}

// GetTR8Tracking returns tracking information for a TR8 shipment
func (s *CrossborderTR8Service) GetTR8Tracking(ctx context.Context, trackingCode string, operatorID string) (*models.TR8TrackingResponse, error) {
	// Apply data governance
	sanitizedTrackingCode := s.governanceSvc.SanitizeField(ctx, trackingCode, operatorID, "tracking_code")

	// Get shipment
	shipment, err := s.repo.GetShipmentByTrackingCode(ctx, sanitizedTrackingCode)
	if err != nil {
		return nil, fmt.Errorf("shipment not found: %w", err)
	}

	// Check if operator has access to this shipment
	if shipment.Operator != operatorID {
		// Check if operator is the origin or destination
		if shipment.OriginCountry != operatorID && shipment.DestinationCountry != operatorID {
			return nil, errors.New("unauthorized access to shipment")
		}
	}

	// Get tracking events
	events, err := s.repo.GetTrackingEvents(ctx, shipment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking events: %w", err)
	}

	// Convert to response format
	eventResponses := make([]models.TR8TrackingEventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = models.TR8TrackingEventResponse{
			Timestamp:   event.Timestamp,
			Location:    event.Location,
			Country:     event.Country,
			Status:      event.Status,
			Operator:    event.Operator,
			Description: event.Description,
		}
	}

	// Get customs documents
	documents, err := s.repo.GetCustomsDocuments(ctx, shipment.ID)
	if err != nil {
		// Log error but continue
		// log.Printf("failed to get customs documents: %v", err)
	}

	docInfos := make([]models.CustomsDocumentInfo, len(documents))
	for i, doc := range documents {
		docInfos[i] = models.CustomsDocumentInfo{
			DocumentType:   doc.DocumentType,
			DocumentNumber: doc.DocumentNumber,
			Status:         doc.Status,
			Verified:       doc.Verified,
			FileURL:        doc.FileURL,
		}
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "GET /v1/crossborder/tr8/tracking/{tracking_code}", "success", map[string]interface{}{
		"tracking_code": trackingCode,
	})

	return &models.TR8TrackingResponse{
		TrackingCode:     shipment.TrackingCode,
		Status:           shipment.Status,
		Events:           eventResponses,
		EstimatedDelivery: shipment.EstimatedDelivery,
		ActualDelivery:   shipment.ActualDelivery,
		Operator:         shipment.Operator,
		Service:          shipment.Service,
		OriginCountry:    shipment.OriginCountry,
		DestinationCountry: shipment.DestinationCountry,
		CustomsStatus:    shipment.Metadata["customs_status"].(string),
		Documents:        docInfos,
	}, nil
}

// UpdateTR8Tracking updates tracking information for a TR8 shipment
func (s *CrossborderTR8Service) UpdateTR8Tracking(ctx context.Context, trackingCode string, status string, location string, country string, description string, operatorID string) error {
	// Get shipment
	shipment, err := s.repo.GetShipmentByTrackingCode(ctx, trackingCode)
	if err != nil {
		return fmt.Errorf("shipment not found: %w", err)
	}

	// Verify operator has write access
	if shipment.Operator != operatorID {
		return errors.New("unauthorized to update tracking")
	}

	// Create tracking event
	event := &models.CrossborderTrackingEvent{
		ID:           uuid.New().String(),
		ShipmentID:   shipment.ID,
		TrackingCode: trackingCode,
		Timestamp:    time.Now().UTC(),
		Location:     location,
		Country:      country,
		Status:       status,
		Operator:     operatorID,
		Description:  description,
	}

	// Update shipment status if needed
	if status == "delivered" {
		now := time.Now().UTC()
		shipment.Status = "delivered"
		shipment.ActualDelivery = &now
		if err := s.repo.UpdateShipment(ctx, shipment); err != nil {
			return fmt.Errorf("failed to update shipment status: %w", err)
		}
	}

	// Save tracking event
	if err := s.repo.CreateTrackingEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to create tracking event: %w", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "POST /v1/crossborder/tr8/tracking/{tracking_code}", "success", map[string]interface{}{
		"tracking_code": trackingCode,
		"status":        status,
		"location":      location,
	})

	return nil
}

// GetTR8Metrics returns metrics for TR8 observatory
func (s *CrossborderTR8Service) GetTR8Metrics(ctx context.Context, period string, operatorID string) (*models.TR8MetricsResponse, error) {
	// Apply data governance
	sanitizedPeriod := s.governanceSvc.SanitizeField(ctx, period, operatorID, "period")

	// Parse period
	startTime, endTime, err := s.parsePeriod(sanitizedPeriod)
	if err != nil {
		return nil, err
	}

	// Get metrics from repository
	metrics, err := s.repo.GetTR8Metrics(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "GET /v1/observatory/tr8/metrics", "success", map[string]interface{}{
		"period": period,
	})

	return metrics, nil
}

// CreateCustomsDocument creates a customs document for TR8 shipment
func (s *CrossborderTR8Service) CreateCustomsDocument(ctx context.Context, shipmentID string, docType string, docNumber string, issuingCountry string, operatorID string) (*models.CrossborderCustomsDocument, error) {
	// Get shipment
	shipment, err := s.repo.GetShipmentByID(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("shipment not found: %w", err)
	}

	// Verify operator has access
	if shipment.Operator != operatorID && shipment.OriginCountry != operatorID {
		return nil, errors.New("unauthorized to create customs document")
	}

	// Create document
	doc := &models.CrossborderCustomsDocument{
		ID:             uuid.New().String(),
		ShipmentID:     shipmentID,
		DocumentType:   docType,
		DocumentNumber: docNumber,
		IssueDate:      time.Now().UTC(),
		IssuingCountry: issuingCountry,
		Status:         "pending",
		Verified:       false,
	}

	if err := s.repo.CreateCustomsDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create customs document: %w", err)
	}

	// Update shipment metadata
	if shipment.Metadata == nil {
		shipment.Metadata = make(map[string]interface{})
	}
	shipment.Metadata["customs_status"] = "documents_submitted"

	if err := s.repo.UpdateShipment(ctx, shipment); err != nil {
		// Log error but continue
		// log.Printf("failed to update shipment metadata: %v", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "POST /v1/crossborder/tr8/customs/documents", "success", map[string]interface{}{
		"shipment_id":   shipmentID,
		"document_type": docType,
	})

	return doc, nil
}

// ValidateCustomsDocument validates a customs document
func (s *CrossborderTR8Service) ValidateCustomsDocument(ctx context.Context, documentID string, verified bool, operatorID string) error {
	// Get document
	doc, err := s.repo.GetCustomsDocumentByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Get shipment
	shipment, err := s.repo.GetShipmentByID(ctx, doc.ShipmentID)
	if err != nil {
		return fmt.Errorf("shipment not found: %w", err)
	}

	// Verify operator has access
	if shipment.Operator != operatorID {
		return errors.New("unauthorized to validate customs document")
	}

	// Update document
	doc.Verified = verified
	if verified {
		doc.Status = "verified"
	} else {
		doc.Status = "rejected"
	}

	if err := s.repo.UpdateCustomsDocument(ctx, doc); err != nil {
		return fmt.Errorf("failed to update customs document: %w", err)
	}

	// Update shipment metadata
	if shipment.Metadata == nil {
		shipment.Metadata = make(map[string]interface{})
	}
	if verified {
		shipment.Metadata["customs_status"] = "verified"
	} else {
		shipment.Metadata["customs_status"] = "rejected"
	}

	if err := s.repo.UpdateShipment(ctx, shipment); err != nil {
		// Log error but continue
		// log.Printf("failed to update shipment metadata: %v", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "PUT /v1/crossborder/tr8/customs/documents/{id}/validate", "success", map[string]interface{}{
		"document_id": documentID,
		"verified":    verified,
	})

	return nil
}

// GetTR8ComplianceReport generates a compliance report
func (s *CrossborderTR8Service) GetTR8ComplianceReport(ctx context.Context, period string, operatorID string) (*models.TR8ComplianceReport, error) {
	// Apply data governance
	sanitizedPeriod := s.governanceSvc.SanitizeField(ctx, period, operatorID, "period")

	// Parse period
	startTime, endTime, err := s.parsePeriod(sanitizedPeriod)
	if err != nil {
		return nil, err
	}

	// Get compliance data
	report, err := s.repo.GetTR8ComplianceReport(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance report: %w", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "GET /v1/governance/tr8/compliance", "success", map[string]interface{}{
		"period": period,
	})

	return report, nil
}

// Helper methods

func (s *CrossborderTR8Service) calculateTR8Quotes(ctx context.Context, request models.TR8QuoteRequest) ([]models.CrossborderQuote, error) {
	// This would integrate with actual operator APIs
	// For now, return mock data
	quotes := []models.CrossborderQuote{
		{
			ID:                uuid.New().String(),
			RequestID:         ctx.Value("request_id").(string),
			Operator:          "correios_international",
			Service:           "tr8_standard",
			OriginCountry:     request.OriginCountry,
			DestinationCountry: request.DestinationCountry,
			OriginZip:         request.OriginZip,
			DestinationZip:    request.DestinationZip,
			WeightKg:          request.WeightKg,
			DimensionsCm:      request.DimensionsCm,
			DeclaredValueBRL:  request.DeclaredValueBRL,
			DeclaredValueUSD:  request.DeclaredValueBRL / 5.0, // Approximate conversion
			PriceBRL:          150.0,
			PriceUSD:          30.0,
			DeliveryDays:      14,
			CO2Kg:            2.5,
			Incoterm:          request.Incoterm,
			HSCode:            request.HSCode,
			DutiesAndTaxesBRL: 25.0,
			DutiesAndTaxesUSD: 5.0,
			ValidUntil:        time.Now().Add(24 * time.Hour),
			CreatedAt:         time.Now().UTC(),
		},
		{
			ID:                uuid.New().String(),
			RequestID:         ctx.Value("request_id").(string),
			Operator:          "dhl_global",
			Service:           "tr8_express",
			OriginCountry:     request.OriginCountry,
			DestinationCountry: request.DestinationCountry,
			OriginZip:         request.OriginZip,
			DestinationZip:    request.DestinationZip,
			WeightKg:          request.WeightKg,
			DimensionsCm:      request.DimensionsCm,
			DeclaredValueBRL:  request.DeclaredValueBRL,
			DeclaredValueUSD:  request.DeclaredValueBRL / 5.0,
			PriceBRL:          200.0,
			PriceUSD:          40.0,
			DeliveryDays:      7,
			CO2Kg:            1.8,
			Incoterm:          request.Incoterm,
			HSCode:            request.HSCode,
			DutiesAndTaxesBRL: 25.0,
			DutiesAndTaxesUSD: 5.0,
			ValidUntil:        time.Now().Add(24 * time.Hour),
			CreatedAt:         time.Now().UTC(),
		},
	}

	return quotes, nil
}

func (s *CrossborderTR8Service) validateTR8ShipmentRequest(request models.TR8ShipmentRequest) error {
	if request.Operator == "" {
		return errors.New("operator is required")
	}
	if request.Service == "" {
		return errors.New("service is required")
	}
	if request.ShipmentID == "" {
		return errors.New("shipment_id is required")
	}
	if request.OriginCountry == "" {
		return errors.New("origin_country is required")
	}
	if request.DestinationCountry == "" {
		return errors.New("destination_country is required")
	}
	if request.WeightKg <= 0 {
		return errors.New("weight_kg must be greater than 0")
	}
	if request.DeclaredValueBRL < 0 {
		return errors.New("declared_value_brl cannot be negative")
	}
	return nil
}

func (s *CrossborderTR8Service) generateTrackingCode() string {
	// Generate a TR8-specific tracking code
	rand.Seed(time.Now().UnixNano())
	prefix := "TR8"
	numbers := ""
	for i := 0; i < 12; i++ {
		numbers += fmt.Sprintf("%d", rand.Intn(10))
	}
	return prefix + numbers
}

func (s *CrossborderTR8Service) generateLabelURL(shipmentID string) string {
	// Generate a secure label URL
	hash := sha256.Sum256([]byte(shipmentID + time.Now().Format(time.RFC3339)))
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("https://rldb.gov.br/labels/%s.pdf", hashStr[:16])
}

func (s *CrossborderTR8Service) generateQuoteCacheKey(request models.TR8QuoteRequest, operatorID string) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%.2f:%s:%s",
		request.OriginCountry,
		request.DestinationCountry,
		request.OriginZip,
		request.DestinationZip,
		request.WeightKg,
		request.Incoterm,
		operatorID,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *CrossborderTR8Service) checkRateLimit(ctx context.Context, operatorID string, endpoint string) error {
	// Implement rate limiting based on configuration
	// For now, just return nil
	return nil
}

func (s *CrossborderTR8Service) parsePeriod(period string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	switch period {
	case "last_24_hours":
		return now.Add(-24 * time.Hour), now, nil
	case "last_7_days":
		return now.Add(-7 * 24 * time.Hour), now, nil
	case "last_30_days":
		return now.Add(-30 * 24 * time.Hour), now, nil
	case "today":
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return today, today.Add(24 * time.Hour), nil
	default:
		return time.Time{}, time.Time{}, errors.New("invalid period")
	}
}
