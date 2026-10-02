package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rldb-br/rldb-api-universal/internal/models"
	"github.com/rldb-br/rldb-api-universal/internal/service"
)

// CrossborderTR8Handler handles HTTP requests for TR8 crossborder operations
type CrossborderTR8Handler struct {
	service *service.CrossborderTR8Service
}

// NewCrossborderTR8Handler creates a new TR8 handler
func NewCrossborderTR8Handler(service *service.CrossborderTR8Service) *CrossborderTR8Handler {
	return &CrossborderTR8Handler{service: service}
}

// RegisterRoutes registers TR8 routes with the router
func (h *CrossborderTR8Handler) RegisterRoutes(router *gin.RouterGroup) {
	// TR8 Crossborder Routes
	tr8Group := router.Group("/v1/crossborder/tr8")
	{
		// Quotes
		tr8Group.POST("/quotes", h.GetTR8Quotes)

		// Shipments
		tr8Group.POST("/shipments", h.CreateTR8Shipment)
		tr8Group.GET("/shipments/:shipment_id", h.GetTR8Shipment)

		// Tracking
		tr8Group.GET("/tracking/:tracking_code", h.GetTR8Tracking)
		tr8Group.POST("/tracking/:tracking_code", h.UpdateTR8Tracking)

		// Customs
		tr8Group.POST("/customs/documents", h.CreateCustomsDocument)
		tr8Group.PUT("/customs/documents/:document_id/validate", h.ValidateCustomsDocument)
		tr8Group.GET("/customs/documents/:document_id", h.GetCustomsDocument)

		// Webhooks
		tr8Group.POST("/webhooks", h.CreateWebhook)
		tr8Group.GET("/webhooks", h.GetWebhooks)
	}

	// Observatory Routes
	observatoryGroup := router.Group("/v1/observatory")
	{
		observatoryGroup.GET("/tr8/metrics", h.GetTR8Metrics)
		observatoryGroup.GET("/tr8/compliance", h.GetTR8ComplianceReport)
	}
}

