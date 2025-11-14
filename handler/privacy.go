package handler

import (
	"strings"
)

// maskEmail mascara um email para proteção de dados (LGPD/GDPR)
// Exemplo: joao.silva@example.com -> jo***@example.com
func maskEmail(email string) string {
	if email == "" {
		return "***@***"
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}

	username := parts[0]
	domain := parts[1]

	// Manter apenas primeiras 2 letras do username
	if len(username) <= 2 {
		return "**@" + domain
	}

	return username[:2] + "***@" + domain
}

// maskIP mascara um IP para proteção de dados
// Exemplo: 192.168.1.100 -> 192.168.***.***
func maskIP(ip string) string {
	if ip == "" {
		return "***.***.***.***"
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		// IPv6 ou formato inválido
		return "***.***.***.***"
	}

	return parts[0] + "." + parts[1] + ".***." + "***"
}
