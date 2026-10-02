# RLDB - API UNIVERSAL

API de Controle da Rede Logística Digital Brasileira (RLDB) - Integração Crossborder TR8

## Visão Geral

A RLDB - API UNIVERSAL é uma API de controle que funciona como um **orquestrador neutro**, permitindo que qualquer participante autorizado troque dados operacionais sem expor informações competitivas. Esta implementação inclui suporte completo para operações **crossborder TR8** (Transporte Rodoviário Internacional de Cargas), permitindo integração internacional com conformidade total às regulamentações de comércio exterior.

## Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                    CAMADA DE APLICAÇÃO                       │
│  (Marketplaces, ERPs, Sistemas de Vendedores)                │
└─────────────────────────────────────────────────────────────┘
                           │ HTTPS + OAuth 2.0 + JWT
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              API GATEWAY DA RLDB                             │
│  • Autenticação / Autorização                                │
│  • Rate Limiting / SLA                                       │
│  • Validação de Schema                                       │
│  • Auditoria e Logs (LGPD + CADE)                            │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│           ORQUESTRADOR LOGÍSTICO CENTRAL                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Rastreamento│  │ Cotação     │  │ Despacho / Etiquetas│ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Reversa     │  │ Fulfillment │  │ Observatório (KPIs) │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              CAMADA DE OPERADORES                            │
│  Correios │ Mercado Livre │ Shopee │ Transportadoras │ Hubs  │
└─────────────────────────────────────────────────────────────┘
```

## Funcionalidades Principais

### 1. Integração Crossborder TR8
- Cotação de frete internacional
- Criação de remessas TR8
- Rastreamento unificado
- Gerenciamento de documentos aduaneiros
- Suporte a Incoterms (FOB, CIF, DDP, etc.)
- Cálculo de impostos e taxas

### 2. Governança de Dados
- Classificação e anonimização de campos
- Controle de acesso (RBAC)
- Firewall de dados competitivos
- Auditoria imutável (blockchain-like)

### 3. Observatório
- Métricas em tempo real
- KPIs de desempenho
- Relatórios de conformidade
- Análise de rotas

## Pré-requisitos

- Go 1.21 ou superior
- PostgreSQL 12 ou superior
- Redis 6 ou superior (opcional, para cache)

## Instalação

### 1. Clonar o repositório

```bash
git clone https://github.com/rldb-br/rldb-api-universal.git
cd rldb-api-universal
```

### 2. Configurar ambiente

Crie um arquivo `.env` com as seguintes variáveis:

```bash
# Server
SERVER_PORT=8080
SERVER_TIMEOUT=30s

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=rldb_user
DB_PASSWORD=rldb_password
DB_NAME=rldb_api
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-very-secure-secret-key
JWT_ISSUER=rldb-api-universal
JWT_AUDIENCE=rldb-clients
JWT_REALM=RLDB API
JWT_EXPIRES_IN=15m

# Rate Limiting
RATE_LIMIT_DEFAULT=100
RATE_LIMIT_PER_MIN=60

# Logging
LOG_LEVEL=info

# API
API_VERSION=v1
BASE_URL=https://api.rldb.gov.br

# TR8 Crossborder
TR8_ENABLED=true
TR8_DEFAULT_OPERATOR=correios_international
TR8_MAX_WEIGHT_KG=100.0
TR8_SUPPORTED_COUNTRIES=BR,US,CN,DE,FR,UK

# Governance
GOVERNANCE_ENABLED=true
AUDIT_ENABLED=true
```

### 3. Instalar dependências

```bash
go mod download
```

### 4. Inicializar banco de dados

```bash
# Criar banco de dados
createdb rldb_api

# Executar migrações (serão executadas automaticamente na primeira inicialização)
```

### 5. Iniciar a API

```bash
go run cmd/main.go
```

## Uso

### Autenticação

Para obter um token de acesso:

```bash
curl -X POST https://api.rldb.gov.br/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "seu_client_id",
    "client_secret": "seu_client_secret",
    "scope": "logistics:read logistics:write tracking:read"
  }'
```

### Cotação TR8

```bash
curl -X POST https://api.rldb.gov.br/api/v1/crossborder/tr8/quotes \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "origin_country": "BR",
    "destination_country": "US",
    "origin_zip": "01001-000",
    "destination_zip": "90210",
    "weight_kg": 2.5,
    "dimensions_cm": {"l": 30, "w": 20, "h": 15},
    "declared_value_brl": 150.00,
    "incoterm": "FOB",
    "hs_code": "6109.10.00",
    "service_level": "standard"
  }'
```

### Criar Remessa TR8

```bash
curl -X POST https://api.rldb.gov.br/api/v1/crossborder/tr8/shipments \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operator": "correios_international",
    "service": "tr8_standard",
    "shipment_id": "ML-987654321",
    "origin_country": "BR",
    "destination_country": "US",
    "origin_zip": "01001-000",
    "destination_zip": "90210",
    "weight_kg": 2.5,
    "dimensions_cm": {"l": 30, "w": 20, "h": 15},
    "declared_value_brl": 150.00,
    "declared_value_usd": 30.00,
    "incoterm": "FOB",
    "hs_code": "6109.10.00",
    "product_category": "electronics",
    "reverse_logistics_enabled": false,
    "recipient": {
      "name_hash": "a3f8c2...",
      "address_hash": "b7e1d4...",
      "country": "US"
    },
    "idempotency_key": "unique-key-123"
  }'
