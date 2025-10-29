package handler

import (
	"net/http"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
)

// GET /items - Lista todos os itens de recibos do usuário autenticado
// @Summary Listar todos os itens
// @Description Lista todos os itens de recibos do usuário autenticado
// @Tags items
// @Produce json
// @Security BearerAuth
// @Success 200 {array} schemas.ReceiptItemResponse
// @Router /items [get]
func GetItemsHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	var items []schemas.ReceiptItem
	// usa variável global db
	db.Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").Where("receipts.user_id = ?", userID).Find(&items)
	ctx.JSON(http.StatusOK, items)
}

// GET /item/:id - Busca item por ID
// @Summary Buscar item por ID
// @Description Busca um item pelo ID
// @Tags items
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do item"
// @Success 200 {object} schemas.ReceiptItemResponse
// @Failure 404 {object} map[string]string
// @Router /item/{id} [get]
func GetItemByIDHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	var item schemas.ReceiptItem
	// usa variável global db
	if err := db.Where("id = ?", id).First(&item).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Item não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

// GET /items/date/:date - Lista itens por data de recibo
// @Summary Listar itens por data de recibo
// @Description Lista itens de recibos do usuário autenticado por data de recibo
// @Tags items
// @Produce json
// @Security BearerAuth
// @Param date path string true "Data (YYYY-MM-DD)"
// @Success 200 {array} schemas.ReceiptItemResponse
// @Router /items/date/{date} [get]
func GetItemsByDateHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	date := ctx.Param("date")
	var items []schemas.ReceiptItem
	// usa variável global db
	db.Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").Where("receipts.user_id = ? AND receipts.date = ?", userID, date).Find(&items)
	ctx.JSON(http.StatusOK, items)
}

// GET /item/:id/date/:date - Busca item por ID e data de recibo
// @Summary Buscar item por ID e data de recibo
// @Description Busca item por ID e data de recibo
// @Tags items
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do item"
// @Param date path string true "Data (YYYY-MM-DD)"
// @Success 200 {object} schemas.ReceiptItemResponse
// @Failure 404 {object} map[string]string
// @Router /item/{id}/date/{date} [get]

// GetItemsByPeriodHandler busca itens por período de recibos do usuário autenticado
// @Summary Listar itens por período
// @Description Lista itens de recibos do usuário autenticado entre query params `start` e `end` (RFC3339 ou YYYY-MM-DD). Ambos obrigatórios.
// @Tags items
// @Produce json
// @Security BearerAuth
// @Param start query string true "Data/hora inicial (RFC3339 ou YYYY-MM-DD)"
// @Param end query string true "Data/hora final (RFC3339 ou YYYY-MM-DD)"
// @Success 200 {array} schemas.ReceiptItemResponse
// @Router /items/period [get]
func GetItemsByPeriodHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	startStr := ctx.Query("start")
	endStr := ctx.Query("end")

	if startStr == "" || endStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Query params 'start' e 'end' são obrigatórios"})
		return
	}

	// Tenta parsear como RFC3339, se falhar tenta YYYY-MM-DD
	start, err1 := time.Parse(time.RFC3339, startStr)
	end, err2 := time.Parse(time.RFC3339, endStr)
	if err1 != nil || err2 != nil {
		s, errS := time.ParseInLocation("2006-01-02", startStr, time.Local)
		e, errE := time.ParseInLocation("2006-01-02", endStr, time.Local)
		if errS != nil || errE != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use RFC3339 ou YYYY-MM-DD"})
			return
		}
		start = s
		end = e.Add(24 * time.Hour)
	}

	var items []schemas.ReceiptItem
	db.Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").Where("receipts.user_id = ? AND receipts.date >= ? AND receipts.date < ?", userID, start, end).Find(&items)
	ctx.JSON(http.StatusOK, items)
}
