package models

import (
	"time"
)

// TR8 Crossborder Models for RLDB API UNIVERSAL

// CrossborderShipment represents a TR8 international shipment
type CrossborderShipment struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	ShipmentID        string    `json:"shipment_id" gorm:"uniqueIndex;not null"`
	Operator          string    `json:"operator" gorm:"index;not null"`
	Service           string    `json:"service" gorm:"not null"`
	OriginCountry     string    `json:"origin_country" gorm:"index;not null"`
	DestinationCountry string   `json:"destination_country" gorm:"index;not null"`
	OriginZip         string    `json:"origin_zip" gorm:"not null"`
	DestinationZip    string    `json:"destination_zip" gorm:"not null"`
	WeightKg          float64   `json:"weight_kg" gorm:"not null"`
	DimensionsCm      Dimensions `json:"dimensions_cm" gorm:"type:jsonb"`
	DeclaredValueBRL  float64   `json:"declared_value_brl" gorm:"not null"`
	DeclaredValueUSD  float64   `json:"declared_value_usd" gorm:"not null"`
	Currency          string    `json:"currency" gorm:"default:'BRL'"`
	Incoterm          string    `json:"incoterm" gorm:"not null"`
	HSCode            string    `json:"hs_code" gorm:"index"`
	ProductCategory   string    `json:"product_category" gorm:"index"`
	ReverseLogistics  bool      `json:"reverse_logistics_enabled" gorm:"default:false"`
	Status            string    `json:"status" gorm:"index;default:'pending'"`
	TrackingCode      string    `json:"tracking_code" gorm:"uniqueIndex"`
	LabelURL          string    `json:"label_url"`
	EstimatedDelivery  time.Time `json:"estimated_delivery" gorm:"type:timestamptz"`
	ActualDelivery    *time.Time `json:"actual_delivery" gorm:"type:timestamptz"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Metadata          map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// CrossborderQuote represents a TR8 international shipping quote
type CrossborderQuote struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	RequestID         string    `json:"request_id" gorm:"uniqueIndex;not null"`
	Operator          string    `json:"operator" gorm:"index;not null"`
	Service           string    `json:"service" gorm:"not null"`
	OriginCountry     string    `json:"origin_country" gorm:"not null"`
	DestinationCountry string   `json:"destination_country" gorm:"not null"`
	OriginZip         string    `json:"origin_zip" gorm:"not null"`
	DestinationZip    string    `json:"destination_zip" gorm:"not null"`
	WeightKg          float64   `json:"weight_kg" gorm:"not null"`
	DimensionsCm      Dimensions `json:"dimensions_cm" gorm:"type:jsonb"`
	DeclaredValueBRL  float64   `json:"declared_value_brl" gorm:"not null"`
	DeclaredValueUSD  float64   `json:"declared_value_usd" gorm:"not null"`
	PriceBRL          float64   `json:"price_brl" gorm:"not null"`
	PriceUSD          float64   `json:"price_usd" gorm:"not null"`
	DeliveryDays      int       `json:"delivery_days" gorm:"not null"`
	CO2Kg            float64   `json:"co2_kg" gorm:"not null"`
	Incoterm          string    `json:"incoterm" gorm:"not null"`
	HSCode            string    `json:"hs_code"`
	DutiesAndTaxesBRL float64   `json:"duties_and_taxes_brl" gorm:"not null"`
	DutiesAndTaxesUSD float64   `json:"duties_and_taxes_usd" gorm:"not null"`
	ValidUntil        time.Time `json:"valid_until" gorm:"type:timestamptz"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// CrossborderTrackingEvent represents a tracking event for TR8 shipments
type CrossborderTrackingEvent struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	ShipmentID    string    `json:"shipment_id" gorm:"index;not null"`
	TrackingCode  string    `json:"tracking_code" gorm:"index;not null"`
	Timestamp     time.Time `json:"timestamp" gorm:"type:timestamptz;index"`
	Location      string    `json:"location" gorm:"not null"`
	Country       string    `json:"country" gorm:"not null"`
	Status        string    `json:"status" gorm:"index;not null"`
	Operator      string    `json:"operator" gorm:"not null"`
	Description   string    `json:"description"`
	CustomsStatus string    `json:"customs_status"`
	Documents     []string  `json:"documents" gorm:"type:text[]"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// CrossborderWebhook represents webhook configuration for TR8
type CrossborderWebhook struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	ClientID  string    `json:"client_id" gorm:"index;not null"`
	URL       string    `json:"url" gorm:"not null"`
	Events    []string  `json:"events" gorm:"type:text[]"`
	Secret    string    `json:"secret" gorm:"not null"`
	Active    bool      `json:"active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CrossborderCustomsDocument represents customs documentation for TR8
