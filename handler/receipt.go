package handler

import (
	"net/http"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
func DeleteReceiptHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Receipt ID is required")
		return
	}

	var receipt schemas.Receipt
	if err := db.First(&receipt, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Receipt not found")
		return
	}

	if err := db.Delete(&receipt).Error; err != nil {
		logger.ErrorF("error deleting receipt: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error deleting receipt")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Receipt deleted successfully"})
}

// @Summary Update a receipt
// @Description Update an existing receipt. All fields are optional.
// @Tags notasfiscais
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Receipt ID"
// @Param request body UpdateReceiptRequest true "Receipt data to update"
// @Success 200 {object} schemas.ReceiptResponse
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

	ctx.JSON(http.StatusOK, receipt.ToResponse())
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
// @Success 200 {object} schemas.ReceiptResponse
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
	ctx.JSON(http.StatusOK, receipt.ToResponse())
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
