package schemas

import (
	"time"

	"gorm.io/gorm"
)

// Product representa um produto no sistema.
// Ele armazena o nome e a unidade do produto.
type Product struct {
	gorm.Model
	Unity        string        `json:"unity" gorm:"size:10;not null"` // Unidade do produto, ex: "un", "kg", "g", "l", "ml"
	Name         string        `json:"name" gorm:"not null;index"`    // Nome do produto
	ReceiptItems []ReceiptItem `json:"-" gorm:"foreignKey:ProductID"` // Relacionamento HasMany com ReceiptItems
	ListItems    []ListItem    `json:"-" gorm:"foreignKey:ProductID"` // Relacionamento HasMany com ListItems
}

// ProductResponse define a estrutura dos dados do produto enviados nas respostas da API.
type ProductResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Unity     string    `json:"unity"`
	Name      string    `json:"name"`
}

// ToResponse converte um modelo Product para o formato ProductResponse.
func (p *Product) ToResponse() ProductResponse {
	return ProductResponse{
		ID:        p.ID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		Unity:     p.Unity,
		Name:      p.Name,
	}
}
