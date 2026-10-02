package nacional

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rldb-br/rldb-api-universal/internal/models/nacional"
	"github.com/rldb-br/rldb-api-universal/internal/repository/nacional"
	"github.com/rldb-br/rldb-api-universal/pkg/governance"
)

// TR8NacionalService handles TR8 nacional operations
type TR8NacionalService struct {
	repo           *nacional.TR8NacionalRepository
	governanceSvc *governance.DataGovernanceService
	cache          CacheService
	config         *nacional.TR8GovernanceConfig
}

// CacheService interface for caching
type CacheService interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
}

// NewTR8NacionalService creates a new TR8 nacional service
func NewTR8NacionalService(
	repo *nacional.TR8NacionalRepository,
	governanceSvc *governance.DataGovernanceService,
	cache CacheService,
	config *nacional.TR8GovernanceConfig,
) *TR8NacionalService {
	return &TR8NacionalService{
		repo:           repo,
		governanceSvc: governanceSvc,
		cache:          cache,
		config:         config,
	}
}

// GetCotacoes returns shipping quotes for TR8 nacional
func (s *TR8NacionalService) GetCotacoes(ctx context.Context, request nacional.SolicitacaoCotacao, operatorID string) ([]nacional.RespostaCotacao, error) {
	// Apply data governance
	sanitizedRequest, err := s.governanceSvc.SanitizeRequest(ctx, request, operatorID, "GET /v1/nacional/tr8/cotacoes")
	if err != nil {
		return nil, fmt.Errorf("governance sanitization failed: %w", err)
	}

	// Check rate limiting
	if err := s.checkRateLimit(ctx, operatorID, "cotacoes"); err != nil {
		return nil, err
	}

	// Generate cache key
	cacheKey := s.generateCotacaoCacheKey(sanitizedRequest, operatorID)

	// Try to get from cache
	var cachedCotacoes []nacional.RespostaCotacao
	if err := s.cache.Get(cacheKey, &cachedCotacoes); err == nil {
		return cachedCotacoes, nil
	}

	// Calculate quotes from operators
	cotacoes, err := s.calculateCotacoes(ctx, sanitizedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cotacoes: %w", err)
	}

	// Cache for 5 minutes
	if err := s.cache.Set(cacheKey, cotacoes, 5*time.Minute); err != nil {
		// Log cache error but continue
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "GET /v1/nacional/tr8/cotacoes", "success", sanitizedRequest)

	return cotacoes, nil
}

