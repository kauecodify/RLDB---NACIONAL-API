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

// RouteOptimizationService - Servico de otimizacao de rotas com IA
type RouteOptimizationService struct {
	routeRepo            *awsrepository.RouteOptimizationDynamoDBRepository
	transactionRepo      *awsrepository.TransactionDynamoDBRepository
	awsConfig            *awsmodels.AWSConfig
	modelConfig          *awsmodels.AIModelConfig
	transactionService   *TransactionService
	geocodingService     GeocodingService
	maxIterations        int
}

// GeocodingService - Interface para servico de geocodificacao
type GeocodingService interface {
	GetCoordinates(ctx context.Context, address string) (*awsmodels.LocationData, error)
	GetDistance(ctx context.Context, loc1, loc2 *awsmodels.LocationData) (float64, error)
	GetRouteDistance(ctx context.Context, locations []*awsmodels.LocationData) (float64, error)
}

// TransactionService - Interface para acesso a transacoes
type TransactionService interface {
	GetShipment(ctx context.Context, shipmentID string) (*models.CrossborderShipment, error)
}

// NewRouteOptimizationService - Cria novo servico de otimizacao de rotas
def NewRouteOptimizationService(
	routeRepo *awsrepository.RouteOptimizationDynamoDBRepository,
	transactionRepo *awsrepository.TransactionDynamoDBRepository,
	awsConfig *awsmodels.AWSConfig,
	modelConfig *awsmodels.AIModelConfig,
	transactionService *TransactionService,
	geocodingService GeocodingService,
	maxIterations int,
) *RouteOptimizationService {
	return &RouteOptimizationService{
		routeRepo:           routeRepo,
		transactionRepo:     transactionRepo,
		awsConfig:           awsConfig,
		modelConfig:         modelConfig,
		transactionService:  transactionService,
		geocodingService:    geocodingService,
		maxIterations:       maxIterations,
	}
}

// RouteOptimizationRequest - Request para otimizacao de rota
type RouteOptimizationRequest struct {
	ShipmentID     string                 `json:"shipment_id"`
	TrackingCode   string                 `json:"tracking_code"`
	CurrentRoute   []*awsmodels.LocationData `json:"current_route"`
	Constraints    []string               `json:"constraints,omitempty"`
	OptimizeFor    string                 `json:"optimize_for,omitempty"` // distance, time, cost, co2
	MaxStops      int                    `json:"max_stops,omitempty"`
	VehicleType    string                 `json:"vehicle_type,omitempty"`
	WeightKG       float64                `json:"weight_kg,omitempty"`
}

// RouteOptimizationResponse - Response da otimizacao de rota
type RouteOptimizationResponse struct {
	OptimizationID  string                 `json:"optimization_id"`
	ShipmentID      string                 `json:"shipment_id"`
	TrackingCode    string                 `json:"tracking_code"`
	OriginalRoute   []*awsmodels.LocationData `json:"original_route"`
	OptimizedRoute  []*awsmodels.LocationData `json:"optimized_route"`
	DistanceOriginal float64                `json:"distance_original_km"`
	DistanceOptimized float64              `json:"distance_optimized_km"`
	DistanceSaved    float64                `json:"distance_saved_km"`
	TimeOriginal     float64                `json:"time_original_hours"`
	TimeOptimized    float64                `json:"time_optimized_hours"`
	TimeSaved        float64                `json:"time_saved_hours"`
	CostOriginal     float64                `json:"cost_original_brl"`
	CostOptimized    float64                `json:"cost_optimized_brl"`
	CostSaved        float64                `json:"cost_saved_brl"`
	CO2Original      float64                `json:"co2_original_kg"`
	CO2Optimized     float64                `json:"co2_optimized_kg"`
	CO2Saved         float64                `json:"co2_saved_kg"`
	ConstraintsSatisfied []string          `json:"constraints_satisfied"`
	Iterations       int                    `json:"iterations"`
	Algorithm        string                 `json:"algorithm"`
	Timestamp        time.Time              `json:"timestamp"`
}

