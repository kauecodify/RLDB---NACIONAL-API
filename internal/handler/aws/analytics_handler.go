package awshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rldb-br/rldb-api-universal/internal/models/aws"
	"github.com/rldb-br/rldb-api-universal/internal/service/analytics"
)

// AnalyticsHandler - Handler para endpoints de analise
type AnalyticsHandler struct {
	analyticsService *analyticsservice.RealTimeAnalyticsService
}

// NewAnalyticsHandler - Cria novo handler de analise
def NewAnalyticsHandler(analyticsService *analyticsservice.RealTimeAnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// RegisterRoutes - Registra rotas de analise
def (h *AnalyticsHandler) RegisterRoutes(router *gin.RouterGroup) {
	analytics := router.Group("/v1/analytics")
	{
		analytics.GET("/metrics", h.GetRealTimeMetrics)
		analytics.GET("/dashboard", h.GetAnalyticsDashboard)
		analytics.POST("/transactions", h.ProcessTransaction)
		analytics.POST("/logs", h.ProcessLog)
		analytics.POST("/batch", h.ProcessBatch)
	}
}

// GetRealTimeMetrics - Obtem metricas em tempo real
// @Summary Obtem metricas em tempo real
// @Description Retorna metricas de processamento em tempo real
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} awsmodels.RealTimeMetrics
// @Router /api/v1/analytics/metrics [get]
func (h *AnalyticsHandler) GetRealTimeMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetRealTimeMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetAnalyticsDashboard - Obtem dashboard de analise
// @Summary Obtem dashboard de analise
// @Description Retorna dashboard completo com todas as metricas e estatisticas
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/dashboard [get]
func (h *AnalyticsHandler) GetAnalyticsDashboard(c *gin.Context) {
	dashboard, err := h.analyticsService.GetAnalyticsDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// ProcessTransaction - Processa uma transacao
// @Summary Processa uma transacao
// @Description Processa uma transacao em tempo real
// @Tags analytics
// @Accept json
// @Produce json
// @Param transaction body awsmodels.TransactionStreamEvent true "Transacao"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/transactions [post]
func (h *AnalyticsHandler) ProcessTransaction(c *gin.Context) {
	var transaction awsmodels.TransactionStreamEvent
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.analyticsService.ProcessTransaction(c.Request.Context(), &transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction processed successfully",
		"event_id": transaction.EventID,
	})
}

// ProcessLog - Processa um log
// @Summary Processa um log
// @Description Processa um log em tempo real
// @Tags analytics
// @Accept json
// @Produce json
// @Param log body awsmodels.LogStreamEvent true "Log"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/logs [post]
func (h *AnalyticsHandler) ProcessLog(c *gin.Context) {
	var logEvent awsmodels.LogStreamEvent
	if err := c.ShouldBindJSON(&logEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.analyticsService.ProcessLog(c.Request.Context(), &logEvent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Log processed successfully",
		"log_id": logEvent.LogID,
	})
}

// ProcessBatch - Processa batch de eventos
// @Summary Processa batch de eventos
// @Description Processa multiplos eventos em batch
// @Tags analytics
// @Accept json
// @Produce json
// @Param batch body []interface{} true "Batch de eventos"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/batch [post]
func (h *AnalyticsHandler) ProcessBatch(c *gin.Context) {
	var batch []interface{}
	if err := c.ShouldBindJSON(&batch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.analyticsService.ProcessBatch(c.Request.Context(), batch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Batch processed successfully, %d events", len(batch)),
	})
}
