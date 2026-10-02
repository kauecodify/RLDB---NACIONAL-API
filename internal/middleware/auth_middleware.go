package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	jwtSecret     string
	issuer        string
	audience      string
	realm         string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtSecret, issuer, audience, realm string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		issuer:    issuer,
		audience:  audience,
		realm:     realm,
	}
}

// Authenticate validates JWT tokens
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Header("WWW-Authenticate", m.realm)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Header("WWW-Authenticate", m.realm)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			// Validate issuer
			if token.Claims.(jwt.MapClaims)["iss"] != m.issuer {
				return nil, jwt.ErrTokenInvalidIssuer
			}

			// Validate audience
			if token.Claims.(jwt.MapClaims)["aud"] != m.audience {
				return nil, jwt.ErrTokenInvalidAudience
			}

			return []byte(m.jwtSecret), nil
		})

		if err != nil {
			c.Header("WWW-Authenticate", m.realm)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Validate token claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.Header("WWW-Authenticate", m.realm)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Check expiration
		if exp, ok := claims["exp"].(float64); ok {
			expTime := time.Unix(int64(exp), 0)
			if time.Now().After(expTime) {
				c.Header("WWW-Authenticate", m.realm)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Token expired",
				})
				return
			}
		}

		// Extract operator ID
		operatorID, ok := claims["sub"].(string)
		if !ok {
			c.Header("WWW-Authenticate", m.realm)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token subject",
			})
			return
		}

		// Extract client ID
		clientID, ok := claims["client_id"].(string)
		if !ok {
			clientID = operatorID
		}

		// Extract scopes
		scopes, ok := claims["scopes"].([]interface{})
		if !ok {
			scopes = []interface{}{}
		}

		// Set context values
		c.Set("operator_id", operatorID)
		c.Set("client_id", clientID)
		c.Set("scopes", scopes)
		c.Set("token", tokenString)

		// Continue
		c.Next()
	}
}

// Authorize checks if the operator has the required scopes
func (m *AuthMiddleware) Authorize(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get scopes from context
		scopes, exists := c.Get("scopes")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "No scopes found in token",
			})
			return
		}

		// Convert to map for easier lookup
		scopeMap := make(map[string]bool)
		for _, scope := range scopes.([]interface{}) {
			if s, ok := scope.(string); ok {
				scopeMap[s] = true
			}
		}

		// Check required scopes
		for _, required := range requiredScopes {
			if !scopeMap[required] {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "Insufficient permissions",
					"required_scope": required,
				})
				return
			}
		}

		// Continue
		c.Next()
	}
}

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	store      RateLimitStore
	defaultLimit int
}

// RateLimitStore interface for rate limiting storage
type RateLimitStore interface {
	Increment(key string) (int, error)
	Reset(key string) error
	Get(key string) (int, error)
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(store RateLimitStore, defaultLimit int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		store:       store,
		defaultLimit: defaultLimit,
	}
}

// LimitByOperator limits requests by operator
func (m *RateLimitMiddleware) LimitByOperator(limit int) gin.HandlerFunc {
	if limit <= 0 {
		limit = m.defaultLimit
	}

	return func(c *gin.Context) {
		// Get operator ID
		operatorID, exists := c.Get("operator_id")
		if !exists {
			c.Next()
			return
		}

		// Generate key
		key := fmt.Sprintf("rate_limit:%s", operatorID)

		// Increment counter
		count, err := m.store.Increment(key)
		if err != nil {
			c.Next()
			return
		}

		// Check limit
		if count > limit {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Unix()+3600))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": 3600,
			})
			return
		}

		// Set headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Unix()+3600))

		// Continue
		c.Next()
	}
}

// AuditMiddleware logs all requests for audit purposes
type AuditMiddleware struct {
	governanceSvc interface {
		LogAudit(ctx context.Context, operatorID string, endpoint string, action string, data interface{})
	}
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(governanceSvc interface{}) *AuditMiddleware {
	return &AuditMiddleware{
		governanceSvc: governanceSvc,
	}
}

// LogRequest logs all requests
func (m *AuditMiddleware) LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Continue
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get operator ID
		operatorID, _ := c.Get("operator_id")
		if operatorID == nil {
			operatorID = "anonymous"
		}

		// Get endpoint
		endpoint := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)

		// Get status code
		statusCode := c.Writer.Status()

		// Log audit
		m.governanceSvc.LogAudit(c.Request.Context(), operatorID.(string), endpoint, "request_completed", map[string]interface{}{
			"status_code":    statusCode,
			"response_time_ms": duration.Milliseconds(),
			"ip_address":    c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
		})
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