```

### Rastreamento

```bash
curl -X GET https://api.rldb.gov.br/api/v1/crossborder/tr8/tracking/TR812345678901 \
  -H "Authorization: Bearer SEU_TOKEN"
```

## Endpoints Principais

### Autenticação
- `POST /api/v1/auth/token` - Obter token JWT

### TR8 Crossborder
- `POST /api/v1/crossborder/tr8/quotes` - Obter cotações
- `POST /api/v1/crossborder/tr8/shipments` - Criar remessa
- `GET /api/v1/crossborder/tr8/shipments/{shipment_id}` - Obter remessa
- `GET /api/v1/crossborder/tr8/tracking/{tracking_code}` - Rastreamento
- `POST /api/v1/crossborder/tr8/tracking/{tracking_code}` - Atualizar rastreamento
- `POST /api/v1/crossborder/tr8/customs/documents` - Criar documento aduaneiro
- `PUT /api/v1/crossborder/tr8/customs/documents/{document_id}/validate` - Validar documento
- `GET /api/v1/crossborder/tr8/customs/documents/{document_id}` - Obter documento
- `POST /api/v1/crossborder/tr8/webhooks` - Criar webhook
- `GET /api/v1/crossborder/tr8/webhooks` - Listar webhooks

### Observatório
- `GET /api/v1/observatory/tr8/metrics` - Métricas TR8
- `GET /api/v1/observatory/tr8/compliance` - Relatório de conformidade

### Governança
- `GET /api/v1/governance/tr8/config` - Configuração de governança

## Segurança e Compliance

### Autenticação
- OAuth 2.0 + JWT com escopos granulares
- Tokens com validade curta (15 minutos)

### Autorização
- RBAC (Role-Based Access Control) por operador
- Controle de acesso baseados em escopos

### Criptografia
- TLS 1.3 obrigatório
- AES-256 em repouso
- Campos sensíveis com HSM (opcional)

### Auditoria
- Logs imutáveis com hash encadeado (tipo blockchain)
- Registro de todos os acessos com timestamp e ID do operador

### Conformidade
- **LGPD**: Pseudonimização automática de dados pessoais
- **CADE**: Relatório automático de acessos cruzados entre concorrentes
- **TR8**: Conformidade com regulamentações de comércio exterior

## Governança de Dados

A camada de governança implementa as seguintes regras:

### Campos Permitidos
- `origin_country`, `destination_country`
- `origin_zip`, `destination_zip`
- `weight_kg`, `dimensions_cm`
- `declared_value_brl`, `declared_value_usd`
- `incoterm`, `hs_code`, `product_category`
- `status`, `timestamp`, `tracking_code`

### Campos Bloqueados
- `product_price`, `margin`
- `customer_email`, `customer_name`
- `full_address`, `sales_strategy`
- `competitor_data`, `campaign_info`
- `internal_cost`, `profit_margin`

### Campos Pseudonimizados
- `name`, `address`, `city`, `phone`, `email`

## Stack Tecnológica

- **API Gateway**: Gin (Go)
- **Backend**: Go 1.21
- **Banco de dados**: PostgreSQL + TimescaleDB
- **Cache**: Redis
- **Fila de eventos**: Apache Kafka (opcional)
- **Monitoramento**: Prometheus + Grafana
- **Segurança**: HashiCorp Vault (opcional)

## Estrutura do Projeto

```
rldb-api-universal/
├── cmd/
│   └── main.go                    # Entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration
│   ├── handler/
│   │   └── crossborder_tr8_handler.go  # HTTP handlers
│   ├── middleware/
│   │   └── auth_middleware.go     # Authentication middleware
│   ├── models/
│   │   └── crossborder_tr8.go      # Data models
│   ├── repository/
│   │   └── crossborder_tr8_repository.go  # Database operations
│   └── service/
│       └── crossborder_tr8_service.go   # Business logic
├── pkg/
│   └── governance/
│       └── data_governance.go     # Data governance
├── docs/
│   └── openapi.yaml               # OpenAPI specification
├── migrations/
├── scripts/
├── go.mod
├── go.sum
└── README.md
```

## Contribuição

1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

## Licença

MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

## Contato

- **Email**: api-support@rldb.gov.br
- **Site**: https://rldb.gov.br
- **Documentação**: https://docs.rldb.gov.br

## Status do Projeto

Este projeto está em desenvolvimento ativo. A integração TR8 crossborder está funcional e pronta para uso em produção.

## Roadmap

| Fase | Prazo | Entregável |
|------|-------|------------|
| Fase 0 | 0-3 meses | Especificação OpenAPI 3.1, schema de dados, modelo de segurança |
| Fase 1 | 3-6 meses | Sandbox com Correios + 1 marketplace |
| Fase 2 | 6-12 meses | Piloto regional (SP + Nordeste) com 3 operadores |
| Fase 3 | 12-24 meses | Produção nacional com todos os operadores |
| Fase 4 | 24-36 meses | Integração cross-border (APIs internacionais) |

## Limitações Jurídicas

A API **NÃO**:
1. Compartilha listas de clientes entre marketplaces
2. Permite que operadores vejam preços de concorrentes
3. Permite coordenação de campanhas ou margens
4. Substitui a análise de atos de concentração pelo CADE
5. Armazena dados pessoais sem base legal (LGPD)
6. Obriga nenhum operador a participar (adesão voluntária)

## Suporte

Para suporte técnico, entre em contato com api-support@rldb.gov.br.

Para relatar problemas de segurança, envie um email para security@rldb.gov.br.
