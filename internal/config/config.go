package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	// Server configuration
	ServerPort    string
	ServerTimeout time.Duration

	// Database configuration
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis configuration
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// JWT configuration
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	JWTRealm     string
	JWTExpiresIn time.Duration

	// Rate limiting
	RateLimitDefault int
	RateLimitPerMin  int

	// Logging
	LogLevel string

	// API configuration
	APIVersion string
	BaseURL    string

	// Security
	EnableHTTPS bool
	TLSCertFile string
	TLSKeyFile  string

	// Crossborder TR8
	TR8Enabled          bool
	TR8DefaultOperator  string
	TR8MaxWeightKg      float64
	TR8SupportedCountries []string

	// Governance
	GovernanceEnabled bool
	AuditEnabled      bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		// Ignore error if .env file doesn't exist
		_ = err
	}

	config := &Config{
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		ServerTimeout:       getEnvDuration("SERVER_TIMEOUT", 30*time.Second),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnvInt("DB_PORT", 5432),
		DBUser:              getEnv("DB_USER", "rldb_user"),
		DBPassword:          getEnv("DB_PASSWORD", "rldb_password"),
		DBName:              getEnv("DB_NAME", "rldb_api"),
		DBSSLMode:           getEnv("DB_SSL_MODE", "disable"),
		RedisHost:           getEnv("REDIS_HOST", "localhost"),
		RedisPort:           getEnvInt("REDIS_PORT", 6379),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		RedisDB:             getEnvInt("REDIS_DB", 0),
		JWTSecret:          getEnv("JWT_SECRET", "rldb-secret-key-change-me"),
		JWTIssuer:          getEnv("JWT_ISSUER", "rldb-api-universal"),
		JWTAudience:        getEnv("JWT_AUDIENCE", "rldb-clients"),
		JWTRealm:           getEnv("JWT_REALM", "RLDB API"),
		JWTExpiresIn:       getEnvDuration("JWT_EXPIRES_IN", 15*time.Minute),
		RateLimitDefault:   getEnvInt("RATE_LIMIT_DEFAULT", 100),
		RateLimitPerMin:    getEnvInt("RATE_LIMIT_PER_MIN", 60),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		APIVersion:         getEnv("API_VERSION", "v1"),
		BaseURL:            getEnv("BASE_URL", "https://api.rldb.gov.br"),
		EnableHTTPS:        getEnvBool("ENABLE_HTTPS", false),
		TLSCertFile:       getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:        getEnv("TLS_KEY_FILE", ""),
		TR8Enabled:         getEnvBool("TR8_ENABLED", true),
		TR8DefaultOperator: getEnv("TR8_DEFAULT_OPERATOR", "correios_international"),
		TR8MaxWeightKg:     getEnvFloat("TR8_MAX_WEIGHT_KG", 100.0),
		TR8SupportedCountries: getEnvSlice("TR8_SUPPORTED_COUNTRIES", []string{"BR", "US", "CN", "DE", "FR", "UK"}),
		GovernanceEnabled:  getEnvBool("GOVERNANCE_ENABLED", true),
		AuditEnabled:       getEnvBool("AUDIT_ENABLED", true),
	}

	return config, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as int with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvFloat gets an environment variable as float64 with a default value
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as bool with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// getEnvDuration gets an environment variable as time.Duration with a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvSlice gets an environment variable as a slice of strings with a default value
func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return splitAndTrim(value, ",")
	}
	return defaultValue
}

// splitAndTrim splits a string by delimiter and trims whitespace
func splitAndTrim(s string, delimiter string) []string {
	result := []string{}
	for _, part := range strings.Split(s, delimiter) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.LogLevel == "prod" || c.LogLevel == "production"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Check required fields
	if c.JWTSecret == "rldb-secret-key-change-me" {
		return fmt.Errorf("JWT_SECRET must be changed from default value")
	}

	// Check database configuration
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}

	// Check TR8 configuration
	if c.TR8Enabled && len(c.TR8SupportedCountries) == 0 {
		return fmt.Errorf("TR8_SUPPORTED_COUNTRIES is required when TR8 is enabled")
	}

	return nil
}
