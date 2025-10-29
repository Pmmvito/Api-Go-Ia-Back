package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ScanQRCodePreviewRequest representa o request de preview
type ScanQRCodePreviewRequest struct {
	QRCodeURL string `json:"qrCodeUrl" binding:"required"` // URL do QR Code da NFC-e
}

// PreviewItem representa um item no preview (antes de salvar)
type PreviewItem struct {
	TempID      int     `json:"tempId"`            // ID temporário para edição
	Description string  `json:"description"`       // Nome do produto
	Quantity    float64 `json:"quantity"`          // Quantidade
	Unit        string  `json:"unit"`              // Unidade (kg, un, ml, etc)
	UnitPrice   float64 `json:"unitPrice"`         // Preço unitário
	Total       float64 `json:"total"`             // Total do item
	Deleted     bool    `json:"deleted,omitempty"` // Se true, item será ignorado (usado no confirm)
}

// PreviewReceiptData representa os dados da nota para preview
type PreviewReceiptData struct {
	StoreName  string        `json:"storeName"`  // Nome do estabelecimento
	Date       string        `json:"date"`       // Data da compra
	Items      []PreviewItem `json:"items"`      // Items extraídos
	ItemsCount int           `json:"itemsCount"` // Total de items
	Subtotal   float64       `json:"subtotal"`   // Subtotal
	Discount   float64       `json:"discount"`   // Desconto
	Total      float64       `json:"total"`      // Total
	AccessKey  string        `json:"accessKey"`  // Chave de acesso da NFC-e
	Number     string        `json:"number"`     // Número da nota
	QRCodeURL  string        `json:"qrCodeUrl"`  // URL original (para confirmação)
}

// ScanQRCodePreviewResponse representa a resposta do preview
type ScanQRCodePreviewResponse struct {
	Message string             `json:"message"`
	Data    PreviewReceiptData `json:"data"`
}

// ScanQRCodePreviewHandler extrai dados do QR Code sem salvar no banco
// @Summary Preview de NFC-e via QR Code (Etapa 1/2)
// @Description Extrai dados da NFC-e sem salvar no banco. Retorna dados para visualização/edição
// @Tags receipts
// @Accept json
// @Produce json
// @Param request body ScanQRCodePreviewRequest true "QR Code URL"
// @Success 200 {object} ScanQRCodePreviewResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /scan-qrcode/preview [post]
func ScanQRCodePreviewHandler(ctx *gin.Context) {
	var request ScanQRCodePreviewRequest

	// Bind JSON
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("error binding json: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Valida URL
	if request.QRCodeURL == "" {
		sendError(ctx, http.StatusBadRequest, "QR Code URL is required")
		return
	}

	// ⚡ Faz scraping da NFC-e (rápido e gratuito!)
	logger.InfoF("🔍 Preview: Scraping NFC-e from URL: %s", request.QRCodeURL)
	startTime := time.Now()

	receiptData, err := scrapeNFCe(request.QRCodeURL)
	if err != nil {
		logger.ErrorF("error scraping NFC-e: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("Error scraping NFC-e: %v", err.Error()))
		return
	}

	scrapingTime := time.Since(startTime)
	logger.InfoF("✅ NFC-e scraped successfully in %.2fs: %s - %d items - Total: R$ %.2f",
		scrapingTime.Seconds(), receiptData.StoreName, len(receiptData.Items), receiptData.Total)

	// Converte para formato de preview (sem categorias ainda)
	previewItems := make([]PreviewItem, len(receiptData.Items))
	for i, item := range receiptData.Items {
		previewItems[i] = PreviewItem{
			TempID:      i + 1, // ID temporário sequencial
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			Total:       item.Total,
		}
	}

	// Monta resposta de preview
	previewData := PreviewReceiptData{
		StoreName:  receiptData.StoreName,
		Date:       receiptData.Date,
		Items:      previewItems,
		ItemsCount: receiptData.ItemsCount,
		Subtotal:   receiptData.Subtotal,
		Discount:   receiptData.Discount,
		Total:      receiptData.Total,
		AccessKey:  receiptData.AccessKey,
		Number:     receiptData.Number,
		QRCodeURL:  request.QRCodeURL,
	}

	logger.InfoF("📋 Preview ready: %d items extracted (not saved yet)", len(previewItems))

	// Retorna preview (SEM salvar no banco)
	ctx.JSON(http.StatusOK, ScanQRCodePreviewResponse{
		Message: fmt.Sprintf("✅ Preview ready! %d items extracted. You can now edit, remove items, or confirm to save.", len(previewItems)),
		Data:    previewData,
	})
}
