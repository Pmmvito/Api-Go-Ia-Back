package config

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

// EmailService gerencia o envio de emails
type EmailService struct {
	SMTPHost    string
	SMTPPort    string
	SenderEmail string
	SenderName  string
	Password    string
	auth        smtp.Auth
}

// NewEmailService cria uma nova instância do serviço de email
func NewEmailService() *EmailService {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	senderEmail := os.Getenv("SMTP_EMAIL")
	senderName := os.Getenv("SMTP_SENDER_NAME")
	password := os.Getenv("SMTP_PASSWORD")

	// Log das configurações (sem mostrar senha completa)
	logger.InfoF("📧 Inicializando EmailService...")
	logger.InfoF("📧 SMTP_HOST: '%s'", smtpHost)
	logger.InfoF("📧 SMTP_PORT: '%s'", smtpPort)
	logger.InfoF("📧 SMTP_EMAIL: '%s'", senderEmail)
	logger.InfoF("📧 SMTP_SENDER_NAME: '%s'", senderName)
	logger.InfoF("📧 SMTP_PASSWORD configurado: %t (tamanho: %d)", password != "", len(password))

	if smtpHost == "" {
		smtpHost = "smtp.gmail.com" // Default para Gmail
		logger.InfoF("📧 Usando SMTP_HOST padrão: %s", smtpHost)
	}
	if smtpPort == "" {
		smtpPort = "587" // Default porta TLS
		logger.InfoF("📧 Usando SMTP_PORT padrão: %s", smtpPort)
	}
	if senderName == "" {
		senderName = "Sistema de Notas Fiscais"
		logger.InfoF("📧 Usando SMTP_SENDER_NAME padrão: %s", senderName)
	}

	if senderEmail == "" || password == "" {
		logger.ErrorF("❌ EmailService NÃO configurado! SMTP_EMAIL ou SMTP_PASSWORD ausentes")
	} else {
		logger.InfoF("✅ EmailService configurado com sucesso")
	}

	auth := smtp.PlainAuth("", senderEmail, password, smtpHost)

	return &EmailService{
		SMTPHost:    smtpHost,
		SMTPPort:    smtpPort,
		SenderEmail: senderEmail,
		SenderName:  senderName,
		Password:    password,
		auth:        auth,
	}
}

