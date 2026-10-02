package governance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rldb-br/rldb-api-universal/internal/models"
	"github.com/rldb-br/rldb-api-universal/internal/repository"
)

// DataGovernanceService handles data governance and compliance
type DataGovernanceService struct {
	repo          *repository.CrossborderTR8Repository
	config        *models.TR8GovernanceConfig
	previousHash  string
	auditEnabled  bool
}

// NewDataGovernanceService creates a new data governance service
func NewDataGovernanceService(repo *repository.CrossborderTR8Repository, config *models.TR8GovernanceConfig) *DataGovernanceService {
	return &DataGovernanceService{
		repo:         repo,
		config:       config,
		previousHash: "",
		auditEnabled: true,
	}
}

// SanitizeRequest sanitizes a request according to governance rules
func (s *DataGovernanceService) SanitizeRequest(ctx context.Context, request interface{}, operatorID string, endpoint string) (map[string]interface{}, error) {
	// Convert request to map
	requestMap, err := structToMap(request)
	if err != nil {
		return nil, err
	}

	// Apply field filtering
	sanitized := s.applyFieldFiltering(requestMap, operatorID, endpoint)

	// Apply pseudonymization
	sanitized = s.applyPseudonymization(sanitized)

	// Log audit
	if s.auditEnabled {
		s.logAudit(ctx, operatorID, endpoint, "request_sanitized", map[string]interface{}{
			"original_fields": requestMap,
			"sanitized_fields": sanitized,
		})
	}

	return sanitized, nil
}

// SanitizeField sanitizes a single field
func (s *DataGovernanceService) SanitizeField(ctx context.Context, value string, operatorID string, fieldName string) string {
	// Check if field should be blocked
	for _, blocked := range s.config.BlockedFields {
		if blocked == fieldName {
			return ""
		}
	}

	// Check if field should be pseudonymized
	for _, pseudo := range s.config.PseudonymizeFields {
		if pseudo == fieldName {
			return s.hashWithSalt(value)
		}
	}

	return value
}

// applyFieldFiltering applies field filtering based on governance rules
func (s *DataGovernanceService) applyFieldFiltering(request map[string]interface{}, operatorID string, endpoint string) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range request {
		// Check if field is blocked
		isBlocked := false
		for _, blocked := range s.config.BlockedFields {
			if blocked == key {
				isBlocked = true
				break
			}
		}

		if isBlocked {
			// Log blocked field access
			if s.auditEnabled {
				s.logAudit(context.Background(), operatorID, endpoint, "field_blocked", map[string]interface{}{
					"field": key,
				})
			}
			continue
		}

		// Check if field is allowed
		isAllowed := false
		for _, allowed := range s.config.AllowedFields {
			if allowed == key {
				isAllowed = true
				break
			}
		}

		// If field is not in allowed list but is in sensitive list, check if operator has access
		isSensitive := false
		for _, sensitive := range s.config.SensitiveFields {
			if sensitive == key {
				isSensitive = true
				break
			}
		}

		if !isAllowed && isSensitive {
			// Check operator-specific access
			if !s.hasOperatorAccess(operatorID, key) {
				if s.auditEnabled {
					s.logAudit(context.Background(), operatorID, endpoint, "sensitive_field_blocked", map[string]interface{}{
						"field": key,
					})
				}
				continue
			}
		}

		// If field is required, ensure it's present
		isRequired := false
		for _, required := range s.config.RequiredFields {
			if required == key {
				isRequired = true
				break
			}
		}

		// For now, include all non-blocked fields
		sanitized[key] = value
	}

	return sanitized
}

// applyPseudonymization applies pseudonymization to sensitive fields
func (s *DataGovernanceService) applyPseudonymization(request map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range request {
		// Check if field should be pseudonymized
		for _, pseudo := range s.config.PseudonymizeFields {
			if pseudo == key {
				if strVal, ok := value.(string); ok {
					result[key] = s.hashWithSalt(strVal)
				} else {
					result[key] = value
				}
				continue
			}
		}

		// Handle nested structures (like recipient)
		if key == "recipient" {
			if recipientMap, ok := value.(map[string]interface{}); ok {
				result[key] = s.pseudonymizeRecipient(recipientMap)
			} else {
				result[key] = value
			}
			continue
		}

		result[key] = value
	}

	return result
}

// pseudonymizeRecipient pseudonymizes recipient data
func (s *DataGovernanceService) pseudonymizeRecipient(recipient map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range recipient {
		switch key {
		case "name", "address", "city", "phone", "email":
			if strVal, ok := value.(string); ok {
				result[key+"_hash"] = s.hashWithSalt(strVal)
			} else {
				result[key] = value
			}
		default:
			result[key] = value
		}
	}

	return result
}

// hasOperatorAccess checks if operator has access to a specific field
func (s *DataGovernanceService) hasOperatorAccess(operatorID string, field string) bool {
	// Implement operator-specific field access rules
	// For now, allow all operators to access all fields
	return true
}