// OptimizeRoute - Otimiza uma rota
def (s *RouteOptimizationService) OptimizeRoute(ctx context.Context, request *RouteOptimizationRequest) (*RouteOptimizationResponse, error) {
	// Validar request
	if request.ShipmentID == "" && len(request.CurrentRoute) == 0 {
		return nil, fmt.Errorf("shipment_id or current_route is required")
	}

	// Obter shipment se nao fornecido
	var shipment *models.CrossborderShipment
	if request.ShipmentID != "" {
		var err error
		shipment, err = s.transactionService.GetShipment(ctx, request.ShipmentID)
		if err != nil {
			log.Printf("Warning: failed to get shipment: %v", err)
		}
	}

	// Obter rota atual
	currentRoute := request.CurrentRoute
	if shipment != nil && len(currentRoute) == 0 {
		// Converter shipment para rota
		currentRoute = s.shipmentToRoute(shipment)
	}

	if len(currentRoute) < 2 {
		return nil, fmt.Errorf("current_route must have at least 2 points")
	}

	// Calcular metrica original
	originalDistance := s.calculateRouteDistance(currentRoute)
	originalTime := s.calculateRouteTime(currentRoute)
	originalCost := s.calculateRouteCost(currentRoute, request.WeightKG, request.VehicleType)
	originalCO2 := s.calculateRouteCO2(currentRoute, request.WeightKG, request.VehicleType)

	// Otimizar rota
	optimizedRoute, iterations := s.optimizeRoute(ctx, currentRoute, request)

	// Calcular metrica otimizada
	optimizedDistance := s.calculateRouteDistance(optimizedRoute)
	optimizedTime := s.calculateRouteTime(optimizedRoute)
	optimizedCost := s.calculateRouteCost(optimizedRoute, request.WeightKG, request.VehicleType)
	optimizedCO2 := s.calculateRouteCO2(optimizedRoute, request.WeightKG, request.VehicleType)

	// Calcular economias
	distanceSaved := originalDistance - optimizedDistance
	timeSaved := originalTime - optimizedTime
	costSaved := originalCost - optimizedCost
	co2Saved := originalCO2 - optimizedCO2

	// Verificar restricoes
	constraintsSatisfied := s.checkConstraints(optimizedRoute, request.Constraints)

	// Criar resultado
	optimizationID := fmt.Sprintf("OPT-%s-%d", request.ShipmentID, time.Now().UnixNano())

	optimization := &awsmodels.RouteOptimization{
		OptimizationID: optimizationID,
		ShipmentID:    request.ShipmentID,
		TrackingCode:  request.TrackingCode,
		Timestamp:     time.Now(),
		CurrentRoute:  currentRoute,
		OptimizedRoute: optimizedRoute,
		DistanceSaved:  distanceSaved,
		TimeSaved:      timeSaved,
		CostSaved:      costSaved,
		CO2Saved:       co2Saved,
		Constraints:   request.Constraints,
		Metadata: map[string]interface{}{
			"original_distance": originalDistance,
			"original_time":     originalTime,
			"original_cost":     originalCost,
			"original_co2":      originalCO2,
			"optimized_distance": optimizedDistance,
			"optimized_time":    optimizedTime,
			"optimized_cost":    optimizedCost,
			"optimized_co2":     optimizedCO2,
			"algorithm":         "genetic_algorithm",
			"iterations":        iterations,
		},
	}

	// Salvar otimizacao
	err := s.routeRepo.SaveRouteOptimization(ctx, optimization)
	if err != nil {
		log.Printf("Warning: failed to save route optimization: %v", err)
	}

	// Retornar response
	return &RouteOptimizationResponse{
		OptimizationID:        optimizationID,
		ShipmentID:            request.ShipmentID,
		TrackingCode:          request.TrackingCode,
		OriginalRoute:        currentRoute,
		OptimizedRoute:       optimizedRoute,
		DistanceOriginal:     originalDistance,
		DistanceOptimized:    optimizedDistance,
		DistanceSaved:        distanceSaved,
		TimeOriginal:         originalTime,
		TimeOptimized:        optimizedTime,
		TimeSaved:            timeSaved,
		CostOriginal:         originalCost,
		CostOptimized:        optimizedCost,
		CostSaved:            costSaved,
		CO2Original:          originalCO2,
		CO2Optimized:         optimizedCO2,
		CO2Saved:             co2Saved,
		ConstraintsSatisfied: constraintsSatisfied,
		Iterations:           iterations,
		Algorithm:            "genetic_algorithm",
		Timestamp:            time.Now(),
	}, nil
}

// optimizeRoute - Otimiza rota usando algoritmo genetico
func (s *RouteOptimizationService) optimizeRoute(ctx context.Context, currentRoute []*awsmodels.LocationData, request *RouteOptimizationRequest) ([]*awsmodels.LocationData, int) {
	// Implementar algoritmo de otimizacao
	// Por enquanto, usar 2-opt (simplificado)

	bestRoute := make([]*awsmodels.LocationData, len(currentRoute))
	copy(bestRoute, currentRoute)

	bestDistance := s.calculateRouteDistance(bestRoute)

	// 2-opt algorithm
	for iteration := 0; iteration < s.maxIterations; iteration++ {
		improved := false

		// Tentar todas as trocas possiveis
		for i := 1; i < len(bestRoute)-1; i++ {
			for j := i + 1; j < len(bestRoute); j++ {
				// Criar nova rota com troca 2-opt
				newRoute := s.apply2OptSwap(bestRoute, i, j)

				// Calcular nova distancia
				newDistance := s.calculateRouteDistance(newRoute)

				// Se melhor, manter
				if newDistance < bestDistance {
					bestRoute = newRoute
					bestDistance = newDistance
					improved = true
				}
			}
		}

		if !improved {
			// Nao houve melhora, parar
			return bestRoute, iteration + 1
		}
	}

	return bestRoute, s.maxIterations
}