// GetTR8Quotes returns TR8 shipping quotes
// @Summary Get TR8 Crossborder Shipping Quotes
// @Description Returns available shipping quotes for TR8 international shipments
// @Tags TR8 Crossborder
// @Accept json
// @Produce json
// @Param request body models.TR8QuoteRequest true "Quote request"
// @Success 200 {array} models.CrossborderQuote
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /v1/crossborder/tr8/quotes [post]
func (h *CrossborderTR8Handler) GetTR8Quotes(c *gin.Context) {
	var request models.TR8QuoteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate request
	if err := validateTR8QuoteRequest(request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Add request ID to context
	ctx := c.Request.Context()
	ctx = contextWithRequestID(ctx, uuid.New().String())

	// Get quotes
	quotes, err := h.service.GetTR8Quotes(ctx, request, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, quotes)
}

// CreateTR8Shipment creates a new TR8 crossborder shipment
// @Summary Create TR8 Crossborder Shipment
// @Description Creates a new TR8 international shipment
// @Tags TR8 Crossborder
// @Accept json
// @Produce json
// @Param request body models.TR8ShipmentRequest true "Shipment request"
// @Success 201 {object} models.TR8ShipmentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /v1/crossborder/tr8/shipments [post]
func (h *CrossborderTR8Handler) CreateTR8Shipment(c *gin.Context) {
	var request models.TR8ShipmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate request
	if err := validateTR8ShipmentRequest(request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Add request ID to context
	ctx := c.Request.Context()
	ctx = contextWithRequestID(ctx, uuid.New().String())

	// Create shipment
	response, err := h.service.CreateTR8Shipment(ctx, request, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetTR8Shipment retrieves a TR8 shipment by ID
// @Summary Get TR8 Shipment
// @Description Retrieves a TR8 shipment by its ID
// @Tags TR8 Crossborder
// @Produce json
// @Param shipment_id path string true "Shipment ID"
// @Success 200 {object} models.CrossborderShipment
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/crossborder/tr8/shipments/{shipment_id} [get]
func (h *CrossborderTR8Handler) GetTR8Shipment(c *gin.Context) {
	shipmentID := c.Param("shipment_id")
	if shipmentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "shipment_id is required"})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get shipment
	shipment, err := h.service.GetTR8Shipment(c.Request.Context(), shipmentID, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, shipment)
}

// GetTR8Tracking retrieves tracking information for a TR8 shipment
// @Summary Get TR8 Tracking Information
// @Description Retrieves tracking information for a TR8 shipment
// @Tags TR8 Crossborder
// @Produce json
// @Param tracking_code path string true "Tracking Code"
// @Success 200 {object} models.TR8TrackingResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/crossborder/tr8/tracking/{tracking_code} [get]
func (h *CrossborderTR8Handler) GetTR8Tracking(c *gin.Context) {
	trackingCode := c.Param("tracking_code")
	if trackingCode == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tracking_code is required"})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get tracking
	tracking, err := h.service.GetTR8Tracking(c.Request.Context(), trackingCode, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tracking)
}

// UpdateTR8Tracking updates tracking information for a TR8 shipment
// @Summary Update TR8 Tracking
// @Description Updates tracking information for a TR8 shipment
// @Tags TR8 Crossborder
// @Accept json
// @Produce json
// @Param tracking_code path string true "Tracking Code"
// @Param request body models.CrossborderTrackingEvent true "Tracking update"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/crossborder/tr8/tracking/{tracking_code} [post]
func (h *CrossborderTR8Handler) UpdateTR8Tracking(c *gin.Context) {
	trackingCode := c.Param("tracking_code")
	if trackingCode == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tracking_code is required"})
		return
	}

	var request struct {
		Status      string `json:"status"`
		Location    string `json:"location"`
		Country     string `json:"country"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Update tracking
	if err := h.service.UpdateTR8Tracking(c.Request.Context(), trackingCode, request.Status, request.Location, request.Country, request.Description, operatorID.(string)); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true, Message: "Tracking updated"})
}

// CreateCustomsDocument creates a customs document for a TR8 shipment
// @Summary Create Customs Document
// @Description Creates a customs document for a TR8 shipment
// @Tags TR8 Crossborder
// @Accept json
// @Produce json
// @Param request body models.CrossborderCustomsDocument true "Customs document"
// @Success 201 {object} models.CrossborderCustomsDocument
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/crossborder/tr8/customs/documents [post]
func (h *CrossborderTR8Handler) CreateCustomsDocument(c *gin.Context) {
	var request struct {
		ShipmentID     string `json:"shipment_id"`
		DocumentType   string `json:"document_type"`
		DocumentNumber string `json:"document_number"`
		IssuingCountry string `json:"issuing_country"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate request
	if request.ShipmentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "shipment_id is required"})
		return
	}
	if request.DocumentType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "document_type is required"})
		return
	}
	if request.DocumentNumber == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "document_number is required"})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Create document
	doc, err := h.service.CreateCustomsDocument(c.Request.Context(), request.ShipmentID, request.DocumentType, request.DocumentNumber, request.IssuingCountry, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// ValidateCustomsDocument validates a customs document
// @Summary Validate Customs Document
// @Description Validates a customs document for a TR8 shipment
// @Tags TR8 Crossborder
// @Accept json
// @Produce json
// @Param document_id path string true "Document ID"
// @Param request body struct{Verified bool} true "Validation request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/crossborder/tr8/customs/documents/{document_id}/validate [put]
func (h *CrossborderTR8Handler) ValidateCustomsDocument(c *gin.Context) {
	documentID := c.Param("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "document_id is required"})
		return
	}

	var request struct {
		Verified bool `json:"verified"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Validate document
	if err := h.service.ValidateCustomsDocument(c.Request.Context(), documentID, request.Verified, operatorID.(string)); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true, Message: "Document validation updated"})
}

// GetCustomsDocument retrieves a customs document
// @Summary Get Customs Document
// @Description Retrieves a customs document by ID
// @Tags TR8 Crossborder
// @Produce json
// @Param document_id path string true "Document ID"
// @Success 200 {object} models.CrossborderCustomsDocument
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/crossborder/tr8/customs/documents/{document_id} [get]
func (h *CrossborderTR8Handler) GetCustomsDocument(c *gin.Context) {
	documentID := c.Param("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "document_id is required"})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get document
	doc, err := h.service.GetCustomsDocument(c.Request.Context(), documentID, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// CreateWebhook creates a webhook configuration
// @Summary Create Webhook
// @Description Creates a webhook configuration for TR8 events
// @Tags TR8 Crossborder
// @Accept json
// @Produce json
// @Param request body models.CrossborderWebhook true "Webhook configuration"
// @Success 201 {object} models.CrossborderWebhook
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/crossborder/tr8/webhooks [post]
func (h *CrossborderTR8Handler) CreateWebhook(c *gin.Context) {
	var request models.CrossborderWebhook
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate request
	if request.URL == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}
	if len(request.Events) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "events are required"})
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Set client ID
	request.ClientID = clientID.(string)
	request.ID = uuid.New().String()
	request.CreatedAt = time.Now().UTC()
	request.UpdatedAt = time.Now().UTC()

	// Create webhook
	if err := h.service.CreateWebhook(c.Request.Context(), &request); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, request)
}

// GetWebhooks retrieves webhooks for the current client
// @Summary Get Webhooks
// @Description Retrieves webhook configurations for the current client
// @Tags TR8 Crossborder
// @Produce json
// @Success 200 {array} models.CrossborderWebhook
// @Failure 401 {object} ErrorResponse
// @Router /v1/crossborder/tr8/webhooks [get]
func (h *CrossborderTR8Handler) GetWebhooks(c *gin.Context) {
	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get webhooks
	webhooks, err := h.service.GetWebhooks(c.Request.Context(), clientID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

// GetTR8Metrics returns TR8 metrics
// @Summary Get TR8 Metrics
// @Description Returns metrics for TR8 crossborder operations
// @Tags Observatory
// @Produce json
// @Param period query string false "Period (last_24_hours, last_7_days, last_30_days, today)"
// @Success 200 {object} models.TR8MetricsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/observatory/tr8/metrics [get]
func (h *CrossborderTR8Handler) GetTR8Metrics(c *gin.Context) {
	period := c.DefaultQuery("period", "last_30_days")

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get metrics
	metrics, err := h.service.GetTR8Metrics(c.Request.Context(), period, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetTR8ComplianceReport returns TR8 compliance report
// @Summary Get TR8 Compliance Report
// @Description Returns compliance report for TR8 operations
// @Tags Observatory
// @Produce json
// @Param period query string false "Period (last_24_hours, last_7_days, last_30_days, today)"
// @Success 200 {object} models.TR8ComplianceReport
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/observatory/tr8/compliance [get]
func (h *CrossborderTR8Handler) GetTR8ComplianceReport(c *gin.Context) {
	period := c.DefaultQuery("period", "last_30_days")

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get compliance report
	report, err := h.service.GetTR8ComplianceReport(c.Request.Context(), period, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// Helper functions

func validateTR8QuoteRequest(request models.TR8QuoteRequest) error {
	if request.OriginCountry == "" {
		return fmt.Errorf("origin_country is required")
	}
	if request.DestinationCountry == "" {
		return fmt.Errorf("destination_country is required")
	}
	if request.OriginZip == "" {
		return fmt.Errorf("origin_zip is required")
	}
	if request.DestinationZip == "" {
		return fmt.Errorf("destination_zip is required")
	}
	if request.WeightKg <= 0 {
		return fmt.Errorf("weight_kg must be greater than 0")
	}
	if request.DeclaredValueBRL < 0 {
		return fmt.Errorf("declared_value_brl cannot be negative")
	}
	if request.Incoterm == "" {
		return fmt.Errorf("incoterm is required")
	}
	if request.ServiceLevel == "" {
		return fmt.Errorf("service_level is required")
	}
	return nil
}

func validateTR8ShipmentRequest(request models.TR8ShipmentRequest) error {
	if request.Operator == "" {
		return fmt.Errorf("operator is required")
	}
	if request.Service == "" {
		return fmt.Errorf("service is required")
	}
	if request.ShipmentID == "" {
		return fmt.Errorf("shipment_id is required")
	}
	if request.OriginCountry == "" {
		return fmt.Errorf("origin_country is required")
	}
	if request.DestinationCountry == "" {
		return fmt.Errorf("destination_country is required")
	}
	if request.OriginZip == "" {
		return fmt.Errorf("origin_zip is required")
	}
	if request.DestinationZip == "" {
		return fmt.Errorf("destination_zip is required")
	}
	if request.WeightKg <= 0 {
		return fmt.Errorf("weight_kg must be greater than 0")
	}
	if request.DeclaredValueBRL < 0 {
		return fmt.Errorf("declared_value_brl cannot be negative")
	}
	if request.Incoterm == "" {
		return fmt.Errorf("incoterm is required")
	}
	return nil
}

func contextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
