# Golang API de Gerenciamento de Notas Fiscais com IA

Uma **API REST** moderna e eficiente com sistema completo de autenticação JWT (Bearer Token), **escaneamento de notas fiscais com IA Google Gemini**, e **banco de dados normalizado** seguindo as melhores práticas acadêmicas, desenvolvida em **Go (Golang)**.

## 🚀 Sobre o Projeto

Este sistema robusto oferece:

- ✅ **Registro de usuários** com hash de senha (bcrypt)
- ✅ **Login com JWT** (JSON Web Tokens)
- ✅ **Sessão Bearer Token** para rotas protegidas
- ✅ **Escaneamento de notas fiscais** com IA Google Gemini
- ✅ **Categorização automática** de produtos com IA
- ✅ **Sistema completo de categorias** (CRUD)
- ✅ **Banco de dados normalizado** (3NF) com relacionamentos adequados
- ✅ **Filtros avançados** por categoria e período
- ✅ **Categorias padrão** pré-configuradas

## 🎯 Funcionalidades Principais

### 📸 Escaneamento de Notas Fiscais
- Envie uma imagem de nota fiscal (base64)
- A IA extrai: estabelecimento, data, itens, preços, totais
- **Categorização automática** de cada item

### 🏷️ Sistema de Categorias
- Categorias padrão com emojis e cores
- CRUD completo (Create, Read, Update, Delete)
- Relacionamento com itens via Foreign Key

### 📊 Banco de Dados Normalizado
- **Users** → **Receipts** → **ReceiptItems** → **Categories**
- Foreign Keys com `CASCADE DELETE`
- Queries otimizadas com `INNER JOINs`

### 🔍 Filtros e Buscas
- Listar todos os itens
- Filtrar por categoria
- Filtrar por período (data de início/fim)

## 🛠️ Tecnologias Utilizadas

- **[Go](https://golang.org/)** - Linguagem de programação
- **[Gin](https://gin-gonic.com/)** - Framework web HTTP
- **[GORM](https://gorm.io/)** - ORM (Object Relational Mapping)
- **[PostgreSQL](https://www.postgresql.org/)** - Banco de dados relacional
- **[JWT](https://github.com/golang-jwt/jwt)** - JSON Web Tokens para autenticação
- **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** - Hash de senhas
- **[Google Gemini AI](https://ai.google.dev/)** - IA para OCR e categorização
- **[Swaggo](https://github.com/swaggo/swag)** - Geração automática de documentação Swagger
- **[godotenv](https://github.com/joho/godotenv)** - Gerenciamento de variáveis de ambiente

## 📋 Pré-requisitos

- Go 1.21 ou superior
- PostgreSQL 12 ou superior
- Chave de API do Google Gemini (gratuita no [Google AI Studio](https://makersuite.google.com/app/apikey))

## 🚀 Como Executar

### 1. Clone o repositório
```bash
git clone https://github.com/Pmmvito/Golang-Api-Exemple.git
cd Golang-Api-Exemple
```

### 2. Configure as variáveis de ambiente
Copie o arquivo `.env.example` para `.env` e preencha com suas credenciais:
```bash
cp .env.example .env
```
Edite o arquivo `.env`:
```env
DATABASE_DSN=postgresql://usuario:senha@localhost:5432/nome_do_banco?sslmode=disable
JWT_SECRET=sua-chave-secreta-super-segura
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
`http://localhost:8080/swagger/index.html`

### Endpoints Principais

#### 🔓 Rotas Públicas
| Método | Endpoint | Descrição |
|---|---|---|
| `POST` | `/api/v1/register` | Registrar novo usuário |
| `POST` | `/api/v1/login` | Login (retorna token JWT) |

#### 🔒 Rotas Protegidas (Requerem Bearer Token)

**Usuário:**
| Método | Endpoint | Descrição |
|---|---|---|
| `GET` | `/api/v1/me` | Obter dados do usuário autenticado |

**Notas Fiscais:**
| Método | Endpoint | Descrição |
|---|---|---|
| `POST` | `/api/v1/scan-qrcode/preview` | Preview de nota fiscal via QR Code |
| `POST` | `/api/v1/scan-qrcode/confirm` | Confirma e salva a nota fiscal |
| `GET` | `/api/v1/receipts` | Listar todos os recibos |
| `GET` | `/api/v1/receipt/:id` | Obter recibo específico |

**Itens:**
| Método | Endpoint | Descrição |
|---|---|---|
| `GET` | `/api/v1/items` | Listar todos os itens |
| `GET` | `/api/v1/item/:id` | Obter item específico |

**Categorias:**
| Método | Endpoint | Descrição |
|---|---|---|
| `POST` | `/api/v1/category` | Criar categoria |
| `GET` | `/api/v1/categories` | Listar todas as categorias |
| `GET` | `/api/v1/category/:id` | Obter categoria específica |
| `PATCH` | `/api/v1/category/:id` | Atualizar categoria |
| `DELETE` | `/api/v1/category/:id` | Deletar categoria |

## 🏗️ Estrutura do Projeto
```
├── config/       # Configurações (DB, Logger)
├── docs/         # Documentação Swagger gerada
├── handler/      # Handlers HTTP (Controllers)
├── router/       # Configuração de rotas
├── schemas/      # Modelos de dados (GORM)
├── main.go       # Ponto de entrada da aplicação
├── .env.example  # Exemplo de variáveis de ambiente
└── README.md     # Documentação do projeto
```

## 🗄️ Banco de Dados

### Relacionamentos
```
Users (1) -> Receipts (N) -> ReceiptItems (N) -> Categories (M)
```
- **Normalização 3NF**: Sem redundância de dados.
- **Foreign Keys**: Integridade referencial garantida.
- **CASCADE Deletes**: Apagar um usuário remove seus dados relacionados.

## 🤝 Contribuindo
1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📝 Licença
Este projeto está sob a licença MIT.

## 👨‍💻 Autor
**Vitor Benevento** - [GitHub](https://github.com/Pmmvito)
