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

// AuthMiddleware valida o token JWT Bearer
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
			// Adiciona o userID no contexto para uso nos handlers
			if userID, ok := claims["user_id"].(float64); ok {
				ctx.Set("user_id", uint(userID))

				// Opcionalmente, busca o usuário completo do banco
				db := config.GetPostgreSQL()
				var user schemas.User
				if err := db.First(&user, uint(userID)).Error; err == nil {
					ctx.Set("user", user)
				}
			}
		}

		ctx.Next()
	}
}
