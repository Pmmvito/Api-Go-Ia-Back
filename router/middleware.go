package router

import (
	"net/http"
	"os"
	"strings"

	"github.com/Pmmvito/Golang-Api-Exemple/config"
	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var logger = config.GetLogger("middleware")

// AuthMiddleware é um middleware Gin que valida o token JWT Bearer.
// Ele verifica a presença e o formato do cabeçalho de autorização,
// valida o token e extrai as informações do usuário para o contexto da solicitação.
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Pega o header Authorization
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message":   "Authorization header is required",
				"errorCode": http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		// Verifica se é Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message":   "Invalid authorization header format. Expected: Bearer <token>",
				"errorCode": http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		tokenString := parts[1]

		// Verifica se o token está na blacklist (logout)
		db := config.GetPostgreSQL()
		var blacklisted schemas.TokenBlacklist
		if err := db.Where("token = ?", tokenString).First(&blacklisted).Error; err == nil {
			// Token foi invalidado (logout)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message":   "Token has been invalidated. Please login again.",
				"errorCode": http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		// Valida o token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verifica o método de assinatura
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message":   "Invalid or expired token",
				"errorCode": http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		// Extrai as claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// 🔒 Verifica se é um access token (não aceita refresh tokens aqui)
			if tokenType, ok := claims["type"].(string); ok && tokenType != "access" {
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"message":   "Invalid token type. Use access token for API requests.",
					"errorCode": http.StatusUnauthorized,
				})
				ctx.Abort()
				return
			}

			// Adiciona o userID no contexto para uso nos handlers
			if userID, ok := claims["user_id"].(float64); ok {
				ctx.Set("user_id", uint(userID))
				ctx.Set("token", tokenString) // Armazena token para usar no logout

				// Opcionalmente, busca o usuário completo do banco
				var user schemas.User
				if err := db.First(&user, uint(userID)).Error; err == nil {
					ctx.Set("user", user)
				}
			}
		}

		ctx.Next()
	}
}
