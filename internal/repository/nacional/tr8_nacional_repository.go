package nacional

import (
	"context"
	"fmt"
	"time"

	"github.com/rldb-br/rldb-api-universal/internal/models/nacional"
	"gorm.io/gorm"
)

// TR8NacionalRepository handles database operations for TR8 nacional
type TR8NacionalRepository struct {
	db *gorm.DB
}

// NewTR8NacionalRepository creates a new TR8 nacional repository
func NewTR8NacionalRepository(db *gorm.DB) *TR8NacionalRepository {
	return &TR8NacionalRepository{db: db}
}

// CreateRemessa creates a new TR8 nacional remessa
func (r *TR8NacionalRepository) CreateRemessa(ctx context.Context, remessa *nacional.RemessaNacional) error {
	return r.db.WithContext(ctx).Create(remessa).Error
}

// GetRemessaByID retrieves a remessa by ID
func (r *TR8NacionalRepository) GetRemessaByID(ctx context.Context, id string) (*nacional.RemessaNacional, error) {
	var remessa nacional.RemessaNacional
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&remessa).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("remessa not found")
		}
		return nil, err
	}
	return &remessa, nil
}

// GetRemessaByCodigoRastreamento retrieves a remessa by tracking code
func (r *TR8NacionalRepository) GetRemessaByCodigoRastreamento(ctx context.Context, codigoRastreamento string) (*nacional.RemessaNacional, error) {
	var remessa nacional.RemessaNacional
	if err := r.db.WithContext(ctx).Where("codigo_rastreamento = ?", codigoRastreamento).First(&remessa).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("remessa not found")
		}
		return nil, err
	}
	return &remessa, nil
}

// GetRemessaByChaveIdempotencia retrieves a remessa by idempotency key
func (r *TR8NacionalRepository) GetRemessaByChaveIdempotencia(ctx context.Context, chaveIdempotencia string) (*nacional.RemessaNacional, error) {
	var remessa nacional.RemessaNacional
	if err := r.db.WithContext(ctx).Where("metadados->>$.chave_idempotencia = ?", chaveIdempotencia).First(&remessa).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &remessa, nil
}

// UpdateRemessa updates a remessa
func (r *TR8NacionalRepository) UpdateRemessa(ctx context.Context, remessa *nacional.RemessaNacional) error {
	return r.db.WithContext(ctx).Save(remessa).Error
}

// CreateEventoRastreamento creates a tracking event
func (r *TR8NacionalRepository) CreateEventoRastreamento(ctx context.Context, evento *nacional.EventoRastreamento) error {
	return r.db.WithContext(ctx).Create(evento).Error
}

