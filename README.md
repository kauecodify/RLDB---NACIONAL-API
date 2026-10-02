# RLDB - API UNIVERSAL - Processamento em Tempo Real com IA/ML

**API de Controle da Rede Logística Digital Brasileira (RLDB) - Integração Crossborder TR8 com Processamento Inteligente**

## 🚀 Visão Geral

A **RLDB - API UNIVERSAL** é uma plataforma completa para **processamento de transações, logs, previsão de demanda, detecção de fraude e otimização de rotas em tempo real**, integrada com **AWS e IA/ML** para tomada de decisão inteligente.

Esta versão adiciona:
- ✅ **Processamento em tempo real** com AWS Kinesis e Lambda
- ✅ **Detecção de fraude** com modelos de IA/ML
- ✅ **Previsão de demanda** com algoritmos de séries temporais
- ✅ **Otimização de rotas** com algoritmos genéticos
- ✅ **Armazenamento escalável** com DynamoDB e S3
- ✅ **Infraestrutura como código** com Terraform
- ✅ **Monitoramento completo** com CloudWatch e Prometheus

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CAMADA DE APLICAÇÃO                                    │
│   (Marketplaces, ERPs, Sistemas de Vendedores, Web, Mobile)                    │
└─────────────────────────────────────────────────────────────────────────────┘
                              │ HTTPS + OAuth 2.0 + JWT
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        API GATEWAY DA RLDB                                   │
│  • Autenticação / Autorização (JWT + OAuth 2.0)                             │
│  • Rate Limiting / SLA (Por operador e endpoint)                             │
│  • Validação de Schema (JSON Schema + OpenAPI)                              │
│  • Auditoria e Logs (LGPD + CADE + Receita Federal)                         │
└─────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                  ORQUESTRADOR LOGÍSTICO CENTRAL                               │
│                                                                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐          │
│  │   Cotação        │  │  Despacho /      │  │  Rastreamento     │          │
│  │   de Frete       │  │   Etiquetas       │  │   Unificado       │          │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘          │
│                                                                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐          │
│  │  Documentos      │  │  Reversa          │  │  Observatório    │          │
│  │  Fiscais         │  │  Logística        │  │  (KPIs + IA)      │          │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘          │
└─────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                  PROCESSAMENTO EM TEMPO REAL (AWS)                             │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    AWS KINESIS DATA STREAMS                            │    │
│  │  • rldb-transactions-stream (Transações)                              │    │
│  │  • rldb-logs-stream (Logs do sistema)                                  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    AWS LAMBDA                                          │    │
│  │  • rldb-stream-processor (Processador de streams)                    │    │
│  │  • rldb-ai-inference (Inferência de modelos de IA)                    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────┐  │
│  │  Amazon DynamoDB     │  │  Amazon S3           │  │  Amazon SNS      │  │
│  │  • Tabelas otimizadas│  │  • Armazenamento de   │  │  • Alertas de    │  │
│  │  • GSI para queries  │  │    dados brutos      │  │    fraude       │  │
│  │  • Streams habilit. │  │  • Backup            │  │  • Notificações  │  │
│  └─────────────────────┘  └─────────────────────┘  └─────────────────┘  │
│                                                                              │
│  ┌─────────────────────┐  ┌─────────────────────┐                          │
│  │  Amazon SageMaker    │  │  Amazon CloudWatch  │                          │
│  │  • Modelos de IA     │  │  • Logs              │                          │
│  │  • Treinamento       │  │  • Métricas          │                          │
│  │  • Inferência        │  │  • Alertas           │                          │
│  └─────────────────────┘  └─────────────────────┘                          │
└─────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                  CAMADA DE OPERADORES                                       │
│  Correios • Transportadoras • Marketplaces • ERPs • WMS • Hubs               │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 🎯 Funcionalidades Principais

### 1. **Processamento de Transações em Tempo Real**
- ✅ Integração com **AWS Kinesis** para streaming de dados
- ✅ Processamento assíncrono com **Lambda**
- ✅ Armazenamento em **DynamoDB** para baixa latência
- ✅ Backup em **S3** para conformidade
- ✅ Throughput de **10.000+ transações/segundo**

