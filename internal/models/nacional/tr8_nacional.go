package nacional

import (
	"time"
)

// TR8Nacional - Modelo consolidado para solução nacional genérica
// Adaptável para qualquer cenário de transporte e logística

// RemessaNacional representa uma remessa TR8 nacional
type RemessaNacional struct {
	ID                     string            `json:"id" gorm:"primaryKey"`
	IDRemessa              string            `json:"id_remessa" gorm:"uniqueIndex;not null"`
	Operador               string            `json:"operador" gorm:"index;not null"`
	Servico                string            `json:"servico" gorm:"not null"`
	TipoOperacao           string            `json:"tipo_operacao" gorm:"index;not null"` // "coleta", "entrega", "transbordo", "reversa"
	OrigemCEP              string            `json:"origem_cep" gorm:"index;not null"`
	DestinoCEP             string            `json:"destino_cep" gorm:"index;not null"`
	OrigemUF               string            `json:"origem_uf" gorm:"index"`
	DestinoUF              string            `json:"destino_uf" gorm:"index"`
	OrigemMunicipio        string            `json:"origem_municipio"`
	DestinoMunicipio       string            `json:"destino_municipio"`
	PesoKg                 float64           `json:"peso_kg" gorm:"not null"`
	DimensoesCm            Dimensoes         `json:"dimensoes_cm" gorm:"type:jsonb"`
	ValorDeclaradoBRL      float64           `json:"valor_declarado_brl" gorm:"not null"`
	ValorFreteBRL          float64           `json:"valor_frete_brl" gorm:"not null"`
	ModalidadeFrete        string            `json:"modalidade_frete" gorm:"not null"` // "cif", "fob", "por_conta"
	TipoProduto            string            `json:"tipo_produto" gorm:"index"`
	NFCe                  string            `json:"nfce" gorm:"index"`
	CTe                   string            `json:"cte" gorm:"index"`
	ReversaLogistica      bool              `json:"reversa_logistica" gorm:"default:false"`
	Status                string            `json:"status" gorm:"index;default:'pendente'"`
	CodigoRastreamento     string            `json:"codigo_rastreamento" gorm:"uniqueIndex"`
	URLEtiqueta           string            `json:"url_etiqueta"`
	PrevisaoEntrega       time.Time         `json:"previsao_entrega" gorm:"type:timestamptz"`
	DataEntrega           *time.Time        `json:"data_entrega" gorm:"type:timestamptz"`
	DataColetado          *time.Time        `json:"data_coletado" gorm:"type:timestamptz"`
	DataPostagem          *time.Time        `json:"data_postagem" gorm:"type:timestamptz"`
	CreatedAt             time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	Metadados             map[string]interface{} `json:"metadados" gorm:"type:jsonb"`
}

// CotacaoNacional representa uma cotação de frete nacional
type CotacaoNacional struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	IDSolicitacao     string    `json:"id_solicitacao" gorm:"uniqueIndex;not null"`
	Operador          string    `json:"operador" gorm:"index;not null"`
	Servico           string    `json:"servico" gorm:"not null"`
	OrigemCEP         string    `json:"origem_cep" gorm:"not null"`
	DestinoCEP        string    `json:"destino_cep" gorm:"not null"`
	OrigemUF          string    `json:"origem_uf"`
	DestinoUF         string    `json:"destino_uf"`
	PesoKg            float64   `json:"peso_kg" gorm:"not null"`
	DimensoesCm       Dimensoes `json:"dimensoes_cm" gorm:"type:jsonb"`
	ValorDeclaradoBRL float64   `json:"valor_declarado_brl" gorm:"not null"`
	ValorFreteBRL     float64   `json:"valor_frete_brl" gorm:"not null"`
	ValorMinimoBRL    float64   `json:"valor_minimo_brl" gorm:"not null"`
	PrazoEntregaDias  int       `json:"prazo_entrega_dias" gorm:"not null"`
	PrazoMinimoDias   int       `json:"prazo_minimo_dias" gorm:"not null"`
	PrazoMaximoDias   int       `json:"prazo_maximo_dias" gorm:"not null"`
	ModalidadeFrete   string    `json:"modalidade_frete" gorm:"not null"`
	TipoServico       string    `json:"tipo_servico" gorm:"not null"` // "normal", "expresso", "economico", "agendado"
	IncluiSeguro      bool      `json:"inclui_seguro" gorm:"default:false"`
	IncluiReversa     bool      `json:"inclui_reversa" gorm:"default:false"`
	CO2Kg            float64   `json:"co2_kg" gorm:"not null"`
	Notas            string    `json:"notas"`
	Validade         time.Time `json:"validade" gorm:"type:timestamptz"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// EventoRastreamento representa um evento de rastreamento
type EventoRastreamento struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	IDRemessa      string    `json:"id_remessa" gorm:"index;not null"`
	CodigoRastreamento string `json:"codigo_rastreamento" gorm:"index;not null"`
	Timestamp      time.Time `json:"timestamp" gorm:"type:timestamptz;index"`
	Local          string    `json:"local" gorm:"not null"`
	UF             string    `json:"uf" gorm:"not null"`
	CEP            string    `json:"cep"`
	Status         string    `json:"status" gorm:"index;not null"`
	Operador       string    `json:"operador" gorm:"not null"`
	Descricao      string    `json:"descricao"`
	LatLong        string    `json:"lat_long"`
	FotoURL        string    `json:"foto_url"`
	Assinatura     string    `json:"assinatura"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// DocumentoFiscal representa documentos fiscais da remessa
