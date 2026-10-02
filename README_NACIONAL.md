# RLDB - API UNIVERSAL - TR8 Nacional

Solução consolidada para o problema da operação logística nacional brasileira.

## Visão Geral

A **RLDB - API UNIVERSAL - TR8 Nacional** é uma solução completa e adaptável para qualquer cenário de transporte e logística no Brasil. Esta implementação consolida o problema da operação em uma solução TR8 plausível, que pode ser adaptada para qualquer contexto de transporte rodoviário, aéreo, marítimo ou ferroviário.

## Problema Resolvido

A solução TR8 Nacional resolve os seguintes problemas do setor logístico brasileiro:

1. **Fragmentação do Mercado**: Integra todos os operadores logísticos (Correios, transportadoras, marketplaces) em uma única plataforma neutra.

2. **Falta de Padrões**: Estabelece padrões únicos para cotação, rastreamento e documentação fiscal.

3. **Baixa Interoperabilidade**: Permite que sistemas de diferentes operadores se comuniquem sem perder informações críticas.

4. **Falta de Transparência**: Fornece visibilidade total do processo logístico para todos os atores autorizados.

5. **Complexidade Regulatória**: Implementa automaticamente conformidade com LGPD, CADE e regulamentações fiscais.

6. **Ineficiência Operacional**: Reduz custos e tempo através de automação e padronização.

## Arquitetura da Solução

```
┌─────────────────────────────────────────────────────────────────┐
│                    CAMADA DE APLICAÇÃO                             │
│  (Marketplaces, ERPs, TMS, WMS, Sistemas de Vendedores)             │
└─────────────────────────────────────────────────────────────────┘
                              │ HTTPS + OAuth 2.0 + JWT
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    API GATEWAY DA RLDB                             │
│  • Autenticação / Autorização (JWT + OAuth 2.0)                   │
│  • Rate Limiting / SLA (Por operador e endpoint)                   │
│  • Validação de Schema (JSON Schema + OpenAPI)                    │
│  • Auditoria e Logs (LGPD + CADE + Receita Federal)              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              ORQUESTRADOR LOGÍSTICO NACIONAL                       │
│                                                                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Cotação       │  │  Despacho /     │  │   Rastreamento   │  │
│  │   de Frete      │  │   Etiquetas      │  │   Unificado      │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│                                                                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Documentos    │  │   Reversa        │  │   Observatório   │  │
│  │   Fiscais       │  │   Logística      │  │   (KPIs)         │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    CAMADA DE OPERADORES                           │
│                                                                      │
│  Correios  │  Transportadoras  │  Marketplaces  │  ERPs  │  WMS    │
│  (Sedex,  │  (ABC, XYZ, etc)  │  (ML, Shopee,  │       │         │
│   PAC)    │                  │   etc)         │       │         │
└─────────────────────────────────────────────────────────────────┘
```

## Funcionalidades Principais

### 1. Cotação de Frete Nacional
- Cotação unificada para todos os operadores
- Comparação de preços, prazos e serviços
- Suporte a modalidades CIF, FOB, Por Conta
- Cálculo de CO2 por remessa
- Cache de cotações para performance

### 2. Criação de Remessas
- Criação de remessas com idempotência
- Geração automática de código de rastreamento
- Suporte a documentos fiscais (NFCe, NFe, CTe)
- Integração com sistemas de etiquetagem
- Validação de dados obrigatórios

### 3. Rastreamento Unificado
- Rastreamento em tempo real
- Histórico completo de eventos
- Notificações automáticas via webhook
- Suporte a múltiplos operadores por remessa
- Geolocalização e assinatura digital

### 4. Documentação Fiscal
- Gerenciamento de NFCe, NFe, CTe, DACTe
- Validação de documentos
- Armazenamento seguro e auditável
- Integração com Receita Federal

### 5. Reversa Logística
- Processo de devolução padronizado
- Rastreamento de reversa
- Integração com centros de distribuição

### 6. Observatório
- Métricas de desempenho por operador
- KPIs de entrega e prazo
- Análise de custos e eficiência
- Relatórios de conformidade

## Princípios de Design

### 1. Neutralidade
- Todos os operadores acessam os mesmos endpoints
- Mesmas regras para todos os participantes
- Sem privilégios ou restrições arbitrárias

### 2. Segregação de Dados
- Dados competitivos nunca trafegam na API
- Cada operador só vê o que é necessário para sua operação
- Firewall de dados sensíveis

### 3. Mínimo Necessário
- Apenas informações essenciais são compartilhadas
- Dados pessoais são pseudonimizados (LGPD)
- Informações fiscais são protegidas

### 4. Auditoria Total
- Toda chamada é logada com hash, timestamp e ID do operador
- Logs imutáveis (blockchain-like)
- Conformidade com CADE e Receita Federal

