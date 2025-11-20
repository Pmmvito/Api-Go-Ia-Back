package schemas

import (
	"time"

	"gorm.io/gorm"
)

// CategorySimple fornece uma representação leve de uma categoria, incluindo apenas ID e Nome.
type CategorySimple struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ReceiptItem representa um item em um recibo. É armazenado em uma tabela separada
// e vinculado a um recibo, um produto e uma categoria.
type ReceiptItem struct {
	gorm.Model
	ReceiptID  uint      `json:"receiptId" gorm:"not null;index"`                 // Chave estrangeira para Receipt
	Receipt    *Receipt  `json:"receipt,omitempty" gorm:"foreignKey:ReceiptID"`   // Relacionamento com Receipt
	CategoryID uint      `json:"categoryId" gorm:"not null;index"`                // Chave estrangeira para Category
	Category   *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"` // Relacionamento com Category
	ProductID  uint      `json:"productId" gorm:"not null;index"`                 // Chave estrangeira para Product
	Product    *Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`   // Relacionamento com Product
	Quantity   float64   `json:"quantity" gorm:"type:decimal(10,3);not null"`     // Quantidade ou peso do item
	UnitPrice  float64   `json:"unitPrice" gorm:"type:decimal(10,2);not null"`    // Preço unitário do item
	Total      float64   `json:"total" gorm:"type:decimal(10,2);not null"`        // Preço total do item
	// Campos legados para compatibilidade, a serem removidos no futuro.
	Description string `json:"description,omitempty" gorm:"-"` // Legado: usar Product.Name
	Unit        string `json:"unit,omitempty" gorm:"-"`        // Legado: usar Product.Unity
}

// Receipt representa um recibo escaneado no banco de dados.
type Receipt struct {
	gorm.Model
	UserID      uint          `json:"userId" gorm:"not null;index"`                                  // Chave estrangeira para User
	User        *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`                       // Relacionamento com User
	StoreName   string        `json:"storeName"`                                                     // Nome da loja
	Date        string        `json:"date" gorm:"type:date;index"`                                   // Data da compra (YYYY-MM-DD)
	Items       []ReceiptItem `json:"items" gorm:"foreignKey:ReceiptID;constraint:OnDelete:CASCADE"` // Relacionamento HasMany com ReceiptItem
	Subtotal    float64       `json:"subtotal" gorm:"type:decimal(10,2)"`                            // Subtotal do recibo
	Discount    float64       `json:"discount" gorm:"type:decimal(10,2)"`                            // Valor do desconto
	Total       float64       `json:"total" gorm:"type:decimal(10,2)"`                               // Total final
	Currency    string        `json:"currency" gorm:"size:3;default:'BRL'"`                          // Moeda (ex: BRL, USD)
	Confidence  float64       `json:"confidence" gorm:"type:decimal(3,2)"`                           // Pontuação de confiança da IA (0-1)
	Notes       string        `json:"notes" gorm:"type:text"`                                        // Notas da IA
	ImageBase64 string        `json:"-" gorm:"type:text"`                                            // Imagem original em base64
}

// ReceiptItemSummary fornece uma visão resumida de um item de recibo, excluindo campos de auditoria.
type ReceiptItemSummary struct {
	ID         uint            `json:"id"`
	CategoryID uint            `json:"categoryId"`
	Category   *CategorySimple `json:"category,omitempty"` // Apenas ID e Nome
	ProductID  uint            `json:"productId"`
	Product    *ProductSimple  `json:"product,omitempty"` // Nome e unidade do produto
	Quantity   float64         `json:"quantity"`
	UnitPrice  float64         `json:"unitPrice"`
	Total      float64         `json:"total"`
	Subtotal   float64         `json:"subtotal"`
	Discount   float64         `json:"discount"`
}

// ProductSimple fornece uma representação leve ede um produto, com apenas nome e unidade.
type ProductSimple struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Unity string `json:"unity"`
}

// ReceiptBasic fornece uma versão ultra-simplificada de um recibo para listagens rápidas.
type ReceiptBasic struct {
	ID        uint    `json:"id"`
	StoreName string  `json:"storeName"`
	Date      string  `json:"date"`
	ItemCount int     `json:"itemCount"` // Número de itens
	Total     float64 `json:"total"`
	Currency  string  `json:"currency"`
}

// ReceiptSummary fornece uma versão leve de um recibo para listagens.
type ReceiptSummary struct {
	ID        uint                 `json:"id"`
	StoreName string               `json:"storeName"`
	Date      string               `json:"date"`
	Items     []ReceiptItemSummary `json:"items"`
	Subtotal  float64              `json:"subtotal"`
	Discount  float64              `json:"discount"`
	Total     float64              `json:"total"`
	Currency  string               `json:"currency"`
}

