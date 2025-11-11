package handler

import (
	"net/http"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateReceiptItemRequest define a estrutura para um item de nota fiscal na criação manual.
type CreateReceiptItemRequest struct {
	ProductName string  `json:"productName" binding:"required" example:"Arroz Integral"`
	ProductUnit string  `json:"productUnit" binding:"required" example:"kg"`
	CategoryID  uint    `json:"categoryId" binding:"required" example:"1"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0" example:"2.5"`
	UnitPrice   float64 `json:"unitPrice" binding:"required,gt=0" example:"15.90"`
	Total       float64 `json:"total" binding:"required,gt=0" example:"39.75"`
}

// CreateReceiptRequest define a estrutura para criar uma nota fiscal manualmente.
type CreateReceiptRequest struct {
	StoreName string                     `json:"storeName" binding:"required" example:"Supermercado Silva"`
	Date      string                     `json:"date" binding:"required" example:"2024-11-11"`
	Items     []CreateReceiptItemRequest `json:"items" binding:"required,min=1"`
	Subtotal  float64                    `json:"subtotal" example:"100.00"`
	Discount  float64                    `json:"discount" example:"5.00"`
	Total     float64                    `json:"total" binding:"required,gt=0" example:"95.00"`
	Currency  string                     `json:"currency" example:"BRL"`
	Notes     string                     `json:"notes" example:"Compra mensal"`
}

// UpdateReceiptRequest define a estrutura para atualizar um recibo.
// Todos os campos são ponteiros para permitir atualizações parciais.
type UpdateReceiptRequest struct {
	StoreName *string  `json:"storeName"`
	Date      *string  `json:"date"`
	Subtotal  *float64 `json:"subtotal"`
	Discount  *float64 `json:"discount"`
	Total     *float64 `json:"total"`
}

// GetReceiptsHandler lida com a requisição para listar todos os recibos do usuário autenticado.
// @Summary Listar todos os recibos
// @Description Lista todos os recibos do usuário autenticado
// @Tags notasfiscais
// @Produce json
// @Security BearerAuth
// @Success 200 {array} schemas.ReceiptSummary
// @Router /receipts [get]
func GetReceiptsHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	var receipts []schemas.Receipt
	// Utiliza a conexão de banco de dados global 'db' com preload otimizado para evitar queries N+1.
	db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Preload("Items.Category").Preload("Items.Product").Where("user_id = ?", userID).Order("date DESC").Find(&receipts)

	// Converte para uma resposta otimizada para listagens.
	summaries := make([]schemas.ReceiptSummary, len(receipts))
	for i, receipt := range receipts {
		summaries[i] = receipt.ToSummary()
	}

	ctx.JSON(http.StatusOK, summaries)
}

// GetReceiptsBasicHandler lida com a requisição para listar todos os recibos do usuário de forma simplificada.
// @Summary Listar recibos básicos
// @Description Lista todos os recibos do usuário autenticado (versão ultra-simplificada para seleção)
// @Tags notasfiscais-basic
// @Produce json
// @Security BearerAuth
// @Success 200 {array} schemas.ReceiptBasic
// @Router /receipts-basic [get]
func GetReceiptsBasicHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	var receipts []schemas.Receipt
	// Query otimizada - seleciona apenas os campos necessários, sem preloading de relacionamentos complexos.
	db.Select("id, store_name, date, total, currency, user_id").
		Where("user_id = ?", userID).
		Order("date DESC").
		Find(&receipts)

	// Para cada recibo, conta os itens associados para preencher o campo ItemCount.
	basics := make([]schemas.ReceiptBasic, len(receipts))
	for i, receipt := range receipts {
		var count int64
		db.Model(&schemas.ReceiptItem{}).Where("receipt_id = ?", receipt.ID).Count(&count)

		basics[i] = schemas.ReceiptBasic{
			ID:        receipt.ID,
			StoreName: receipt.StoreName,
			Date:      receipt.Date,
			ItemCount: int(count),
			Total:     receipt.Total,
			Currency:  receipt.Currency,
		}
	}

	ctx.JSON(http.StatusOK, basics)
}