### 5. Idempotência
- Retries seguros sem duplicar remessas
- Chaves de idempotência para prevenção de duplicação

### 6. Versionamento
- Suporte a múltiplas versões de API
- Migração suave entre versões

## Modelos de Dados

### Remessa Nacional
Representa uma remessa completa com todas as informações necessárias:
- Operador e serviço
- Origem e destino (CEP, UF, Município)
- Peso e dimensões
- Valores (declarado, frete)
- Modalidade de frete
- Documentos fiscais (NFCe, CTe)
- Status e rastreamento
- Metadados adicionais

### Cotação Nacional
Representa uma cotação de frete com:
- Operador e serviço
- Preços (frete, mínimo)
- Prazos (entrega, mínimo, máximo)
- Modalidade e tipo de serviço
- Informações de inclusão (seguro, reversa, coleta)
- Cálculo de CO2

### Evento de Rastreamento
Representa um evento no ciclo de vida da remessa:
- Timestamp e localização
- Status e descrição
- Operador responsável
- Geolocalização e assinatura

### Documento Fiscal
Representa documentos fiscais brasileiros:
- Tipo (NFCe, NFe, CTe, DACTe)
- Número, série, chave de acesso
- Valores e impostos (ICMS, IPI, PIS, COFINS)
- Emitente e destinatário
- Status e verificação

## Endpoints Principais

### Autenticação
- `POST /api/v1/auth/token` - Obter token JWT

### TR8 Nacional
- `POST /api/v1/nacional/tr8/cotacoes` - Obter cotações de frete
- `POST /api/v1/nacional/tr8/remessas` - Criar remessa
- `GET /api/v1/nacional/tr8/remessas/{id_remessa}` - Obter remessa
- `GET /api/v1/nacional/tr8/rastreamento/{codigo_rastreamento}` - Rastreamento
- `POST /api/v1/nacional/tr8/rastreamento/{codigo_rastreamento}` - Atualizar rastreamento
- `POST /api/v1/nacional/tr8/documentos/fiscais` - Criar documento fiscal
- `PUT /api/v1/nacional/tr8/documentos/fiscais/{documento_id}/validar` - Validar documento
- `GET /api/v1/nacional/tr8/documentos/fiscais/{documento_id}` - Obter documento
- `POST /api/v1/nacional/tr8/webhooks` - Criar webhook
- `GET /api/v1/nacional/tr8/webhooks` - Listar webhooks

### Observatório
- `GET /api/v1/observatorio/nacional/tr8/metricas` - Métricas nacionais
- `GET /api/v1/observatorio/nacional/tr8/conformidade` - Relatório de conformidade

## Governança de Dados

### Campos Permitidos
Todos os campos necessários para a operação logística:
- Identificação (remessa, operador, serviço)
- Localização (CEP, UF, Município)
- Dimensões (peso, comprimento, largura, altura)
- Valores (declarado, frete)
- Modalidades e tipos
- Status e timestamps

### Campos Bloqueados
Dados que não podem ser acessados por concorrentes:
- Custos internos
- Margens de lucro
- Estratégias comerciais
- Listas de clientes
- Dados de concorrentes
- Informações financeiras privadas

### Campos Pseudonimizados
Dados pessoais que são automaticamente pseudonimizados:
- Nomes (remetente, destinatário)
- Endereços
- Telefones
- E-mails
- CPF/CNPJ
- Inscrição Estadual

### Controle de Acesso (RBAC)
- Operadores: acesso a suas próprias remessas
- Marketplaces: acesso a remessas de seus vendedores
- Transportadoras: acesso a remessas sob sua responsabilidade
- Auditores: acesso a relatórios de conformidade

## Segurança e Compliance

### Autenticação
- OAuth 2.0 + JWT
- Tokens com escopos granulares
- Validade curta (15 minutos)
- Refresh tokens opcionais

### Autorização
- RBAC (Role-Based Access Control)
- Controle por operador
- Restrição por endpoint

### Criptografia
- TLS 1.3 obrigatório
- AES-256 em repouso
- Hash SHA-256 para logs

### Auditoria
- Logs imutáveis com hash encadeado
- Registro de todos os acessos
- Timestamp e ID do operador
- Conformidade com LGPD, CADE e Receita Federal

### Conformidade Regulatória
- **LGPD**: Pseudonimização automática
- **CADE**: Firewall de dados competitivos
- **Receita Federal**: Validação de documentos fiscais
- **ANTT**: Regulamentação de transporte

## Integração com Operadores

### Correios
- Serviços: Sedex, PAC, Sedex 10, Sedex Hoje
- Modalidades: CIF, FOB
- Documentos: CTe quando aplicável