// SendPasswordResetEmail envia email com código de recuperação
func (e *EmailService) SendPasswordResetEmail(toEmail, userName, resetCode string) error {
	subject := "Recuperação de Senha - Código de Verificação"

	// Template HTML do email
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #2d3748;
            max-width: 600px;
            margin: 0 auto;
            padding: 0;
            background-color: #f7fafc;
        }
        .container {
            background-color: #ffffff;
            margin: 20px;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
            background-color: #ffffff;
        }
        .greeting {
            font-size: 18px;
            margin-bottom: 20px;
            color: #1a202c;
        }
        .code-box {
            background: linear-gradient(135deg, #f6f8fb 0%, #edf2f7 100%);
            border: 2px solid #667eea;
            padding: 30px;
            text-align: center;
            margin: 30px 0;
            border-radius: 10px;
        }
        .code-label {
            font-size: 14px;
            color: #718096;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }
        .code {
            font-size: 42px;
            font-weight: 700;
            color: #667eea;
            letter-spacing: 8px;
            font-family: 'Courier New', monospace;
        }
        .info-box {
            background-color: #fef5e7;
            border-left: 4px solid #f39c12;
            padding: 15px 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-box p {
            margin: 0;
            color: #856404;
            font-size: 14px;
        }
        .warning-box {
            background-color: #fee;
            border-left: 4px solid #e53e3e;
            padding: 15px 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .warning-box p {
            margin: 0;
            color: #742a2a;
            font-weight: 600;
            font-size: 14px;
        }
        .footer {
            background-color: #edf2f7;
            padding: 25px 30px;
            text-align: center;
            font-size: 13px;
            color: #718096;
        }
        .footer p {
            margin: 5px 0;
        }
        strong {
            color: #667eea;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Recuperação de Senha</h1>
        </div>
        <div class="content">
            <p class="greeting">Olá, <strong>{{.UserName}}</strong>!</p>
            
            <p>Recebemos uma solicitação para redefinir a senha da sua conta no Sistema de Notas Fiscais.</p>
            
            <div class="code-box">
                <div class="code-label">Seu Código de Verificação</div>
                <div class="code">{{.ResetCode}}</div>
            </div>
            
            <div class="info-box">
                <p><strong>Atenção:</strong> Este código expira em 15 minutos e só pode ser usado uma vez.</p>
            </div>
            
            <p style="margin-top: 25px;">Se você não solicitou a recuperação de senha, pode ignorar este email com segurança. Sua senha permanecerá inalterada.</p>
            
            <div class="warning-box">
                <p>IMPORTANTE: Nunca compartilhe este código com ninguém, nem mesmo com nossa equipe de suporte.</p>
            </div>
        </div>
        <div class="footer">
            <p>Este é um email automático, por favor não responda.</p>
            <p>&copy; 2025 Sistema de Notas Fiscais. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>
`

	// Processa o template
	tmpl, err := template.New("passwordReset").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("erro ao processar template: %v", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]string{
		"UserName":  userName,
		"ResetCode": resetCode,
	})
	if err != nil {
		return fmt.Errorf("erro ao executar template: %v", err)
	}

	return e.sendEmail(toEmail, subject, body.String())
}

// SendPasswordChangedEmail notifica o usuário sobre mudança de senha
func (e *EmailService) SendPasswordChangedEmail(toEmail, userName string) error {
	subject := "Senha Alterada com Sucesso"

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #2d3748;
            max-width: 600px;
            margin: 0 auto;
            padding: 0;
            background-color: #f7fafc;
        }
        .container {
            background-color: #ffffff;
            margin: 20px;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
            background-color: #ffffff;
        }
        .greeting {
            font-size: 18px;
            margin-bottom: 20px;
            color: #1a202c;
        }
        .success-icon {
            width: 80px;
            height: 80px;
            margin: 20px auto;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .success-icon svg {
            width: 50px;
            height: 50px;
            fill: white;
        }
        .success-box {
            background: linear-gradient(135deg, #f6f8fb 0%, #edf2f7 100%);
            border: 2px solid #667eea;
            padding: 25px;
            text-align: center;
            margin: 25px 0;
            border-radius: 10px;
        }
        .success-box h2 {
            color: #667eea;
            margin: 0 0 10px 0;
            font-size: 24px;
        }
        .success-box p {
            color: #4a5568;
            margin: 0;
            font-size: 16px;
        }
        .info-box {
            background-color: #fff5f5;
            border-left: 4px solid #e53e3e;
            padding: 15px 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-box p {
            margin: 0;
            color: #742a2a;
            font-size: 14px;
        }
        .info-box strong {
            color: #e53e3e;
        }
        .footer {
            background-color: #edf2f7;
            padding: 25px 30px;
            text-align: center;
            font-size: 13px;
            color: #718096;
        }
        .footer p {
            margin: 5px 0;
        }
        strong {
            color: #667eea;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Senha Alterada com Sucesso</h1>
        </div>
        <div class="content">
            <div class="success-icon">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                    <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                </svg>
            </div>
            
            <p class="greeting">Olá, <strong>{{.UserName}}</strong>!</p>
            
            <div class="success-box">
                <h2>Tudo pronto!</h2>
                <p>Sua senha foi alterada com sucesso.</p>
            </div>
            
            <p style="margin: 20px 0;">A partir de agora, você deve usar sua nova senha para acessar o Sistema de Notas Fiscais.</p>
            
            <p>Por segurança, todas as sessões anteriores foram encerradas. Se você estava conectado em outros dispositivos, será necessário fazer login novamente.</p>
            
            <div class="info-box">
                <p><strong>Você não reconhece esta alteração?</strong></p>
                <p>Se você não solicitou essa mudança, sua conta pode estar comprometida. Entre em contato imediatamente com nosso suporte para recuperar o acesso.</p>
            </div>
        </div>
        <div class="footer">
            <p>Este é um email automático, por favor não responda.</p>
            <p>&copy; 2025 Sistema de Notas Fiscais. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>
`

	tmpl, err := template.New("passwordChanged").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("erro ao processar template: %v", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]string{
		"UserName": userName,
	})
	if err != nil {
		return fmt.Errorf("erro ao executar template: %v", err)
	}

	return e.sendEmail(toEmail, subject, body.String())
}

// SendEmailVerificationCode envia código para verificação de email
func (e *EmailService) SendEmailVerificationCode(toEmail, userName, verificationCode string) error {
	subject := "Verificação de Email - Código de Confirmação"

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #2d3748;
            max-width: 600px;
            margin: 0 auto;
            padding: 0;
            background-color: #f7fafc;
        }
        .container {
            background-color: #ffffff;
            margin: 20px;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
            background-color: #ffffff;
        }
        .greeting {
            font-size: 18px;
            margin-bottom: 20px;
            color: #1a202c;
        }
        .code-box {
            background: linear-gradient(135deg, #f6f8fb 0%, #edf2f7 100%);
            border: 2px solid #667eea;
            padding: 30px;
            text-align: center;
            margin: 30px 0;
            border-radius: 10px;
        }
        .code-label {
            font-size: 14px;
            color: #718096;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }
        .code {
            font-size: 42px;
            font-weight: 700;
            color: #667eea;
            letter-spacing: 8px;
            font-family: 'Courier New', monospace;
        }
        .info-box {
            background-color: #fef5e7;
            border-left: 4px solid #f39c12;
            padding: 15px 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-box p {
            margin: 0;
            color: #856404;
            font-size: 14px;
        }
        .footer {
            background-color: #edf2f7;
            padding: 25px 30px;
            text-align: center;
            font-size: 13px;
            color: #718096;
        }
        .footer p {
            margin: 5px 0;
        }
        strong {
            color: #667eea;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Verificação de Email</h1>
        </div>
        <div class="content">
            <p class="greeting">Olá, <strong>{{.UserName}}</strong>!</p>
            
            <p>Para confirmar a alteração do seu endereço de email no Sistema de Notas Fiscais, utilize o código de verificação abaixo:</p>
            
            <div class="code-box">
                <div class="code-label">Código de Verificação</div>
                <div class="code">{{.VerificationCode}}</div>
            </div>
            
            <div class="info-box">
                <p><strong>Atenção:</strong> Este código expira em 15 minutos.</p>
            </div>
            
            <p style="margin-top: 25px;">Se você não solicitou esta verificação, pode ignorar este email com segurança.</p>
        </div>
        <div class="footer">
            <p>Este é um email automático, por favor não responda.</p>
            <p>&copy; 2025 Sistema de Notas Fiscais. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>
`

	tmpl, err := template.New("emailVerification").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("erro ao processar template: %v", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]string{
		"UserName":         userName,
		"VerificationCode": verificationCode,
	})
	if err != nil {
		return fmt.Errorf("erro ao executar template: %v", err)
	}

	return e.sendEmail(toEmail, subject, body.String())
}

// sendEmail é o método privado que realmente envia o email
func (e *EmailService) sendEmail(to, subject, htmlBody string) error {
	logger.InfoF("📧 Tentando enviar email para: %s", to)
	logger.InfoF("📧 SMTP Host: %s:%s", e.SMTPHost, e.SMTPPort)
	logger.InfoF("📧 Sender: %s", e.SenderEmail)

	if e.SenderEmail == "" || e.Password == "" {
		logger.ErrorF("❌ Configurações de email não definidas! SMTP_EMAIL: '%s', SMTP_PASSWORD configurado: %t",
			e.SenderEmail, e.Password != "")
		return fmt.Errorf("configurações de email não definidas. Configure SMTP_EMAIL e SMTP_PASSWORD")
	}

	// Monta o email no formato MIME
	header := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n",
		e.SenderName, e.SenderEmail, to, subject)
	message := []byte(header + htmlBody)

	addr := e.SMTPHost + ":" + e.SMTPPort

	// Se porta for 587 usamos STARTTLS
	if e.SMTPPort == "587" {
		logger.InfoF("📧 Usando STARTTLS (porta 587)")
		c, err := smtp.Dial(addr)
		if err != nil {
			logger.ErrorF("❌ Erro ao conectar ao servidor SMTP: %v", err)
			return fmt.Errorf("erro ao conectar ao servidor SMTP: %v", err)
		}
		defer c.Close()

		host := e.SMTPHost
		if ok, _ := c.Extension("STARTTLS"); ok {
			tlsconfig := &tls.Config{
				ServerName: host,
			}
			if err := c.StartTLS(tlsconfig); err != nil {
				logger.ErrorF("❌ Erro ao iniciar STARTTLS: %v", err)
				return fmt.Errorf("erro ao iniciar STARTTLS: %v", err)
			}
		} else {
			logger.WarnF("⚠️ Servidor SMTP não suporta STARTTLS")
		}

		auth := smtp.PlainAuth("", e.SenderEmail, e.Password, e.SMTPHost)
		if err := c.Auth(auth); err != nil {
			logger.ErrorF("❌ Erro na autenticação SMTP: %v", err)
			return fmt.Errorf("erro na autenticação SMTP: %v", err)
		}

		if err := c.Mail(e.SenderEmail); err != nil {
			logger.ErrorF("❌ erro Mail: %v", err)
			return err
		}
		if err := c.Rcpt(to); err != nil {
			logger.ErrorF("❌ erro Rcpt: %v", err)
			return err
		}

		w, err := c.Data()
		if err != nil {
			logger.ErrorF("❌ erro Data: %v", err)
			return err
		}
		_, err = w.Write(message)
		if err != nil {
			logger.ErrorF("❌ erro escrevendo mensagem: %v", err)
			return err
		}
		err = w.Close()
		if err != nil {
			logger.ErrorF("❌ erro fechando writer: %v", err)
			return err
		}

		if err := c.Quit(); err != nil {
			logger.WarnF("⚠️ erro no Quit SMTP: %v", err)
		}

		logger.InfoF("✅ Email enviado com sucesso (STARTTLS) para: %s", to)
		return nil
	}

	// Caso geral: tentar conexão TLS direta (porta 465) ou fallback
	logger.InfoF("📧 Tentando conexão TLS direta para: %s", addr)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         e.SMTPHost,
	}
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		logger.ErrorF("❌ Erro ao conectar TLS: %v", err)
		// Fallback para smtp.SendMail como última tentativa
		logger.InfoF("📧 Tentando fallback smtp.SendMail para: %s", addr)
		err2 := smtp.SendMail(addr, smtp.PlainAuth("", e.SenderEmail, e.Password, e.SMTPHost), e.SenderEmail, []string{to}, message)
		if err2 != nil {
			logger.ErrorF("❌ Fallback smtp.SendMail também falhou: %v", err2)
			return fmt.Errorf("erro ao enviar email: %v (tls: %v, fallback: %v)", err2, err, err2)
		}
		logger.InfoF("✅ Email enviado com sucesso (fallback smtp.SendMail) para: %s", to)
		return nil
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, e.SMTPHost)
	if err != nil {
		logger.ErrorF("❌ Erro criando cliente SMTP: %v", err)
		return fmt.Errorf("erro criando cliente SMTP: %v", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", e.SenderEmail, e.Password, e.SMTPHost)
	if err := client.Auth(auth); err != nil {
		logger.ErrorF("❌ Erro na autenticação SMTP (TLS): %v", err)
		return fmt.Errorf("erro na autenticação SMTP (TLS): %v", err)
	}

	if err := client.Mail(e.SenderEmail); err != nil {
		logger.ErrorF("❌ erro Mail (TLS): %v", err)
		return err
	}
	if err := client.Rcpt(to); err != nil {
		logger.ErrorF("❌ erro Rcpt (TLS): %v", err)
		return err
	}

	w, err := client.Data()
	if err != nil {
		logger.ErrorF("❌ erro Data (TLS): %v", err)
		return err
	}
	_, err = w.Write(message)
	if err != nil {
		logger.ErrorF("❌ erro escrevendo mensagem (TLS): %v", err)
		return err
	}
	err = w.Close()
	if err != nil {
		logger.ErrorF("❌ erro fechando writer (TLS): %v", err)
		return err
	}

	if err := client.Quit(); err != nil {
		logger.WarnF("⚠️ erro no Quit SMTP (TLS): %v", err)
	}

	logger.InfoF("✅ Email enviado com sucesso (TLS) para: %s", to)
	return nil
}

// IsConfigured verifica se o serviço de email está configurado
func (e *EmailService) IsConfigured() bool {
	return e.SenderEmail != "" && e.Password != ""
}
