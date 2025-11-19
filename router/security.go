package router

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SecureMiddleware adiciona headers de segurança e força HTTPS em produção
func SecureMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Log useful headers for debugging HTTPS/proxy forwarding
		proto := ctx.Request.Header.Get("X-Forwarded-Proto")
		logger.Debugf("SecurityMiddleware: X-Forwarded-Proto=%s, TLS=%v, URL=%s", proto, ctx.Request.TLS != nil, ctx.Request.URL.String())
		// 🔒 Forçar HTTPS em produção
		if os.Getenv("ENV") == "production" {
			// Verifica se está usando HTTP (não HTTPS)
			if ctx.Request.Header.Get("X-Forwarded-Proto") != "https" && ctx.Request.TLS == nil {
				httpsURL := "https://" + ctx.Request.Host + ctx.Request.RequestURI
				logger.WarnF("Redirecionando HTTP para HTTPS: %s -> %s", ctx.Request.URL, httpsURL)
				ctx.Redirect(http.StatusMovedPermanently, httpsURL)
				ctx.Abort()
				return
			}
		}

		// 🔒 Headers de Segurança (aplicar sempre)

		// HSTS: Forçar HTTPS por 1 ano (incluindo subdomínios)
		ctx.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// X-Content-Type-Options: Prevenir MIME type sniffing
		ctx.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Prevenir clickjacking
		ctx.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection: Proteção XSS (legacy browsers)
		ctx.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Controlar informação de referrer
		ctx.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content-Security-Policy: Prevenir XSS e data injection
		ctx.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'")

		// Permissions-Policy: Controlar features do navegador
		ctx.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// X-Permitted-Cross-Domain-Policies: Prevenir Flash/PDF cross-domain
		ctx.Header("X-Permitted-Cross-Domain-Policies", "none")

		ctx.Next()
	}
}

// CORSMiddleware configuração segura de CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			// Default: permitir apenas localhost em desenvolvimento
			if os.Getenv("ENV") == "production" {
				allowedOrigins = "https://yourdomain.com" // ⚠️ CONFIGURAR SEU DOMÍNIO
			} else {
				allowedOrigins = "http://localhost:3000,http://localhost:5173" // React/Vite
			}
		}

		// Verificar se origin está na lista permitida
		// Em produção, validar exatamente. Em dev, permitir localhost
		ctx.Header("Access-Control-Allow-Origin", allowedOrigins)
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Max-Age", "86400") // 24 horas

		// Preflight request
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
