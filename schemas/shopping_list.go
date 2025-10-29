package schemas

import (
	"time"

	"gorm.io/gorm"
)

// ShoppingList representa uma lista de compras
type ShoppingList struct {
	gorm.Model
	UserID uint       `json:"userId" gorm:"not null;index"`                                       // FK para User
	User   *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`                            // Relacionamento com User
	Name   string     `json:"name" gorm:"not null"`                                               // Nome da lista
	Items  []ListItem `json:"items" gorm:"foreignKey:ShoppingListID;constraint:OnDelete:CASCADE"` // Relacionamento HasMany
}

// ListItem representa um item de uma lista de compras
type ListItem struct {
	gorm.Model
	ShoppingListID uint      `json:"shoppingListId" gorm:"not null;index"`            // FK para ShoppingList
	CategoryID     uint      `json:"categoryId" gorm:"not null;index"`                // FK para Category
	Category       *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"` // Relacionamento com Category
	ProductID      uint      `json:"productId" gorm:"not null;index"`                 // FK para Product
	Product        *Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`   // Relacionamento com Product
	Quantity       float64   `json:"quantity" gorm:"type:decimal(10,3);not null"`     // Quantidade desejada
	UnitPrice      float64   `json:"unitPrice" gorm:"type:decimal(10,2)"`             // Preço unitário (opcional)
	Total          float64   `json:"total" gorm:"type:decimal(10,2)"`                 // Total do item (opcional)
}

// ShoppingListResponse representa a resposta da API de lista de compras
type ShoppingListResponse struct {
	ID        uint               `json:"id"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
	UserID    uint               `json:"userId"`
	Name      string             `json:"name"`
	Items     []ListItemResponse `json:"items"`
}

// ListItemResponse representa um item na resposta da API
type ListItemResponse struct {
	ID             uint             `json:"id"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
	ShoppingListID uint             `json:"shoppingListId"`
	CategoryID     uint             `json:"categoryId"`
	Category       *CategorySimple  `json:"category,omitempty"` // Apenas ID e Nome
	ProductID      uint             `json:"productId"`
	Product        *ProductResponse `json:"product,omitempty"` // Produto completo
	Quantity       float64          `json:"quantity"`
	UnitPrice      float64          `json:"unitPrice"`
	Total          float64          `json:"total"`
}

// ToResponse converte ShoppingList para ShoppingListResponse
func (s *ShoppingList) ToResponse() ShoppingListResponse {
	items := make([]ListItemResponse, len(s.Items))
	for i, item := range s.Items {
		itemResponse := ListItemResponse{
			ID:             item.ID,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			ShoppingListID: item.ShoppingListID,
			CategoryID:     item.CategoryID,
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			Total:          item.Total,
		}

		// Adiciona categoria se existir
		if item.Category != nil {
			itemResponse.Category = &CategorySimple{
				ID:   item.Category.ID,
				Name: item.Category.Name,
			}
		}

		// Adiciona produto se existir
		if item.Product != nil {
			productResponse := item.Product.ToResponse()
			itemResponse.Product = &productResponse
		}

		items[i] = itemResponse
	}

	return ShoppingListResponse{
		ID:        s.ID,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		UserID:    s.UserID,
		Name:      s.Name,
		Items:     items,
	}
}

// ToResponse converte ListItem para ListItemResponse
func (item *ListItem) ToResponse() ListItemResponse {
	itemResponse := ListItemResponse{
		ID:             item.ID,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
		ShoppingListID: item.ShoppingListID,
		CategoryID:     item.CategoryID,
		ProductID:      item.ProductID,
		Quantity:       item.Quantity,
		UnitPrice:      item.UnitPrice,
		Total:          item.Total,
	}

	// Adiciona categoria se existir
	if item.Category != nil {
		itemResponse.Category = &CategorySimple{
			ID:   item.Category.ID,
			Name: item.Category.Name,
		}
	}

	// Adiciona produto se existir
	if item.Product != nil {
		productResponse := item.Product.ToResponse()
		itemResponse.Product = &productResponse
	}

	return itemResponse
}
