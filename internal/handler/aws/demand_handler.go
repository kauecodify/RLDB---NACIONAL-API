package awshandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rldb-br/rldb-api-universal/internal/models/aws"
	"github.com/rldb-br/rldb-api-universal/internal/service/ai"
)

// DemandHandler - Handler para endpoints de previsao de demanda
type DemandHandler struct {
	demandService *aiservice.DemandForecastService
}

// NewDemandHandler - Cria novo handler de demanda
def NewDemandHandler(demandService *aiservice.DemandForecastService) *DemandHandler {
	return &DemandHandler{
		demandService: demandService,
	}
}

// RegisterRoutes - Registra rotas de demanda
def (h *DemandHandler) RegisterRoutes(router *gin.RouterGroup) {
	demand := router.Group("/v1/demand")
	{
		demand.POST("/forecast", h.GenerateForecast)
		demand.GET("/forecasts/:region", h.GetForecastsByRegion)
		demand.GET("/forecast/:forecast_id", h.GetForecast)
		demand.GET("/statistics", h.GetDemandStatistics)
	}
}

// GenerateForecast - Gera previsao de demanda
// @Summary Gera previsao de demanda
// @Description Gera previsao de demanda usando IA/ML
// @Tags demand
// @Accept json
// @Produce json
// @Param request body aiservice.DemandForecastRequest true "Request de previsao de demanda"
// @Success 200 {object} aiservice.DemandForecastResponse
// @Router /api/v1/demand/forecast [post]
func (h *DemandHandler) GenerateForecast(c *gin.Context) {
	var request aiservice.DemandForecastRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validar request
	if request.Region == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "region is required",
		})
		return
	}

	response, err := h.demandService.GenerateForecast(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetForecastsByRegion - Obtem previsoes por regiao
// @Summary Obtem previsoes por regiao
// @Description Obtem todas as previsoes de demanda para uma regiao
// @Tags demand
// @Accept json
// @Produce json
// @Param region path string true "Regiao"
// @Param limit query int false "Limite de resultados"
// @Success 200 {object} []awsmodels.DemandForecast
// @Router /api/v1/demand/forecasts/{region} [get]
func (h *DemandHandler) GetForecastsByRegion(c *gin.Context) {
	region := c.Param("region")
	if region == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "region is required",
		})
		return
	}

	limit := c.DefaultQuery("limit", "100")
	limitInt := 100
	if l, err := parseInt(limit); err == nil {
		limitInt = l
	}

	forecasts, err := h.demandService.GetForecastsByRegion(c.Request.Context(), region, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, forecasts)
}

// GetForecast - Obtem uma previsao especifica
// @Summary Obtem uma previsao
// @Description Obtem uma previsao de demanda especifica
// @Tags demand
// @Accept json
// @Produce json
// @Param forecast_id path string true "Forecast ID"
// @Success 200 {object} awsmodels.DemandForecast
// @Router /api/v1/demand/forecast/{forecast_id} [get]
func (h *DemandHandler) GetForecast(c *gin.Context) {
	forecastID := c.Param("forecast_id")
	if forecastID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "forecast_id is required",
		})
		return
	}

	forecast, err := h.demandService.GetForecast(c.Request.Context(), forecastID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if forecast == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "forecast not found",
		})
		return
	}

	c.JSON(http.StatusOK, forecast)
}

// GetDemandStatistics - Obtem estatisticas de demanda
// @Summary Obtem estatisticas de demanda
// @Description Obtem estatisticas de previsao de demanda
// @Tags demand
// @Accept json
// @Produce json
// @Param region query string false "Regiao"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/demand/statistics [get]
func (h *DemandHandler) GetDemandStatistics(c *gin.Context) {
	region := c.DefaultQuery("region", "BR")
	startTimeStr := c.DefaultQuery("start_time", "")
	endTimeStr := c.DefaultQuery("end_time", "")

	// Parsear times (simplificado)
	startTime := parseTime(startTimeStr)
	endTime := parseTime(endTimeStr)

	stats, err := h.demandService.GetDemandStatistics(c.Request.Context(), region, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// parseInt - Parseia string para int
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

import (
	"fmt"
)

// parseTime - Parseia string para time.Time (simplificado)
func parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Now().AddDate(0, 0, -30) // 30 dias atras
	}

	// Tentar parsear RFC3339
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t
	}

	// Tentar outros formatos
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err = time.Parse(format, timeStr)
		if err == nil {
			return t
		}
	}

	return time.Now().AddDate(0, 0, -30)
}
