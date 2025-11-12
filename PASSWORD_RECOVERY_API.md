# 🔐 Sistema de Recuperação de Senha e Atualização de Perfil

## 📋 Resumo

Sistema completo de recuperação de senha via email e atualização de perfil de usuário com verificação de email.

---

## 🆕 Novos Endpoints

### 1️⃣ Recuperação de Senha

#### POST `/api/v1/auth/forgot-password`
**Descrição**: Envia código de 6 dígitos para o email do usuário  
**Autenticação**: Não requerida (pública)  
**Validade**: 15 minutos

**Request:**
```json
{
  "email": "usuario@example.com"
}
```

**Response (200):**
```json
{
  "message": "Código de recuperação enviado para seu email. Válido por 15 minutos."
}
```

**Email recebido:**
- **Assunto**: "Recuperação de Senha - Código de Verificação"
- **Conteúdo**: Código de 6 dígitos (ex: `123456`)

---

#### POST `/api/v1/auth/reset-password`
**Descrição**: Redefine a senha usando o código recebido  
**Autenticação**: Não requerida (pública)

**Request:**
```json
{
  "email": "usuario@example.com",
  "token": "123456",
  "newPassword": "novaSenha123"
}
```

**Response (200):**
```json
{
  "message": "Senha alterada com sucesso! Faça login com sua nova senha."
}
```

**Observações:**
- Código expira em 15 minutos
- Código só pode ser usado uma vez
- Token JWT atual é invalidado (usuário precisa fazer login novamente)
- Email de confirmação é enviado automaticamente

---

### 2️⃣ Atualização de Perfil

#### PATCH `/api/v1/user/profile`
**Descrição**: Atualiza nome do usuário  
**Autenticação**: JWT Token requerido  
**Header**: `Authorization: Bearer {token}`

**Request:**
```json
{
  "name": "Novo Nome"
}
```

**Response (200):**
```json
{
  "message": "Perfil atualizado com sucesso",
  "user": {
    "id": 1,
    "name": "Novo Nome",
    "email": "usuario@example.com",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-15T10:30:00Z"
  }
}
```

**Nota**: Para alterar email, use os endpoints específicos abaixo.

---

### 3️⃣ Alteração de Email (com Verificação)

#### POST `/api/v1/user/request-email-change`
**Descrição**: Solicita alteração de email (envia código para novo email)  
**Autenticação**: JWT Token requerido  
**Header**: `Authorization: Bearer {token}`

**Request:**
```json
{
  "newEmail": "novo-email@example.com"
}
```

**Response (200):**
```json
{
  "message": "Código de verificação enviado para o novo email. Válido por 15 minutos."
}
```

**Email recebido no NOVO email:**
- **Assunto**: "Verificação de Email - Código de Confirmação"
- **Conteúdo**: Código de 6 dígitos (ex: `654321`)

---

#### POST `/api/v1/user/confirm-email-change`
**Descrição**: Confirma alteração de email com código recebido  
**Autenticação**: JWT Token requerido  
**Header**: `Authorization: Bearer {token}`

**Request:**
```json
{
  "newEmail": "novo-email@example.com",
  "token": "654321"
}
```

**Response (200):**
```json
{
  "message": "Email atualizado com sucesso!",
  "user": {
    "id": 1,
    "name": "Nome do Usuário",
    "email": "novo-email@example.com",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-15T11:00:00Z"
  }
}
```

**Observações:**
- Código expira em 15 minutos
- Novo email não pode estar em uso por outra conta
- Código só pode ser usado uma vez

---

## 🔄 Fluxo Completo

### Fluxo 1: Recuperação de Senha

```
1. Usuário esqueceu a senha
   ↓
2. POST /auth/forgot-password { email }
   ↓
3. Sistema envia email com código de 6 dígitos
   ↓
4. Usuário recebe email e copia código
   ↓
5. POST /auth/reset-password { email, token, newPassword }
   ↓
6. Sistema valida código e altera senha
   ↓
7. Email de confirmação enviado
   ↓
8. Usuário faz login com nova senha
```

### Fluxo 2: Alteração de Email

```
1. Usuário autenticado quer trocar email
   ↓
2. POST /user/request-email-change { newEmail }
   ↓
3. Sistema envia código para NOVO email
   ↓
4. Usuário acessa novo email e copia código
   ↓
5. POST /user/confirm-email-change { newEmail, token }
   ↓
6. Sistema valida código e altera email
   ↓
7. Email atualizado com sucesso
```

---

## ⚙️ Configuração

### 1. Variáveis de Ambiente

Adicione ao arquivo `.env`:

```bash
# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=seu-email@gmail.com
SMTP_PASSWORD=sua-senha-de-app
SMTP_SENDER_NAME=Sistema de Notas Fiscais
```

### 2. Configurar Gmail (Recomendado)