type DocumentoFiscal struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	IDRemessa    string    `json:"id_remessa" gorm:"index;not null"`
	Tipo         string    `json:"tipo" gorm:"index;not null"` // "nfce", "cte", "nfe", "dacte"
	Numero       string    `json:"numero" gorm:"not null"`
	Serie        string    `json:"serie"`
	ChaveAcesso  string    `json:"chave_acesso" gorm:"uniqueIndex"`
	DataEmissao  time.Time `json:"data_emissao" gorm:"type:timestamptz"`
	ValorBRL     float64   `json:"valor_brl" gorm:"not null"`
	CFOP         string    `json:"cfop"`
	ICMS         float64   `json:"icms"`
	IPI          float64   `json:"ipi"`
	PIS         float64   `json:"pis"`
	COFINS       float64   `json:"cofins"`
	Emitente     string    `json:"emitente"`
	Destinatario string    `json:"destinatario"`
	Status       string    `json:"status" gorm:"default:'pendente'"`
	ArquivoURL   string    `json:"arquivo_url"`
	XMLURL       string    `json:"xml_url"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// WebhookNacional representa configuração de webhook
type WebhookNacional struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	ClientID  string    `json:"client_id" gorm:"index;not null"`
	URL       string    `json:"url" gorm:"not null"`
	Eventos   []string  `json:"eventos" gorm:"type:text[]"`
	Segredo   string    `json:"segredo" gorm:"not null"`
	Ativo     bool      `json:"ativo" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SolicitacaoCotacao representa solicitação de cotação
type SolicitacaoCotacao struct {
	OrigemCEP         string    `json:"origem_cep" validate:"required"`
	DestinoCEP        string    `json:"destino_cep" validate:"required"`
	PesoKg            float64   `json:"peso_kg" validate:"required,gt=0"`
	DimensoesCm       Dimensoes `json:"dimensoes_cm"`
	ValorDeclaradoBRL float64   `json:"valor_declarado_brl" validate:"required,gte=0"`
	ModalidadeFrete   string    `json:"modalidade_frete" validate:"required"`
	TipoServico       string    `json:"tipo_servico" validate:"required"`
	ReversaLogistica  bool      `json:"reversa_logistica"`
	SeguroObrigatorio bool      `json:"seguro_obrigatorio"`
	Coletar          bool      `json:"coletar"`
	Entregar         bool      `json:"entregar"`
	Agendamento       string    `json:"agendamento"`
	Observacoes       string    `json:"observacoes"`
}

// RespostaCotacao representa resposta de cotação
type RespostaCotacao struct {
	Operador          string    `json:"operador"`
	Servico           string    `json:"servico"`
	ValorFreteBRL     float64   `json:"valor_frete_brl"`
	ValorMinimoBRL    float64   `json:"valor_minimo_brl"`
	PrazoEntregaDias  int       `json:"prazo_entrega_dias"`
	PrazoMinimoDias   int       `json:"prazo_minimo_dias"`
	PrazoMaximoDias   int       `json:"prazo_maximo_dias"`
	ModalidadeFrete   string    `json:"modalidade_frete"`
	TipoServico       string    `json:"tipo_servico"`
	IncluiSeguro      bool      `json:"inclui_seguro"`
	IncluiReversa     bool      `json:"inclui_reversa"`
	IncluiColeta      bool      `json:"inclui_coleta"`
	IncluiEntrega     bool      `json:"inclui_entrega"`
	CO2Kg            float64   `json:"co2_kg"`
	Notas            string    `json:"notas"`
	Disponivel       bool      `json:"disponivel"`
	IDCotacao         string    `json:"id_cotacao"`
}

