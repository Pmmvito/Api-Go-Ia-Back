# 🔐 AUTENTICAÇÃO BEARER CONFIGURADA NO SWAGGER!

## ✅ FEITO! Botão "Authorize" está disponível no Swagger!

---

## 🎯 Como Usar (Passo a Passo Visual)

### **Passo 1: Abra o Swagger**
```
http://localhost:8080/swagger/index.html
```

### **Passo 2: Localize o Botão "Authorize"**
No topo da página, você verá:
```
┌─────────────────────────────────────────┐
│  🔒 Authorize                           │
└─────────────────────────────────────────┘
```

### **Passo 3: Faça Login**
Role até o endpoint **POST /api/v1/login**:
```
Auth ▼
  POST /api/v1/login - Login user
```

Clique em **"Try it out"** e execute com suas credenciais:
```json
{
  "email": "seu-email@exemplo.com",
  "password": "sua-senha"
}
```

### **Passo 4: Copie o Token**
Da resposta, copie o valor do campo `token`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### **Passo 5: Clique em "Authorize"** 🔑
1. Clique no botão **"Authorize"** no topo
2. Uma janela modal vai abrir

### **Passo 6: Cole o Token**
No campo **"Value"**, digite:
```
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**⚠️ IMPORTANTE:** Não esqueça da palavra `Bearer` antes do token!

### **Passo 7: Autorizar**
1. Clique no botão **"Authorize"** dentro da modal
2. Clique em **"Close"**

### **Passo 8: Pronto! 🎉**
Agora você verá que o cadeado mudou de 🔓 (aberto) para 🔒 (fechado).

Todos os endpoints protegidos agora funcionarão!

---

## 🔒 Endpoints Protegidos

Estes endpoints agora mostram o cadeado 🔒 e funcionam após autorizar:

```
Auth
  ✅ GET  /api/v1/me

Receipts  
  ✅ POST  /api/v1/scan-receipt
  ✅ GET   /api/v1/receipts
  ✅ GET   /api/v1/receipt/:id
  ✅ PATCH /api/v1/receipt/:id
  ✅ PATCH /api/v1/receipt/:id/item/:itemIndex
```

---

## 🔓 Endpoints Públicos (Sem Cadeado)

Estes funcionam sem autenticação:

```
Auth
  ⭕ POST /api/v1/register
  ⭕ POST /api/v1/login
```

---

## 📱 Visualização no Swagger

### Antes de Autorizar:
```
┌──────────────────────────────┐
│  🔓 Authorize                │  ← Cadeado ABERTO
└──────────────────────────────┘

🔒 GET /api/v1/me              ← Mostra cadeado (requer auth)
```

### Depois de Autorizar:
```
┌──────────────────────────────┐
│  🔒 Authorize    [Logout]    │  ← Cadeado FECHADO
└──────────────────────────────┘

🔒 GET /api/v1/me              ← Agora funciona!
```

---

## ❌ Problemas Comuns

### 1. Erro 401 ao testar endpoint protegido
**Causa:** Você não autorizou ou o token está errado  
**Solução:** Clique em "Authorize" e cole o token corretamente

### 2. "Invalid or expired token"
**Causa:** Token expirou (7 dias de validade)  
**Solução:** Faça login novamente e pegue novo token

### 3. Colei o token mas não funciona
**Causa:** Esqueceu de colocar "Bearer " antes do token  
**Formato correto:** `Bearer SEU_TOKEN`  
**Formato errado:** `SEU_TOKEN`

---

## 🎯 Teste Rápido

1. Acesse: http://localhost:8080/swagger/index.html
2. Execute: **POST /login** (copie o token)
3. Clique: **"Authorize"** no topo
4. Cole: `Bearer TOKEN_COPIADO`
5. Clique: **"Authorize"** na modal
6. Teste: **GET /me** (deve funcionar! ✅)

---

## 📋 Formato do Token

```
Correto:   Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
           ↑      ↑
           |      Seu token JWT
           Palavra "Bearer" + espaço

Errado:    eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
           ↑
           Faltou "Bearer "
```

---

## 🔄 Para Desautorizar (Logout)

1. Clique em **"Authorize"** no topo
2. Clique em **"Logout"**
3. O cadeado volta para 🔓 (aberto)

---

## 💾 Configuração Técnica Implementada

**No arquivo `main.go`:**
```go
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Digite "Bearer" seguido do seu token JWT
```

**Nos handlers protegidos:**
```go
// @Security BearerAuth
```

**Gerado com:**
```bash
swag init
```

---

## ✨ Informações da API

- **Título:** API de Autenticação e Scan de Notas Fiscais
- **Versão:** 1.0
- **Host:** localhost:8080
- **Base Path:** /api/v1
- **Licença:** MIT
- **Autenticação:** Bearer JWT (7 dias de validade)

---

**🎉 Pronto! Agora você pode testar toda a API diretamente no Swagger com autenticação!**

Veja também: `SWAGGER_AUTH_GUIDE.md` para guia detalhado.
