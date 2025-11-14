package schemas

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User define o modelo de usuário para o banco de dados.
// Email é único globalmente - mesmo usuários deletados não podem ter email reutilizado
type User struct {
	gorm.Model
	Name          string         `gorm:"not null"`
	Email         string         `gorm:"not null;index:idx_email_unique,unique"` // Índice único que impede reuso de email mesmo após soft delete
	Password      string         `gorm:"not null"`
	ActiveToken   *string        `gorm:"type:text" json:"-"`                            // Token JWT ativo atual (null após logout)
	Receipts      []Receipt      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // Relacionamento HasMany com Receipts
	ShoppingLists []ShoppingList `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // Relacionamento HasMany com ShoppingLists
}

// UserResponse define a estrutura de dados do usuário para respostas da API, omitindo a senha..
type UserResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

// HashPassword gera o hash da senha do usuário usando bcrypt.
// 🔒 SEGURANÇA: Usa cost 12 (4x mais seguro que default 10)
func (u *User) HashPassword(password string) error {
	// Bcrypt cost 12 = 4096 iterações (vs cost 10 = 1024 iterações)
	// Cada +1 no cost dobra o tempo de processamento
	// Cost 12 é recomendado para 2024+ (balanço segurança/performance)
	const bcryptCost = 12
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifica se a senha fornecida corresponde ao hash armazenado.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// ToResponse converte um modelo User para o formato UserResponse, omitindo a senha.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Name:      u.Name,
		Email:     u.Email,
	}
}