### Transportadoras
- Serviços personalizados
- Modalidades flexíveis
- Documentos: CTe, NFCe

### Marketplaces
- Integração com vendedores
- Cotação unificada
- Rastreamento consolidado

### ERPs e WMS
- Automação de processos
- Sincronização de dados
- Relatórios gerenciais

## Casos de Uso

### 1. E-commerce
- Cotação de frete na finalização de pedido
- Geração de etiqueta de envio
- Rastreamento para o cliente
- Devolução automatizada

### 2. Transportadoras
- Gestão de frota
- Otimização de rotas
- Documentação fiscal
- Monitoramento de entregas

### 3. Indústria
- Logística de entrada e saída
- Gestão de estoque
- Rastreamento de matérias-primas
- Distribuição de produtos acabados

### 4. Varejo
- Transferência entre lojas
- Reposição automática
- Logística reversa
- Integração com fornecedores

## Implantação

### Pré-requisitos
- Go 1.21 ou superior
- PostgreSQL 12 ou superior
- Redis 6 ou superior (opcional)

### Configuração
1. Clonar o repositório
2. Configurar variáveis de ambiente
3. Inicializar banco de dados
4. Executar migrações
5. Iniciar a API

### Docker
```bash
docker-compose up -d
```

### Kubernetes
```yaml
# Deployment example
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rldb-api-nacional
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rldb-api-nacional
  template:
    metadata:
      labels:
        app: rldb-api-nacional
    spec:
      containers:
      - name: api
        image: rldb-br/rldb-api-universal:nacional
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: rldb-config
        - secretRef:
            name: rldb-secrets
```

## Monitoramento

### Métricas
- Taxa de requisições
- Tempo de resposta
- Erros por endpoint
- Uso de cache

### Alertas
- Falhas de autenticação
- Rate limiting excedido
- Erros de banco de dados
- Falhas de integração

### Logging
- Logs estruturados
- Níveis de severidade
- Filtros por operador
- Exportação para SIEM

## Suporte

### Documentação
- OpenAPI 3.1: `/docs/openapi_nacional.yaml`
- Swagger UI: `/docs`
- Guia de integração: `/docs/integracao.md`

### Contato
- Email: api-support@rldb.gov.br
- Site: https://rldb.gov.br
- Telefone: +55 XX XXXX-XXXX

### SLA
- Disponibilidade: 99.95%
- Tempo de resposta: < 500ms (cotacoes), < 2s (remessas)
- Suporte: 24/7 para operadores críticos

## Roadmap

### Fase 1 (0-3 meses)
- Especificação OpenAPI 3.1
- Schema de dados nacional
- Modelo de segurança

### Fase 2 (3-6 meses)
- Sandbox com Correios + 2 transportadoras
- Integração com 1 marketplace
- Piloto em São Paulo

### Fase 3 (6-12 meses)
- Produção nacional
- Integração com todos os operadores principais
- Observatório completo

### Fase 4 (12-24 meses)
- Otimização de performance
- Novos recursos (IA, ML)
- Integração internacional (crossborder)

## Contribuição

1. Fork o repositório
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## Licença

MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

## Anexos

### A. Operadores Suportados
- Correios (Sedex, PAC, etc.)
- Transportadora ABC
- Transportadora XYZ
- Loggi
- Azul Cargo
- Gol Log
- Total Express
- JadLog
- E mais...

### B. Tipos de Serviço
- Normal
- Expresso
- Econômico
- Agendado
- Fulfilment
- Reversa

### C. Modalidades de Frete
- CIF (Frete pago pelo remetente)
- FOB (Frete por conta do destinatário)
- Por Conta (Frete a ser cobrado)
- Frete Cobrar (Frete a cobrar)

### D. Tipos de Produto
- Eletroeletrônicos
- Vestuário
- Alimentos
- Livros
- Cosméticos
- Móveis
- Brinquedos
- Automóveis
- Perigosos
- Frágeis
- Perecíveis
- Grande Portes

### E. Documentos Fiscais
- NFCe (Nota Fiscal de Consumidor Eletrônica)
- NFe (Nota Fiscal Eletrônica)
- CTe (Conhecimento de Transporte Eletrônico)
- DACTe (Documento Auxiliar do CTe)
- MDFe (Manifesto Eletrônico de Documentos Fiscais)

## Conclusão

A **RLDB - API UNIVERSAL - TR8 Nacional** é a solução definitiva para os problemas da operação logística nacional brasileira. Ela fornece uma plataforma neutra, segura, escalável e conformidade regulatória, que pode ser adaptada para qualquer cenário de transporte e logística.

Com esta solução, o Brasil terá uma infraestrutura logística digital moderna, eficiente e competitiva, capaz de impulsionar o comércio eletrônico, reduzir custos e melhorar a experiência do cliente.