// @Summary Create a receipt manually
// @Description Create a new receipt with items manually (without QR code scanning)
// @Tags notasfiscais
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateReceiptRequest true "Receipt data"
// @Success 201 {object} map[string]interface{} "Receipt created successfully"
// @Failure 400 {object} ErrorResponse "Dados inválidos | Categoria não encontrada ou não pertence ao usuário"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Erro ao criar nota fiscal. Por favor, tente novamente"
// @Router /receipt [post]
func CreateReceiptHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	// Valida o body da requisição
	var request CreateReceiptRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		sendError(ctx, http.StatusBadRequest, "Dados inválidos. Verifique os campos obrigatórios: storeName, date, items (com productName, productUnit, categoryId, quantity, unitPrice, total) e total")
		return
	}

	// Valida se o usuário tem acesso às categorias fornecidas
	categoryIDs := make([]uint, len(request.Items))
	for i, item := range request.Items {
		categoryIDs[i] = item.CategoryID
	}

	var categoryCount int64
	db.Model(&schemas.Category{}).
		Where("id IN ? AND user_id = ?", categoryIDs, userID).
		Count(&categoryCount)

	if int(categoryCount) != len(categoryIDs) {
		sendError(ctx, http.StatusBadRequest, "Uma ou mais categorias não foram encontradas ou não pertencem ao usuário autenticado")
		return
	}

	// Define valores padrão
	if request.Currency == "" {
		request.Currency = "BRL"
	}

	// Inicia transação para garantir consistência
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Cria o receipt
	receipt := schemas.Receipt{
		UserID:     userID.(uint),
		StoreName:  request.StoreName,
		Date:       request.Date,
		Subtotal:   request.Subtotal,
		Discount:   request.Discount,
		Total:      request.Total,
		Currency:   request.Currency,
		Confidence: 1.0, // Criação manual = 100% de confiança
		Notes:      request.Notes,
	}

	if err := tx.Create(&receipt).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error creating receipt: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao criar nota fiscal. Por favor, tente novamente")
		return
	}

	// Processa cada item
	for _, itemReq := range request.Items {
		// Busca ou cria o produto
		var product schemas.Product
		if err := tx.Where("name = ? AND unity = ?", itemReq.ProductName, itemReq.ProductUnit).
			First(&product).Error; err != nil {
			// Produto não existe, cria um novo
			product = schemas.Product{
				Name:  itemReq.ProductName,
				Unity: itemReq.ProductUnit,
			}
			if err := tx.Create(&product).Error; err != nil {
				tx.Rollback()
				logger.ErrorF("error creating product: %v", err.Error())
				sendError(ctx, http.StatusInternalServerError, "Erro ao criar produto. Por favor, tente novamente")
				return
			}
		}

		// Cria o item do recibo
		receiptItem := schemas.ReceiptItem{
			ReceiptID:  receipt.ID,
			CategoryID: itemReq.CategoryID,
			ProductID:  product.ID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  itemReq.UnitPrice,
			Total:      itemReq.Total,
		}

		if err := tx.Create(&receiptItem).Error; err != nil {
			tx.Rollback()
			logger.ErrorF("error creating receipt item: %v", err.Error())
			sendError(ctx, http.StatusInternalServerError, "Erro ao criar item da nota fiscal. Por favor, tente novamente")
			return
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error committing transaction: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao salvar nota fiscal. Por favor, tente novamente")
		return
	}

	// Busca o receipt completo com todos os relacionamentos
	var completeReceipt schemas.Receipt
	db.Preload("Items.Category").
		Preload("Items.Product").
		First(&completeReceipt, receipt.ID)

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Nota fiscal criada com sucesso",
		"data":    completeReceipt.ToResponse(),
	})
}