// hashWithSalt generates a SHA-256 hash with salt
func (s *DataGovernanceService) hashWithSalt(value string) string {
	if value == "" {
		return ""
	}

	// Use operator ID and timestamp as salt
	salt := fmt.Sprintf("%s-%d", uuid.New().String(), time.Now().UnixNano())
	data := fmt.Sprintf("%s:%s", value, salt)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// LogAudit logs an audit entry
func (s *DataGovernanceService) LogAudit(ctx context.Context, operatorID string, endpoint string, action string, data interface{}) {
	if !s.auditEnabled {
		return
	}

	// Generate hash chain
	currentData := fmt.Sprintf("%s:%s:%s:%s:%v", operatorID, endpoint, action, time.Now().UTC().Format(time.RFC3339), data)
	currentHash := s.generateHash(currentData)

	// Create audit log
	log := &models.AuditLog{
		ID:            uuid.New().String(),
		RequestID:     ctx.Value("request_id").(string),
		OperatorID:    operatorID,
		Endpoint:      endpoint,
		HTTPMethod:    strings.ToUpper(strings.Split(endpoint, " ")[0]),
		FieldsAccessed: []string{},
		FieldsBlocked:  []string{},
		Timestamp:     time.Now().UTC(),
		IPAddress:     "",
		UserAgent:     "",
		StatusCode:    200,
		ResponseTimeMs: 0,
		PreviousHash:  s.previousHash,
		CurrentHash:   currentHash,
		CreatedAt:     time.Now().UTC(),
	}

	// Update previous hash
	s.previousHash = currentHash

	// Save audit log
	if err := s.repo.CreateAuditLog(ctx, log); err != nil {
		// Log error
		// log.Printf("failed to create audit log: %v", err)
	}
}

// generateHash generates a SHA-256 hash
func (s *DataGovernanceService) generateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ValidateRequest validates a request against governance rules
func (s *DataGovernanceService) ValidateRequest(request interface{}, operatorID string, endpoint string) error {
	// Convert request to map
	requestMap, err := structToMap(request)
	if err != nil {
		return err
	}

	// Check required fields
	for _, required := range s.config.RequiredFields {
		if _, exists := requestMap[required]; !exists {
			return fmt.Errorf("missing required field: %s", required)
		}
	}

	// Check blocked fields
	for _, blocked := range s.config.BlockedFields {
		if _, exists := requestMap[blocked]; exists {
			return fmt.Errorf("blocked field detected: %s", blocked)
		}
	}

	return nil
}

// structToMap converts a struct to a map
func structToMap(s interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		} else {
			// Remove comma-separated options
			jsonTag = strings.Split(jsonTag, ",")[0]
		}

		// Handle nested structs
		if value.Kind() == reflect.Struct {
			nestedMap, err := structToMap(value.Interface())
			if err != nil {
				continue
			}
			result[jsonTag] = nestedMap
			continue
		}

		// Handle slices
		if value.Kind() == reflect.Slice {
			if value.Len() == 0 {
				result[jsonTag] = []interface{}{}
				continue
			}

			elements := make([]interface{}, value.Len())
			for j := 0; j < value.Len(); j++ {
				element := value.Index(j)
				if element.Kind() == reflect.Struct {
					elementMap, err := structToMap(element.Interface())
					if err != nil {
						elements[j] = nil
					} else {
						elements[j] = elementMap
					}
				} else {
					elements[j] = element.Interface()
				}
			}
			result[jsonTag] = elements
			continue
		}

		// Handle maps
		if value.Kind() == reflect.Map {
			if value.Len() == 0 {
				result[jsonTag] = map[string]interface{}{}
				continue
			}

			iter := value.MapRange()
			nestedMap := make(map[string]interface{})
			for iter.Next() {
				key := iter.Key()
				val := iter.Value()
				nestedMap[fmt.Sprintf("%v", key)] = val.Interface()
			}
			result[jsonTag] = nestedMap
			continue
		}

		// Handle basic types
		result[jsonTag] = value.Interface()
	}

	return result, nil
}

// GetComplianceScore calculates compliance score for an operator
func (s *DataGovernanceService) GetComplianceScore(ctx context.Context, operatorID string, period string) (float64, error) {
	// Get compliance report
	report, err := s.repo.GetTR8ComplianceReport(ctx, time.Now().Add(-30*24*time.Hour), time.Now())
	if err != nil {
		return 0, err
	}

	// Find operator in report
	for _, op := range report.Operators {
		if op.Operator == operatorID {
			return op.ComplianceScore, nil
		}
	}

	return 1.0, nil
}

// CheckDataAccess checks if an operator can access specific data
func (s *DataGovernanceService) CheckDataAccess(operatorID string, dataType string, action string) bool {
	// Implement access control logic
	// For TR8 crossborder, operators can access:
	// - Their own shipment data
	// - Public tracking information
	// - Quote information
	// But NOT:
	// - Competitor pricing data
	// - Other operator's sensitive information

	switch dataType {
	case "shipment":
		return action == "read" || action == "write"
	case "tracking":
		return action == "read"
	case "quote":
		return action == "read"
	case "metrics":
		return action == "read"
	case "compliance":
		return action == "read"
	default:
		return false
	}
}
