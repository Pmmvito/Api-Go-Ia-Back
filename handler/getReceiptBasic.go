package handler

import (
	"net/http"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
)

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