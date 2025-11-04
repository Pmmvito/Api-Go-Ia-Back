package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// sendError envia uma resposta de erro JSON com um código de status e mensagem específicos.
func sendError(ctx *gin.Context, code int, msg string) {
	ctx.Header("Content-type", "application/json")
	ctx.JSON(code, gin.H{
		"message":   msg,
		"errorCode": code,
	})
}

// sendSuccess envia uma resposta de sucesso JSON com os dados fornecidos.
func sendSucces(ctx *gin.Context, op string, data interface{}) {
	ctx.Header("Content-type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("operation from handler: %s successfull", op),
		"data":    data,
	})
}

// ErrorResponse define a estrutura para respostas de erro da API.
type ErrorResponse struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}