type CrossborderCustomsDocument struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ShipmentID     string    `json:"shipment_id" gorm:"index;not null"`
	DocumentType   string    `json:"document_type" gorm:"index;not null"`
	DocumentNumber string    `json:"document_number" gorm:"not null"`
	IssueDate      time.Time `json:"issue_date" gorm:"type:timestamptz"`
	ExpiryDate     *time.Time `json:"expiry_date" gorm:"type:timestamptz"`
	IssuingCountry string    `json:"issuing_country" gorm:"not null"`
	Status         string    `json:"status" gorm:"default:'pending'"`
	FileURL        string    `json:"file_url"`
	Verified       bool      `json:"verified" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TR8QuoteRequest represents a quote request for TR8 crossborder
type TR8QuoteRequest struct {
	OriginCountry     string    `json:"origin_country" validate:"required"`
	DestinationCountry string    `json:"destination_country" validate:"required"`
	OriginZip         string    `json:"origin_zip" validate:"required"`
	DestinationZip    string    `json:"destination_zip" validate:"required"`
	WeightKg          float64   `json:"weight_kg" validate:"required,gt=0"`
	DimensionsCm      Dimensions `json:"dimensions_cm"`
	DeclaredValueBRL  float64   `json:"declared_value_brl" validate:"required,gte=0"`
	Incoterm          string    `json:"incoterm" validate:"required"`
	HSCode            string    `json:"hs_code"`
	ProductCategory   string    `json:"product_category"`
	ServiceLevel      string    `json:"service_level" validate:"required"`
	CustomsRequired   bool      `json:"customs_required"`
	InsuranceRequired bool      `json:"insurance_required"`
}

// TR8ShipmentRequest represents a shipment request for TR8 crossborder
type TR8ShipmentRequest struct {
	Operator          string    `json:"operator" validate:"required"`
	Service           string    `json:"service" validate:"required"`
	ShipmentID        string    `json:"shipment_id" validate:"required"`
	OriginCountry     string    `json:"origin_country" validate:"required"`
	DestinationCountry string   `json:"destination_country" validate:"required"`
	OriginZip         string    `json:"origin_zip" validate:"required"`
	DestinationZip    string    `json:"destination_zip" validate:"required"`
	WeightKg          float64   `json:"weight_kg" validate:"required,gt=0"`
	DimensionsCm      Dimensions `json:"dimensions_cm"`
	DeclaredValueBRL  float64   `json:"declared_value_brl" validate:"required,gte=0"`
	DeclaredValueUSD  float64   `json:"declared_value_usd"`
	Incoterm          string    `json:"incoterm" validate:"required"`
	HSCode            string    `json:"hs_code"`
	ProductCategory   string    `json:"product_category"`
	ReverseLogistics  bool      `json:"reverse_logistics_enabled"`
	Recipient         Recipient `json:"recipient"`
	CustomsDocuments  []string  `json:"customs_documents"`
	IdempotencyKey    string    `json:"idempotency_key"`
}

// TR8ShipmentResponse represents the response for TR8 shipment creation
type TR8ShipmentResponse struct {
	TrackingCode     string    `json:"tracking_code"`
	LabelURL         string    `json:"label_url"`
	EstimatedDelivery time.Time `json:"estimated_delivery"`
	HubRoute         []string  `json:"hub_route"`
	ShipmentID       string    `json:"shipment_id"`
	Operator         string    `json:"operator"`
	Service          string    `json:"service"`
	CustomsStatus    string    `json:"customs_status"`
	DutiesAndTaxesBRL float64   `json:"duties_and_taxes_brl"`
	DutiesAndTaxesUSD float64   `json:"duties_and_taxes_usd"`
	CreatedAt        time.Time `json:"created_at"`
}

// TR8TrackingResponse represents tracking information for TR8
type TR8TrackingResponse struct {
	TrackingCode     string                   `json:"tracking_code"`
	Status           string                   `json:"status"`
	Events           []TR8TrackingEventResponse `json:"events"`
	EstimatedDelivery time.Time                `json:"estimated_delivery"`
	ActualDelivery   *time.Time               `json:"actual_delivery"`
	Operator         string                   `json:"operator"`
	Service          string                   `json:"service"`
	OriginCountry    string                   `json:"origin_country"`
	DestinationCountry string                `json:"destination_country"`
	CustomsStatus    string                   `json:"customs_status"`
	Documents        []CustomsDocumentInfo    `json:"documents"`
}

// TR8TrackingEventResponse represents a tracking event in response
type TR8TrackingEventResponse struct {
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Country     string    `json:"country"`
	Status      string    `json:"status"`
	Operator    string    `json:"operator"`
	Description string    `json:"description"`
}

// CustomsDocumentInfo represents customs document information
type CustomsDocumentInfo struct {
	DocumentType   string    `json:"document_type"`
	DocumentNumber string    `json:"document_number"`
	Status         string    `json:"status"`
	Verified       bool      `json:"verified"`
	FileURL        string    `json:"file_url"`
}

// TR8MetricsResponse represents metrics for TR8 observatory
type TR8MetricsResponse struct {
	Period                string  `json:"period"`
	TotalShipments        int64   `json:"total_shipments"`
	TotalValueBRL         float64 `json:"total_value_brl"`
	TotalValueUSD         float64 `json:"total_value_usd"`
	AverageDeliveryDays   float64 `json:"avg_delivery_days"`
	OnTimeDeliveryRate    float64 `json:"on_time_delivery_rate"`
	AverageCostPerKgBRL   float64 `json:"avg_cost_per_kg_brl"`
	AverageCostPerKgUSD   float64 `json:"avg_cost_per_kg_usd"`
	CO2PerPackageKg      float64 `json:"co2_per_package_kg"`
	CustomsClearanceRate  float64 `json:"customs_clearance_rate"`
	ReverseLogisticsRate float64 `json:"reverse_logistics_rate"`
	TopDestinations       []CountryMetric `json:"top_destinations"`
	TopOperators          []OperatorMetric `json:"top_operators"`
}

// CountryMetric represents metrics for a specific country
type CountryMetric struct {
	Country      string  `json:"country"`
	ShipmentCount int64   `json:"shipment_count"`
	ValueBRL      float64 `json:"value_brl"`
	ValueUSD      float64 `json:"value_usd"`
	AvgDeliveryDays float64 `json:"avg_delivery_days"`
}

// OperatorMetric represents metrics for a specific operator
type OperatorMetric struct {
	Operator      string  `json:"operator"`
	ShipmentCount int64   `json:"shipment_count"`
	ValueBRL      float64 `json:"value_brl"`
	ValueUSD      float64 `json:"value_usd"`
	AvgDeliveryDays float64 `json:"avg_delivery_days"`
}

// TR8ComplianceReport represents compliance report for TR8
type TR8ComplianceReport struct {
	Period           string    `json:"period"`
	GeneratedAt      time.Time `json:"generated_at"`
	TotalRequests    int64     `json:"total_requests"`
	BlockedRequests  int64     `json:"blocked_requests"`
	SanitizedFields  int64     `json:"sanitized_fields"`
	DataAccessViolations int64  `json:"data_access_violations"`
	AuditLogsCount   int64     `json:"audit_logs_count"`
	Operators        []OperatorCompliance `json:"operators"`
}

// OperatorCompliance represents compliance data for a specific operator
type OperatorCompliance struct {
	Operator         string `json:"operator"`
	RequestsMade     int64  `json:"requests_made"`
	RequestsBlocked  int64  `json:"requests_blocked"`
	DataFieldsAccessed []string `json:"data_fields_accessed"`
	ComplianceScore  float64 `json:"compliance_score"`
}

// TR8GovernanceConfig represents governance configuration for TR8
type TR8GovernanceConfig struct {
	AllowedFields    []string `json:"allowed_fields"`
	BlockedFields    []string `json:"blocked_fields"`
	RequiredFields   []string `json:"required_fields"`
	SensitiveFields  []string `json:"sensitive_fields"`
	PseudonymizeFields []string `json:"pseudonymize_fields"`
	RateLimits       map[string]int `json:"rate_limits"`
	SLARequirements   map[string]string `json:"sla_requirements"`
}

// Dimensions represents package dimensions
type Dimensions struct {
	Length float64 `json:"l"`
	Width  float64 `json:"w"`
	Height float64 `json:"h"`
}

// Recipient represents shipment recipient with hashed data
type Recipient struct {
	NameHash      string `json:"name_hash"`
	AddressHash   string `json:"address_hash"`
	CityHash      string `json:"city_hash"`
	State         string `json:"state"`
	Country       string `json:"country"`
	PostalCode    string `json:"postal_code"`
	PhoneHash     string `json:"phone_hash"`
	EmailHash     string `json:"email_hash"`
}

// AuditLog represents audit log entry for TR8
type AuditLog struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	RequestID     string    `json:"request_id" gorm:"index"`
	OperatorID    string    `json:"operator_id" gorm:"index;not null"`
	Endpoint      string    `json:"endpoint" gorm:"index;not null"`
	HTTPMethod    string    `json:"http_method" gorm:"not null"`
	FieldsAccessed []string  `json:"fields_accessed" gorm:"type:text[]"`
	FieldsBlocked  []string  `json:"fields_blocked" gorm:"type:text[]"`
	Timestamp     time.Time `json:"timestamp" gorm:"type:timestamptz;index"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	StatusCode    int       `json:"status_code"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	PreviousHash  string    `json:"previous_hash" gorm:"not null"`
	CurrentHash   string    `json:"current_hash" gorm:"uniqueIndex;not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}