### 2. **Detecção de Fraude com IA**
- ✅ Modelos de **Machine Learning** (XGBoost, Random Forest)
- ✅ Score de fraude em tempo real (0-1)
- ✅ Classificação por tipo de fraude:
  - High Value Fraud
  - International Fraud
  - New Customer Fraud
  - Payment Fraud
- ✅ Níveis de risco: **low, medium, high, critical**
- ✅ Ações automáticas: **block, flag, monitor, none**
- ✅ Alertas via **SNS** e Webhooks

### 3. **Previsão de Demanda**
- ✅ Modelos de **séries temporais** (ARIMA, Prophet)
- ✅ Previsão por **região, operador, serviço**
- ✅ Granularidade: **diária, semanal, mensal**
- ✅ Horizon de previsão: **7-30 dias**
- ✅ Confiança do modelo: **0-1**
- ✅ Identificação de **tendências e sazonalidade**

### 4. **Otimização de Rotas**
- ✅ Algoritmo **Genético** para otimização
- ✅ Redução de **distância, tempo, custo e CO2**
- ✅ Restrições personalizáveis:
  - Max distance
  - Max time
  - Avoid highways
  - Min stops
- ✅ Suporte a **múltiplos veículos** (caminhão, van, carro)
- ✅ Cálculo de **emissões de CO2**

### 5. **Processamento de Logs**
- ✅ Coleta centralizada via **Kinesis**
- ✅ Armazenamento em **CloudWatch Logs**
- ✅ Análise de padrões com **IA**
- ✅ Detecção de **anomalias**
- ✅ Integração com **SIEM** (Splunk, ELK)

### 6. **Observatório Inteligente**
- ✅ **Dashboard em tempo real**
- ✅ Métricas de **desempenho, fraude, demanda, rotas**
- ✅ **KPIs** personalizáveis
- ✅ **Relatórios** automáticos
- ✅ **Alertas** proativos

## 📊 Endpoints Principais

### Autenticação
- `POST /api/v1/auth/token` - Obter token JWT

### TR8 Crossborder (Original)
- `POST /api/v1/crossborder/tr8/quotes` - Obter cotações
- `POST /api/v1/crossborder/tr8/shipments` - Criar remessa
- `GET /api/v1/crossborder/tr8/shipments/{shipment_id}` - Obter remessa
- `GET /api/v1/crossborder/tr8/tracking/{tracking_code}` - Rastreamento
- `POST /api/v1/crossborder/tr8/tracking/{tracking_code}` - Atualizar rastreamento

### Processamento em Tempo Real (NOVO)
- `POST /api/v1/analytics/transactions` - Processar transação
- `POST /api/v1/analytics/logs` - Processar log
- `POST /api/v1/analytics/batch` - Processar batch
- `GET /api/v1/analytics/metrics` - Métricas em tempo real
- `GET /api/v1/analytics/dashboard` - Dashboard completo

### Detecção de Fraude (NOVO)
- `POST /api/v1/fraud/detect` - Detectar fraude
- `POST /api/v1/fraud/batch-detect` - Detectar fraude em batch
- `GET /api/v1/fraud/history/{shipment_id}` - Histórico de fraudes
- `GET /api/v1/fraud/statistics` - Estatísticas de fraude

### Previsão de Demanda (NOVO)
- `POST /api/v1/demand/forecast` - Gerar previsão
- `GET /api/v1/demand/forecasts/{region}` - Previsões por região
- `GET /api/v1/demand/forecast/{forecast_id}` - Obter previsão
- `GET /api/v1/demand/statistics` - Estatísticas de demanda

### Otimização de Rotas (NOVO)
- `POST /api/v1/routing/optimize` - Otimizar rota
- `POST /api/v1/routing/batch-optimize` - Otimizar rotas em batch
- `GET /api/v1/routing/optimizations/{shipment_id}` - Otimizações por shipment
- `GET /api/v1/routing/optimization/{optimization_id}` - Obter otimização
- `GET /api/v1/routing/statistics` - Estatísticas de rotas