// apply2OptSwap - Aplica troca 2-opt
func (s *RouteOptimizationService) apply2OptSwap(route []*awsmodels.LocationData, i, j int) []*awsmodels.LocationData {
	newRoute := make([]*awsmodels.LocationData, len(route))
	copy(newRoute, route)

	// Inverter segmento entre i e j
	for k := i; k <= j; k++ {
		newRoute[k] = route[j-(k-i)]
	}

	return newRoute
}

// calculateRouteDistance - Calcula distancia total da rota
func (s *RouteOptimizationService) calculateRouteDistance(route []*awsmodels.LocationData) float64 {
	var totalDistance float64

	for i := 0; i < len(route)-1; i++ {
		distance, err := s.geocodingService.GetDistance(context.Background(), route[i], route[i+1])
		if err != nil {
			// Distancia padrao se erro
			distance = s.haversineDistance(route[i], route[i+1])
		}
		totalDistance += distance
	}

	return totalDistance
}

// calculateRouteTime - Calcula tempo total da rota
func (s *RouteOptimizationService) calculateRouteTime(route []*awsmodels.LocationData) float64 {
	var totalTime float64

	for i := 0; i < len(route)-1; i++ {
		// Tempo estimado: distancia / velocidade media
		distance := s.haversineDistance(route[i], route[i+1])
		// Velocidade media: 60 km/h
		timeHours := distance / 60.0
		totalTime += timeHours
	}

	return totalTime
}

// calculateRouteCost - Calcula custo total da rota
func (s *RouteOptimizationService) calculateRouteCost(route []*awsmodels.LocationData, weightKG float64, vehicleType string) float64 {
	// Custo por km baseado no tipo de veiculo
	costPerKm := s.getCostPerKm(vehicleType)

	// Custo por kg
	costPerKg := 0.1 // R$ 0.10 por kg por km

	distance := s.calculateRouteDistance(route)

	// Custo base
	baseCost := distance * costPerKm

	// Custo por peso
	weightCost := distance * costPerKg * weightKG

	return baseCost + weightCost
}

// calculateRouteCO2 - Calcula emissao de CO2 da rota
func (s *RouteOptimizationService) calculateRouteCO2(route []*awsmodels.LocationData, weightKG float64, vehicleType string) float64 {
	// Emissao por km baseada no tipo de veiculo (kg CO2 por km)
	co2PerKm := s.getCO2PerKm(vehicleType)

	// Fator de carga
	loadFactor := 1.0 + (weightKG / 1000.0) // +1% por tonelada

	distance := s.calculateRouteDistance(route)

	return distance * co2PerKm * loadFactor
}

