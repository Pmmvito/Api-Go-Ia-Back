package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ScanQRCodeConfirmRequest representa o request de confirmação (IGUAL ao PreviewReceiptData)
type ScanQRCodeConfirmRequest struct {
	StoreName  string        `json:"storeName" binding:"required"` // Nome do estabelecimento
	Date       string        `json:"date"`                         // Data da compra (opcional, usa data atual se vazio)
	Items      []PreviewItem `json:"items" binding:"required"`     // Items (editados) - USA PreviewItem
	ItemsCount int           `json:"itemsCount"`                   // Total de items
	Subtotal   float64       `json:"subtotal"`                     // Subtotal
	Discount   float64       `json:"discount"`                     // Desconto
	Total      float64       `json:"total"`                        // Total (opcional, calcula se zero)
	AccessKey  string        `json:"accessKey"`                    // Chave de acesso
	Number     string        `json:"number"`                       // Número da nota
	QRCodeURL  string        `json:"qrCodeUrl" binding:"required"` // URL original
}

// ScanQRCodeConfirmResponse representa a resposta após salvar (apenas mensagem)
type ScanQRCodeConfirmResponse struct {
	Message string `json:"message"`
}

// ScanQRCodeConfirmHandler confirma, categoriza com IA e salva no banco (Etapa 2/2)
// @Summary Confirmar e salvar NFC-e (Etapa 2/2)
// @Description Categoriza items com IA e salva nota fiscal no banco de dados. Envie APENAS o campo 'data' da resposta do preview, NÃO envie 'message'.
// @Tags receipts
// @Accept json
// @Produce json
// @Param request body ScanQRCodeConfirmRequest true "Dados da nota (envie APENAS o campo 'data' do preview, sem 'message')"
// @Success 200 {object} ScanQRCodeConfirmResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /scan-qrcode/confirm [post]
func ScanQRCodeConfirmHandler(ctx *gin.Context) {
	var request ScanQRCodeConfirmRequest

	// Bind JSON
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("error binding json: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Validações básicas
	if len(request.Items) == 0 {
		sendError(ctx, http.StatusBadRequest, "At least one item is required")
		return
	}

	// Obtém o User ID do contexto (JWT)
	userID, exists := ctx.Get("user_id")
	if !exists {
		sendError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	logger.InfoF("📝 Confirming receipt: %s - %d items - Total: R$ %.2f",
		request.StoreName, len(request.Items), request.Total)

	// Filtra items deletados
	activeItems := []PreviewItem{}
	for _, item := range request.Items {
		if !item.Deleted {
			activeItems = append(activeItems, item)
		}
	}

	if len(activeItems) == 0 {
		sendError(ctx, http.StatusBadRequest, "All items were deleted. Cannot save empty receipt.")
		return
	}

	logger.InfoF("✅ Active items after filtering: %d (deleted: %d)",
		len(activeItems), len(request.Items)-len(activeItems))

	// 🤖 ETAPA 1: Categorização com IA (em lote)
	startAI := time.Now()
	logger.InfoF("🤖 Starting AI categorization for %d items...", len(activeItems))

	// Converte para formato NFCeItem para usar a função existente
	nfceItems := make([]NFCeItem, len(activeItems))
	for i, item := range activeItems {
		nfceItems[i] = NFCeItem{
			ItemNumber:  item.TempID,
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			Total:       item.Total,
		}
	}

	// Usa a função de categorização existente
	categorizedItems, err := categorizeItemsWithAI(nfceItems)
	if err != nil {
		logger.ErrorF("❌ AI categorization failed: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("AI categorization error: %v", err.Error()))
		return
	}

	aiTime := time.Since(startAI)
	logger.InfoF("✅ AI categorization completed in %.2fs", aiTime.Seconds())

	// Monta mapa tempID -> categoryID
	categoryMap := make(map[int]uint)
	for i, categorizedItem := range categorizedItems {
		if i < len(activeItems) {
			categoryMap[activeItems[i].TempID] = categorizedItem.CategoryID
			logger.InfoF("✓ Item #%d (%s) -> CategoryID: %d",
				activeItems[i].TempID, activeItems[i].Description, categorizedItem.CategoryID)
		}
	}

	// 💾 ETAPA 2: Salvar no banco de dados (em background)
	go func() {
		startSave := time.Now()
		logger.InfoF("💾 [Background] Saving receipt to database...")

		// Cria Receipt
		receipt := schemas.Receipt{
			UserID:      userID.(uint),
			StoreName:   request.StoreName,
			Date:        request.Date,
			Total:       request.Total,
			Subtotal:    request.Subtotal,
			Discount:    request.Discount,
			Currency:    "BRL",
			Confidence:  1.0,
			Notes:       fmt.Sprintf("NFC-e #%s - Chave: %s", request.Number, request.AccessKey),
			ImageBase64: request.QRCodeURL,
		}

		// Usa transação para salvar
		err := db.Transaction(func(tx *gorm.DB) error {
			// Salva Receipt
			if err := tx.Create(&receipt).Error; err != nil {
				return fmt.Errorf("error creating receipt: %w", err)
			}

			logger.InfoF("✓ [Background] Receipt created with ID: %d", receipt.ID)

			// Salva Items com categorias da IA
			for _, item := range activeItems {
				categoryID := categoryMap[item.TempID]
				if categoryID == 0 {
					// Fallback para categoria "Outros"
					var defaultCategory schemas.Category
					if err := tx.Where("name = ?", "Outros").First(&defaultCategory).Error; err == nil {
						categoryID = defaultCategory.ID
						logger.InfoF("⚠️  [Background] Using default category 'Outros' (ID: %d) for item #%d", categoryID, item.TempID)
					}
				}

				// Busca ou cria o produto
				var product schemas.Product
				normalizedUnit := normalizeUnit(item.Unit)
				if err := tx.Where("name = ? AND unity = ?", item.Description, normalizedUnit).First(&product).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						// Cria novo produto
						product = schemas.Product{
							Name:  item.Description,
							Unity: normalizedUnit,
						}
						if err := tx.Create(&product).Error; err != nil {
							return fmt.Errorf("error creating product: %w", err)
						}
						logger.InfoF("[Background] Produto criado: %s (%s) - ID: %d", product.Name, product.Unity, product.ID)
					} else {
						return fmt.Errorf("error finding product: %w", err)
					}
				}

				receiptItem := schemas.ReceiptItem{
					ReceiptID:   receipt.ID,
					CategoryID:  categoryID,
					ProductID:   product.ID,
					Description: item.Description,
					Quantity:    item.Quantity,
					Unit:        item.Unit,
					UnitPrice:   item.UnitPrice,
					Total:       item.Total,
				}

				if err := tx.Create(&receiptItem).Error; err != nil {
					return fmt.Errorf("error creating receipt item: %w", err)
				}
			}

			return nil
		})

		if err != nil {
			logger.ErrorF("❌ [Background] Error saving receipt: %v", err.Error())
			return
		}

		saveTime := time.Since(startSave)
		totalTime := time.Since(startAI)
		logger.InfoF("🎉 [Background] Complete! Receipt ID: %d, Items: %d, Total time: %.2fs (AI: %.2fs, Save: %.2fs)",
			receipt.ID, len(activeItems), totalTime.Seconds(), aiTime.Seconds(), saveTime.Seconds())
	}()

	// Retorna imediatamente apenas mensagem de sucesso
	ctx.JSON(http.StatusOK, gin.H{
		"message": "✅ Nota fiscal processada! ",
	})
}
