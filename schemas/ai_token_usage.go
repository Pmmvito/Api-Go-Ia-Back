package schemas

import (
	"time"

	"gorm.io/gorm"
)

// AITokenUsage rastreia o uso de tokens da IA por usuário
type AITokenUsage struct {
	gorm.Model
	UserID         uint      `json:"userId" gorm:"not null;index"`                           // FK para User
	User           *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`                // Relacionamento
	PromptTokens   int       `json:"promptTokens" gorm:"not null"`                           // Tokens usados no prompt
	ResponseTokens int       `json:"responseTokens" gorm:"not null"`                         // Tokens usados na resposta
	TotalTokens    int       `json:"totalTokens" gorm:"not null"`                            // Total de tokens
	AIModel        string    `json:"model" gorm:"size:100;column:model"`                     // Modelo usado (ex: gemini-2.5-flash)
	Endpoint       string    `json:"endpoint" gorm:"size:255"`                               // Endpoint que fez a chamada
	CostCents      float64   `json:"costCents"`                                              // Custo estimado em centavos
	UsedAt         time.Time `json:"usedAt" gorm:"not null;index;default:CURRENT_TIMESTAMP"` // Data/hora do uso
}

// AITokenUsageResponse representa a resposta da API
type AITokenUsageResponse struct {
	ID             uint      `json:"id"`
	UserID         uint      `json:"userId"`
	PromptTokens   int       `json:"promptTokens"`
	ResponseTokens int       `json:"responseTokens"`
	TotalTokens    int       `json:"totalTokens"`
	Model          string    `json:"model"`
	Endpoint       string    `json:"endpoint"`
	CostCents      float64   `json:"costCents"`
	UsedAt         time.Time `json:"usedAt"`
	CreatedAt      time.Time `json:"createdAt"`
}

// AITokenUsageSummary representa o resumo de uso de um usuário
type AITokenUsageSummary struct {
	UserID              uint    `json:"userId"`
	TotalPromptTokens   int     `json:"totalPromptTokens"`
	TotalResponseTokens int     `json:"totalResponseTokens"`
	TotalTokens         int     `json:"totalTokens"`
	TotalCostCents      float64 `json:"totalCostCents"`
	RequestCount        int     `json:"requestCount"`
}

// ToResponse converte AITokenUsage para AITokenUsageResponse
func (a *AITokenUsage) ToResponse() AITokenUsageResponse {
	return AITokenUsageResponse{
		ID:             a.ID,
		UserID:         a.UserID,
		PromptTokens:   a.PromptTokens,
		ResponseTokens: a.ResponseTokens,
		TotalTokens:    a.TotalTokens,
		Model:          a.AIModel,
		Endpoint:       a.Endpoint,
		CostCents:      a.CostCents,
		UsedAt:         a.UsedAt,
		CreatedAt:      a.CreatedAt,
	}
}