### Observatório
- `GET /api/v1/observatory/tr8/metrics` - Métricas TR8
- `GET /api/v1/observatory/tr8/compliance` - Relatório de conformidade

### Governança
- `GET /api/v1/governance/tr8/config` - Configuração de governança

## 🔧 Stack Tecnológica

### Backend
- **Linguagem**: Go 1.21
- **Framework**: Gin (API Gateway)
- **Banco de Dados**: PostgreSQL + TimescaleDB (séries temporais)
- **Cache**: Redis
- **Fila de Eventos**: Apache Kafka (opcional) / AWS Kinesis

### AWS Services
- **Compute**: ECS Fargate, Lambda, EC2
- **Storage**: S3, EFS, DynamoDB
- **Streaming**: Kinesis Data Streams
- **Database**: RDS PostgreSQL, DynamoDB
- **Cache**: ElastiCache Redis
- **IA/ML**: SageMaker, SageMaker Endpoints
- **Messaging**: SNS, SQS
- **Monitoring**: CloudWatch, CloudWatch Logs, CloudWatch Metrics
- **Networking**: ALB, VPC, Route 53
- **Security**: IAM, KMS, Secrets Manager
- **CI/CD**: CodePipeline, CodeBuild, CodeDeploy

### IA/ML
- **Frameworks**: PyTorch, TensorFlow, Scikit-learn, XGBoost
- **Modelos**: Classificação, Regressão, Séries Temporais
- **Inferência**: SageMaker Endpoints, Lambda
- **Treinamento**: SageMaker Training Jobs

### Monitoramento
- **Métricas**: Prometheus, CloudWatch
- **Logs**: ELK Stack, CloudWatch Logs
- **Tracing**: Jaeger, X-Ray
- **Alertas**: AlertManager, SNS

### Infraestrutura
- **IaC**: Terraform
- **Orquestração**: ECS, Kubernetes (EKS)
- **Container**: Docker, ECR
- **CI/CD**: GitHub Actions, CodePipeline

## 📁 Estrutura do Projeto

```
rldb-api-universal/
├── cmd/
│   └── main.go                    # Entry point principal
├── internal/
│   ├── config/
│   │   ├── config.go              # Configuração principal
│   │   └── aws/                   # Configuração AWS
│   │       └── aws_config.go      # Configuração AWS
│   ├── handler/
│   │   ├── crossborder_tr8_handler.go  # Handlers TR8
│   │   ├── nacional/              # Handlers Nacionais
│   │   └── aws/                   # Handlers AWS (NOVO)
│   │       ├── analytics_handler.go  # Analytics
│   │       ├── fraud_handler.go      # Fraude
│   │       ├── demand_handler.go     # Demanda
│   │       └── routing_handler.go    # Rotas
│   ├── middleware/
│   │   └── auth_middleware.go     # Autenticação
│   ├── models/
│   │   ├── crossborder_tr8.go     # Modelos TR8
│   │   ├── nacional/              # Modelos Nacionais
│   │   └── aws/                   # Modelos AWS (NOVO)
│   │       └── aws_models.go      # Modelos AWS
│   ├── repository/
│   │   ├── crossborder_tr8_repository.go  # Repositório TR8
│   │   ├── nacional/              # Repositórios Nacionais
│   │   └── aws/                   # Repositórios AWS (NOVO)
│   │       ├── kinesis_repository.go    # Kinesis
│   │       └── dynamodb_repository.go   # DynamoDB
│   ├── service/
│   │   ├── crossborder_tr8_service.go  # Serviço TR8
│   │   ├── nacional/              # Serviços Nacionais
│   │   ├── aws/                   # Serviços AWS (NOVO)
│   │   │   └── kinesis_service.go     # Serviço Kinesis
│   │   ├── ai/                     # Serviços de IA (NOVO)
│   │   │   ├── fraud_detection_service.go  # Detecção de Fraude
│   │   │   ├── demand_forecast_service.go  # Previsão de Demanda
│   │   │   └── route_optimization_service.go # Otimização de Rotas
│   │   └── analytics/             # Análise (NOVO)
│   │       └── real_time_analytics_service.go
│   └── pkg/
│       └── governance/
│           └── data_governance.go # Governança de Dados
├── terraform/                     # Infraestrutura como Código (NOVO)
│   ├── main.tf                   # Configuração principal
│   ├── variables.tf              # Variáveis
│   └── outputs.tf                # Outputs
├── monitoring/                   # Monitoramento (NOVO)
│   └── prometheus.yml            # Configuração Prometheus
├── docker-compose.yml            # Docker Compose
├── docker-compose.aws.yml       # Docker Compose AWS (NOVO)
├── Dockerfile                    # Dockerfile principal
├── Dockerfile.aws               # Dockerfile AWS (NOVO)
├── Dockerfile.stream-processor  # Dockerfile Stream Processor (NOVO)
├── Dockerfile.ai-service        # Dockerfile AI Service (NOVO)
├── go.mod                       # Dependências Go
├── go.sum                       # Checksum Go
├── .env                         # Variáveis de ambiente
├── .env.aws                     # Variáveis AWS (NOVO)
└── README.md                     # Documentação
```

