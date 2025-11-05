package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
)

// recordAITokenUsageInternal é uma função interna para registrar uso de tokens da IA
// É chamada automaticamente pelos handlers que usam IA (como ScanQRCodeConfirmHandler)
func recordAITokenUsageInternal(userID uint, promptTokens, responseTokens int, model, endpoint string) error {
	// Calcula custo baseado nas variáveis de ambiente
	promptCostPer1K := getEnvFloat("GEMINI_PROMPT_COST_PER_1K_CENTS", 10.0)
	responseCostPer1K := getEnvFloat("GEMINI_RESPONSE_COST_PER_1K_CENTS", 20.0)

	promptCost := (float64(promptTokens) / 1000.0) * promptCostPer1K
	responseCost := (float64(responseTokens) / 1000.0) * responseCostPer1K
	totalCost := promptCost + responseCost

	usage := schemas.AITokenUsage{
		UserID:         userID,
		PromptTokens:   promptTokens,
		ResponseTokens: responseTokens,
		TotalTokens:    promptTokens + responseTokens,
		AIModel:        model,
		Endpoint:       endpoint,
		CostCents:      totalCost,
		UsedAt:         time.Now(),
	}

	return db.Create(&usage).Error
}

// getEnvFloat pega uma variável de ambiente como float64 ou retorna um valor padrão
func getEnvFloat(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

// checkAITokenLimit verifica se o usuário ainda tem limite disponível de tokens
func checkAITokenLimit(userID uint) error {
	// Pega limite do .env (0 = sem limite)
	limitStr := os.Getenv("AI_TOKEN_LIMIT_PER_USER")
	if limitStr == "" || limitStr == "0" {
		// Sem limite configurado
		return nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		// Limite inválido ou zero = sem limite
		return nil
	}

	// Busca uso total do usuário
	var summary schemas.AITokenUsageSummary
	err = db.Model(&schemas.AITokenUsage{}).
		Where("user_id = ?", userID).
		Select("SUM(total_tokens) as total_tokens").
		Scan(&summary).Error

	if err != nil {
		logger.ErrorF("Erro ao verificar limite de tokens: %v", err)
		return nil // Em caso de erro, permite continuar
	}

	if summary.TotalTokens >= limit {
		return fmt.Errorf("limite de tokens excedido. Usado: %d, Limite: %d", summary.TotalTokens, limit)
	}

	logger.InfoF("✓ Token limit check - User %d: %d/%d tokens used", userID, summary.TotalTokens, limit)
	return nil
}

// GetAITokenUsageHandler retorna o histórico de uso de tokens da IA do usuário autenticado
// @Summary Obter histórico de uso de tokens
// @Description Retorna todo o histórico de uso de tokens da IA do usuário autenticado
// @Tags ai-usage
// @Produce json
// @Security BearerAuth
// @Success 200 {array} schemas.AITokenUsageResponse
// @Failure 500 {object} map[string]string
// @Router /ai-usage [get]
func GetAITokenUsageHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var usages []schemas.AITokenUsage
	if err := db.Where("user_id = ?", userID).Order("used_at DESC").Find(&usages).Error; err != nil {
		logger.ErrorF("Erro ao buscar uso de tokens: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar histórico"})
		return
	}

	// Converte para response
	var responses []schemas.AITokenUsageResponse
	for _, usage := range usages {
		responses = append(responses, usage.ToResponse())
	}

	ctx.JSON(http.StatusOK, responses)
}

// GetAITokenUsageSummaryHandler retorna o resumo de uso de tokens do usuário
// @Summary Obter resumo de uso de tokens
// @Description Retorna estatísticas consolidadas de uso de tokens da IA do usuário autenticado
// @Tags ai-usage
// @Produce json
// @Security BearerAuth
// @Success 200 {object} schemas.AITokenUsageSummary
// @Failure 500 {object} map[string]string
// @Router /ai-usage/summary [get]
func GetAITokenUsageSummaryHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var summary schemas.AITokenUsageSummary
	summary.UserID = userID.(uint)

	// Query agregada
	err := db.Model(&schemas.AITokenUsage{}).
		Where("user_id = ?", userID).
		Select("SUM(prompt_tokens) as total_prompt_tokens, SUM(response_tokens) as total_response_tokens, SUM(total_tokens) as total_tokens, SUM(cost_cents) as total_cost_cents, COUNT(*) as request_count").
		Scan(&summary).Error

	if err != nil {
		logger.ErrorF("Erro ao buscar resumo de uso: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar resumo"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}
