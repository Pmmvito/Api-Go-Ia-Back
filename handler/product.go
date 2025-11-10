package handler

import (
	"net/http"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
)

// UpdateProductRequest define a estrutura para atualizar um produto.
// Todos os campos são ponteiros para permitir atualizações parciais.
type UpdateProductRequest struct {
	Name  *string `json:"name"`
	Unity *string `json:"unity"`
}

// GetProductsHandler lida com a requisição para listar todos os produtos cadastrados no sistema.
// @Summary Listar todos os produtos
// @Description Lista todos os produtos cadastrados
// @Tags products
// @Produce json
// @Security BearerAuth
// @Success 200 {array} schemas.ProductResponse
// @Router /products [get]
func GetProductsHandler(ctx *gin.Context) {
	var products []schemas.Product
	db.Find(&products)
	ctx.JSON(http.StatusOK, products)
}

// GetProductByIDHandler lida com a requisição para buscar um produto pelo seu ID.
// @Summary Buscar produto por ID
// @Description Busca produto pelo ID
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do produto"
// @Success 200 {object} schemas.ProductResponse
// @Failure 404 {object} map[string]string
// @Router /products/{id} [get]
func GetProductByIDHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	var product schemas.Product

	if err := db.Where("id = ?", id).First(&product).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, product)
}

// GetProductsByDateHandler busca todos os produtos de uma data específica
// @Summary Buscar produtos por data
// @Description Retorna todos os produtos criados em uma data específica (YYYY-MM-DD)
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param date path string true "Data no formato YYYY-MM-DD"
// @Success 200 {array} schemas.ProductResponse
// @Router /products/date/{date} [get]
func GetProductsByDateHandler(ctx *gin.Context) {
	dateStr := ctx.Param("date")
	d, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use YYYY-MM-DD"})
		return
	}

	start := d
	end := d.Add(24 * time.Hour)

	var products []schemas.Product
	db.Where("created_at >= ? AND created_at < ?", start, end).Find(&products)
	ctx.JSON(http.StatusOK, products)
}

// @Summary Update a product
// @Description Update an existing product. All fields are optional.
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body UpdateProductRequest true "Product data to update"
// @Success 200 {object} schemas.ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [patch]
func UpdateProductHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Product ID is required")
		return
	}

	var request UpdateProductRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var product schemas.Product
	if err := db.First(&product, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Product not found")
		return
	}

	// Atualiza apenas os campos fornecidos
	if request.Name != nil {
		product.Name = *request.Name
	}
	if request.Unity != nil {
		product.Unity = *request.Unity
	}

	if err := db.Save(&product).Error; err != nil {
		logger.ErrorF("error updating product: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error updating product")
		return
	}

	ctx.JSON(http.StatusOK, product.ToResponse())
}

// @Summary Delete a product
// @Description Delete an existing product by its ID.
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [delete]
// @Summary Delete product
// @Description Soft delete a product and ALL its associated receipt items across all user's receipts (sets deleted_at timestamp). Warning: This will delete ALL occurrences of this product in ALL your receipts.
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{} "Product deleted successfully"
// @Failure 404 {object} ErrorResponse "Product not found or not owned by user"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /product/{id} [delete]
func DeleteProductHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Product ID is required")
		return
	}

	userID, _ := ctx.Get("user_id")

	// Verifica se o produto existe e pertence ao usuário (através de algum receipt_item)
	var product schemas.Product
	if err := db.Joins("JOIN receipt_items ON receipt_items.product_id = products.id").
		Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("products.id = ? AND receipts.user_id = ?", id, userID).
		First(&product).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Product not found or you don't have permission to delete it")
		return
	}

	// Inicia transação
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Soft delete APENAS dos receipt_items do usuário que usam este produto
	result := tx.Where("product_id = ? AND receipt_id IN (SELECT id FROM receipts WHERE user_id = ?)", product.ID, userID).
		Delete(&schemas.ReceiptItem{})
	if result.Error != nil {
		tx.Rollback()
		logger.ErrorF("error deleting receipt items: %v", result.Error.Error())
		sendError(ctx, http.StatusInternalServerError, "Error deleting receipt items")
		return
	}
	itemsDeleted := result.RowsAffected
	logger.InfoF("Soft deleted %d receipt items for product %d", itemsDeleted, product.ID)

	// 2. Verifica se ainda há items de outros usuários usando este produto
	var remainingItems int64
	if err := tx.Model(&schemas.ReceiptItem{}).Where("product_id = ?", product.ID).Count(&remainingItems).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error counting remaining items: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error checking remaining items")
		return
	}

	// 3. Se não há mais items usando este produto, pode deletar o produto
	if remainingItems == 0 {
		if err := tx.Delete(&product).Error; err != nil {
			tx.Rollback()
			logger.ErrorF("error deleting product: %v", err.Error())
			sendError(ctx, http.StatusInternalServerError, "Error deleting product")
			return
		}
		logger.InfoF("Product %d soft deleted (no more references)", product.ID)
	} else {
		logger.InfoF("Product %d NOT deleted (%d items from other users still reference it)", product.ID, remainingItems)
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		logger.ErrorF("error committing transaction: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error committing deletion")
		return
	}

	logger.InfoF("Product deletion completed: %d items deleted", itemsDeleted)
	ctx.JSON(http.StatusOK, gin.H{
		"message":        "All your items using this product were deleted successfully",
		"itemsDeleted":   itemsDeleted,
		"productDeleted": remainingItems == 0,
	})
}

// GetProductsByPeriodHandler busca todos os produtos dentro de um período de tempo
// @Summary Buscar produtos por período
// @Description Retorna todos os produtos criados entre as query params `start` e `end` (RFC3339 ou YYYY-MM-DD). Ambos são obrigatórios.
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param start query string true "Data/hora inicial (RFC3339 ou YYYY-MM-DD)"
// @Param end query string true "Data/hora final (RFC3339 ou YYYY-MM-DD)"
// @Success 200 {array} schemas.ProductResponse
// @Router /products/period [get]
func GetProductsByPeriodHandler(ctx *gin.Context) {
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
		// para incluir o dia final até 23:59:59
		end = e.Add(24 * time.Hour)
	}

	var products []schemas.Product
	db.Where("created_at >= ? AND created_at < ?", start, end).Find(&products)
	ctx.JSON(http.StatusOK, products)
}