// SolicitacaoRemessa representa solicitação de criação de remessa
type SolicitacaoRemessa struct {
	Operador          string    `json:"operador" validate:"required"`
	Servico           string    `json:"servico" validate:"required"`
	IDRemessa         string    `json:"id_remessa" validate:"required"`
	OrigemCEP         string    `json:"origem_cep" validate:"required"`
	DestinoCEP        string    `json:"destino_cep" validate:"required"`
	OrigemUF          string    `json:"origem_uf"`
	DestinoUF         string    `json:"destino_uf"`
	OrigemMunicipio   string    `json:"origem_municipio"`
	DestinoMunicipio  string    `json:"destino_municipio"`
	PesoKg            float64   `json:"peso_kg" validate:"required,gt=0"`
	DimensoesCm       Dimensoes `json:"dimensoes_cm"`
	ValorDeclaradoBRL float64   `json:"valor_declarado_brl" validate:"required,gte=0"`
	ValorFreteBRL     float64   `json:"valor_frete_brl" validate:"required,gte=0"`
	ModalidadeFrete   string    `json:"modalidade_frete" validate:"required"`
	TipoProduto       string    `json:"tipo_produto"`
	NFCe             string    `json:"nfce"`
	CTe              string    `json:"cte"`
	ReversaLogistica bool      `json:"reversa_logistica"`
	Coletar         bool      `json:"coletar"`
	Entregar        bool      `json:"entregar"`
	Destinatario     Destinatario `json:"destinatario"`
	Remetente       Remetente    `json:"remetente"`
	ChaveIdempotencia string    `json:"chave_idempotencia"`
}

// RespostaRemessa representa resposta de criação de remessa
type RespostaRemessa struct {
	CodigoRastreamento string    `json:"codigo_rastreamento"`
	URLEtiqueta        string    `json:"url_etiqueta"`
	PrevisaoEntrega    time.Time `json:"previsao_entrega"`
	RotaHubs           []string  `json:"rota_hubs"`
	IDRemessa          string    `json:"id_remessa"`
	Operador           string    `json:"operador"`
	Servico            string    `json:"servico"`
	ValorFreteBRL       float64   `json:"valor_frete_brl"`
	CreatedAt          time.Time `json:"created_at"`
}

// RespostaRastreamento representa resposta de rastreamento
type RespostaRastreamento struct {
	CodigoRastreamento string                  `json:"codigo_rastreamento"`
	Status             string                  `json:"status"`
	Eventos            []EventoRastreamento   `json:"eventos"`
	PrevisaoEntrega    time.Time               `json:"previsao_entrega"`
	DataEntrega        *time.Time              `json:"data_entrega"`
	Operador           string                  `json:"operador"`
	Servico            string                  `json:"servico"`
	OrigemCEP          string                  `json:"origem_cep"`
	DestinoCEP         string                  `json:"destino_cep"`
	PesoKg            float64                 `json:"peso_kg"`
	ValorDeclaradoBRL  float64                 `json:"valor_declarado_brl"`
	Documentos         []DocumentoInfo         `json:"documentos"`
	HistoricoStatus    []StatusHistorico       `json:"historico_status"`
}

// EventoRastreamentoSimples representa evento simplificado
type EventoRastreamentoSimples struct {
	Timestamp   time.Time `json:"timestamp"`
	Local      string    `json:"local"`
	UF         string    `json:"uf"`
	Status     string    `json:"status"`
	Operador   string    `json:"operador"`
	Descricao string    `json:"descricao"`
}

// DocumentoInfo representa informações de documento
type DocumentoInfo struct {
	Tipo       string `json:"tipo"`
	Numero     string `json:"numero"`
	Status     string `json:"status"`
	Verificado bool   `json:"verificado"`
	ArquivoURL string `json:"arquivo_url"`
}

// StatusHistorico representa histórico de status
type StatusHistorico struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Responsavel string    `json:"responsavel"`
	Motivo      string    `json:"motivo"`
}

// MetricasNacional representa métricas do sistema
type MetricasNacional struct {
	Periodo                string  `json:"periodo"`
	TotalRemessas          int64   `json:"total_remessas"`
	TotalValorBRL          float64 `json:"total_valor_brl"`
	MediaDiasEntrega       float64 `json:"media_dias_entrega"`
	TaxaEntregaNoPrazo     float64 `json:"taxa_entrega_no_prazo"`
	MediaCustoPorKgBRL     float64 `json:"media_custo_por_kg_brl"`
	CO2PorPacoteKg        float64 `json:"co2_por_pacote_kg"`
	TaxaReversaLogistica   float64 `json:"taxa_reversa_logistica"`
	TaxaColetasRealizadas  float64 `json:"taxa_coletas_realizadas"`
	TopDestinos            []UFMetric `json:"top_destinos"`
	TopOperadores          []OperadorMetric `json:"top_operadores"`
	TopTiposServico        []ServicoMetric `json:"top_tipos_servico"`
}