// ReceiptItemResponse define a estrutura de um item de recibo nas respostas da API, excluindo gorm.Model para o Swagger.
type ReceiptItemResponse struct {
	ID         uint            `json:"id"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
	ReceiptID  uint            `json:"receiptId"`
	CategoryID uint            `json:"categoryId"`
	Category   *CategorySimple `json:"category,omitempty"` // Apenas ID e Nome
	ProductID  uint            `json:"productId"`
	Product    *ProductSimple  `json:"product,omitempty"` // Nome e unidade do produto
	Quantity   float64         `json:"quantity"`
	UnitPrice  float64         `json:"unitPrice"`
	Total      float64         `json:"total"`
}

// ReceiptResponse representa a resposta para uma chamada de API de recibo escaneado.
type ReceiptResponse struct {
	ID         uint                  `json:"id"`
	CreatedAt  time.Time             `json:"createdAt"`
	UpdatedAt  time.Time             `json:"updatedAt"`
	UserID     uint                  `json:"userId"`
	StoreName  string                `json:"storeName"`
	Date       string                `json:"date"`
	Items      []ReceiptItemResponse `json:"items"`
	Subtotal   float64               `json:"subtotal"`
	Discount   float64               `json:"discount"`
	Total      float64               `json:"total"`
	Currency   string                `json:"currency"`
	Confidence float64               `json:"confidence"`
	Notes      string                `json:"notes"`
}

// ToBasic converte um Receipt para um ReceiptBasic, uma versão ultra-simplificada para listagens rápidas.
func (r *Receipt) ToBasic() ReceiptBasic {
	return ReceiptBasic{
		ID:        r.ID,
		StoreName: r.StoreName,
		Date:      r.Date,
		ItemCount: len(r.Items), // Conta os itens
		Total:     r.Total,
		Currency:  r.Currency,
	}
}

// ToSummary converte um Receipt para um ReceiptSummary, uma versão leve para listagens.
func (r *Receipt) ToSummary() ReceiptSummary {
	// Converte itens para uma versão resumida
	items := make([]ReceiptItemSummary, len(r.Items))
	for i, item := range r.Items {
		itemSummary := ReceiptItemSummary{
			ID:         item.ID,
			CategoryID: item.CategoryID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			Total:      item.Total,
			Subtotal:   r.Subtotal, // usar o subtotal do receipt
			Discount:   r.Discount, // usar o discount do receipt
		}

		// Adiciona categoria se existir (APENAS ID e Nome)
		if item.Category != nil {
			itemSummary.Category = &CategorySimple{
				ID:   item.Category.ID,
				Name: item.Category.Name,
			}
		}

		// Adiciona produto se existir (APENAS ID, Nome e Unidade)
		if item.Product != nil {
			itemSummary.Product = &ProductSimple{
				ID:    item.Product.ID,
				Name:  item.Product.Name,
				Unity: item.Product.Unity,
			}
		}

		items[i] = itemSummary
	}

	return ReceiptSummary{
		ID:        r.ID,
		StoreName: r.StoreName,
		Date:      r.Date,
		Items:     items,
		Subtotal:  r.Subtotal,
		Discount:  r.Discount,
		Total:     r.Total,
		Currency:  r.Currency,
	}
}

// ToResponse converte um Receipt para um ReceiptResponse.
func (r *Receipt) ToResponse() ReceiptResponse {
	// Converte itens
	items := make([]ReceiptItemResponse, len(r.Items))
	for i, item := range r.Items {
		itemResponse := ReceiptItemResponse{
			ID:         item.ID,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
			ReceiptID:  item.ReceiptID,
			CategoryID: item.CategoryID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			Total:      item.Total,
		}

		// Adiciona categoria se existir (APENAS ID e Nome - resposta leve!)
		if item.Category != nil {
			itemResponse.Category = &CategorySimple{
				ID:   item.Category.ID,
				Name: item.Category.Name,
			}
		}

		// Adiciona produto se existir (APENAS ID, Nome e Unidade)
		if item.Product != nil {
			itemResponse.Product = &ProductSimple{
				ID:    item.Product.ID,
				Name:  item.Product.Name,
				Unity: item.Product.Unity,
			}
		}

		items[i] = itemResponse
	}

	return ReceiptResponse{
		ID:         r.ID,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
		UserID:     r.UserID,
		StoreName:  r.StoreName,
		Date:       r.Date,
		Items:      items,
		Subtotal:   r.Subtotal,
		Discount:   r.Discount,
		Total:      r.Total,
		Currency:   r.Currency,
		Confidence: r.Confidence,
		Notes:      r.Notes,
	}
}

// ToResponse converte um ReceiptItem para um ReceiptItemResponse.
func (item *ReceiptItem) ToResponse() ReceiptItemResponse {
	itemResponse := ReceiptItemResponse{
		ID:         item.ID,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
		ReceiptID:  item.ReceiptID,
		CategoryID: item.CategoryID,
		ProductID:  item.ProductID,
		Quantity:   item.Quantity,
		UnitPrice:  item.UnitPrice,
		Total:      item.Total,
	}

	// Adiciona categoria se existir (APENAS ID e Nome)
	if item.Category != nil {
		itemResponse.Category = &CategorySimple{
			ID:   item.Category.ID,
			Name: item.Category.Name,
		}
	}

	// Adiciona produto se existir (APENAS ID, Nome e Unidade)
	if item.Product != nil {
		itemResponse.Product = &ProductSimple{
			ID:    item.Product.ID,
			Name:  item.Product.Name,
			Unity: item.Product.Unity,
		}
	}

	return itemResponse
}
