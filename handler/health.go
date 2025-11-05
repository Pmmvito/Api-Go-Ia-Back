package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse define a estrutura da resposta do health check
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
}

// HealthHandler retorna o status de saúde da API
// @Summary Health check da API
// @Description Retorna o status de saúde da API para monitoramento
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func HealthHandler(ctx *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Service:   "TCC API - Gestão de Notas Fiscais",
	}

	ctx.JSON(http.StatusOK, response)
}