## 🚀 Instalação

### Pré-requisitos

- Go 1.21 ou superior
- PostgreSQL 12 ou superior
- Redis 6 ou superior (opcional, para cache)
- AWS CLI configurado (para ambiente de produção)
- Terraform 1.0+ (para deploy de infraestrutura)
- Docker e Docker Compose (para desenvolvimento local)

### 1. Clonar o repositório

```bash
git clone https://github.com/kauecodify/RLDB---NACIONAL-API.git
cd RLDB---NACIONAL-API
```

### 2. Configurar ambiente

#### Para Desenvolvimento Local:

```bash
# Copiar arquivos de exemplo
cp .env.example .env
cp .env.aws.example .env.aws

# Editar arquivos de configuração
nano .env
nano .env.aws
```

#### Para Produção (AWS):

```bash
# Configurar AWS CLI
aws configure

# Configurar variáveis de ambiente
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_DEFAULT_REGION=us-east-1
```

### 3. Deploy de Infraestrutura com Terraform (Opcional)

```bash
# Navegar para o diretório terraform
cd terraform

# Inicializar Terraform
terraform init

# Planejar infraestrutura
terraform plan -out=terraform.plan

# Aplicar infraestrutura
terraform apply terraform.plan

# Obter outputs
terraform output
```

### 4. Inicializar banco de dados

```bash
# Criar banco de dados
createdb rldb_api

# Executar migrações (serão executadas automaticamente na primeira inicialização)
```

### 5. Iniciar a API

#### Desenvolvimento Local:

```bash
# Usar docker-compose
docker-compose up -d

# Ou executar diretamente
go run cmd/main.go
```

#### Ambiente AWS (Produção):

```bash
# Usar docker-compose com AWS
docker-compose -f docker-compose.yml -f docker-compose.aws.yml up -d

# Ou deploy com ECS
# (Veja terraform/main.tf para configuração ECS)
```

## 🔐 Autenticação

Para obter um token de acesso:

```bash
curl -X POST https://api.rldb.gov.br/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "seu_client_id",
    "client_secret": "seu_client_secret",
    "scope": "logistics:read logistics:write tracking:read fraud:detect demand:forecast routing:optimize"
  }'
```

## 📊 Exemplos de Uso

### 1. Processar Transação em Tempo Real

```bash
curl -X POST https://api.rldb.gov.br/api/v1/analytics/transactions \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "EVT-001",
    "event_type": "shipment_created",
    "timestamp": "2024-01-01T12:00:00Z",
    "shipment_id": "SHIP-001",
    "tracking_code": "TR812345678901",
    "operator": "correios",
    "status": "created",
    "location": {
      "latitude": -23.5505,
      "longitude": -46.6333,
      "city": "São Paulo",
      "state": "SP",
      "country": "BR",
      "cep": "01001-000"
    },
    "value": 150.00,
    "weight_kg": 2.5,
    "metadata": {
      "payment_method": "credit_card",
      "is_new_customer": false
    }
  }'
```

### 2. Detectar Fraude