// @Summary Delete a receipt
// @Description Delete an existing receipt by its ID.
// @Tags notasfiscais
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Receipt ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /receipt/{id} [delete]
// @Summary Delete a receipt
// @Description Soft delete a receipt and all its related items and products (sets deleted_at timestamp)
// @Tags notasfiscais
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Receipt ID"
// @Success 200 {object} map[string]interface{} "Receipt deleted successfully"
// @Failure 404 {object} ErrorResponse "Receipt not found"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /receipt/{id} [delete]
func DeleteReceiptHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Receipt ID is required")
		return
	}

	userID, _ := ctx.Get("user_id")

	// Busca o recibo do usuário
	var receipt schemas.Receipt
	if err := db.Where("user_id = ?", userID).First(&receipt, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Receipt not found")
		return
	}

	// Inicia transação para garantir consistência
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Busca todos os itens do recibo
	var receiptItems []schemas.ReceiptItem
	if err := tx.Where("receipt_id = ?", receipt.ID).Find(&receiptItems).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error finding receipt items: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error deleting receipt items")
		return
	}

	// 2. Soft delete dos itens do recibo
	// NOTA: NÃO deletamos produtos pois eles podem estar sendo usados por outros items
	if err := tx.Where("receipt_id = ?", receipt.ID).Delete(&schemas.ReceiptItem{}).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error deleting receipt items: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error deleting receipt items")
		return
	}
	logger.InfoF("Soft deleted %d receipt items", len(receiptItems))

	// 3. Soft delete do recibo
	if err := tx.Delete(&receipt).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error deleting receipt: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error deleting receipt")
		return
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		logger.ErrorF("error committing transaction: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error committing deletion")
		return
	}

	logger.InfoF("Receipt %s and all related items soft deleted successfully", id)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Receipt and all related items deleted successfully",
		"details": gin.H{
			"receiptId":    receipt.ID,
			"itemsDeleted": len(receiptItems),
		},
	})
}

// @Summary Update a receipt
// @Description Update an existing receipt. All fields are optional.
// @Tags notasfiscais
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Receipt ID"
// @Param request body UpdateReceiptRequest true "Receipt data to update"
// @Success 200 {object} schemas.ReceiptSummary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /receipt/{id} [patch]
func UpdateReceiptHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Receipt ID is required")
		return
	}

	var request UpdateReceiptRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var receipt schemas.Receipt
	if err := db.First(&receipt, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Receipt not found")
		return
	}

	// Atualiza apenas os campos fornecidos
	if request.StoreName != nil {
		receipt.StoreName = *request.StoreName
	}
	if request.Date != nil {
		receipt.Date = *request.Date
	}
	if request.Subtotal != nil {
		receipt.Subtotal = *request.Subtotal
	}
	if request.Discount != nil {
		receipt.Discount = *request.Discount
	}
	if request.Total != nil {
		receipt.Total = *request.Total
	}

	if err := db.Save(&receipt).Error; err != nil {
		logger.ErrorF("error updating receipt: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error updating receipt")
		return
	}

	// Recarrega o receipt com os relacionamentos para retornar o summary completo
	db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Preload("Items.Category").Preload("Items.Product").First(&receipt, id)

	ctx.JSON(http.StatusOK, receipt.ToSummary())
}

// GetReceiptsBasicByPeriodHandler lista recibos básicos por período
// @Summary Listar recibos básicos por período
// @Description Lista recibos básicos do usuário autenticado por período de datas
// @Tags notasfiscais-basic
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Data inicial (YYYY-MM-DD)"
// @Param end_date query string true "Data final (YYYY-MM-DD)"
// @Success 200 {array} schemas.ReceiptBasic
// @Router /receipts-basic/period [get]
func GetReceiptsBasicByPeriodHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	if startDate == "" || endDate == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date e end_date são obrigatórios"})
		return
	}

	var receipts []schemas.Receipt
	// Query otimizada
	db.Select("id, store_name, date, total, currency, user_id").
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Order("date DESC").Find(&receipts)

	// Conta items para cada receipt
	basics := make([]schemas.ReceiptBasic, len(receipts))
	for i, receipt := range receipts {
		var count int64
		db.Model(&schemas.ReceiptItem{}).Where("receipt_id = ?", receipt.ID).Count(&count)

		basics[i] = schemas.ReceiptBasic{
			ID:        receipt.ID,
			StoreName: receipt.StoreName,
			Date:      receipt.Date,
			ItemCount: int(count),
			Total:     receipt.Total,
			Currency:  receipt.Currency,
		}
	}

	ctx.JSON(http.StatusOK, basics)
}

