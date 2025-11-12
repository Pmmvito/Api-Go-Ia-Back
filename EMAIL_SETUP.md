# 📧 Configuração de Email - Sistema de Recuperação de Senha

## 📋 Visão Geral

O sistema utiliza SMTP para envio de emails de recuperação de senha e verificação de email. Suporta os principais provedores:
- **Gmail** (recomendado)
- **Outlook/Hotmail**
- **Yahoo Mail**
- Qualquer provedor SMTP

---

## 🔧 Configuração por Provedor

### 1️⃣ Gmail (Recomendado)

#### Passo 1: Ativar 2FA
1. Acesse [Configurações do Google](https://myaccount.google.com/security)
2. Ative a **Verificação em duas etapas**

#### Passo 2: Criar Senha de App
1. Acesse [Senhas de App](https://myaccount.google.com/apppasswords)
2. Selecione **Email** como app
3. Selecione **Outro (nome personalizado)** como dispositivo
4. Digite "API Sistema Notas Fiscais"
5. Clique em **Gerar**
6. **Copie a senha gerada** (16 caracteres sem espaços)

#### Passo 3: Configurar Variáveis de Ambiente
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=seu-email@gmail.com
SMTP_PASSWORD=xxxx xxxx xxxx xxxx  # Senha de app gerada
SMTP_SENDER_NAME=Sistema de Notas Fiscais
```

#### ⚠️ Problemas Comuns Gmail
- **"Invalid credentials"**: Use senha de app, não sua senha normal
- **"Less secure app access"**: Não é mais necessário com senha de app
- **"SMTP AUTH disabled"**: Certifique-se que 2FA está ativo

---

### 2️⃣ Outlook/Hotmail

#### Configuração
```bash
SMTP_HOST=smtp-mail.outlook.com
SMTP_PORT=587
SMTP_EMAIL=seu-email@outlook.com
SMTP_PASSWORD=sua-senha-normal  # Use sua senha normal
SMTP_SENDER_NAME=Sistema de Notas Fiscais
```

#### ⚠️ Notas Outlook
- Use sua senha normal (não precisa de senha de app)
- Se tiver 2FA, pode precisar de senha de app
- Verifique se SMTP está habilitado nas configurações

---

### 3️⃣ Yahoo Mail

#### Configuração
```bash
SMTP_HOST=smtp.mail.yahoo.com
SMTP_PORT=587
SMTP_EMAIL=seu-email@yahoo.com
SMTP_PASSWORD=sua-senha-de-app  # Senha de app necessária
SMTP_SENDER_NAME=Sistema de Notas Fiscais
```

#### Passo para Senha de App Yahoo
1. Acesse [Segurança da Conta Yahoo](https://login.yahoo.com/account/security)
2. Ative **Verificação em duas etapas**
3. Clique em **Gerar senha de app**
4. Selecione "Outro app" e dê um nome
5. Use a senha gerada

---

### 4️⃣ Provedor Personalizado

Para qualquer outro provedor SMTP:

```bash
SMTP_HOST=smtp.seu-provedor.com
SMTP_PORT=587  # ou 465 para SSL
SMTP_EMAIL=seu-email@dominio.com
SMTP_PASSWORD=sua-senha
SMTP_SENDER_NAME=Seu Nome ou Sistema
```

#### Consulte a documentação do seu provedor:
- **Porta 587**: TLS (StartTLS) - **Recomendado**
- **Porta 465**: SSL direto
- **Porta 25**: Não seguro (evitar)

---

## 🧪 Testando a Configuração

### 1. Configurar Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```bash
# Copie .env.example para .env
cp .env.example .env

# Edite o arquivo .env com suas credenciais
nano .env  # ou vim, code, notepad, etc.
```

### 2. Testar Envio de Email

Inicie o servidor e teste o endpoint:

```bash
# Inicie o servidor
go run main.go

# Em outro terminal, teste o endpoint
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu-email-teste@example.com"
  }'
```

### 3. Verificar Logs

Verifique os logs do servidor para mensagens de erro:

```
✅ Email enviado com sucesso
❌ Error sending email: ...
```

---

## 📨 Templates de Email

O sistema envia 3 tipos de email:

### 1. Recuperação de Senha
- **Assunto**: "Recuperação de Senha - Código de Verificação"
- **Conteúdo**: Código de 6 dígitos
- **Validade**: 15 minutos
- **Arquivo**: `config/email.go` → `SendPasswordResetEmail()`

### 2. Confirmação de Alteração de Senha
- **Assunto**: "Senha Alterada com Sucesso"
- **Conteúdo**: Notificação de segurança
- **Arquivo**: `config/email.go` → `SendPasswordChangedEmail()`

### 3. Verificação de Email
- **Assunto**: "Verificação de Email - Código de Confirmação"
- **Conteúdo**: Código de 6 dígitos para trocar email
- **Validade**: 15 minutos
- **Arquivo**: `config/email.go` → `SendEmailVerificationCode()`

---

## 🔒 Segurança

### Boas Práticas

1. **Nunca commitar credenciais**
   ```bash
   # .gitignore já contém
   .env
   ```

2. **Use variáveis de ambiente em produção**
   - No Heroku: `heroku config:set SMTP_EMAIL=...`
   - No Railway: Configure nas variáveis do projeto
   - No Docker: Use secrets ou env files

3. **Proteja senhas de app**
   - Trate como senhas normais
   - Revogue se comprometida
   - Use diferentes senhas de app para diferentes apps

4. **Monitore uso de email**
   - Gmail: 500 emails/dia (limite gratuito)
   - Outlook: 300 emails/dia
   - Considere serviço dedicado em produção

---

## 🚀 Produção

### Recomendações para Produção

Para ambientes de produção, considere usar serviços especializados:

1. **SendGrid** (Recomendado)
   - 100 emails/dia grátis
   - API simples
   - [Documentação](https://sendgrid.com/docs/)

2. **Mailgun**
   - 5.000 emails/mês grátis
   - Ótima deliverability
   - [Documentação](https://www.mailgun.com/)

3. **Amazon SES**
   - Muito barato
   - Escalável
   - [Documentação](https://aws.amazon.com/ses/)

### Exemplo com SendGrid

```go
// Modifique config/email.go para usar SendGrid API
// ao invés de SMTP direto
```

---

## 🐛 Troubleshooting

### Erro: "dial tcp: lookup smtp.gmail.com: no such host"
**Causa**: Sem conexão com internet ou DNS incorreto  
**Solução**: Verifique conexão e DNS

### Erro: "535-5.7.8 Username and Password not accepted"
**Causa**: Credenciais inválidas  
**Solução**: 
- Gmail: Use senha de app, não senha normal
- Verifique se email está correto

### Erro: "454 4.7.0 Too many login attempts"
**Causa**: Muitas tentativas de login falhas  
**Solução**: Aguarde 15 minutos e tente novamente

### Erro: "SMTP service not configured"
**Causa**: Variáveis de ambiente não configuradas  
**Solução**: Configure `SMTP_EMAIL` e `SMTP_PASSWORD` no `.env`

### Email não chega na caixa de entrada
**Causas possíveis**:
1. Caiu na pasta de spam (verifique!)
2. Email bloqueado pelo provedor
3. Configuração incorreta do remetente

**Soluções**:
- Adicione o email remetente aos contatos
- Configure SPF/DKIM em produção (requer domínio próprio)
- Use serviço dedicado em produção

---

## 📊 Limites por Provedor

| Provedor | Emails/Dia (Grátis) | Emails/Hora | Requer Senha App |
|----------|---------------------|-------------|-----------------|
| Gmail    | 500                 | ~50         | ✅ Sim          |
| Outlook  | 300                 | ~30         | ⚠️ Se tiver 2FA |
| Yahoo    | 500                 | ~50         | ✅ Sim          |
| SendGrid | 100                 | 100         | ❌ Usa API      |
| Mailgun  | 5.000/mês           | Sem limite  | ❌ Usa API      |

---

## 📝 Checklist de Configuração

- [ ] Variáveis de ambiente configuradas no `.env`
- [ ] Senha de app criada (Gmail/Yahoo)
- [ ] 2FA ativado na conta de email
- [ ] Teste de envio realizado com sucesso
- [ ] Email recebido e código funciona
- [ ] `.env` adicionado ao `.gitignore`
- [ ] Documentação do provedor consultada
- [ ] Limites de envio conhecidos

---

## 🎯 Próximos Passos

Após configurar o email:

1. ✅ Teste recuperação de senha: `POST /auth/forgot-password`
2. ✅ Teste reset de senha: `POST /auth/reset-password`
3. ✅ Teste alteração de email: `POST /user/request-email-change`
4. ✅ Teste confirmação de email: `POST /user/confirm-email-change`
5. 📄 Consulte a documentação Swagger em `/swagger/index.html`

---

## 💡 Dicas

- **Desenvolvimento**: Use Gmail com senha de app (fácil e rápido)
- **Produção**: Migre para SendGrid ou Mailgun (mais confiável)
- **Templates**: Personalize os HTMLs em `config/email.go`
- **Logs**: Monitore logs para problemas de entrega
- **Backup**: Tenha um email secundário configurado

---

## 📞 Suporte

Se encontrar problemas:

1. Verifique logs do servidor
2. Consulte documentação do provedor
3. Teste credenciais manualmente
4. Verifique firewall/antivírus
5. Abra issue no GitHub com logs (sem expor credenciais!)

---

**Configuração completa! 🎉**  
Seu sistema de recuperação de senha está pronto para uso.