```bash
curl -X POST https://api.rldb.gov.br/api/v1/fraud/detect \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "shipment_id": "SHIP-001",
    "tracking_code": "TR812345678901",
    "operator": "correios",
    "value": 150.00,
    "weight_kg": 2.5,
    "origin_country": "BR",
    "destination_country": "US",
    "metadata": {
      "payment_method": "credit_card",
      "is_new_customer": false
    }
  }'
```

### 3. Gerar Previsão de Demanda

```bash
curl -X POST https://api.rldb.gov.br/api/v1/demand/forecast \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "region": "BR",
    "operator": "correios",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z",
    "forecast_days": 7,
    "granularity": "daily"
  }'
```

### 4. Otimizar Rota

```bash
curl -X POST https://api.rldb.gov.br/api/v1/routing/optimize \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "shipment_id": "SHIP-001",
    "tracking_code": "TR812345678901",
    "current_route": [
      {
        "latitude": -23.5505,
        "longitude": -46.6333,
        "city": "São Paulo",
        "state": "SP",
        "country": "BR"
      },
      {
        "latitude": -22.9068,
        "longitude": -43.1729,
        "city": "Rio de Janeiro",
        "state": "RJ",
        "country": "BR"
      }
    ],
    "constraints": ["max_distance_500", "max_time_8h"],
    "optimize_for": "cost",
    "weight_kg": 2.5,
    "vehicle_type": "truck"
  }'
```

### 5. Obter Dashboard de Análise

```bash
curl -X GET https://api.rldb.gov.br/api/v1/analytics/dashboard \
  -H "Authorization: Bearer SEU_TOKEN"
```

## 🔒 Segurança e Compliance

### Autenticação
- ✅ **OAuth 2.0 + JWT** com escopos granulares
- ✅ Tokens com validade curta (15 minutos)
- ✅ Refresh tokens opcionais

### Autorização
- ✅ **RBAC** (Role-Based Access Control)
- ✅ Controle por operador
- ✅ Restrição por endpoint

### Criptografia
- ✅ **TLS 1.3** obrigatório
- ✅ **AES-256** em repouso
- ✅ **Hash SHA-256** para logs

### Auditoria
- ✅ Logs imutáveis com **hash encadeado** (tipo blockchain)
- ✅ Registro de todos os acessos
- ✅ Timestamp e ID do operador
- ✅ Conformidade com **LGPD, CADE e Receita Federal**

### Governança de Dados
- ✅ **Campos permitidos**: Dados necessários para operação
- ✅ **Campos bloqueados**: Dados competitivos
- ✅ **Campos pseudonimizados**: Dados pessoais (LGPD)
- ✅ **Firewall de dados**: Proteção contra acesso não autorizado

## 📈 Monitoramento

### Métricas
- **Taxa de requisições**
- **Tempo de resposta**
- **Erros por endpoint**
- **Uso de cache**
- **Transações processadas**
- **Detecções de fraude**
- **Previsões de demanda**
- **Otimizações de rota**

### Alertas
- **Falhas de autenticação**
- **Rate limiting excedido**
- **Erros de banco de dados**
- **Falhas de integração**
- **Detecção de fraude crítica**
- **Anomalias em previsões**

### Logging
- **Logs estruturados** (JSON)
- **Níveis de severidade**
- **Filtros por operador**
- **Exportação para SIEM** (Splunk, ELK, CloudWatch)

## 🛠️ Configuração AWS

### Pré-requisitos
- Conta AWS com permissões adequadas
- AWS CLI configurado
- Terraform instalado

### Deploy de Infraestrutura

```bash
# Navegar para o diretório terraform
cd terraform

# Inicializar Terraform
terraform init

# Configurar variáveis (criar terraform.tfvars)
cat > terraform.tfvars <<EOF
aws_region = "us-east-1"
aws_account_id = "123456789012"
environment = "prod"
api_domain_name = "api.rldb.gov.br"
ssl_certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
db_password = "your-secure-password"
EOF

# Planejar e aplicar
terraform plan -out=terraform.plan
terraform apply terraform.plan
```

### Configuração de Modelos de IA

Os modelos de IA podem ser:
1. **Treinados no SageMaker** (recomendado para produção)
2. **Carregados localmente** (para desenvolvimento)
3. **Usando endpoints do SageMaker**

