package schemas

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// RefreshToken armazena tokens de refresh para autenticação
type RefreshToken struct {
	gorm.Model
	Token     string     `gorm:"type:varchar(255);uniqueIndex;not null"` // Token UUID único
	UserID    uint       `gorm:"not null;index"`                         // ID do usuário
	ExpiresAt time.Time  `gorm:"not null"`                               // Data de expiração (7 dias)
	Used      bool       `gorm:"default:false;not null"`                 // Se já foi usado
	RevokedAt *time.Time `gorm:"index"`                                  // Data de revogação (logout)
}

// GenerateRefreshToken cria um novo refresh token criptograficamente seguro
func GenerateRefreshToken() (string, error) {
	// Gerar 32 bytes aleatórios (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Converter para string hexadecimal (64 caracteres)
	return hex.EncodeToString(bytes), nil
}

// IsValid verifica se o refresh token ainda é válido
func (rt *RefreshToken) IsValid() bool {
	now := time.Now()

	// Token não pode estar expirado
	if rt.ExpiresAt.Before(now) {
		return false
	}

	// Token não pode estar revogado
	if rt.RevokedAt != nil {
		return false
	}

	// Token não pode ter sido usado (one-time use)
	if rt.Used {
		return false
	}

	return true
}

// Revoke marca o token como revogado (usado no logout)
func (rt *RefreshToken) Revoke(db *gorm.DB) error {
	now := time.Now()
	rt.RevokedAt = &now
	return db.Save(rt).Error
}

// MarkAsUsed marca o token como usado (após renovação)
func (rt *RefreshToken) MarkAsUsed(db *gorm.DB) error {
	rt.Used = true
	return db.Save(rt).Error
}

// CreateRefreshToken cria e salva um novo refresh token no banco
func CreateRefreshToken(db *gorm.DB, userID uint) (*RefreshToken, error) {
	token, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshToken := &RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 dias
		Used:      false,
	}

	if err := db.Create(refreshToken).Error; err != nil {
		return nil, err
	}

	return refreshToken, nil
}

// CleanupExpiredTokens remove tokens expirados ou revogados antigos (rotina de limpeza)
func CleanupExpiredTokens(db *gorm.DB) error {
	// Deletar tokens expirados há mais de 30 dias
	cutoff := time.Now().Add(-30 * 24 * time.Hour)

	return db.Unscoped().Where("expires_at < ? OR revoked_at < ?", cutoff, cutoff).
		Delete(&RefreshToken{}).Error
}