// GetEventosRastreamento retrieves all tracking events for a remessa
func (r *TR8NacionalRepository) GetEventosRastreamento(ctx context.Context, remessaID string) ([]nacional.EventoRastreamento, error) {
	var events []nacional.EventoRastreamento
	if err := r.db.WithContext(ctx).Where("id_remessa = ?", remessaID).Order("timestamp asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// CreateDocumentoFiscal creates a fiscal document
func (r *TR8NacionalRepository) CreateDocumentoFiscal(ctx context.Context, doc *nacional.DocumentoFiscal) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// GetDocumentosFiscais retrieves all fiscal documents for a remessa
func (r *TR8NacionalRepository) GetDocumentosFiscais(ctx context.Context, remessaID string) ([]nacional.DocumentoFiscal, error) {
	var docs []nacional.DocumentoFiscal
	if err := r.db.WithContext(ctx).Where("id_remessa = ?", remessaID).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// GetDocumentoFiscalByID retrieves a fiscal document by ID
func (r *TR8NacionalRepository) GetDocumentoFiscalByID(ctx context.Context, id string) (*nacional.DocumentoFiscal, error) {
	var doc nacional.DocumentoFiscal
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document not found")
		}
		return nil, err
	}
	return &doc, nil
}

// UpdateDocumentoFiscal updates a fiscal document
func (r *TR8NacionalRepository) UpdateDocumentoFiscal(ctx context.Context, doc *nacional.DocumentoFiscal) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

// GetMetricasNacional retrieves TR8 nacional metrics for a period
func (r *TR8NacionalRepository) GetMetricasNacional(ctx context.Context, startTime, endTime time.Time) (*nacional.MetricasNacional, error) {
	// Count remessas
	var totalRemessas int64
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&totalRemessas).Error; err != nil {
		return nil, err
	}

	// Sum declared values
	var totalValorBRL float64
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("COALESCE(SUM(valor_declarado_brl), 0)").
		Scan(&totalValorBRL).Error; err != nil {
		return nil, err
	}

	// Average delivery days
	var avgDeliveryDays float64
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ? AND data_entrega IS NOT NULL", startTime, endTime).
		Select("COALESCE(AVG(EXTRACT(DAY FROM (data_entrega - created_at))), 0)").
		Scan(&avgDeliveryDays).Error; err != nil {
		return nil, err
	}

	// On-time delivery rate
	var onTimeCount, totalDelivered int64
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ? AND data_entrega IS NOT NULL", startTime, endTime).
		Count(&totalDelivered).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ? AND data_entrega IS NOT NULL AND data_entrega <= previsao_entrega", startTime, endTime).
		Count(&onTimeCount).Error; err != nil {
		return nil, err
	}

	onTimeRate := float64(0)
	if totalDelivered > 0 {
		onTimeRate = float64(onTimeCount) / float64(totalDelivered)
	}

	// Average cost per kg
	var avgCostPerKg float64
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("COALESCE(AVG(valor_declarado_brl / NULLIF(peso_kg, 0)), 0)").
		Scan(&avgCostPerKg).Error; err != nil {
		return nil, err
	}

	// Top UFs
	var topUFs []nacional.UFMetric
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("destino_uf as uf, COUNT(*) as remessas_count, SUM(valor_declarado_brl) as valor_brl, AVG(EXTRACT(DAY FROM (data_entrega - created_at))) as media_dias").
		Group("destino_uf").
		Order("remessas_count DESC").
		Limit(10).
		Scan(&topUFs).Error; err != nil {
		return nil, err
	}

	// Top operators
	var topOperadores []nacional.OperadorMetric
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("operador as operador, COUNT(*) as remessas_count, SUM(valor_declarado_brl) as valor_brl, AVG(EXTRACT(DAY FROM (data_entrega - created_at))) as media_dias").
		Group("operador").
		Order("remessas_count DESC").
		Limit(10).
		Scan(&topOperadores).Error; err != nil {
		return nil, err
	}

	// Top service types
	var topServicos []nacional.ServicoMetric
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("servico as servico, COUNT(*) as remessas_count, SUM(valor_declarado_brl) as valor_brl, AVG(EXTRACT(DAY FROM (data_entrega - created_at))) as media_dias").
		Group("servico").
		Order("remessas_count DESC").
		Limit(10).
		Scan(&topServicos).Error; err != nil {
		return nil, err
	}

	// Calculate CO2 per package
	var totalCO2, totalPackages float64
	if err := r.db.WithContext(ctx).
		Model(&nacional.CotacaoNacional{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("COALESCE(SUM(co2_kg), 0)").
		Scan(&totalCO2).Error; err != nil {
		return nil, err
	}
	totalPackages = float64(totalRemessas)
	co2PerPackage := float64(0)
	if totalPackages > 0 {
		co2PerPackage = totalCO2 / totalPackages
	}

	// Reverse logistics rate
	var reverseCount int64
	if err := r.db.WithContext(ctx).
		Model(&nacional.RemessaNacional{}).
		Where("created_at BETWEEN ? AND ? AND reversa_logistica = true", startTime, endTime).
		Count(&reverseCount).Error; err != nil {
		return nil, err
	}

	reverseLogisticsRate := float64(0)
	if totalRemessas > 0 {
		reverseLogisticsRate = float64(reverseCount) / float64(totalRemessas)
	}

	return &nacional.MetricasNacional{
		Periodo:               fmt.Sprintf("%s/%s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		TotalRemessas:         totalRemessas,
		TotalValorBRL:         totalValorBRL,
		MediaDiasEntrega:      avgDeliveryDays,
		TaxaEntregaNoPrazo:    onTimeRate,
		MediaCustoPorKgBRL:    avgCostPerKg,
		CO2PorPacoteKg:       co2PerPackage,
		TaxaReversaLogistica:  reverseLogisticsRate,
		TopDestinos:           topUFs,
		TopOperadores:         topOperadores,
		TopTiposServico:       topServicos,
	}, nil
}

// GetRelatorioConformidade generates a compliance report
func (r *TR8NacionalRepository) GetRelatorioConformidade(ctx context.Context, startTime, endTime time.Time) (*nacional.RelatorioConformidade, error) {
	// Count total requests (from audit logs)
	var totalRequests int64
	if err := r.db.WithContext(ctx).
		Model(&nacional.AuditLogNacional{}).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Count(&totalRequests).Error; err != nil {
		return nil, err
	}

	// Count blocked requests
	var blockedRequests int64
	if err := r.db.WithContext(ctx).
		Model(&nacional.AuditLogNacional{}).
		Where("timestamp BETWEEN ? AND ? AND status_code >= 400", startTime, endTime).
		Count(&blockedRequests).Error; err != nil {
		return nil, err
	}

	// Get operators
	var operators []nacional.OperadorConformidade
	if err := r.db.WithContext(ctx).
		Model(&nacional.AuditLogNacional{}).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Select("operador_id as operador, COUNT(*) as solicitacoes, SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as bloqueadas").
		Group("operador_id").
		Scan(&operators).Error; err != nil {
		return nil, err
	}

	// Calculate compliance scores
	for i := range operators {
		total := operators[i].Solicitacoes
		blocked := operators[i].Bloqueadas
		if total > 0 {
			operators[i].ScoreConformidade = 1.0 - (float64(blocked) / float64(total))
		} else {
			operators[i].ScoreConformidade = 1.0
		}
	}

	return &nacional.RelatorioConformidade{
		Periodo:           fmt.Sprintf("%s/%s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		DataGeracao:       time.Now().UTC(),
		TotalSolicitacoes: totalRequests,
		SolicitacoesBloqueadas: blockedRequests,
		Operadores:        operators,
	}, nil
}

// CreateCotacao creates a cotacao record
func (r *TR8NacionalRepository) CreateCotacao(ctx context.Context, cotacao *nacional.CotacaoNacional) error {
	return r.db.WithContext(ctx).Create(cotacao).Error
}

// CreateWebhook creates a webhook configuration
func (r *TR8NacionalRepository) CreateWebhook(ctx context.Context, webhook *nacional.WebhookNacional) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

// GetWebhooks retrieves webhooks for a client
func (r *TR8NacionalRepository) GetWebhooks(ctx context.Context, clientID string) ([]nacional.WebhookNacional, error) {
	var webhooks []nacional.WebhookNacional
	if err := r.db.WithContext(ctx).Where("client_id = ? AND ativo = true", clientID).Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}

// GetAuditLogs retrieves audit logs for a period
func (r *TR8NacionalRepository) GetAuditLogs(ctx context.Context, startTime, endTime time.Time) ([]nacional.AuditLogNacional, error) {
	var logs []nacional.AuditLogNacional
	if err := r.db.WithContext(ctx).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Order("timestamp desc").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// CreateAuditLog creates an audit log entry
func (r *TR8NacionalRepository) CreateAuditLog(ctx context.Context, log *nacional.AuditLogNacional) error {
	return r.db.WithContext(ctx).Create(log).Error
}
