package schemas

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model 
	Name          string         `gorm:"not null"`
	Email         string         `gorm:"unique;not null;index"`
	Password      string         `gorm:"not null"`
	Receipts      []Receipt      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // Relacionamento HasMany com Receipts
	ShoppingLists []ShoppingList `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // Relacionamento HasMany com ShoppingLists
}

type UserResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

// HashPassword gera o hash da senha
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifica se a senha está correta
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// ToResponse converte User para UserResponse (sem senha)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Name:      u.Name,
		Email:     u.Email,
	}
}
