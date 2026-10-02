package nacional

import (
	"time"
)

// TR8GovernanceConfig represents governance configuration for TR8 Nacional
type TR8GovernanceConfig struct {
	// Fields that are allowed to be accessed
	CamposPermitidos []string `json:"campos_permitidos"`

	// Fields that are blocked
	CamposBloqueados []string `json:"campos_bloqueados"`

	// Fields that are required
	CamposObrigatorios []string `json:"campos_obrigatorios"`

	// Sensitive fields that require special access
	CamposSensiveis []string `json:"campos_sensiveis"`

	// Fields that should be pseudonymized
	CamposPseudonimizar []string `json:"campos_pseudonimizar"`

	// Rate limits by endpoint
	LimitesTaxa map[string]int `json:"limites_taxa"`

	// SLA requirements by endpoint
	RequisitosSLA map[string]string `json:"requisitos_sla"`

	// Maximum weight for TR8 operations
	PesoMaximoKg float64 `json:"peso_maximo_kg"`

	// Supported Brazilian states
	UFsSuportadas []string `json:"ufs_suportadas"`

	// Supported operators
	OperadoresSuportados []string `json:"operadores_suportados"`

	// Supported service types
	TiposServicoSuportados []string `json:"tipos_servico_suportados"`

	// Supported freight modalities
	ModalidadesFreteSuportadas []string `json:"modalidades_frete_suportadas"`

	// Supported product types
	TiposProdutoSuportados []string `json:"tipos_produto_suportados"`

	// Supported fiscal document types
	TiposDocumentosFiscais []string `json:"tipos_documentos_fiscais"`

	// Cache TTL for quotes
	CacheTTLCotacoes time.Duration `json:"cache_ttl_cotacoes"`

	// Idempotency key TTL
	IdempotencyTTL time.Duration `json:"idempotency_ttl"`
}

// NewTR8GovernanceConfig creates a new governance configuration with default values
func NewTR8GovernanceConfig() *TR8GovernanceConfig {
	return &TR8GovernanceConfig{
		CamposPermitidos: []string{
			"id_remessa", "operador", "servico", "tipo_operacao",
			"origem_cep", "destino_cep", "origem_uf", "destino_uf",
			"origem_municipio", "destino_municipio",
			"peso_kg", "dimensoes_cm", "valor_declarado_brl", "valor_frete_brl",
			"modalidade_frete", "tipo_produto", "nfce", "cte",
			"reversa_logistica", "status", "codigo_rastreamento",
			"url_etiqueta", "previsao_entrega", "data_entrega",
			"data_coletado", "data_postagem", "created_at", "updated_at",
		},
		CamposBloqueados: []string{
			"custo_interno", "margem_lucro", "estrategia_comercial",
			"lista_clientes", "dados_concorrente", "informacoes_campanha",
			"preco_concorrente", "desconto_especial", "condicoes_pagamento",
			"dados_financeiros", "informacoes_privadas",
		},
		CamposObrigatorios: []string{
			"operador", "servico", "id_remessa",
			"origem_cep", "destino_cep", "peso_kg",
			"valor_declarado_brl", "modalidade_frete",
		},
		CamposSensiveis: []string{
			"valor_declarado_brl", "valor_frete_brl", "nfce", "cte",
			"tipo_produto", "reversa_logistica",
		},
		CamposPseudonimizar: []string{
			"nome", "endereco", "bairro", "cidade", "telefone",
			"email", "cpf", "cnpj", "ie",
		},
		LimitesTaxa: map[string]int{
			"cotacoes":    200,
			"remessas":    100,
			"rastreamento": 500,
			"documentos":   50,
			"webhooks":    20,
		},
		RequisitosSLA: map[string]string{
			"cotacoes":    "500ms",
			"remessas":    "2s",
			"rastreamento": "200ms",
			"documentos":   "1s",
		},
		PesoMaximoKg: 100.0,
		UFsSuportadas: []string{
			"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO",
			"MA", "MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI",
			"RJ", "RN", "RS", "RO", "RR", "SC", "SP", "SE", "TO",
		},
		OperadoresSuportados: []string{
			"correios", "mercado_livre_logistics", "shopee_logistics",
			"transportadora_abc", "transportadora_xyz", "loggi",
			"azul_cargo", "gol_log", "total_express", "jadlog",
		},
		TiposServicoSuportados: []string{
			"normal", "expresso", "economico", "agendado",
			"sedex", "pac", "fulfillment", "reversa",
		},
		ModalidadesFreteSuportadas: []string{
			"cif", "fob", "por_conta", "frete_cobrar",
		},
		TiposProdutoSuportados: []string{
			"eletroeletronicos", "vestuario", "alimentos", "livros",
			"cosmeticos", "moveis", "brinquedos", "automoveis",
			"perigosos", "fragil", "perecivel", "grande_portes",
		},
		TiposDocumentosFiscais: []string{
			"nfce", "nfe", "cte", "dacte", "mdfe",
		},
		CacheTTLCotacoes: 5 * time.Minute,
		IdempotencyTTL:   24 * time.Hour,
	}
}

// Validate validates the configuration
func (c *TR8GovernanceConfig) Validate() error {
	// Check if required operators are present
	if len(c.OperadoresSuportados) == 0 {
		return fmt.Errorf("operadores_suportados não pode ser vazio")
	}

	// Check if required UFs are present
	if len(c.UFsSuportadas) == 0 {
		return fmt.Errorf("ufs_suportadas não pode ser vazio")
	}

	// Check if required fields are present
	if len(c.CamposPermitidos) == 0 {
		return fmt.Errorf("campos_permitidos não pode ser vazio")
	}

	return nil
}

// HasOperador checks if an operator is supported
func (c *TR8GovernanceConfig) HasOperador(operador string) bool {
	for _, op := range c.OperadoresSuportados {
		if op == operador {
			return true
		}
	}
	return false
}

// HasUF checks if a UF is supported
func (c *TR8GovernanceConfig) HasUF(uf string) bool {
	for _, u := range c.UFsSuportadas {
		if u == uf {
			return true
		}
	}
	return false
}

// HasServico checks if a service type is supported
func (c *TR8GovernanceConfig) HasServico(servico string) bool {
	for _, s := range c.TiposServicoSuportados {
		if s == servico {
			return true
		}
	}
	return false
}

// HasModalidadeFrete checks if a freight modality is supported
func (c *TR8GovernanceConfig) HasModalidadeFrete(modalidade string) bool {
	for _, m := range c.ModalidadesFreteSuportadas {
		if m == modalidade {
			return true
		}
	}
	return false
}

// GetLimiteTaxa returns the rate limit for an endpoint
func (c *TR8GovernanceConfig) GetLimiteTaxa(endpoint string) int {
	if limit, ok := c.LimitesTaxa[endpoint]; ok {
		return limit
	}
	return 100 // Default limit
}

// GetSLA returns the SLA requirement for an endpoint
func (c *TR8GovernanceConfig) GetSLA(endpoint string) string {
	if sla, ok := c.RequisitosSLA[endpoint]; ok {
		return sla
	}
	return "1s" // Default SLA
}
