# Golang API de Gerenciamento de Notas Fiscais com IA

Uma **API REST** moderna e eficiente com sistema completo de autenticação JWT (Bearer Token), **escaneamento de notas fiscais com IA Google Gemini**, e **banco de dados normalizado** seguindo as melhores práticas acadêmicas, desenvolvida em **Go (Golang)**.

## 🚀 Sobre o Projeto

Esta## 🔒 Segurança

### Autenticação JWT
- Tokens JWT válidos por 7 dias
- Assinatura HMAC-SHA256
- Tokens enviados via header `Authorization: Bearer <token>`

### Senhas
- Hash bcrypt com salt automático
- Custo padrão do bcrypt (10 rounds)
- Senha nunca retornada nas respostas da API

### Variáveis de Ambiente
- `JWT_SECRET`: Chave secreta para assinatura de tokens (MUDE EM PRODUÇÃO!)
- `DATABASE_DSN`: String de conexão do PostgreSQL
- `GEMINI_API_KEY`: Chave API do Google Gemini (obtenha gratuitamente)

### Proteção de Dados
- ✅ Imagens de notas armazenadas apenas como base64
- ✅ Relacionamento User → Receipt garante isolamento de dados
- ✅ Todas as queries verificam ownership via JWT

## 🤖 Google Gemini AI

### Como Funciona
1. Usuário envia imagem da nota fiscal (base64)
2. API envia para Google Gemini com prompt estruturado
3. IA retorna JSON com:
   - Nome do estabelecimento
   - Data da compra
   - Lista de items com descrição, quantidade, preço
   - **Categoria sugerida para cada item**
   - Subtotal, descontos, total
4. API salva tudo no banco de dados normalizado

### Prompt da IA
```text
Você é um assistente especializado em analisar notas fiscais brasileiras...
Para cada item, você DEVE categorizar usando uma dessas opções:
- Alimentação 🍽️
- Bebidas 🥤
- Frutas 🍎
... (15 categorias)
```

