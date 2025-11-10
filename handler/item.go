package handler

import (
	"net/http"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
)

// UpdateItemRequest define a estrutura para atualizar um item de recibo.
// Todos os campos são ponteiros para permitir atualizações parciais.
type UpdateItemRequest struct {
	CategoryID *uint    `json:"categoryId"`
	ProductID  *uint    `json:"productId"`
	Quantity   *float64 `json:"quantity"`
	UnitPrice  *float64 `json:"unitPrice"`
	Total      *float64 `json:"total"`
}

// GetItemsHandler lida com a requisição para listar todos os itens de recibos do usuário autenticado.
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
	// Utiliza a conexão de banco de dados global 'db'
	db.Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").Where("receipts.user_id = ?", userID).Find(&items)
	ctx.JSON(http.StatusOK, items)
}

// GetItemByIDHandler lida com a requisição para buscar um item de recibo pelo seu ID.
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
	// Utiliza a conexão de banco de dados global 'db'
	if err := db.Where("id = ?", id).First(&item).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Item não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

// GetItemsByDateHandler lida com a requisição para listar itens de recibos do usuário autenticado por uma data específica.
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
	// Utiliza a conexão de banco de dados global 'db'
	db.Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").Where("receipts.user_id = ? AND receipts.date = ?", userID, date).Find(&items)
	ctx.JSON(http.StatusOK, items)
}

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

// @Summary Update an item
// @Description Update an existing receipt item. All fields are optional.
// @Tags items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Param request body UpdateItemRequest true "Item data to update"
// @Success 200 {object} schemas.ReceiptItemResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /item/{id} [patch]
func UpdateItemHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Item ID is required")
		return
	}

	var request UpdateItemRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var item schemas.ReceiptItem
	if err := db.First(&item, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Item not found")
		return
	}

	// Atualiza apenas os campos fornecidos
	if request.CategoryID != nil {
		item.CategoryID = *request.CategoryID
	}
	if request.ProductID != nil {
		item.ProductID = *request.ProductID
	}
	if request.Quantity != nil {
		item.Quantity = *request.Quantity
	}
	if request.UnitPrice != nil {
		item.UnitPrice = *request.UnitPrice
	}
	if request.Total != nil {
		item.Total = *request.Total
	}

	if err := db.Save(&item).Error; err != nil {
		logger.ErrorF("error updating item: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error updating item")
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// @Summary Delete an item
// @Description Delete an existing receipt item by its ID.
// @Tags items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /item/{id} [delete]
// @Summary Delete receipt item
// @Description Soft delete a receipt item (sets deleted_at timestamp). Note: The associated product is NOT deleted as it may be referenced by other items.
// @Tags items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Receipt Item ID"
// @Success 200 {object} map[string]interface{} "Item deleted successfully"
// @Failure 404 {object} ErrorResponse "Item not found"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /item/{id} [delete]
func DeleteItemHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Item ID is required")
		return
	}

	userID, _ := ctx.Get("user_id")

	// Busca o item através do recibo do usuário
	var item schemas.ReceiptItem
	if err := db.Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipt_items.id = ? AND receipts.user_id = ?", id, userID).
		First(&item).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Item not found")
		return
	}

	// Soft delete do item
	// NOTA: NÃO deletamos o produto pois ele pode estar sendo usado por outros items
	if err := db.Delete(&item).Error; err != nil {
		logger.ErrorF("error deleting item: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error deleting item")
		return
	}

	logger.InfoF("Item %s soft deleted successfully", id)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Item deleted successfully",
	})
}