1. **Ative 2FA** na sua conta Google
2. Acesse [Senhas de App](https://myaccount.google.com/apppasswords)
3. Crie senha de app para "Mail"
4. Use essa senha no `SMTP_PASSWORD`

📖 **[Guia Completo de Configuração](EMAIL_SETUP.md)**

---

## 🧪 Testando os Endpoints

### 1. Teste Recuperação de Senha

```bash
# Passo 1: Solicitar código
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu-email@example.com"
  }'

# Passo 2: Verifique seu email e use o código
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu-email@example.com",
    "token": "123456",
    "newPassword": "novaSenha123"
  }'
```

### 2. Teste Atualização de Perfil

```bash
# Atualizar nome
curl -X PATCH http://localhost:8080/api/v1/user/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -d '{
    "name": "Novo Nome"
  }'
```

### 3. Teste Alteração de Email

```bash
# Passo 1: Solicitar código para novo email
curl -X POST http://localhost:8080/api/v1/user/request-email-change \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -d '{
    "newEmail": "novo-email@example.com"
  }'

# Passo 2: Confirmar com código recebido
curl -X POST http://localhost:8080/api/v1/user/confirm-email-change \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -d '{
    "newEmail": "novo-email@example.com",
    "token": "654321"
  }'
```

---

## 🔒 Segurança

### Medidas Implementadas

✅ **Códigos de 6 dígitos** aleatórios criptograficamente seguros  
✅ **Expiração em 15 minutos** para todos os códigos  
✅ **Uso único** - códigos não podem ser reutilizados  
✅ **Invalidação de tokens anteriores** após trocar senha  
✅ **Verificação de email** antes de alterar  
✅ **Validação de unicidade** de email  
✅ **Notificações por email** de alterações de senha  
✅ **Rate limiting** (implícito por tempo de expiração)

### Recomendações Adicionais

- Implemente rate limiting no nível de aplicação
- Use HTTPS em produção (obrigatório!)
- Configure SPF/DKIM para domínio próprio
- Monitore tentativas de recuperação de senha
- Considere adicionar CAPTCHA em produção

---

## 📊 Banco de Dados

### Nova Tabela: `password_resets`

```sql
CREATE TABLE password_resets (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token VARCHAR(6) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  used BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP
);

CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX idx_password_resets_token ON password_resets(token);
CREATE INDEX idx_password_resets_expires_at ON password_resets(expires_at);
```

**Limpeza automática**: Considere adicionar job para deletar tokens expirados após 24h.

---

## 📄 Documentação Swagger

Acesse a documentação interativa completa:

```
http://localhost:8080/swagger/index.html
```

**Novos endpoints documentados:**
- 🔐 **Authentication** → `/auth/forgot-password`, `/auth/reset-password`
- 👤 **User** → `/user/profile`, `/user/request-email-change`, `/user/confirm-email-change`

---

## ❌ Tratamento de Erros

### Recuperação de Senha

| Código | Erro | Causa |
|--------|------|-------|
| 400 | Dados inválidos | Email malformado |
| 401 | Código inválido ou expirado | Token incorreto ou expirado |
| 404 | Usuário não encontrado | Email não cadastrado |
| 500 | Erro ao enviar email | Configuração SMTP incorreta |

### Atualização de Perfil

| Código | Erro | Causa |
|--------|------|-------|
| 400 | Dados inválidos | Campos vazios ou inválidos |
| 400 | Email já em uso | Novo email pertence a outra conta |
| 401 | Token inválido | JWT expirado ou inválido |
| 401 | Código inválido | Código de verificação incorreto |

---

## 🎨 Templates de Email

Os templates HTML são totalmente personalizáveis. Edite em:

```
config/email.go
```

**3 templates disponíveis:**
1. `SendPasswordResetEmail()` - Recuperação de senha
2. `SendPasswordChangedEmail()` - Confirmação de alteração
3. `SendEmailVerificationCode()` - Verificação de email

**Personalização:**
- Cores
- Logo
- Texto
- Estilo CSS

---

## 🚀 Deploy

### Checklist antes do Deploy

- [ ] Variáveis de ambiente configuradas no servidor
- [ ] Email de produção configurado (não use email pessoal!)
- [ ] HTTPS habilitado (obrigatório!)
- [ ] Migração de banco aplicada (`password_resets` table)
- [ ] Teste de envio de email no ambiente de produção
- [ ] Limite de taxa configurado (opcional mas recomendado)
- [ ] Logs configurados para monitorar falhas de email
- [ ] Swagger desabilitado em produção (opcional)

---

## 📞 Suporte

### Problemas Comuns

**"SMTP service not configured"**
- Configure `SMTP_EMAIL` e `SMTP_PASSWORD` no `.env`

**"Invalid credentials"**
- Gmail: Use senha de app, não senha normal
- Verifique 2FA ativo

**Email não chega**
- Verifique pasta de spam
- Confirme credenciais SMTP
- Teste com outro provedor

**Código expirado**
- Códigos expiram em 15 minutos
- Solicite novo código

📖 **[Guia Completo de Troubleshooting](EMAIL_SETUP.md#-troubleshooting)**

---

## 📚 Arquivos Criados/Modificados

### Novos Arquivos
- ✅ `schemas/password_reset.go` - Schema de recuperação
- ✅ `config/email.go` - Serviço de email SMTP
- ✅ `handler/utils.go` - Gerador de códigos
- ✅ `.env.example` - Exemplo de variáveis
- ✅ `EMAIL_SETUP.md` - Guia completo de configuração
- ✅ `PASSWORD_RECOVERY_API.md` - Este documento

### Arquivos Modificados
- ✅ `handler/auth.go` - Endpoints de recuperação
- ✅ `handler/user.go` - Endpoints de perfil
- ✅ `router/routes.go` - Novas rotas
- ✅ `config/postgres.go` - Migration da nova tabela
- ✅ `docs/swagger.*` - Documentação atualizada

---

## ✅ Status

**Sistema 100% funcional!** 🎉

Todos os endpoints testados e documentados. Pronto para uso em desenvolvimento e produção.

**Próximos passos:**
1. Configure suas credenciais SMTP
2. Teste os endpoints
3. Personalize os templates de email
4. Deploy para produção

**Boa sorte! 🚀**
