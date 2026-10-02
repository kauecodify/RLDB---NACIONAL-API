package nacional

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rldb-br/rldb-api-universal/internal/models/nacional"
	"github.com/rldb-br/rldb-api-universal/internal/service/nacional"
)

// TR8NacionalHandler handles HTTP requests for TR8 nacional operations
type TR8NacionalHandler struct {
	service *nacional.TR8NacionalService
}

// NewTR8NacionalHandler creates a new TR8 nacional handler
func NewTR8NacionalHandler(service *nacional.TR8NacionalService) *TR8NacionalHandler {
	return &TR8NacionalHandler{service: service}
}

// RegisterRoutes registers TR8 nacional routes with the router
func (h *TR8NacionalHandler) RegisterRoutes(router *gin.RouterGroup) {
	// TR8 Nacional Routes
	nacionalGroup := router.Group("/v1/nacional/tr8")
	{
		// Cotações
		nacionalGroup.POST("/cotacoes", h.GetCotacoes)

		// Remessas
		nacionalGroup.POST("/remessas", h.CreateRemessa)
		nacionalGroup.GET("/remessas/:id_remessa", h.GetRemessa)

		// Rastreamento
		nacionalGroup.GET("/rastreamento/:codigo_rastreamento", h.GetRastreamento)
		nacionalGroup.POST("/rastreamento/:codigo_rastreamento", h.UpdateRastreamento)

		// Documentos Fiscais
		nacionalGroup.POST("/documentos/fiscais", h.CreateDocumentoFiscal)
		nacionalGroup.PUT("/documentos/fiscais/:documento_id/validar", h.ValidateDocumentoFiscal)
		nacionalGroup.GET("/documentos/fiscais/:documento_id", h.GetDocumentoFiscal)

		// Webhooks
		nacionalGroup.POST("/webhooks", h.CreateWebhook)
		nacionalGroup.GET("/webhooks", h.GetWebhooks)
	}

	// Observatory Routes
	observatoryGroup := router.Group("/v1/observatorio")
	{
		observatoryGroup.GET("/nacional/tr8/metricas", h.GetMetricas)
		observatoryGroup.GET("/nacional/tr8/conformidade", h.GetRelatorioConformidade)
	}
}

// GetCotacoes returns TR8 nacional shipping quotes
// @Summary Get TR8 Nacional Shipping Quotes
// @Description Returns available shipping quotes for TR8 nacional shipments
// @Tags TR8 Nacional
// @Accept json
// @Produce json
// @Param request body models.SolicitacaoCotacao true "Quote request"
// @Success 200 {array} models.RespostaCotacao
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /v1/nacional/tr8/cotacoes [post]
func (h *TR8NacionalHandler) GetCotacoes(c *gin.Context) {
	var request nacional.SolicitacaoCotacao
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate request
	if err := validateCotacaoRequest(request); err != nil {
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

	// Get cotacoes
	cotacoes, err := h.service.GetCotacoes(ctx, request, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, cotacoes)
}

// CreateRemessa creates a new TR8 nacional remessa
// @Summary Create TR8 Nacional Remessa
// @Description Creates a new TR8 nacional shipment
// @Tags TR8 Nacional
// @Accept json
// @Produce json
// @Param request body models.SolicitacaoRemessa true "Remessa request"
// @Success 201 {object} models.RespostaRemessa
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /v1/nacional/tr8/remessas [post]
func (h *TR8NacionalHandler) CreateRemessa(c *gin.Context) {
	var request nacional.SolicitacaoRemessa
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate request
	if err := validateRemessaRequest(request); err != nil {
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

	// Create remessa
	response, err := h.service.CreateRemessa(ctx, request, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetRemessa retrieves a TR8 nacional remessa by ID
// @Summary Get TR8 Nacional Remessa
// @Description Retrieves a TR8 nacional remessa by its ID
// @Tags TR8 Nacional
// @Produce json
// @Param id_remessa path string true "Remessa ID"
// @Success 200 {object} models.RemessaNacional
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/nacional/tr8/remessas/{id_remessa} [get]
func (h *TR8NacionalHandler) GetRemessa(c *gin.Context) {
	idRemessa := c.Param("id_remessa")
	if idRemessa == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "id_remessa is required"})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get remessa
	remessa, err := h.service.GetRemessa(c.Request.Context(), idRemessa, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, remessa)
}

// GetRastreamento retrieves tracking information for a TR8 nacional remessa
// @Summary Get TR8 Nacional Tracking Information
// @Description Retrieves tracking information for a TR8 nacional remessa
// @Tags TR8 Nacional
// @Produce json
// @Param codigo_rastreamento path string true "Tracking Code"
// @Success 200 {object} models.RespostaRastreamento
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/nacional/tr8/rastreamento/{codigo_rastreamento} [get]
func (h *TR8NacionalHandler) GetRastreamento(c *gin.Context) {
	codigoRastreamento := c.Param("codigo_rastreamento")
	if codigoRastreamento == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "codigo_rastreamento is required"})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get rastreamento
	rastreamento, err := h.service.GetRastreamento(c.Request.Context(), codigoRastreamento, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, rastreamento)
}

