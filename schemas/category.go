package schemas

import (
	"time"

	"gorm.io/gorm"
)

// Category representa uma categoria de produto
type Category struct {
	gorm.Model
	Name         string        `json:"name" gorm:"unique;not null"`    // Nome da categoria
	Description  string        `json:"description"`                    // Descrição da categoria
	Icon         string        `json:"icon"`                           // Emoji ou ícone da categoria
	Color        string        `json:"color"`                          // Cor em hexadecimal (#FF5733)
	ReceiptItems []ReceiptItem `json:"-" gorm:"foreignKey:CategoryID"` // Relacionamento HasMany com ReceiptItems
	ListItems    []ListItem    `json:"-" gorm:"foreignKey:CategoryID"` // Relacionamento HasMany com ListItems
}

// CategoryResponse representa a resposta da API de categoria
type CategoryResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
}

// ToResponse converte Category para CategoryResponse
func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:          c.ID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		Name:        c.Name,
		Description: c.Description,
		Icon:        c.Icon,
		Color:       c.Color,
	}
}