// UFMetric representa métricas por UF
type UFMetric struct {
	UF            string  `json:"uf"`
	RemessasCount int64   `json:"remessas_count"`
	ValorBRL      float64 `json:"valor_brl"`
	MediaDias     float64 `json:"media_dias"`
}

// OperadorMetric representa métricas por operador
type OperadorMetric struct {
	Operador      string  `json:"operador"`
	RemessasCount int64   `json:"remessas_count"`
	ValorBRL      float64 `json:"valor_brl"`
	MediaDias     float64 `json:"media_dias"`
}

// ServicoMetric representa métricas por tipo de serviço
type ServicoMetric struct {
	Servico       string  `json:"servico"`
	RemessasCount int64   `json:"remessas_count"`
	ValorBRL      float64 `json:"valor_brl"`
	MediaDias     float64 `json:"media_dias"`
}

// RelatorioConformidade representa relatório de conformidade
type RelatorioConformidade struct {
	Periodo              string    `json:"periodo"`
	DataGeracao          time.Time `json:"data_geracao"`
	TotalSolicitacoes    int64     `json:"total_solicitacoes"`
	SolicitacoesBloqueadas int64    `json:"solicitacoes_bloqueadas"`
	CamposSanitizados     int64     `json:"campos_sanitizados"`
	ViolacoesAcesso       int64     `json:"violacoes_acesso"`
	LogsAuditados         int64     `json:"logs_auditados"`
	Operadores            []OperadorConformidade `json:"operadores"`
}

// OperadorConformidade representa conformidade por operador
type OperadorConformidade struct {
	Operador        string  `json:"operador"`
	Solicitacoes   int64   `json:"solicitacoes"`
	Bloqueadas     int64   `json:"bloqueadas"`
	CamposAcessados []string `json:"campos_acessados"`
	ScoreConformidade float64 `json:"score_conformidade"`
}

// ConfiguracaoGovernanca representa configuração de governança
type ConfiguracaoGovernanca struct {
	CamposPermitidos    []string          `json:"campos_permitidos"`
	CamposBloqueados    []string          `json:"campos_bloqueados"`
	CamposObrigatorios  []string          `json:"campos_obrigatorios"`
	CamposSensiveis     []string          `json:"campos_sensiveis"`
	CamposPseudonimizar []string          `json:"campos_pseudonimizar"`
	LimitesTaxa         map[string]int    `json:"limites_taxa"`
	RequisitosSLA       map[string]string `json:"requisitos_sla"`
}

// Dimensoes representa dimensões do pacote
type Dimensoes struct {
	Comprimento float64 `json:"comprimento"`
	Largura     float64 `json:"largura"`
	Altura      float64 `json:"altura"`
}

// Destinatario representa informações do destinatário
type Destinatario struct {
	NomeHash      string `json:"nome_hash"`
	EnderecoHash  string `json:"endereco_hash"`
	BairroHash    string `json:"bairro_hash"`
	CidadeHash    string `json:"cidade_hash"`
	UF            string `json:"uf"`
	CEP          string `json:"cep"`
	TelefoneHash  string `json:"telefone_hash"`
	EmailHash     string `json:"email_hash"`
	CPFCNPJHash   string `json:"cpf_cnpj_hash"`
}

// Remetente representa informações do remetente
type Remetente struct {
	NomeHash      string `json:"nome_hash"`
	EnderecoHash  string `json:"endereco_hash"`
	BairroHash    string `json:"bairro_hash"`
	CidadeHash    string `json:"cidade_hash"`
	UF            string `json:"uf"`
	CEP          string `json:"cep"`
	TelefoneHash  string `json:"telefone_hash"`
	EmailHash     string `json:"email_hash"`
	CPFCNPJHash   string `json:"cpf_cnpj_hash"`
	IEHash        string `json:"ie_hash"`
}

// AuditLogNacional representa log de auditoria
type AuditLogNacional struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	RequestID      string    `json:"request_id" gorm:"index"`
	OperadorID     string    `json:"operador_id" gorm:"index;not null"`
	Endpoint       string    `json:"endpoint" gorm:"index;not null"`
	MetodoHTTP     string    `json:"metodo_http" gorm:"not null"`
	CamposAcessados []string  `json:"campos_acessados" gorm:"type:text[]"`
	CamposBloqueados []string `json:"campos_bloqueados" gorm:"type:text[]"`
	Timestamp      time.Time `json:"timestamp" gorm:"type:timestamptz;index"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	StatusCode     int       `json:"status_code"`
	TempoRespostaMs int64     `json:"tempo_resposta_ms"`
	HashAnterior   string    `json:"hash_anterior" gorm:"not null"`
	HashAtual      string    `json:"hash_atual" gorm:"uniqueIndex;not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}