// UpdateRastreamento updates tracking information for a TR8 nacional remessa
// @Summary Update TR8 Nacional Tracking
// @Description Updates tracking information for a TR8 nacional remessa
// @Tags TR8 Nacional
// @Accept json
// @Produce json
// @Param codigo_rastreamento path string true "Tracking Code"
// @Param request body models.EventoRastreamento true "Tracking update"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/nacional/tr8/rastreamento/{codigo_rastreamento} [post]
func (h *TR8NacionalHandler) UpdateRastreamento(c *gin.Context) {
	codigoRastreamento := c.Param("codigo_rastreamento")
	if codigoRastreamento == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "codigo_rastreamento is required"})
		return
	}

	var request struct {
		Status    string `json:"status"`
		Local     string `json:"local"`
		UF        string `json:"uf"`
		Descricao string `json:"descricao"`
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

	// Update rastreamento
	if err := h.service.UpdateRastreamento(c.Request.Context(), codigoRastreamento, request.Status, request.Local, request.UF, request.Descricao, operatorID.(string)); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true, Message: "Rastreamento atualizado"})
}

// CreateDocumentoFiscal creates a fiscal document for a TR8 nacional remessa
// @Summary Create Fiscal Document
// @Description Creates a fiscal document for a TR8 nacional remessa
// @Tags TR8 Nacional
// @Accept json
// @Produce json
// @Param request body models.DocumentoFiscal true "Fiscal document"
// @Success 201 {object} models.DocumentoFiscal
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/nacional/tr8/documentos/fiscais [post]
func (h *TR8NacionalHandler) CreateDocumentoFiscal(c *gin.Context) {
	var request struct {
		IDRemessa string `json:"id_remessa"`
		Tipo     string `json:"tipo"`
		Numero   string `json:"numero"`
		Emitente string `json:"emitente"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate request
	if request.IDRemessa == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "id_remessa is required"})
		return
	}
	if request.Tipo == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tipo is required"})
		return
	}
	if request.Numero == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "numero is required"})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Create document
	doc, err := h.service.CreateDocumentoFiscal(c.Request.Context(), request.IDRemessa, request.Tipo, request.Numero, request.Emitente, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// ValidateDocumentoFiscal validates a fiscal document
// @Summary Validate Fiscal Document
// @Description Validates a fiscal document for a TR8 nacional remessa
// @Tags TR8 Nacional
// @Accept json
// @Produce json
// @Param documento_id path string true "Document ID"
// @Param request body struct{Verificado bool} true "Validation request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/nacional/tr8/documentos/fiscais/{documento_id}/validar [put]
func (h *TR8NacionalHandler) ValidateDocumentoFiscal(c *gin.Context) {
	documentoID := c.Param("documento_id")
	if documentoID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "documento_id is required"})
		return
	}

	var request struct {
		Verificado bool `json:"verificado"`
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
	if err := h.service.ValidateDocumentoFiscal(c.Request.Context(), documentoID, request.Verificado, operatorID.(string)); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true, Message: "Documento validado"})
}

// GetDocumentoFiscal retrieves a fiscal document
// @Summary Get Fiscal Document
// @Description Retrieves a fiscal document by ID
// @Tags TR8 Nacional
// @Produce json
// @Param documento_id path string true "Document ID"
// @Success 200 {object} models.DocumentoFiscal
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/nacional/tr8/documentos/fiscais/{documento_id} [get]
func (h *TR8NacionalHandler) GetDocumentoFiscal(c *gin.Context) {
	documentoID := c.Param("documento_id")
	if documentoID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "documento_id is required"})
		return
	}

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get document
	doc, err := h.service.GetDocumentoFiscal(c.Request.Context(), documentoID, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// CreateWebhook creates a webhook configuration
// @Summary Create Webhook
// @Description Creates a webhook configuration for TR8 nacional events
// @Tags TR8 Nacional
// @Accept json
// @Produce json
// @Param request body models.WebhookNacional true "Webhook configuration"
// @Success 201 {object} models.WebhookNacional
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/nacional/tr8/webhooks [post]
func (h *TR8NacionalHandler) CreateWebhook(c *gin.Context) {
	var request nacional.WebhookNacional
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate request
	if request.URL == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}
	if len(request.Eventos) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "eventos are required"})
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
// @Tags TR8 Nacional
// @Produce json
// @Success 200 {array} models.WebhookNacional
// @Failure 401 {object} ErrorResponse
// @Router /v1/nacional/tr8/webhooks [get]
func (h *TR8NacionalHandler) GetWebhooks(c *gin.Context) {
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

// GetMetricas returns TR8 nacional metrics
// @Summary Get TR8 Nacional Metrics
// @Description Returns metrics for TR8 nacional operations
// @Tags Observatory
// @Produce json
// @Param periodo query string false "Period (ultimas_24_horas, ultimos_7_dias, ultimos_30_dias, hoje)"
// @Success 200 {object} models.MetricasNacional
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/observatorio/nacional/tr8/metricas [get]
func (h *TR8NacionalHandler) GetMetricas(c *gin.Context) {
	periodo := c.DefaultQuery("periodo", "ultimos_30_dias")

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get metrics
	metrics, err := h.service.GetMetricas(c.Request.Context(), periodo, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetRelatorioConformidade returns TR8 nacional compliance report
// @Summary Get TR8 Nacional Compliance Report
// @Description Returns compliance report for TR8 nacional operations
// @Tags Observatory
// @Produce json
// @Param periodo query string false "Period (ultimas_24_horas, ultimos_7_dias, ultimos_30_dias, hoje)"
// @Success 200 {object} models.RelatorioConformidade
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/observatorio/nacional/tr8/conformidade [get]
func (h *TR8NacionalHandler) GetRelatorioConformidade(c *gin.Context) {
	periodo := c.DefaultQuery("periodo", "ultimos_30_dias")

	// Get operator ID from context
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Get compliance report
	report, err := h.service.GetRelatorioConformidade(c.Request.Context(), periodo, operatorID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// Helper functions

func validateCotacaoRequest(request nacional.SolicitacaoCotacao) error {
	if request.OrigemCEP == "" {
		return fmt.Errorf("origem_cep é obrigatório")
	}
	if request.DestinoCEP == "" {
		return fmt.Errorf("destino_cep é obrigatório")
	}
	if request.PesoKg <= 0 {
		return fmt.Errorf("peso_kg deve ser maior que 0")
	}
	if request.ValorDeclaradoBRL < 0 {
		return fmt.Errorf("valor_declarado_brl não pode ser negativo")
	}
	if request.ModalidadeFrete == "" {
		return fmt.Errorf("modalidade_frete é obrigatório")
	}
	if request.TipoServico == "" {
		return fmt.Errorf("tipo_servico é obrigatório")
	}
	return nil
}

func validateRemessaRequest(request nacional.SolicitacaoRemessa) error {
	if request.Operador == "" {
		return fmt.Errorf("operador é obrigatório")
	}
	if request.Servico == "" {
		return fmt.Errorf("serviço é obrigatório")
	}
	if request.IDRemessa == "" {
		return fmt.Errorf("id_remessa é obrigatório")
	}
	if request.OrigemCEP == "" {
		return fmt.Errorf("origem_cep é obrigatório")
	}
	if request.DestinoCEP == "" {
		return fmt.Errorf("destino_cep é obrigatório")
	}
	if request.PesoKg <= 0 {
		return fmt.Errorf("peso_kg deve ser maior que 0")
	}
	if request.ValorDeclaradoBRL < 0 {
		return fmt.Errorf("valor_declarado_brl não pode ser negativo")
	}
	if request.ValorFreteBRL < 0 {
		return fmt.Errorf("valor_frete_brl não pode ser negativo")
	}
	if request.ModalidadeFrete == "" {
		return fmt.Errorf("modalidade_frete é obrigatório")
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
