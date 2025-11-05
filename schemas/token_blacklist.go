package schemas

import (
	"time"

	"gorm.io/gorm"
)

// TokenBlacklist armazena tokens JWT invalidados (logout)
type TokenBlacklist struct {
	gorm.Model
	Token     string    `json:"token" gorm:"unique;not null;index"` // Token JWT completo
	UserID    uint      `json:"userId" gorm:"not null;index"`       // ID do usuário que fez logout
	ExpiresAt time.Time `json:"expiresAt" gorm:"not null;index"`    // Quando o token expira naturalmente
}

// TokenBlacklistResponse representa a resposta da API
type TokenBlacklistResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToResponse converte TokenBlacklist para TokenBlacklistResponse
func (t *TokenBlacklist) ToResponse() TokenBlacklistResponse {
	return TokenBlacklistResponse{
		ID:        t.ID,
		UserID:    t.UserID,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
	}
}