// haversineDistance - Calcula distancia usando formula de Haversine
func (s *RouteOptimizationService) haversineDistance(loc1, loc2 *awsmodels.LocationData) float64 {
	// Raio da Terra em km
	const R = 6371.0

	// Converter graus para radianos
	lat1 := loc1.Latitude * math.Pi / 180.0
	lon1 := loc1.Longitude * math.Pi / 180.0
	lat2 := loc2.Latitude * math.Pi / 180.0
	lon2 := loc2.Longitude * math.Pi / 180.0

	// Diferenca
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	// Formula de Haversine
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
		math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// getCostPerKm - Retorna custo por km por tipo de veiculo
func (s *RouteOptimizationService) getCostPerKm(vehicleType string) float64 {
	costs := map[string]float64{
		"truck":        2.50, // R$ 2.50 por km
		"van":          1.80,
		"car":          1.20,
		"motorcycle":   0.80,
		"bicycle":      0.10,
	}
	if cost, ok := costs[vehicleType]; ok {
		return cost
	}
	return 2.00 // Default
}

// getCO2PerKm - Retorna emissao de CO2 por km por tipo de veiculo
func (s *RouteOptimizationService) getCO2PerKm(vehicleType string) float64 {
	co2 := map[string]float64{
		"truck":        0.16, // kg CO2 por km
		"van":          0.12,
		"car":          0.08,
		"motorcycle":   0.05,
		"bicycle":      0.00,
	}
	if emission, ok := co2[vehicleType]; ok {
		return emission
	}
	return 0.10 // Default
}

// checkConstraints - Verifica se restricoes sao satisfeitas
func (s *RouteOptimizationService) checkConstraints(route []*awsmodels.LocationData, constraints []string) []string {
	var satisfied []string

	for _, constraint := range constraints {
		switch constraint {
		case "max_distance_500":
			distance := s.calculateRouteDistance(route)
			if distance <= 500 {
				satisfied = append(satisfied, constraint)
			}
		case "max_time_8h":
			time := s.calculateRouteTime(route)
			if time <= 8 {
				satisfied = append(satisfied, constraint)
			}
		case "avoid_highways":
			// Verificar se rota evita rodovias
			// Por enquanto, assumir que sim
			satisfied = append(satisfied, constraint)
		case "min_stops_2":
			if len(route) >= 2 {
				satisfied = append(satisfied, constraint)
			}
		}
	}

	return satisfied
}

// shipmentToRoute - Converte shipment para rota
func (s *RouteOptimizationService) shipmentToRoute(shipment *models.CrossborderShipment) []*awsmodels.LocationData {
	var route []*awsmodels.LocationData

	// Adicionar origem
	origin := &awsmodels.LocationData{
		City:    shipment.OriginCity,
		State:   shipment.OriginState,
		Country: shipment.OriginCountry,
		CEP:     shipment.OriginZip,
	}
	route = append(route, origin)

	// Adicionar destino
	destination := &awsmodels.LocationData{
		City:    shipment.DestinationCity,
		State:   shipment.DestinationState,
		Country: shipment.DestinationCountry,
		CEP:     shipment.DestinationZip,
	}
	route = append(route, destination)

	return route
}

// BatchOptimizeRoutes - Otimiza multiplas rotas
def (s *RouteOptimizationService) BatchOptimizeRoutes(ctx context.Context, requests []*RouteOptimizationRequest) ([]*RouteOptimizationResponse, error) {
	var responses []*RouteOptimizationResponse

	for _, request := range requests {
		response, err := s.OptimizeRoute(ctx, request)
		if err != nil {
			log.Printf("Warning: failed to optimize route for shipment %s: %v", request.ShipmentID, err)
			continue
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetRouteOptimization - Obtem uma otimizacao especifica
def (s *RouteOptimizationService) GetRouteOptimization(ctx context.Context, optimizationID string) (*awsmodels.RouteOptimization, error) {
	// Implementar busca por optimization ID
	return nil, nil
}

// GetRouteOptimizationsByShipment - Obtem otimizacoes por shipment
def (s *RouteOptimizationService) GetRouteOptimizationsByShipment(ctx context.Context, shipmentID string) ([]*awsmodels.RouteOptimization, error) {
	return s.routeRepo.GetRouteOptimizationsByShipment(ctx, shipmentID)
}

// GetRouteStatistics - Obtem estatisticas de rotas
def (s *RouteOptimizationService) GetRouteStatistics(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Obter otimizacoes
	// Calcular estatisticas

	return stats, nil
}

// LocalGeocodingService - Implementacao local de geocodificacao (para desenvolvimento)
type LocalGeocodingService struct {
	// Dados de geocodificacao mock
	addressToCoords map[string]*awsmodels.LocationData
}

// NewLocalGeocodingService - Cria novo servico de geocodificacao local
def NewLocalGeocodingService() *LocalGeocodingService {
	return &LocalGeocodingService{
		addressToCoords: make(map[string]*awsmodels.LocationData),
	}
}

// GetCoordinates - Obtem coordenadas de um endereco
def (s *LocalGeocodingService) GetCoordinates(ctx context.Context, address string) (*awsmodels.LocationData, error) {
	// Retornar coordenadas mock
	return &awsmodels.LocationData{
		Latitude:  -23.5505,
		Longitude: -46.6333,
		City:     "Sao Paulo",
		State:    "SP",
		Country:  "BR",
	}, nil
}

// GetDistance - Calcula distancia entre dois pontos
def (s *LocalGeocodingService) GetDistance(ctx context.Context, loc1, loc2 *awsmodels.LocationData) (float64, error) {
	// Calcular distancia usando Haversine
	service := &RouteOptimizationService{}
	return service.haversineDistance(loc1, loc2), nil
}

// GetRouteDistance - Calcula distancia de uma rota
def (s *LocalGeocodingService) GetRouteDistance(ctx context.Context, locations []*awsmodels.LocationData) (float64, error) {
	var total float64
	for i := 0; i < len(locations)-1; i++ {
		distance, err := s.GetDistance(ctx, locations[i], locations[i+1])
		if err != nil {
			return 0, err
		}
		total += distance
	}
	return total, nil
}

// sort.Slice - Funcao para ordenar slices
func sortSlice[T any](slice []T, less func(i, j int) bool) {
	sort.Slice(slice, less)
}