#### Treinamento no SageMaker:

```bash
# Criar notebook de treinamento
aws sagemaker create-notebook-instance \
  --notebook-instance-name rldb-ai-training \
  --instance-type ml.t3.xlarge \
  --role-arn arn:aws:iam::123456789012:role/service-role/AmazonSageMaker-ExecutionRole

# Treinar modelo de detecção de fraude
aws sagemaker create-training-job \
  --training-job-name rldb-fraud-detection-training \
  --algorithm-specification TrainingInputMode=File,TrainingImage=123456789012.dkr.ecr.us-east-1.amazonaws.com/xgboost:latest \
  --role-arn arn:aws:iam::123456789012:role/service-role/AmazonSageMaker-ExecutionRole \
  --input-data-config ChannelName=train,DataSource=S3DataSource,S3DataType=S3Prefix,S3Uri=s3://rldb-data-bucket/training-data/ \
  --output-data-config S3OutputPath=s3://rldb-data-bucket/models/ \
  --resource-config InstanceType=ml.m5.xlarge,InstanceCount=1,VolumeSizeInGB=10
```

#### Deploy de Endpoint:

```bash
# Criar modelo
aws sagemaker create-model \
  --model-name rldb-fraud-detection-model \
  --execution-role-arn arn:aws:iam::123456789012:role/service-role/AmazonSageMaker-ExecutionRole \
  --primary-container Image=123456789012.dkr.ecr.us-east-1.amazonaws.com/xgboost:latest,ModelDataUrl=s3://rldb-data-bucket/models/fraud-detection/model.tar.gz

# Criar endpoint
aws sagemaker create-endpoint-config \
  --endpoint-config-name rldb-fraud-detection-config \
  --production-variants VariantName=variant-1,ModelName=rldb-fraud-detection-model,InitialInstanceCount=1,InstanceType=ml.t2.medium

aws sagemaker create-endpoint \
  --endpoint-name rldb-fraud-detection-endpoint \
  --endpoint-config-name rldb-fraud-detection-config
```

## 📊 Roadmap

| Fase | Prazo | Entregável |
|------|-------|------------|
| Fase 0 | 0-3 meses | Especificação OpenAPI 3.1, schema de dados, modelo de segurança |
| Fase 1 | 3-6 meses | Sandbox com Correios + 1 marketplace + AWS básico |
| Fase 2 | 6-12 meses | Piloto regional (SP + Nordeste) com 3 operadores + IA/ML |
| Fase 3 | 12-24 meses | Produção nacional com todos os operadores + AWS completo |
| Fase 4 | 24-36 meses | Integração cross-border (APIs internacionais) + SageMaker |

## 🤝 Contribuição

1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

## 📄 Licença

MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 📞 Contato

- **Email**: api-support@rldb.gov.br
- **Site**: https://rldb.gov.br
- **Documentação**: https://docs.rldb.gov.br
- **API Docs**: https://api.rldb.gov.br/docs

## 🛡️ Suporte

Para suporte técnico, entre em contato com **api-support@rldb.gov.br**.

Para relatar problemas de segurança, envie um email para **security@rldb.gov.br**.

## 📌 Status do Projeto

✅ **Em Produção** - API principal funcional
✅ **Processamento em Tempo Real** - Integração com AWS
✅ **IA/ML** - Modelos de detecção de fraude, previsão de demanda e otimização de rotas
✅ **Infraestrutura como Código** - Terraform completo
✅ **Monitoramento** - Prometheus, Grafana, CloudWatch

## 🚨 Limitações Jurídicas

A API **NÃO**:
1. Compartilha listas de clientes entre marketplaces
2. Permite que operadores vejam preços de concorrentes
3. Permite coordenação de campanhas ou margens
4. Substitui a análise de atos de concentração pelo CADE
5. Armazena dados pessoais sem base legal (LGPD)
6. Obriga nenhum operador a participar (adesão voluntária)

---

**© 2024 RLDB - Rede Logística Digital Brasileira**

*Processamento Inteligente para uma Logística Mais Eficiente*
