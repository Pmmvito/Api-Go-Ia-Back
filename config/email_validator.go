package config

import (
	"context"
	"net"
	"regexp"
	"strings"
	"time"
)

// EmailValidator valida emails de forma robusta
type EmailValidator struct {
	// Cache de domínios já verificados (evita verificar gmail.com 1000x)
	domainCache map[string]bool
}

// NewEmailValidator cria uma nova instância do validador
func NewEmailValidator() *EmailValidator {
	return &EmailValidator{
		domainCache: make(map[string]bool),
	}
}

// ValidateEmail valida email em 3 etapas
func (v *EmailValidator) ValidateEmail(email string) (bool, string) {
	logger.InfoF("🔍 Validando email: %s", email)

	// Etapa 1: Formato básico
	if !v.isValidFormat(email) {
		logger.WarnF("❌ Email com formato inválido: %s", email)
		return false, "Formato de email inválido"
	}

	// Etapa 2: Domínios descartáveis conhecidos
	if v.isDisposableEmail(email) {
		logger.WarnF("❌ Email descartável detectado: %s", email)
		return false, "Emails temporários/descartáveis não são permitidos"
	}

	// Etapa 3: Verificar MX Records (domínio aceita emails?)
	domain := v.getDomain(email)

	// Verificar cache primeiro
	if valid, exists := v.domainCache[domain]; exists {
		if !valid {
			logger.WarnF("❌ Domínio inválido (cache): %s", domain)
			return false, "Domínio de email inválido ou inexistente"
		}
		logger.InfoF("✅ Domínio válido (cache): %s", domain)
		return true, ""
	}

	// Verificar MX records
	valid := v.verifyMXRecords(domain)
	v.domainCache[domain] = valid // Salvar no cache

	if !valid {
		logger.WarnF("❌ Domínio inválido (MX): %s", domain)
		return false, "Domínio de email inválido ou inexistente. Verifique se digitou corretamente."
	}

	logger.InfoF("✅ Email válido: %s", email)
	return true, ""
}

// isValidFormat verifica formato com regex
func (v *EmailValidator) isValidFormat(email string) bool {
	// Regex padrão RFC 5322 (simplificado)
	regex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(regex, email)
	return match
}

// getDomain extrai domínio do email
func (v *EmailValidator) getDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

// verifyMXRecords verifica se domínio tem servidor de email
func (v *EmailValidator) verifyMXRecords(domain string) bool {
	logger.InfoF("🔍 Verificando MX records para: %s", domain)

	// Criar resolver com timeout de 5 segundos
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	// Criar contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Resolver MX records
	mxRecords, err := resolver.LookupMX(ctx, domain)
	if err != nil {
		logger.WarnF("⚠️ Erro ao buscar MX records para %s: %v", domain, err)

		// Fallback: verificar se domínio existe (A record)
		ips, err := resolver.LookupHost(ctx, domain)
		if err != nil || len(ips) == 0 {
			logger.WarnF("❌ Domínio %s não existe", domain)
			return false
		}

		logger.InfoF("✅ Domínio %s existe (mas sem MX explícito) - IP: %v", domain, ips[0])
		return true // Domínio existe, pode aceitar emails
	}

	if len(mxRecords) == 0 {
		logger.WarnF("❌ Nenhum MX record encontrado para %s", domain)
		return false
	}

	logger.InfoF("✅ MX records encontrados para %s: %d servidor(es) - Primeiro: %s",
		domain, len(mxRecords), mxRecords[0].Host)
	return true
}

// isDisposableEmail verifica se é email descartável
func (v *EmailValidator) isDisposableEmail(email string) bool {
	domain := v.getDomain(email)

	// Lista de domínios descartáveis conhecidos (top 20 mais usados)
	disposableDomains := []string{
		"10minutemail.com",
		"10minutemail.net",
		"guerrillamail.com",
		"guerrillamail.net",
		"mailinator.com",
		"tempmail.com",
		"throwaway.email",
		"temp-mail.org",
		"temp-mail.io",
		"maildrop.cc",
		"yopmail.com",
		"mohmal.com",
		"sharklasers.com",
		"trashmail.com",
		"getnada.com",
		"tempr.email",
		"minuteinbox.com",
		"dispostable.com",
		"fakeinbox.com",
		"mailnesia.com",
		"emailondeck.com",
	}

	for _, disposable := range disposableDomains {
		if domain == disposable {
			return true
		}
	}

	return false
}