// GetReceiptsBasicByDateHandler lista recibos básicos por data específica
// @Summary Listar recibos básicos por data específica
// @Description Lista recibos básicos do usuário autenticado por data específica (YYYY-MM-DD)
// @Tags notasfiscais-basic
// @Produce json
// @Security BearerAuth
// @Param date path string true "Data (YYYY-MM-DD)"
// @Success 200 {array} schemas.ReceiptBasic
// @Router /receipts-basic/date/{date} [get]
func GetReceiptsBasicByDateHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	date := ctx.Param("date")
	var receipts []schemas.Receipt
	// Query otimizada
	db.Select("id, store_name, date, total, currency, user_id").
		Where("user_id = ? AND date = ?", userID, date).
		Order("date DESC").Find(&receipts)

	// Conta items para cada receipt
	basics := make([]schemas.ReceiptBasic, len(receipts))
	for i, receipt := range receipts {
		var count int64
		db.Model(&schemas.ReceiptItem{}).Where("receipt_id = ?", receipt.ID).Count(&count)

		basics[i] = schemas.ReceiptBasic{
			ID:        receipt.ID,
			StoreName: receipt.StoreName,
			Date:      receipt.Date,
			ItemCount: int(count),
			Total:     receipt.Total,
			Currency:  receipt.Currency,
		}
	}

	ctx.JSON(http.StatusOK, basics)
}

// GetReceiptByIDHandler lida com a requisição para buscar um recibo pelo seu ID.
// @Summary Buscar recibo por ID
// @Description Busca um recibo pelo ID do usuário autenticado
// @Tags notasfiscais
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do recibo"
// @Success 200 {object} schemas.ReceiptSummary
// @Failure 404 {object} map[string]string
// @Router /receipt/{id} [get]
func GetReceiptByIDHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	id := ctx.Param("id")
	var receipt schemas.Receipt
	// Utiliza a conexão de banco de dados global 'db' com preload otimizado.
	if err := db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Preload("Items.Category").Preload("Items.Product").Where("id = ? AND user_id = ?", id, userID).First(&receipt).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Recibo não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, receipt.ToSummary())
}

// GetReceiptsByDateHandler busca recibos por data específica
// @Summary Buscar recibos por data específica
// @Description Lista recibos do usuário autenticado por data específica (YYYY-MM-DD)
// @Tags notasfiscais
// @Produce json
// @Security BearerAuth
// @Param date path string true "Data (YYYY-MM-DD)"
// @Success 200 {array} schemas.ReceiptSummary
// @Router /receipts/date/{date} [get]
func GetReceiptsByDateHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	date := ctx.Param("date")
	var receipts []schemas.Receipt
	// usa variável global db - preload otimizado
	db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Preload("Items.Category").Preload("Items.Product").
		Where("user_id = ? AND date = ?", userID, date).
		Order("date DESC").Find(&receipts)

	// Converte para resposta otimizada
	summaries := make([]schemas.ReceiptSummary, len(receipts))
	for i, receipt := range receipts {
		summaries[i] = receipt.ToSummary()
	}

	ctx.JSON(http.StatusOK, summaries)
}

// GetReceiptsByPeriodHandler busca recibos por período
// @Summary Buscar recibos por período
// @Description Lista recibos do usuário autenticado por período de datas
// @Tags notasfiscais
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Data inicial (YYYY-MM-DD)"
// @Param end_date query string true "Data final (YYYY-MM-DD)"
// @Success 200 {array} schemas.ReceiptSummary
// @Router /receipts/period [get]
func GetReceiptsByPeriodHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	if startDate == "" || endDate == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date e end_date são obrigatórios"})
		return
	}

	var receipts []schemas.Receipt
	// usa variável global db - preload otimizado
	db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Preload("Items.Category").Preload("Items.Product").
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Order("date DESC").Find(&receipts)

	// Converte para resposta otimizada
	summaries := make([]schemas.ReceiptSummary, len(receipts))
	for i, receipt := range receipts {
		summaries[i] = receipt.ToSummary()
	}

	ctx.JSON(http.StatusOK, summaries)
}
