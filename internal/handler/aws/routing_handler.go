package awshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rldb-br/rldb-api-universal/internal/models/aws"
	"github.com/rldb-br/rldb-api-universal/internal/service/ai"
)

// RoutingHandler - Handler para endpoints de otimizacao de rotas
type RoutingHandler struct {
	routeService *aiservice.RouteOptimizationService
}

// NewRoutingHandler - Cria novo handler de rotas
def NewRoutingHandler(routeService *aiservice.RouteOptimizationService) *RoutingHandler {
	return &RoutingHandler{
		routeService: routeService,
	}
}

// RegisterRoutes - Registra rotas de otimizacao
def (h *RoutingHandler) RegisterRoutes(router *gin.RouterGroup) {
	routing := router.Group("/v1/routing")
	{
		routing.POST("/optimize", h.OptimizeRoute)
		routing.POST("/batch-optimize", h.BatchOptimizeRoutes)
		routing.GET("/optimizations/:shipment_id", h.GetRouteOptimizationsByShipment)
		routing.GET("/optimization/:optimization_id", h.GetRouteOptimization)
		routing.GET("/statistics", h.GetRouteStatistics)
	}
}

// OptimizeRoute - Otimiza uma rota
// @Summary Otimiza uma rota
// @Description Otimiza uma rota usando IA/ML
// @Tags routing
// @Accept json
// @Produce json
// @Param request body aiservice.RouteOptimizationRequest true "Request de otimizacao de rota"
// @Success 200 {object} aiservice.RouteOptimizationResponse
// @Router /api/v1/routing/optimize [post]
func (h *RoutingHandler) OptimizeRoute(c *gin.Context) {
	var request aiservice.RouteOptimizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validar request
	if request.ShipmentID == "" && len(request.CurrentRoute) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "shipment_id or current_route is required",
		})
		return
	}

	response, err := h.routeService.OptimizeRoute(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// BatchOptimizeRoutes - Otimiza multiplas rotas
// @Summary Otimiza multiplas rotas
// @Description Otimiza multiplas rotas em batch
// @Tags routing
// @Accept json
// @Produce json
// @Param requests body []aiservice.RouteOptimizationRequest true "Requests de otimizacao de rotas"
// @Success 200 {object} []aiservice.RouteOptimizationResponse
// @Router /api/v1/routing/batch-optimize [post]
func (h *RoutingHandler) BatchOptimizeRoutes(c *gin.Context) {
	var requests []*aiservice.RouteOptimizationRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	responses, err := h.routeService.BatchOptimizeRoutes(c.Request.Context(), requests)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses)
}

// GetRouteOptimizationsByShipment - Obtem otimizacoes por shipment
// @Summary Obtem otimizacoes por shipment
// @Description Obtem todas as otimizacoes de rota para um shipment
// @Tags routing
// @Accept json
// @Produce json
// @Param shipment_id path string true "Shipment ID"
// @Success 200 {object} []awsmodels.RouteOptimization
// @Router /api/v1/routing/optimizations/{shipment_id} [get]
func (h *RoutingHandler) GetRouteOptimizationsByShipment(c *gin.Context) {
	shipmentID := c.Param("shipment_id")
	if shipmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "shipment_id is required",
		})
		return
	}

	optimizations, err := h.routeService.GetRouteOptimizationsByShipment(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, optimizations)
}

// GetRouteOptimization - Obtem uma otimizacao especifica
// @Summary Obtem uma otimizacao
// @Description Obtem uma otimizacao de rota especifica
// @Tags routing
// @Accept json
// @Produce json
// @Param optimization_id path string true "Optimization ID"
// @Success 200 {object} awsmodels.RouteOptimization
// @Router /api/v1/routing/optimization/{optimization_id} [get]
func (h *RoutingHandler) GetRouteOptimization(c *gin.Context) {
	optimizationID := c.Param("optimization_id")
	if optimizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "optimization_id is required",
		})
		return
	}

	optimization, err := h.routeService.GetRouteOptimization(c.Request.Context(), optimizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if optimization == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "optimization not found",
		})
		return
	}

	c.JSON(http.StatusOK, optimization)
}

// GetRouteStatistics - Obtem estatisticas de rotas
// @Summary Obtem estatisticas de rotas
// @Description Obtem estatisticas de otimizacao de rotas
// @Tags routing
// @Accept json
// @Produce json
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/routing/statistics [get]
func (h *RoutingHandler) GetRouteStatistics(c *gin.Context) {
	startTimeStr := c.DefaultQuery("start_time", "")
	endTimeStr := c.DefaultQuery("end_time", "")

	// Parsear times (simplificado)
	startTime := parseTime(startTimeStr)
	endTime := parseTime(endTimeStr)

	stats, err := h.routeService.GetRouteStatistics(c.Request.Context(), startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// parseTime - Parseia string para time.Time (simplificado)
func parseTime(timeStr string) time.Time {
	// Implementar parsing de time
	// Por enquanto, retornar time.Now()
	return time.Now().AddDate(0, 0, -7) // 7 dias atras
}

import "time"
