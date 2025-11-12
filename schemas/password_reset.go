package schemas

import (
	"time"

	"gorm.io/gorm"
)

// PasswordReset armazena tokens de recuperação de senha
// Tokens expiram em 15 minutos e só podem ser usados uma vez
type PasswordReset struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"not null;size:6;index"` // Código de 6 dígitos
	ExpiresAt time.Time `gorm:"not null;index"`
	Used      bool      `gorm:"default:false;not null"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// IsValid verifica se o token ainda é válido
func (pr *PasswordReset) IsValid() bool {
	return !pr.Used && time.Now().Before(pr.ExpiresAt)
}

// MarkAsUsed marca o token como usado
func (pr *PasswordReset) MarkAsUsed(db *gorm.DB) error {
	pr.Used = true
	return db.Save(pr).Error
}
