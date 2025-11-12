package handler

import (
	"net/http"

	"github.com/Pmmvito/Golang-Api-Exemple/config"
	"github.com/gin-gonic/gin"
)

// @Summary Get AI Worker Pool status
// @Description Get current status and statistics of the AI Worker Pool (queue size, processing, etc)
// @Tags 🤖 AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Worker Pool status"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 503 {object} ErrorResponse "Worker Pool not initialized"
// @Router /ai-worker-pool/status [get]
func GetAIWorkerPoolStatusHandler(ctx *gin.Context) {
	workerPool := config.GetAIWorkerPool()
	if workerPool == nil {
		sendError(ctx, http.StatusServiceUnavailable, "Worker Pool não está inicializado")
		return
	}

	stats := workerPool.GetStats()
	avgTime := workerPool.GetAverageProcessingTime()

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Worker Pool status retrieved successfully",
		"status": gin.H{
			"isHealthy":         !workerPool.IsQueueFull(),
			"queueSize":         workerPool.GetQueueSize(),
			"queueCapacity":     workerPool.GetQueueCapacity(),
			"queueUsagePercent": float64(workerPool.GetQueueSize()) / float64(workerPool.GetQueueCapacity()) * 100,
		},
		"stats": gin.H{
			"totalProcessed":     stats.TotalProcessed,
			"totalFailed":        stats.TotalFailed,
			"totalQueued":        stats.TotalQueued,
			"currentInQueue":     stats.CurrentInQueue,
			"currentProcessing":  stats.CurrentProcessing,
			"averageTimeSeconds": avgTime.Seconds(),
			"successRate":        calculateSuccessRate(stats.TotalProcessed, stats.TotalFailed),
		},
		"limits": gin.H{
			"maxWorkers": 3,
			"rateLimit":  "10 requests per minute",
			"model":      "gemini-2.5-flash-preview-05-20",
			"tier":       "Free",
		},
		"recommendations": getRecommendations(workerPool),
	})
}

// calculateSuccessRate calcula taxa de sucesso em porcentagem
func calculateSuccessRate(processed, failed int64) float64 {
	total := processed + failed
	if total == 0 {
		return 100.0
	}
	return float64(processed) / float64(total) * 100.0
}

// getRecommendations retorna recomendações baseadas no estado do pool
func getRecommendations(pool *config.AIWorkerPool) []string {
	recommendations := []string{}

	queuePercent := float64(pool.GetQueueSize()) / float64(pool.GetQueueCapacity()) * 100

	if queuePercent > 80 {
		recommendations = append(recommendations, "⚠️ Fila está acima de 80% da capacidade. Considere aguardar alguns minutos antes de enviar novas requisições.")
	}

	if queuePercent > 95 {
		recommendations = append(recommendations, "🚨 Fila quase cheia! Novas requisições serão rejeitadas em breve.")
	}

	stats := pool.GetStats()
	if stats.TotalFailed > 0 {
		failRate := float64(stats.TotalFailed) / float64(stats.TotalProcessed+stats.TotalFailed) * 100
		if failRate > 10 {
			recommendations = append(recommendations, "⚠️ Taxa de falha elevada. Verifique logs do servidor.")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "✅ Sistema operando normalmente")
	}

	return recommendations
}