### Obter API Key Gratuita
1. Acesse [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Faça login com sua conta Google
3. Clique em "Create API Key"
4. Copie a chave e adicione no `.env`

## 🎓 Adequado para TCC

Esta API foi desenvolvida seguindo as melhores práticas acadêmicas:

### ✅ Banco de Dados Normalizado
- Terceira Forma Normal (3NF)
- Relacionamentos com Foreign Keys
- Diagramas ER completos
- Queries otimizadas com INNER JOINs

### ✅ Documentação Completa
- README detalhado
- Swagger para API
- Diagrama de relacionamentos
- Changelog de mudanças

### ✅ Arquitetura Limpa
- Separação de responsabilidades (MVC-like)
- Handlers, Schemas, Config separados
- Middleware de autenticação
- Validações de input

### ✅ Tecnologias Modernas
- Go (linguagem compilada, performática)
- PostgreSQL (banco robusto)
- Google Gemini AI (IA de última geração)
- JWT (padrão de mercado)

Veja mais detalhes em:
- [`DATABASE_STRUCTURE.md`](DATABASE_STRUCTURE.md) - Estrutura completa do banco
- [`CHANGELOG_NORMALIZATION.md`](CHANGELOG_NORMALIZATION.md) - Processo de normalização sistema robusto com:

- ✅ **Registro de usuários** com hash de senha (bcrypt)
- ✅ **Login com JWT** (JSON Web Tokens)
- ✅ **Sessão Bearer Token** para rotas protegidas
- ✅ **Escaneamento de notas fiscais** com Google Gemini AI
- ✅ **Categorização automática** de produtos com IA
- ✅ **Sistema completo de categorias** (CRUD)
- ✅ **Banco de dados normalizado** (3NF) com relacionamentos adequados
- ✅ **Filtros avançados** por categoria e período
- ✅ **15 categorias padrão** pré-configuradas

## 🎯 Funcionalidades Principais

### 📸 Escaneamento de Notas Fiscais
- Envie uma imagem de nota fiscal (base64)
- IA extrai: estabelecimento, data, items, preços, totais
- **Categorização automática** de cada item

### 🏷️ Sistema de Categorias
- 15 categorias padrão com emojis e cores
- CRUD completo (Create, Read, Update, Delete)
- Relacionamento com items via Foreign Key

### 📊 Banco de Dados Normalizado
- **Users** → **Receipts** → **ReceiptItems** → **Categories**
- Foreign Keys com CASCADE delete
- Queries otimizadas com INNER JOINs
- Preparado para aprovação em TCC

### 🔍 Filtros e Buscas
- Listar todos os items
- Filtrar por categoria
- Filtrar por período (data início/fim)
- Agrupar por categoria com estatísticas

## 🛠️ Tecnologias Utilizadas

- **[Go](https://golang.org/)** - Linguagem de programação
- **[Gin](https://gin-gonic.com/)** - Framework web HTTP
- **[GORM](https://gorm.io/)** - ORM (Object Relational Mapping)
- **[PostgreSQL](https://www.postgresql.org/)** - Banco de dados relacional
- **[JWT](https://github.com/golang-jwt/jwt)** - JSON Web Tokens para autenticação
- **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** - Hash de senhas
- **[Google Gemini AI](https://ai.google.dev/)** - Inteligência Artificial para OCR e categorização
- **[Swaggo](https://github.com/swaggo/swag)** - Geração automática de documentação Swagger
- **[godotenv](https://github.com/joho/godotenv)** - Gerenciamento de variáveis de ambiente

## 📋 Pré-requisitos

- Go 1.21 ou superior
- PostgreSQL 12 ou superior
- Google Gemini API Key (gratuita em [Google AI Studio](https://makersuite.google.com/app/apikey))
- Make (opcional, para usar os comandos do Makefile)

## 🚀 Como Executar

### 1. Clone o repositório
```bash
git clone https://github.com/Pmmvito/Golang-Api-Exemple.git
cd Golang-Api-Exemple
```

### 2. Configure as variáveis de ambiente
Copie o arquivo `.env.example` para `.env` e configure suas credenciais:
```bash
cp .env.example .env
```

Edite o arquivo `.env`:
```env
DATABASE_DSN=postgresql://usuario:senha@localhost:5432/nome_do_banco?sslmode=disable
JWT_SECRET=sua-chave-secreta-super-segura-mude-isso-em-producao
GEMINI_API_KEY=sua-chave-api-do-google-gemini
```

### 3. Instale as dependências
```bash
go mod tidy
```

### 4. Execute a aplicação
```bash
go run main.go
```

A API estará disponível em: `http://localhost:8080`

## 📚 Documentação da API

### Swagger UI
A documentação interativa da API está disponível em:
```
http://localhost:8080/swagger/index.html
```

### Endpoints Principais

#### 🔓 Rotas Públicas (Sem Autenticação)

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `POST` | `/api/v1/register` | Registrar novo usuário |
| `POST` | `/api/v1/login` | Login (retorna JWT token) |

#### 🔒 Rotas Protegidas (Requerem Bearer Token)

**Usuário:**
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/api/v1/me` | Obter dados do usuário autenticado |

**Notas Fiscais:**
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `POST` | `/api/v1/scan-receipt` | Escanear nota fiscal com IA |
| `GET` | `/api/v1/receipts` | Listar todos os recibos |
| `GET` | `/api/v1/receipt/:id` | Obter recibo específico |
| `PATCH` | `/api/v1/receipt/:id` | Atualizar recibo |
| `DELETE` | `/api/v1/receipt/:id` | Deletar recibo |

**Items:**
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/api/v1/items` | Listar todos os items |
| `GET` | `/api/v1/items/filter` | Filtrar items (categoria, data) |
| `GET` | `/api/v1/receipt/:id/item/:itemId` | Obter item específico |
| `PATCH` | `/api/v1/receipt/:id/item/:itemId` | Atualizar item |

**Categorias:**
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `POST` | `/api/v1/category` | Criar categoria |
| `GET` | `/api/v1/categories` | Listar todas as categorias |
| `GET` | `/api/v1/category/:id` | Obter categoria específica |
| `PATCH` | `/api/v1/category/:id` | Atualizar categoria |
| `DELETE` | `/api/v1/category/:id` | Deletar categoria |

### Exemplo de Uso

#### 1. Registrar novo usuário
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "João Silva",
    "email": "usuario@exemplo.com",
    "password": "senha123"
  }'
```

#### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@exemplo.com",
    "password": "senha123"
  }'
```

#### 3. Escanear Nota Fiscal
```bash
curl -X POST http://localhost:8080/api/v1/scan-receipt \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "imageBase64": "data:image/jpeg;base64,/9j/4AAQSkZJRg...",
    "currency": "BRL",
    "locale": "pt-BR"
  }'
```

**Resposta:**
```json
{
  "message": "Receipt scanned successfully",
  "data": {
    "id": 1,
    "storeName": "Supermercado Exemplo",
    "date": "2024-01-15",
    "items": [
      {
        "id": 1,
        "description": "Arroz 5kg",
        "quantity": 1,
        "unit": "un",
        "unitPrice": 25.90,
        "total": 25.90,
        "category": {
          "id": 1,
          "name": "Alimentação",
          "icon": "🍽️",
          "color": "#FF6B6B"
        }
      }
    ],
    "subtotal": 150.50,
    "discount": 5.00,
    "total": 145.50,
    "currency": "BRL",
    "confidence": 0.95
  }
}
```

#### 4. Listar Categorias
```bash
curl -X GET http://localhost:8080/api/v1/categories \
  -H "Authorization: Bearer SEU_TOKEN_JWT"
```

**Resposta:**
```json
{
  "message": "Categories retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Alimentação",
      "description": "Alimentos em geral",
      "icon": "🍽️",
      "color": "#FF6B6B"
    },
    {
      "id": 2,
      "name": "Bebidas",
      "description": "Bebidas alcoólicas e não alcoólicas",
      "icon": "🥤",
      "color": "#4ECDC4"
    }
  ]
}
```

#### 5. Filtrar Items por Categoria
```bash
curl -X GET "http://localhost:8080/api/v1/items/filter?category=Alimentação&startDate=2024-01-01&endDate=2024-01-31" \
  -H "Authorization: Bearer SEU_TOKEN_JWT"
```

**Resposta:**
```json
{
  "message": "Items retrieved successfully",
  "totalItems": 25,
  "filters": {
    "category": "Alimentação",
    "startDate": "2024-01-01",
    "endDate": "2024-01-31"
  },
  "groupedByCategory": [
    {
      "categoryName": "Alimentação",
      "items": [...],
      "totalAmount": 450.75,
      "itemCount": 25
    }
  ]
}
```

## 🏗️ Estrutura do Projeto

```
├── config/             # Configurações (DB, Logger)
│   ├── config.go       # Inicialização geral
│   ├── logger.go       # Sistema de logging
│   ├── postgres.go     # Conexão PostgreSQL + Migrations
│   └── sqlite.go       # Conexão SQLite (backup)
├── handler/            # Handlers HTTP (Controllers)
│   ├── handler.go      # Inicializador de handlers
│   ├── auth.go         # Handlers de autenticação
│   ├── scanReceipt.go  # Handlers de notas fiscais e items
│   ├── category.go     # Handlers de categorias (CRUD)
│   ├── request.go      # Validação de requests
│   └── response.go     # Respostas padronizadas
├── router/             # Configuração de rotas
│   ├── router.go       # Inicialização do Gin
│   ├── routes.go       # Definição de rotas (16 endpoints)
│   └── middleware.go   # Middleware de autenticação JWT
├── schemas/            # Modelos de dados (GORM)
│   ├── user.go         # User (1) → Receipts (N)
│   ├── receipt.go      # Receipt (1) → ReceiptItems (N)
│   ├── category.go     # Category (1) ← ReceiptItems (N)
│   └── opening.go      # Modelo de vagas (legado)
├── docs/               # Documentação Swagger gerada
├── main.go             # Ponto de entrada da aplicação
├── .env                # Variáveis de ambiente (não commitado)
├── .env.example        # Exemplo de variáveis de ambiente
├── DATABASE_STRUCTURE.md # Documentação completa do banco de dados
├── CHANGELOG_NORMALIZATION.md # Histórico de mudanças
└── README.md           # Documentação do projeto
```

## 🗄️ Banco de Dados

### Relacionamentos (ER Diagram)
```
Users (1) ──────→ Receipts (N) ──────→ ReceiptItems (N) ──────→ Categories (M)
   │                  │                      │                       │
   └─ FK: user_id     └─ FK: receipt_id     └─ FK: category_id      └─ PK: id
```

### Características
- ✅ **Normalização 3NF**: Sem redundância de dados
- ✅ **Foreign Keys**: Integridade referencial garantida
- ✅ **CASCADE Deletes**: Apagar usuário remove tudo automaticamente
- ✅ **Indexes**: Performance otimizada em todas as FKs
- ✅ **Transações ACID**: Garantia de consistência

Veja documentação completa em [`DATABASE_STRUCTURE.md`](DATABASE_STRUCTURE.md)

## 🏷️ Categorias Padrão

15 categorias pré-configuradas com emojis e cores:

| Categoria | Emoji | Cor |
|-----------|-------|-----|
| Alimentação | 🍽️ | #FF6B6B |
| Bebidas | 🥤 | #4ECDC4 |
| Frutas | 🍎 | #95E1D3 |
| Verduras e Legumes | 🥬 | #38A169 |
| Carnes e Peixes | 🥩 | #E53E3E |
| Laticínios | 🥛 | #F6E05E |
| Padaria | 🍞 | #D69E2E |
| Limpeza | 🧹 | #3182CE |
| Higiene Pessoal | 🧴 | #805AD5 |
| Bebê | 👶 | #FBB6CE |
| Pet | 🐾 | #F6AD55 |
| Congelados | ❄️ | #63B3ED |
| Snacks | 🍪 | #FC8181 |
| Temperos | 🧂 | #68D391 |
| Outros | 📦 | #A0AEC0 |

## � Segurança

### Autenticação JWT
- Tokens JWT válidos por 7 dias
- Assinatura HMAC-SHA256
- Tokens enviados via header `Authorization: Bearer <token>`

### Senhas
- Hash bcrypt com salt automático
- Custo padrão do bcrypt (10 rounds)
- Senha nunca retornada nas respostas da API

### Variáveis de Ambiente
- `JWT_SECRET`: Chave secreta para assinatura de tokens (MUDE EM PRODUÇÃO!)
- `DATABASE_DSN`: String de conexão do PostgreSQL

## �📖 Gerando Documentação Swagger

Este projeto utiliza **Swaggo** para gerar automaticamente a documentação Swagger a partir de comentários no código.

### Instalação do Swag CLI
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Gerar documentação
```bash
swag init
```

## 🧪 Testando a API

### Usando cURL
Veja os exemplos na seção "Exemplo de Uso" acima.

### Usando Postman/Insomnia
1. Importe a coleção do Swagger: `http://localhost:8080/swagger/doc.json`
2. Configure o token Bearer nas rotas protegidas

## 🤝 Contribuindo

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## 👨‍💻 Autor

**Vitor Benevento** - [GitHub](https://github.com/Pmmvito)

---