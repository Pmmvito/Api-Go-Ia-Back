package handler

import (
	"net/http"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
