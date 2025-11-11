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
// GetProductsHandler lida com a requisição para listar todos os produtos cadastrados no sistema.
// Retorna apenas produtos que possuem items ativos (não deletados) do usuário
// @Summary Listar todos os produtos
// @Description Lista todos os produtos que o usuário possui em suas notas fiscais ativas
// @Tags products
// @Produce json
// @Security BearerAuth
// @Success 200 {array} schemas.ProductResponse
// @Router /products [get]
func GetProductsHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var products []schemas.Product
	// Busca apenas produtos que tenham items ativos (não deletados) em receipts do usuário
	err := db.Distinct("products.*").
		Joins("INNER JOIN receipt_items ON receipt_items.product_id = products.id").
		Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.user_id = ?", userID).
		Find(&products).Error

	if err != nil {
		logger.ErrorF("error getting products: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error getting products")
		return
	}

	ctx.JSON(http.StatusOK, products)
}

// GetProductByIDHandler lida com a requisição para buscar um produto pelo seu ID.
// Verifica se o produto pertence ao usuário através de receipt_items ativos
// @Summary Buscar produto por ID
// @Description Busca produto pelo ID (apenas se o usuário tiver este produto em alguma nota ativa)
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do produto"
// @Success 200 {object} schemas.ProductResponse
// @Failure 404 {object} map[string]string
// @Router /products/{id} [get]
func GetProductByIDHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	userID, _ := ctx.Get("user_id")

	var product schemas.Product
	// Busca apenas se o produto estiver em algum item ativo do usuário
	err := db.Joins("INNER JOIN receipt_items ON receipt_items.product_id = products.id").
		Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("products.id = ? AND receipts.user_id = ?", id, userID).
		First(&product).Error

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado ou você não tem acesso a ele"})
		return
	}

	ctx.JSON(http.StatusOK, product)
}

// GetProductsByDateHandler busca todos os produtos de uma data específica
// Retorna produtos que foram comprados (têm items) na data especificada
// @Summary Buscar produtos por data
// @Description Retorna todos os produtos comprados em uma data específica (YYYY-MM-DD) através de notas fiscais
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param date path string true "Data no formato YYYY-MM-DD"
// @Success 200 {array} schemas.ProductResponse
// @Router /products/date/{date} [get]
func GetProductsByDateHandler(ctx *gin.Context) {
	dateStr := ctx.Param("date")
	userID, _ := ctx.Get("user_id")

	_, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use YYYY-MM-DD"})
		return
	}

	var products []schemas.Product
	// Busca produtos através de receipt_items de receipts da data específica
	err = db.Distinct("products.*").
		Joins("INNER JOIN receipt_items ON receipt_items.product_id = products.id").
		Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.user_id = ? AND receipts.date = ?", userID, dateStr).
		Find(&products).Error

	if err != nil {
		logger.ErrorF("error getting products by date: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error getting products")
		return
	}

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

	userID, _ := ctx.Get("user_id")

	var request UpdateProductRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Verifica se o produto existe e pertence ao usuário (através de algum receipt_item)
	var product schemas.Product
	if err := db.Joins("JOIN receipt_items ON receipt_items.product_id = products.id").
		Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("products.id = ? AND receipts.user_id = ?", id, userID).
		First(&product).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Produto não encontrado ou você não tem permissão para atualizá-lo")
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
		sendError(ctx, http.StatusInternalServerError, "Erro ao atualizar produto no banco de dados. Por favor, tente novamente")
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

	// 1. Buscar IDs dos receipts do usuário usando GORM
	var userReceipts []schemas.Receipt
	if err := tx.Where("user_id = ?", userID).Select("id").Find(&userReceipts).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error finding user receipts: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error finding user receipts")
		return
	}

	var receiptIDs []uint
	for _, receipt := range userReceipts {
		receiptIDs = append(receiptIDs, receipt.ID)
	}

	// 2. Soft delete APENAS dos receipt_items do usuário que usam este produto
	var itemsDeleted int64
	if len(receiptIDs) > 0 {
		result := tx.Where("product_id = ? AND receipt_id IN ?", product.ID, receiptIDs).
			Delete(&schemas.ReceiptItem{})
		if result.Error != nil {
			tx.Rollback()
			logger.ErrorF("error deleting receipt items: %v", result.Error.Error())
			sendError(ctx, http.StatusInternalServerError, "Error deleting receipt items")
			return
		}
		itemsDeleted = result.RowsAffected
		logger.InfoF("Soft deleted %d receipt items for product %d", itemsDeleted, product.ID)
	}

	// 3. Verifica se ainda há items de outros usuários usando este produto
	var remainingItems int64
	if err := tx.Model(&schemas.ReceiptItem{}).Where("product_id = ?", product.ID).Count(&remainingItems).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error counting remaining items: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error checking remaining items")
		return
	}

	// 4. Se não há mais items usando este produto, pode deletar o produto
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
// Retorna produtos que foram comprados (têm items) no período especificado
// @Summary Buscar produtos por período
// @Description Retorna todos os produtos comprados entre as query params `start` e `end` (YYYY-MM-DD). Ambos são obrigatórios.
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param start query string true "Data inicial (YYYY-MM-DD)"
// @Param end query string true "Data final (YYYY-MM-DD)"
// @Success 200 {array} schemas.ProductResponse
// @Router /products/period [get]
func GetProductsByPeriodHandler(ctx *gin.Context) {
	startStr := ctx.Query("start")
	endStr := ctx.Query("end")
	userID, _ := ctx.Get("user_id")

	if startStr == "" || endStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Query params 'start' e 'end' são obrigatórios (formato YYYY-MM-DD)"})
		return
	}

	// Valida formato das datas
	_, err1 := time.ParseInLocation("2006-01-02", startStr, time.Local)
	_, err2 := time.ParseInLocation("2006-01-02", endStr, time.Local)
	if err1 != nil || err2 != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use YYYY-MM-DD"})
		return
	}

	var products []schemas.Product
	// Busca produtos através de receipt_items de receipts no período
	err := db.Distinct("products.*").
		Joins("INNER JOIN receipt_items ON receipt_items.product_id = products.id").
		Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.user_id = ? AND receipts.date >= ? AND receipts.date <= ?", userID, startStr, endStr).
		Find(&products).Error

	if err != nil {
		logger.ErrorF("error getting products by period: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error getting products")
		return
	}

	ctx.JSON(http.StatusOK, products)
}
