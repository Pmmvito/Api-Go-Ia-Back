package router

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter armazena limitadores por IP
type RateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit // requests por segundo
	b   int        // burst size
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	i := &RateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   rate.Limit(requestsPerSecond),
		b:   burst,
	}

	// Limpar limiters antigos a cada 1 hora
	go i.cleanupStaleEntries()

	return i
}

// getLimiter retorna o limiter para o IP, criando se necessário
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.ips[ip] = limiter
	}

	return limiter
}

// cleanupStaleEntries remove limiters inativos para liberar memória
func (rl *RateLimiter) cleanupStaleEntries() {
	for {
		time.Sleep(time.Hour)
		rl.mu.Lock()

		// Limpa todos (reset a cada hora)
		rl.ips = make(map[string]*rate.Limiter)

		rl.mu.Unlock()
	}
}

// RateLimitMiddleware middleware global de rate limiting
// Uso: router.Use(RateLimitMiddleware(100, 200)) // 100 req/s, burst 200
func RateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerSecond, burst)

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		if !limiter.getLimiter(ip).Allow() {
			logger.WarnF("Rate limit excedido para IP: %s", ip)
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Muitas requisições. Por favor, aguarde alguns segundos e tente novamente",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// StrictRateLimitMiddleware rate limit estrito para endpoints sensíveis
// Uso: router.POST("/login", StrictRateLimitMiddleware(5, time.Minute), handler)
// Permite 5 requisições por minuto por IP
func StrictRateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	type clientInfo struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	clients := make(map[string]*clientInfo)
	mu := &sync.RWMutex{}

	// Cleanup a cada 5 minutos
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()

			now := time.Now()
			for ip, info := range clients {
				// Remove clientes inativos há mais de 10 minutos
				if now.Sub(info.lastSeen) > 10*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		mu.Lock()
		client, exists := clients[ip]
		if !exists {
			// rate.Every calcula o intervalo: window/maxRequests
			// Ex: 1 minuto / 5 requisições = 1 requisição a cada 12 segundos
			client = &clientInfo{
				limiter:  rate.NewLimiter(rate.Every(window/time.Duration(maxRequests)), 1),
				lastSeen: time.Now(),
			}
			clients[ip] = client
		}
		client.lastSeen = time.Now()
		mu.Unlock()

		if !client.limiter.Allow() {
			logger.WarnF("Rate limit estrito excedido para IP: %s (endpoint: %s)", ip, ctx.Request.URL.Path)
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Você atingiu o limite de tentativas. Por favor, aguarde alguns instantes e tente novamente",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// LoginRateLimitMiddleware rate limit específico para login (5 tentativas por minuto)
func LoginRateLimitMiddleware() gin.HandlerFunc {
	return StrictRateLimitMiddleware(5, time.Minute)
}

// RegisterRateLimitMiddleware rate limit para registro (2 cadastros por minuto)
func RegisterRateLimitMiddleware() gin.HandlerFunc {
	return StrictRateLimitMiddleware(2, time.Minute)
}

// ForgotPasswordRateLimitMiddleware rate limit para forgot password (3 tentativas por hora)
func ForgotPasswordRateLimitMiddleware() gin.HandlerFunc {
	return StrictRateLimitMiddleware(3, time.Hour)
}