// CreateRemessa creates a new TR8 nacional remessa
func (s *TR8NacionalService) CreateRemessa(ctx context.Context, request nacional.SolicitacaoRemessa, operatorID string) (*nacional.RespostaRemessa, error) {
	// Apply data governance
	sanitizedRequest, err := s.governanceSvc.SanitizeRequest(ctx, request, operatorID, "POST /v1/nacional/tr8/remessas")
	if err != nil {
		return nil, fmt.Errorf("governance sanitization failed: %w", err)
	}

	// Check idempotency
	if request.ChaveIdempotencia != "" {
		existing, err := s.repo.GetRemessaByChaveIdempotencia(ctx, request.ChaveIdempotencia)
		if err == nil && existing != nil {
			return &nacional.RespostaRemessa{
				CodigoRastreamento: existing.CodigoRastreamento,
				URLEtiqueta:        existing.URLEtiqueta,
				PrevisaoEntrega:    existing.PrevisaoEntrega,
				RotaHubs:           existing.RotaHubs,
				IDRemessa:          existing.IDRemessa,
				Operador:           existing.Operador,
				Servico:            existing.Servico,
				ValorFreteBRL:       existing.ValorFreteBRL,
				CreatedAt:          existing.CreatedAt,
			}, nil
		}
	}

	// Check rate limiting
	if err := s.checkRateLimit(ctx, operatorID, "remessas"); err != nil {
		return nil, err
	}

	// Validate request
	if err := s.validateRemessaRequest(sanitizedRequest); err != nil {
		return nil, err
	}

	// Generate tracking code
	codigoRastreamento := s.generateCodigoRastreamento()

	// Create remessa record
	remessa := &nacional.RemessaNacional{
		ID:                uuid.New().String(),
		IDRemessa:         request.IDRemessa,
		Operador:          request.Operador,
		Servico:           request.Servico,
		TipoOperacao:      "entrega",
		OrigemCEP:         request.OrigemCEP,
		DestinoCEP:        request.DestinoCEP,
		OrigemUF:          request.OrigemUF,
		DestinoUF:         request.DestinoUF,
		OrigemMunicipio:   request.OrigemMunicipio,
		DestinoMunicipio:  request.DestinoMunicipio,
		PesoKg:            request.PesoKg,
		DimensoesCm:       request.DimensoesCm,
		ValorDeclaradoBRL: request.ValorDeclaradoBRL,
		ValorFreteBRL:     request.ValorFreteBRL,
		ModalidadeFrete:   request.ModalidadeFrete,
		TipoProduto:       request.TipoProduto,
		NFCe:              request.NFCe,
		CTe:               request.CTe,
		ReversaLogistica:  request.ReversaLogistica,
		Status:            "pendente",
		CodigoRastreamento: codigoRastreamento,
		PrevisaoEntrega:   time.Now().AddDate(0, 0, 7), // Default 7 days
		Metadados: map[string]interface{}{
			"operador_id": operatorID,
			"request_id":   ctx.Value("request_id"),
		},
	}

	// Save remessa
	if err := s.repo.CreateRemessa(ctx, remessa); err != nil {
		return nil, fmt.Errorf("failed to create remessa: %w", err)
	}

	// Generate label
	urlEtiqueta := s.generateEtiquetaURL(remessa.ID)
	remessa.URLEtiqueta = urlEtiqueta

	// Update with label URL
	if err := s.repo.UpdateRemessa(ctx, remessa); err != nil {
		// Log error but continue
	}

	// Create tracking events
	initialEvent := &nacional.EventoRastreamento{
		ID:              uuid.New().String(),
		IDRemessa:       remessa.ID,
		CodigoRastreamento: codigoRastreamento,
		Timestamp:       time.Now().UTC(),
		Local:           request.OrigemCEP,
		UF:              request.OrigemUF,
		CEP:             request.OrigemCEP,
		Status:          "postado",
		Operador:        request.Operador,
		Descricao:       "Remessa criada",
	}

	if err := s.repo.CreateEventoRastreamento(ctx, initialEvent); err != nil {
		// Log error but continue
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "POST /v1/nacional/tr8/remessas", "success", sanitizedRequest)

	// Return response
	response := &nacional.RespostaRemessa{
		CodigoRastreamento: codigoRastreamento,
		URLEtiqueta:        urlEtiqueta,
		PrevisaoEntrega:    remessa.PrevisaoEntrega,
		RotaHubs:           []string{"Hub Origem", "Hub Destino"},
		IDRemessa:          remessa.IDRemessa,
		Operador:           remessa.Operador,
		Servico:            remessa.Servico,
		ValorFreteBRL:       remessa.ValorFreteBRL,
		CreatedAt:          remessa.CreatedAt,
	}

	return response, nil
}

// GetRastreamento returns tracking information for a TR8 nacional remessa
func (s *TR8NacionalService) GetRastreamento(ctx context.Context, codigoRastreamento string, operatorID string) (*nacional.RespostaRastreamento, error) {
	// Apply data governance
	sanitizedTrackingCode := s.governanceSvc.SanitizeField(ctx, codigoRastreamento, operatorID, "codigo_rastreamento")

	// Get remessa
	remessa, err := s.repo.GetRemessaByCodigoRastreamento(ctx, sanitizedTrackingCode)
	if err != nil {
		return nil, fmt.Errorf("remessa not found: %w", err)
	}

	// Check if operator has access to this remessa
	if remessa.Operador != operatorID {
		// Check if operator is the origin or destination
		if remessa.OrigemUF != operatorID && remessa.DestinoUF != operatorID {
			return nil, errors.New("unauthorized access to remessa")
		}
	}

	// Get tracking events
	events, err := s.repo.GetEventosRastreamento(ctx, remessa.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking events: %w", err)
	}

	// Convert to response format
	eventResponses := make([]nacional.EventoRastreamentoSimples, len(events))
	for i, event := range events {
		eventResponses[i] = nacional.EventoRastreamentoSimples{
			Timestamp:   event.Timestamp,
			Local:      event.Local,
			UF:         event.UF,
			Status:     event.Status,
			Operador:   event.Operador,
			Descricao:  event.Descricao,
		}
	}

	// Get fiscal documents
	documents, err := s.repo.GetDocumentosFiscais(ctx, remessa.ID)
	if err != nil {
		// Log error but continue
	}

	docInfos := make([]nacional.DocumentoInfo, len(documents))
	for i, doc := range documents {
		docInfos[i] = nacional.DocumentoInfo{
			Tipo:       doc.Tipo,
			Numero:     doc.Numero,
			Status:     doc.Status,
			Verificado: doc.Verificado,
			ArquivoURL: doc.ArquivoURL,
		}
	}

	// Get status history
	historicoStatus := []nacional.StatusHistorico{
		{Status: "pendente", Timestamp: remessa.CreatedAt, Responsavel: remessa.Operador, Motivo: "Remessa criada"},
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "GET /v1/nacional/tr8/rastreamento/{codigo_rastreamento}", "success", map[string]interface{}{
		"codigo_rastreamento": codigoRastreamento,
	})

	return &nacional.RespostaRastreamento{
		CodigoRastreamento: remessa.CodigoRastreamento,
		Status:             remessa.Status,
		Eventos:            eventResponses,
		PrevisaoEntrega:    remessa.PrevisaoEntrega,
		DataEntrega:       remessa.DataEntrega,
		Operador:           remessa.Operador,
		Servico:            remessa.Servico,
		OrigemCEP:          remessa.OrigemCEP,
		DestinoCEP:         remessa.DestinoCEP,
		PesoKg:            remessa.PesoKg,
		ValorDeclaradoBRL:  remessa.ValorDeclaradoBRL,
		Documentos:         docInfos,
		HistoricoStatus:    historicoStatus,
	}, nil
}

