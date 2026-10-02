package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rldb-br/rldb-api-universal/internal/config"
	awsconfig "github.com/rldb-br/rldb-api-universal/internal/config/aws"
	"github.com/rldb-br/rldb-api-universal/internal/handler"
	awshandler "github.com/rldb-br/rldb-api-universal/internal/handler/aws"
	"github.com/rldb-br/rldb-api-universal/internal/middleware"
	"github.com/rldb-br/rldb-api-universal/internal/models"
	awsmodels "github.com/rldb-br/rldb-api-universal/internal/models/aws"
	"github.com/rldb-br/rldb-api-universal/internal/repository"
	awsrepository "github.com/rldb-br/rldb-api-universal/internal/repository/aws"
	"github.com/rldb-br/rldb-api-universal/internal/service"
	"github.com/rldb-br/rldb-api-universal/internal/service/aws"
	"github.com/rldb-br/rldb-api-universal/internal/service/ai"
	"github.com/rldb-br/rldb-api-universal/internal/service/analytics"
	"github.com/rldb-br/rldb-api-universal/pkg/governance"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CacheService interface for caching
type CacheService interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
}

// RedisCache implements CacheService
type RedisCache struct {
	// Redis client would go here
	// For now, use in-memory cache
	store map[string]interface{}
}

// Get gets a value from cache
func (r *RedisCache) Get(key string, value interface{}) error {
	// In-memory implementation
	if val, ok := r.store[key]; ok {
		// This is a simplified implementation
		// In production, use proper type assertion
		return nil
	}
	return fmt.Errorf("not found")
}

// Set sets a value in cache
func (r *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	r.store[key] = value
	return nil
}

