package config

import (
	"bytes"
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
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #4CAF50;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border: 1px solid #ddd;
        }
        .code-box {
            background-color: #fff;
            border: 2px dashed #4CAF50;
            padding: 20px;
            text-align: center;
            margin: 20px 0;
            border-radius: 5px;
        }
        .code {
            font-size: 32px;
            font-weight: bold;
            color: #4CAF50;
            letter-spacing: 5px;
        }
        .footer {
            background-color: #f1f1f1;
            padding: 15px;
            text-align: center;
            font-size: 12px;
            color: #666;
            border-radius: 0 0 5px 5px;
        }
        .warning {
            color: #ff6b6b;
            font-weight: bold;
            margin-top: 15px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔐 Recuperação de Senha</h1>
    </div>
    <div class="content">
        <p>Olá, <strong>{{.UserName}}</strong>!</p>
        
        <p>Recebemos uma solicitação para redefinir a senha da sua conta.</p>
        
        <p>Use o código abaixo para recuperar sua senha:</p>
        
        <div class="code-box">
            <div class="code">{{.ResetCode}}</div>
        </div>
        
        <p><strong>⏰ Este código expira em 15 minutos.</strong></p>
        
        <p>Se você não solicitou a recuperação de senha, ignore este email. Sua senha permanecerá inalterada.</p>
        
        <div class="warning">
            ⚠️ Nunca compartilhe este código com ninguém!
        </div>
    </div>
    <div class="footer">
        <p>Este é um email automático, por favor não responda.</p>
        <p>© 2024 Sistema de Notas Fiscais. Todos os direitos reservados.</p>
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
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #4CAF50;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border: 1px solid #ddd;
        }
        .success-icon {
            font-size: 60px;
            text-align: center;
            margin: 20px 0;
        }
        .footer {
            background-color: #f1f1f1;
            padding: 15px;
            text-align: center;
            font-size: 12px;
            color: #666;
            border-radius: 0 0 5px 5px;
        }
        .warning {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>✅ Senha Alterada</h1>
    </div>
    <div class="content">
        <div class="success-icon">🎉</div>
        
        <p>Olá, <strong>{{.UserName}}</strong>!</p>
        
        <p>Sua senha foi alterada com sucesso!</p>
        
        <p>Agora você já pode fazer login com sua nova senha.</p>
        
        <div class="warning">
            <strong>⚠️ Você não fez essa alteração?</strong><br>
            Se você não solicitou essa mudança, entre em contato imediatamente com nosso suporte.
        </div>
    </div>
    <div class="footer">
        <p>Este é um email automático, por favor não responda.</p>
        <p>© 2024 Sistema de Notas Fiscais. Todos os direitos reservados.</p>
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
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #2196F3;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border: 1px solid #ddd;
        }
        .code-box {
            background-color: #fff;
            border: 2px dashed #2196F3;
            padding: 20px;
            text-align: center;
            margin: 20px 0;
            border-radius: 5px;
        }
        .code {
            font-size: 32px;
            font-weight: bold;
            color: #2196F3;
            letter-spacing: 5px;
        }
        .footer {
            background-color: #f1f1f1;
            padding: 15px;
            text-align: center;
            font-size: 12px;
            color: #666;
            border-radius: 0 0 5px 5px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📧 Verificação de Email</h1>
    </div>
    <div class="content">
        <p>Olá, <strong>{{.UserName}}</strong>!</p>
        
        <p>Use o código abaixo para verificar seu email:</p>
        
        <div class="code-box">
            <div class="code">{{.VerificationCode}}</div>
        </div>
        
        <p><strong>⏰ Este código expira em 15 minutos.</strong></p>
        
        <p>Se você não solicitou esta verificação, ignore este email.</p>
    </div>
    <div class="footer">
        <p>Este é um email automático, por favor não responda.</p>
        <p>© 2024 Sistema de Notas Fiscais. Todos os direitos reservados.</p>
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
	// Log início
	logger.InfoF("📧 Tentando enviar email para: %s", to)
	logger.InfoF("📧 SMTP Host: %s:%s", e.SMTPHost, e.SMTPPort)
	logger.InfoF("📧 Sender: %s", e.SenderEmail)

	if e.SenderEmail == "" || e.Password == "" {
		logger.ErrorF("❌ Configurações de email não definidas! SMTP_EMAIL: '%s', SMTP_PASSWORD: %t",
			e.SenderEmail, e.Password != "")
		return fmt.Errorf("configurações de email não definidas. Configure SMTP_EMAIL e SMTP_PASSWORD")
	}

	// Monta o email no formato MIME
	message := []byte("From: " + e.SenderName + " <" + e.SenderEmail + ">\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		htmlBody)

	// Envia o email
	addr := e.SMTPHost + ":" + e.SMTPPort
	logger.InfoF("📧 Conectando em: %s", addr)

	err := smtp.SendMail(addr, e.auth, e.SenderEmail, []string{to}, message)
	if err != nil {
		logger.ErrorF("❌ Erro ao enviar email: %v", err)
		return fmt.Errorf("erro ao enviar email: %v", err)
	}

	logger.InfoF("✅ Email enviado com sucesso para: %s", to)
	return nil
}

// IsConfigured verifica se o serviço de email está configurado
func (e *EmailService) IsConfigured() bool {
	return e.SenderEmail != "" && e.Password != ""
}