// UpdateRastreamento updates tracking information for a TR8 nacional remessa
func (s *TR8NacionalService) UpdateRastreamento(ctx context.Context, codigoRastreamento string, status string, local string, uf string, descricao string, operatorID string) error {
	// Get remessa
	remessa, err := s.repo.GetRemessaByCodigoRastreamento(ctx, codigoRastreamento)
	if err != nil {
		return fmt.Errorf("remessa not found: %w", err)
	}

	// Verify operator has write access
	if remessa.Operador != operatorID {
		return errors.New("unauthorized to update rastreamento")
	}

	// Create tracking event
	event := &nacional.EventoRastreamento{
		ID:              uuid.New().String(),
		IDRemessa:       remessa.ID,
		CodigoRastreamento: codigoRastreamento,
		Timestamp:       time.Now().UTC(),
		Local:           local,
		UF:              uf,
		Status:          status,
		Operador:        operatorID,
		Descricao:       descricao,
	}

	// Update remessa status if needed
	if status == "entregue" {
		now := time.Now().UTC()
		remessa.Status = "entregue"
		remessa.DataEntrega = &now
		if err := s.repo.UpdateRemessa(ctx, remessa); err != nil {
			return fmt.Errorf("failed to update remessa status: %w", err)
		}
	} else if status == "coletado" {
		now := time.Now().UTC()
		remessa.Status = "coletado"
		remessa.DataColetado = &now
		if err := s.repo.UpdateRemessa(ctx, remessa); err != nil {
			return fmt.Errorf("failed to update remessa status: %w", err)
		}
	} else if status == "postado" {
		now := time.Now().UTC()
		remessa.Status = "postado"
		remessa.DataPostagem = &now
		if err := s.repo.UpdateRemessa(ctx, remessa); err != nil {
			return fmt.Errorf("failed to update remessa status: %w", err)
		}
	}

	// Save tracking event
	if err := s.repo.CreateEventoRastreamento(ctx, event); err != nil {
		return fmt.Errorf("failed to create tracking event: %w", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "POST /v1/nacional/tr8/rastreamento/{codigo_rastreamento}", "success", map[string]interface{}{
		"codigo_rastreamento": codigoRastreamento,
		"status":             status,
		"local":             local,
	})

	return nil
}

// GetMetricas returns metrics for TR8 nacional
func (s *TR8NacionalService) GetMetricas(ctx context.Context, periodo string, operatorID string) (*nacional.MetricasNacional, error) {
	// Apply data governance
	sanitizedPeriod := s.governanceSvc.SanitizeField(ctx, periodo, operatorID, "periodo")

	// Parse period
	startTime, endTime, err := s.parsePeriod(sanitizedPeriod)
	if err != nil {
		return nil, err
	}

	// Get metrics from repository
	metrics, err := s.repo.GetMetricasNacional(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "GET /v1/observatorio/nacional/tr8/metricas", "success", map[string]interface{}{
		"periodo": periodo,
	})

	return metrics, nil
}

// CreateDocumentoFiscal creates a fiscal document for TR8 nacional remessa
func (s *TR8NacionalService) CreateDocumentoFiscal(ctx context.Context, idRemessa string, tipo string, numero string, emitente string, operatorID string) (*nacional.DocumentoFiscal, error) {
	// Get remessa
	remessa, err := s.repo.GetRemessaByID(ctx, idRemessa)
	if err != nil {
		return nil, fmt.Errorf("remessa not found: %w", err)
	}

	// Verify operator has access
	if remessa.Operador != operatorID && remessa.OrigemUF != operatorID {
		return nil, errors.New("unauthorized to create fiscal document")
	}

	// Create document
	doc := &nacional.DocumentoFiscal{
		ID:          uuid.New().String(),
		IDRemessa:   idRemessa,
		Tipo:        tipo,
		Numero:      numero,
		DataEmissao: time.Now().UTC(),
		ValorBRL:    remessa.ValorDeclaradoBRL,
		Emitente:    emitente,
		Destinatario: remessa.Metadados["destinatario"].(string),
		Status:      "pendente",
		Verificado:  false,
	}

	if err := s.repo.CreateDocumentoFiscal(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create fiscal document: %w", err)
	}

	// Update remessa metadata
	if remessa.Metadados == nil {
		remessa.Metadados = make(map[string]interface{})
	}
	remessa.Metadados["documentos_fiscais"] = true

	if err := s.repo.UpdateRemessa(ctx, remessa); err != nil {
		// Log error but continue
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "POST /v1/nacional/tr8/documentos/fiscais", "success", map[string]interface{}{
		"id_remessa": idRemessa,
		"tipo":       tipo,
	})

	return doc, nil
}

// ValidateDocumentoFiscal validates a fiscal document
func (s *TR8NacionalService) ValidateDocumentoFiscal(ctx context.Context, documentID string, verificado bool, operatorID string) error {
	// Get document
	doc, err := s.repo.GetDocumentoFiscalByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Get remessa
	remessa, err := s.repo.GetRemessaByID(ctx, doc.IDRemessa)
	if err != nil {
		return fmt.Errorf("remessa not found: %w", err)
	}

	// Verify operator has access
	if remessa.Operador != operatorID {
		return errors.New("unauthorized to validate fiscal document")
	}

	// Update document
	doc.Verificado = verificado
	if verificado {
		doc.Status = "verificado"
	} else {
		doc.Status = "rejeitado"
	}

	if err := s.repo.UpdateDocumentoFiscal(ctx, doc); err != nil {
		return fmt.Errorf("failed to update fiscal document: %w", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "PUT /v1/nacional/tr8/documentos/fiscais/{id}/validar", "success", map[string]interface{}{
		"document_id": documentID,
		"verificado":  verificado,
	})

	return nil
}

// GetRelatorioConformidade generates a compliance report
func (s *TR8NacionalService) GetRelatorioConformidade(ctx context.Context, periodo string, operatorID string) (*nacional.RelatorioConformidade, error) {
	// Apply data governance
	sanitizedPeriod := s.governanceSvc.SanitizeField(ctx, periodo, operatorID, "periodo")

	// Parse period
	startTime, endTime, err := s.parsePeriod(sanitizedPeriod)
	if err != nil {
		return nil, err
	}

	// Get compliance data
	report, err := s.repo.GetRelatorioConformidade(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance report: %w", err)
	}

	// Log audit
	s.governanceSvc.LogAudit(ctx, operatorID, "GET /v1/governanca/nacional/tr8/conformidade", "success", map[string]interface{}{
		"periodo": periodo,
	})

	return report, nil
}

// Helper methods

func (s *TR8NacionalService) calculateCotacoes(ctx context.Context, request nacional.SolicitacaoCotacao) ([]nacional.RespostaCotacao, error) {
	// This would integrate with actual operator APIs
	// For now, return mock data for major Brazilian operators
	cotacoes := []nacional.RespostaCotacao{
		{
			Operador:          "correios",
			Servico:           "sedex",
			ValorFreteBRL:     48.90,
			ValorMinimoBRL:    45.00,
			PrazoEntregaDias:  5,
			PrazoMinimoDias:   3,
			PrazoMaximoDias:   7,
			ModalidadeFrete:   "cif",
			TipoServico:       "expresso",
			IncluiSeguro:      true,
			IncluiReversa:     false,
			IncluiColeta:      false,
			IncluiEntrega:     true,
			CO2Kg:            1.2,
			Notas:            "Entrega em até 5 dias úteis",
			Disponivel:       true,
			IDCotacao:         uuid.New().String(),
		},
		{
			Operador:          "mercado_livre_logistics",
			Servico:           "fulfillment_express",
			ValorFreteBRL:     52.30,
			ValorMinimoBRL:    50.00,
			PrazoEntregaDias:  3,
			PrazoMinimoDias:   2,
			PrazoMaximoDias:   5,
			ModalidadeFrete:   "por_conta",
			TipoServico:       "expresso",
			IncluiSeguro:      true,
			IncluiReversa:     true,
			IncluiColeta:      true,
			IncluiEntrega:     true,
			CO2Kg:            0.9,
			Notas:            "Inclui coleta e reversa",
			Disponivel:       true,
			IDCotacao:         uuid.New().String(),
		},
		{
			Operador:          "transportadora_abc",
			Servico:           "rodonaves",
			ValorFreteBRL:     42.50,
			ValorMinimoBRL:    40.00,
			PrazoEntregaDias:  7,
			PrazoMinimoDias:   5,
			PrazoMaximoDias:   10,
			ModalidadeFrete:   "fob",
			TipoServico:       "normal",
			IncluiSeguro:      false,
			IncluiReversa:     false,
			IncluiColeta:      false,
			IncluiEntrega:     false,
			CO2Kg:            1.5,
			Notas:            "Para volumes acima de 100kg",
			Disponivel:       true,
			IDCotacao:         uuid.New().String(),
		},
	}

	return cotacoes, nil
}

func (s *TR8NacionalService) validateRemessaRequest(request nacional.SolicitacaoRemessa) error {
	if request.Operador == "" {
		return errors.New("operador é obrigatório")
	}
	if request.Servico == "" {
		return errors.New("serviço é obrigatório")
	}
	if request.IDRemessa == "" {
		return errors.New("id_remessa é obrigatório")
	}
	if request.OrigemCEP == "" {
		return errors.New("origem_cep é obrigatório")
	}
	if request.DestinoCEP == "" {
		return errors.New("destino_cep é obrigatório")
	}
	if request.PesoKg <= 0 {
		return errors.New("peso_kg deve ser maior que 0")
	}
	if request.ValorDeclaradoBRL < 0 {
		return errors.New("valor_declarado_brl não pode ser negativo")
	}
	if request.ValorFreteBRL < 0 {
		return errors.New("valor_frete_brl não pode ser negativo")
	}
	if request.ModalidadeFrete == "" {
		return errors.New("modalidade_frete é obrigatório")
	}
	return nil
}

func (s *TR8NacionalService) generateCodigoRastreamento() string {
	// Generate a TR8 nacional tracking code
	rand.Seed(time.Now().UnixNano())
	prefix := "BR"
	numbers := ""
	for i := 0; i < 12; i++ {
		numbers += fmt.Sprintf("%d", rand.Intn(10))
	}
	return prefix + numbers
}

func (s *TR8NacionalService) generateEtiquetaURL(remessaID string) string {
	// Generate a secure label URL
	hash := sha256.Sum256([]byte(remessaID + time.Now().Format(time.RFC3339)))
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("https://rldb.gov.br/etiquetas/%s.pdf", hashStr[:16])
}

func (s *TR8NacionalService) generateCotacaoCacheKey(request nacional.SolicitacaoCotacao, operatorID string) string {
	data := fmt.Sprintf("%s:%s:%.2f:%s:%s",
		request.OrigemCEP,
		request.DestinoCEP,
		request.PesoKg,
		request.ModalidadeFrete,
		operatorID,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *TR8NacionalService) checkRateLimit(ctx context.Context, operatorID string, endpoint string) error {
	// Implement rate limiting based on configuration
	// For now, just return nil
	return nil
}

func (s *TR8NacionalService) parsePeriod(period string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	switch period {
	case "ultimas_24_horas":
		return now.Add(-24 * time.Hour), now, nil
	case "ultimos_7_dias":
		return now.Add(-7 * 24 * time.Hour), now, nil
	case "ultimos_30_dias":
		return now.Add(-30 * 24 * time.Hour), now, nil
	case "hoje":
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return today, today.Add(24 * time.Hour), nil
	default:
		return time.Time{}, time.Time{}, errors.New("período inválido")
	}
}
