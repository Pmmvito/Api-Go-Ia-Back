package schemas

import (
	"time"

	"gorm.io/gorm"
)

// Category representa uma categoria de produto no banco de dados.
// Inclui detalhes como nome, descrição, ícone e cor.
// Cada usuário tem suas próprias categorias individuais.
type Category struct {
	gorm.Model
	UserID       uint          `json:"userId" gorm:"not null;index:idx_user_category"`                        // ID do usuário dono da categoria
	Name         string        `json:"name" gorm:"not null;index:idx_user_category"`                          // Nome da categoria
	Description  string        `json:"description"`                                                           // Descrição da categoria
	Icon         string        `json:"icon"`                                                                  // Emoji ou ícone representando a categoria
	Color        string        `json:"color"`                                                                 // Código de cor hexadecimal para a categoria (ex: #FF5733)
	User         User          `json:"-" gorm:"foreignKey:UserID"`                                            // Relacionamento BelongsTo com User
	ReceiptItems []ReceiptItem `json:"-" gorm:"foreignKey:CategoryID"`                                        // Relacionamento HasMany com ReceiptItems
	ListItems    []ListItem    `json:"-" gorm:"foreignKey:CategoryID"`                                        // Relacionamento HasMany com ListItems
}

// CategoryResponse define a estrutura dos dados da categoria enviados nas respostas da API.
// Omite campos sensíveis ou desnecessários do modelo Category.
type CategoryResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
}

// ToResponse converte um modelo Category para o formato CategoryResponse.
// Isso é útil para garantir respostas de API consistentes.
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
