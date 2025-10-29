package handler

import (
	"net/http"

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
func GetItemByIDAndDateHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	date := ctx.Param("date")
	var item schemas.ReceiptItem
	// usa variável global db
	db.Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").Where("receipt_items.id = ? AND receipts.date = ?", id, date).First(&item)
	if item.ID == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Item não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, item)
}
