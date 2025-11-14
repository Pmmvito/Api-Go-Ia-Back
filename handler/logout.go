package handler

import (
	"net/http"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// LogoutHandler invalida o token JWT adicionando-o à blacklist
// @Summary Fazer logout
// @Description Invalida o token JWT atual, impedindo seu uso futuro
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string "Token não encontrado no contexto de autenticação"
// @Failure 500 {object} map[string]string "Erro ao adicionar o token à lista de tokens invalidados. Por favor, tente novamente"
// @Router /logout [post]
func LogoutHandler(ctx *gin.Context) {
	// Pega o token do contexto (foi armazenado pelo middleware)
	tokenString, exists := ctx.Get("token")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token não encontrado no contexto de autenticação"})
		return
	}

	// Pega o userID do contexto
	userID, _ := ctx.Get("user_id")

	// Parse do token para obter a expiração
	token, err := jwt.Parse(tokenString.(string), func(token *jwt.Token) (interface{}, error) {
		return []byte(""), nil // Não precisa validar aqui, já foi validado no middleware
	})

	var expiresAt time.Time
	if err == nil && token != nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				expiresAt = time.Unix(int64(exp), 0)
			}
		}
	}

	// Se não conseguiu pegar a expiração, usa um padrão de 7 dias
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	// Adiciona à blacklist
	blacklist := schemas.TokenBlacklist{
		Token:     tokenString.(string),
		UserID:    userID.(uint),
		ExpiresAt: expiresAt,
	}

	if err := db.Create(&blacklist).Error; err != nil {
		logger.ErrorF("Erro ao adicionar token à blacklist: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar o token à lista de tokens invalidados. Por favor, tente novamente"})
		return
	}

	// 🔒 Limpa o active_token do usuário
	if err := db.Model(&schemas.User{}).Where("id = ?", userID).Update("active_token", nil).Error; err != nil {
		logger.ErrorF("Erro ao limpar active_token: %v", err)
		// Não falha o logout por isso, já adicionou na blacklist
	}

	// 🔒 NOVO: Revoga todos os refresh tokens do usuário (força re-login)
	if err := db.Model(&schemas.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", time.Now()).Error; err != nil {
		logger.ErrorF("Erro ao revogar refresh tokens: %v", err)
		// Não falha o logout por isso, access token já foi invalidado
	} else {
		logger.InfoF("Todos os refresh tokens do usuário %d foram revogados", userID.(uint))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Logout realizado com sucesso",
	})
}