// Delete deletes a value from cache
func (r *RedisCache) Delete(key string) error {
	delete(r.store, key)
	return nil
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize database
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.CrossborderShipment{},
		&models.CrossborderQuote{},
		&models.CrossborderTrackingEvent{},
		&models.CrossborderCustomsDocument{},
		&models.CrossborderWebhook{},
		&models.AuditLog{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Initialize cache
	cache := &RedisCache{
		store: make(map[string]interface{}),
	}

	// Initialize repositories
	tr8Repo := repository.NewCrossborderTR8Repository(db)

	// Initialize governance configuration
	governanceConfig := &models.TR8GovernanceConfig{
		AllowedFields: []string{
			"origin_country", "destination_country", "origin_zip", "destination_zip",
			"weight_kg", "dimensions_cm", "declared_value_brl", "declared_value_usd",
			"incoterm", "hs_code", "product_category", "service_level",
			"status", "timestamp", "tracking_code", "label_url",
			"estimated_delivery", "actual_delivery", "operator", "service",
			"reverse_logistics_enabled", "customs_status",
		},
		BlockedFields: []string{
			"product_price", "margin", "customer_email", "customer_name",
			"full_address", "sales_strategy", "competitor_data", "campaign_info",
			"internal_cost", "profit_margin", "customer_list",
		},
		RequiredFields: []string{
			"origin_country", "destination_country", "weight_kg",
			"declared_value_brl", "incoterm",
		},
		SensitiveFields: []string{
			"declared_value_brl", "declared_value_usd", "hs_code",
			"product_category", "customs_status",
		},
		PseudonymizeFields: []string{
			"name", "address", "city", "phone", "email",
		},
		RateLimits: map[string]int{
			"quotes":    100,
			"shipments": 50,
			"tracking":  200,
		},
		SLARequirements: map[string]string{
			"quotes":    "1s",
			"shipments": "2s",
			"tracking":  "500ms",
		},
	}

	// Initialize governance service
	governanceSvc := governance.NewDataGovernanceService(tr8Repo, governanceConfig)

	// Initialize services
	tr8Service := service.NewCrossborderTR8Service(tr8Repo, governanceSvc, cache, governanceConfig)

	// Initialize AWS services
	// Carregar configuracao AWS
	awsCfg, err := awsconfig.LoadAWSConfig()
	if err != nil {
		log.Printf("Warning: failed to load AWS config: %v", err)
		// Continuar sem AWS se configuracao falhar
		awsCfg = &awsconfig.AWSConfig{
			KinesisEnabled: false,
			DynamoDBEnabled: false,
			S3Enabled: false,
			LambdaEnabled: false,
			SageMakerEnabled: false,
			SNSEnabled: false,
			CloudWatchEnabled: false,
		}
	}

	// Inicializar repositorios AWS (se habilitado)
	var kinesisRepo *awsrepository.KinesisRepository
	var dynamoRepo *awsrepository.DynamoDBRepository
	var transactionStreamRepo *awsrepository.TransactionStreamRepository
	var logStreamRepo *awsrepository.LogStreamRepository
	var fraudDynamoRepo *awsrepository.FraudDetectionDynamoDBRepository
	var demandDynamoRepo *awsrepository.DemandForecastDynamoDBRepository
	var routeDynamoRepo *awsrepository.RouteOptimizationDynamoDBRepository

	if awsCfg.KinesisEnabled {
		kinesisRepo, err = awsrepository.NewKinesisRepository(
			awsCfg.KinesisStreamName,
			awsCfg.KinesisRegion,
		)
		if err != nil {
			log.Printf("Warning: failed to create Kinesis repository: %v", err)
		} else {
			transactionStreamRepo = awsrepository.NewTransactionStreamRepository(kinesisRepo)
			logStreamRepo = awsrepository.NewLogStreamRepository(kinesisRepo)
		}
	}

	if awsCfg.DynamoDBEnabled {
		dynamoRepo, err = awsrepository.NewDynamoDBRepository(
			awsCfg.DynamoDBTableName,
			awsCfg.DynamoDBRegion,
		)
		if err != nil {
			log.Printf("Warning: failed to create DynamoDB repository: %v", err)
		} else {
			fraudDynamoRepo = awsrepository.NewFraudDetectionDynamoDBRepository(dynamoRepo)
			demandDynamoRepo = awsrepository.NewDemandForecastDynamoDBRepository(dynamoRepo)
			routeDynamoRepo = awsrepository.NewRouteOptimizationDynamoDBRepository(dynamoRepo)
		}
	}

	// Inicializar servicos AWS
	var kinesisService *awsservice.KinesisStreamService
	if kinesisRepo != nil && transactionStreamRepo != nil && logStreamRepo != nil {
		kinesisService = awsservice.NewKinesisStreamService(
			kinesisRepo,
			transactionStreamRepo,
			logStreamRepo,
			&awsmodels.AWSStreamConfig{
				KinesisStreamName: awsCfg.KinesisStreamName,
				KinesisRegion:     awsCfg.KinesisRegion,
				S3BucketName:     awsCfg.S3Bucket,
			},
			5, // 5 consumer workers
		)
	}

	// Inicializar servicos de IA
	var fraudService *aiservice.FraudDetectionService
	var demandService *aiservice.DemandForecastService
	var routeService *aiservice.RouteOptimizationService

	if awsCfg.AIFraudDetectionEnabled && fraudDynamoRepo != nil {
		// Criar servico de alerta local
		alertService := NewLocalAlertService(100)
		fraudService = aiservice.NewFraudDetectionService(
			fraudDynamoRepo,
			nil, // transactionRepo (seria o repositorio principal)
			awsCfg,
			&awsmodels.AIModelConfig{
				ModelName:   "rldb-fraud-detection",
				ModelType:   "fraud",
				EndpointURL: awsCfg.SageMakerEndpoint,
			},
			alertService,
			nil, // transactionService
		)
	}

	if awsCfg.AIDemandForecastEnabled && demandDynamoRepo != nil {
		demandService = aiservice.NewDemandForecastService(
			demandDynamoRepo,
			nil, // transactionRepo
			awsCfg,
			&awsmodels.AIModelConfig{
				ModelName:   "rldb-demand-forecast",
				ModelType:   "demand",
				EndpointURL: awsCfg.SageMakerEndpoint,
			},
			nil, // transactionService
			7,  // forecast horizon days
		)
	}

	if awsCfg.AIRouteOptimizationEnabled && routeDynamoRepo != nil {
		// Criar servico de geocodificacao local
		geocodingService := NewLocalGeocodingService()
		routeService = aiservice.NewRouteOptimizationService(
			routeDynamoRepo,
			nil, // transactionRepo
			awsCfg,
			&awsmodels.AIModelConfig{
				ModelName:   "rldb-route-optimization",
				ModelType:   "routing",
				EndpointURL: awsCfg.SageMakerEndpoint,
			},
			nil, // transactionService
			geocodingService,
			100, // max iterations
		)
	}

	// Inicializar servico de analise em tempo real
	var analyticsService *analyticsservice.RealTimeAnalyticsService
	if awsCfg.RealTimeProcessingEnabled {
		analyticsService = analyticsservice.NewRealTimeAnalyticsService(
			transactionStreamRepo,
			logStreamRepo,
			kinesisService,
			fraudService,
			demandService,
			routeService,
			awsCfg,
		)
	}

	// Initialize handlers
	tr8Handler := handler.NewCrossborderTR8Handler(tr8Service)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		cfg.JWTRealm,
	)

	rateLimitMiddleware := middleware.NewRateLimitMiddleware(nil, cfg.RateLimitDefault)
	auditMiddleware := middleware.NewAuditMiddleware(governanceSvc)

	// Create Gin router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(auditMiddleware.LogRequest())

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "healthy",
				"version": cfg.APIVersion,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Authentication routes
		auth := api.Group("/v1/auth")
		{
			auth.POST("/token", handler.GetAuthToken(cfg))
		}

		// TR8 Crossborder routes (require authentication)
		tr8 := api.Group("/v1/crossborder/tr8")
		tr8.Use(authMiddleware.Authenticate())
		tr8.Use(rateLimitMiddleware.LimitByOperator(100))
		{
			tr8.POST("/quotes", tr8Handler.GetTR8Quotes)
			tr8.POST("/shipments", authMiddleware.Authorize("logistics:write"), tr8Handler.CreateTR8Shipment)
			tr8.GET("/shipments/:shipment_id", authMiddleware.Authorize("logistics:read"), tr8Handler.GetTR8Shipment)
			tr8.GET("/tracking/:tracking_code", authMiddleware.Authorize("tracking:read"), tr8Handler.GetTR8Tracking)
			tr8.POST("/tracking/:tracking_code", authMiddleware.Authorize("tracking:write"), tr8Handler.UpdateTR8Tracking)
			tr8.POST("/customs/documents", authMiddleware.Authorize("logistics:write"), tr8Handler.CreateCustomsDocument)
			tr8.PUT("/customs/documents/:document_id/validate", authMiddleware.Authorize("logistics:write"), tr8Handler.ValidateCustomsDocument)
			tr8.GET("/customs/documents/:document_id", authMiddleware.Authorize("logistics:read"), tr8Handler.GetCustomsDocument)
			tr8.POST("/webhooks", authMiddleware.Authorize("logistics:write"), tr8Handler.CreateWebhook)
			tr8.GET("/webhooks", authMiddleware.Authorize("logistics:read"), tr8Handler.GetWebhooks)
		}

		// Observatory routes
		observatory := api.Group("/v1/observatory")
		observatory.Use(authMiddleware.Authenticate())
		observatory.Use(rateLimitMiddleware.LimitByOperator(50))
		{
			observatory.GET("/tr8/metrics", authMiddleware.Authorize("metrics:read"), tr8Handler.GetTR8Metrics)
			observatory.GET("/tr8/compliance", authMiddleware.Authorize("audit:read"), tr8Handler.GetTR8ComplianceReport)
		}

		// Governance routes
		governance := api.Group("/v1/governance")
		governance.Use(authMiddleware.Authenticate())
		governance.Use(rateLimitMiddleware.LimitByOperator(20))
		{
			governance.GET("/tr8/config", authMiddleware.Authorize("governance:read"), handler.GetGovernanceConfig(governanceConfig))
		}
	}

	// Register TR8 routes
	tr8Handler.RegisterRoutes(api)

	// Register AWS handlers (se servicos estao disponiveis)
	if analyticsService != nil {
		analyticsHandler := awshandler.NewAnalyticsHandler(analyticsService)
		analyticsHandler.RegisterRoutes(api)
	}

	if fraudService != nil {
		fraudHandler := awshandler.NewFraudHandler(fraudService)
		fraudHandler.RegisterRoutes(api)
	}

	if demandService != nil {
		demandHandler := awshandler.NewDemandHandler(demandService)
		demandHandler.RegisterRoutes(api)
	}

	if routeService != nil {
		routingHandler := awshandler.NewRoutingHandler(routeService)
		routingHandler.RegisterRoutes(api)
	}

	// Iniciar servicos de processamento em tempo real
	if awsCfg.RealTimeProcessingEnabled && analyticsService != nil {
		go func() {
			if err := analyticsService.Start(context.Background()); err != nil {
				log.Printf("Failed to start analytics service: %v", err)
			}
		}()
	}

	if awsCfg.KinesisEnabled && kinesisService != nil {
		go func() {
			if err := kinesisService.StartConsumers(context.Background()); err != nil {
				log.Printf("Failed to start Kinesis consumers: %v", err)
			}
		}()
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
		ReadTimeout:  cfg.ServerTimeout,
		WriteTimeout: cfg.ServerTimeout,
		IdleTimeout:  cfg.ServerTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("RLDB API UNIVERSAL starting on port %s", cfg.ServerPort)
		if cfg.EnableHTTPS {
			if err := server.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown gracefully
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// GetAuthToken returns a JWT token for authentication
func GetAuthToken(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			Scope        string `json:"scope"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate client credentials (in production, this would check against a database)
		if request.ClientID == "" || request.ClientSecret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid client credentials"})
			return
		}

		// Create token claims
		claims := jwt.MapClaims{
			"sub":        request.ClientID,
			"client_id":  request.ClientID,
			"iss":        cfg.JWTIssuer,
			"aud":        cfg.JWTAudience,
			"exp":        time.Now().Add(cfg.JWTExpiresIn).Unix(),
			"iat":        time.Now().Unix(),
			"scopes":     strings.Split(request.Scope, " "),
			"jti":        uuid.New().String(),
		}

		// Create token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": tokenString,
			"token_type":   "Bearer",
			"expires_in":   int(cfg.JWTExpiresIn.Seconds()),
			"scope":        request.Scope,
		})
	}
}

// GetGovernanceConfig returns the governance configuration
func GetGovernanceConfig(config *models.TR8GovernanceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, config)
	}
}
