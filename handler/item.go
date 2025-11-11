package handler

import (
	"fmt"
	"net/http"
	"strings"
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
	userID, _ := ctx.Get("user_id")

	var item schemas.ReceiptItem
	// Busca o item e garante que pertence ao usuário autenticado
	if err := db.Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipt_items.id = ? AND receipts.user_id = ?", id, userID).
		First(&item).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Item não encontrado ou não pertence ao usuário autenticado"})
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

	userID, _ := ctx.Get("user_id")

	var request UpdateItemRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var item schemas.ReceiptItem
	// Busca o item e garante que pertence ao usuário autenticado
	if err := db.Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipt_items.id = ? AND receipts.user_id = ?", id, userID).
		First(&item).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Item não encontrado ou não pertence ao usuário autenticado")
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
		sendError(ctx, http.StatusInternalServerError, "Erro ao atualizar item no banco de dados. Por favor, tente novamente")
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

// RecategorizeItemsRequest define a estrutura para requisição de recategorização
type RecategorizeItemsRequest struct {
	ItemIDs []uint `json:"itemIds" binding:"required"` // IDs dos items a serem recategorizados
}

// RecategorizeItemsResponse define a estrutura da resposta
type RecategorizeItemsResponse struct {
	Message            string                       `json:"message"`
	ItemsRecategorized int                          `json:"itemsRecategorized"`
	Results            []ItemRecategorizationResult `json:"results"`
}

type ItemRecategorizationResult struct {
	ItemID          uint   `json:"itemId"`
	ProductName     string `json:"productName"`
	OldCategoryID   uint   `json:"oldCategoryId"`
	OldCategoryName string `json:"oldCategoryName"`
	NewCategoryID   uint   `json:"newCategoryId"`
	NewCategoryName string `json:"newCategoryName"`
	Changed         bool   `json:"changed"`
}

// @Summary Recategorize items using AI
// @Description Use Gemini AI to recategorize items. Useful for items in "Não categorizado" or items that need recategorization.
// @Tags items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RecategorizeItemsRequest true "Item IDs to recategorize"
// @Success 200 {object} RecategorizeItemsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /items/recategorize [post]
func RecategorizeItemsHandler(ctx *gin.Context) {
	var request RecategorizeItemsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userID, _ := ctx.Get("user_id")

	if len(request.ItemIDs) == 0 {
		sendError(ctx, http.StatusBadRequest, "At least one item ID is required")
		return
	}

	// Busca todos os items do usuário com preload de product e category
	var items []schemas.ReceiptItem
	err := db.Preload("Product").Preload("Category").
		Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipt_items.id IN ? AND receipts.user_id = ?", request.ItemIDs, userID).
		Find(&items).Error

	if err != nil {
		logger.ErrorF("error finding items: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error finding items")
		return
	}

	if len(items) == 0 {
		sendError(ctx, http.StatusNotFound, "No items found")
		return
	}

	// Busca todas as categorias DO USUÁRIO
	var categories []schemas.Category
	if err := db.Where("user_id = ?", userID).Find(&categories).Error; err != nil {
		logger.ErrorF("error finding categories: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error finding categories")
		return
	}

	// Prepara o prompt para o Gemini
	prompt := buildRecategorizationPrompt(items, categories)

	// Chama o Gemini AI para recategorizar
	response, err := callGeminiForRecategorization(prompt)
	if err != nil {
		logger.ErrorF("error calling Gemini: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error calling AI service")
		return
	}

	// Processa a resposta e atualiza os items
	results, itemsUpdated := applyRecategorization(items, response, categories)

	ctx.JSON(http.StatusOK, RecategorizeItemsResponse{
		Message:            "Items recategorized successfully",
		ItemsRecategorized: itemsUpdated,
		Results:            results,
	})
}

func buildRecategorizationPrompt(items []schemas.ReceiptItem, categories []schemas.Category) string {
	var builder strings.Builder

	builder.WriteString("Você é um assistente de finanças que recategoriza produtos de compras.\n")
	builder.WriteString("IMPORTANTE: Retorne APENAS um JSON válido e bem formatado, sem comentários, texto adicional ou vírgulas extras.\n")
	builder.WriteString("IDIOMA: Todas as descrições devem estar em PORTUGUÊS (PT-BR).\n\n")

	builder.WriteString("Formato esperado:\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"categorizations\": [\n")
	builder.WriteString("    {\n")
	builder.WriteString("      \"itemId\": number - ID do item,\n")
	builder.WriteString("      \"categoryId\": number - ID da categoria (apenas o número, não o nome)\n")
	builder.WriteString("    }\n")
	builder.WriteString("  ]\n")
	builder.WriteString("}\n\n")

	// Adiciona lista de categorias disponíveis COM IDs
	builder.WriteString("CATEGORIAS DISPONÍVEIS (use o ID para categoryId):\n")
	for _, cat := range categories {
		if cat.Name == "Não categorizado" {
			continue // Não deve recategorizar para esta categoria
		}
		builder.WriteString(fmt.Sprintf("ID %d: %s", cat.ID, cat.Name))
		if cat.Icon != "" {
			builder.WriteString(fmt.Sprintf(" %s", cat.Icon))
		}
		if cat.Description != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", cat.Description))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")

	// Lista os produtos que precisam ser recategorizados
	builder.WriteString("PRODUTOS PARA CATEGORIZAR:\n")
	for _, item := range items {
		if item.Product != nil {
			builder.WriteString(fmt.Sprintf("ItemID %d: %s", item.ID, item.Product.Name))
			if item.Product.Unity != "" {
				builder.WriteString(fmt.Sprintf(" (%s)", item.Product.Unity))
			}
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\n")

	builder.WriteString("Regras importantes:\n")
	builder.WriteString("- NUNCA deixe vírgulas extras antes de fechar objetos } ou arrays ]\n")
	builder.WriteString("- Garanta que o JSON seja válido e possa ser parseado sem erros\n")
	builder.WriteString("- Para cada item, use categoryId com APENAS O NÚMERO do ID da categoria (ex: 1, 2, 3)\n")
	builder.WriteString("- NÃO use o nome da categoria, APENAS o ID numérico\n")
	builder.WriteString("- NUNCA use a categoria 'Não categorizado' para recategorização\n")
	builder.WriteString("\n")
	builder.WriteString("⚠️ CATEGORIZAÇÃO ÚNICA E PRECISA (REGRA CRÍTICA):\n")
	builder.WriteString("  * CADA item deve estar em APENAS UMA categoria - escolha a MAIS ESPECÍFICA\n")
	builder.WriteString("  * Analise o produto e identifique sua categoria PRINCIPAL e ÚNICA\n")
	builder.WriteString("  * NUNCA coloque o mesmo produto em 2 categorias diferentes\n")
	builder.WriteString("\n")
	builder.WriteString("  📋 GUIA DE CATEGORIZAÇÃO (use para decidir):\n")
	builder.WriteString("  • Cerveja, Vinho, Whisky → 'Bebidas Alcoólicas' (NÃO 'Bebidas')\n")
	builder.WriteString("  • Café, Chá, Mate → 'Café e Chá' (NÃO 'Bebidas')\n")
	builder.WriteString("  • Refrigerante, Suco, Água → 'Bebidas' (NÃO 'Café e Chá')\n")
	builder.WriteString("  • Presunto, Mortadela, Salsicha → 'Frios e Embutidos' (NÃO 'Carnes e Proteínas')\n")
	builder.WriteString("  • Frango, Carne Bovina, Peixe → 'Carnes e Proteínas' (NÃO 'Frios e Embutidos')\n")
	builder.WriteString("  • Macarrão, Lasanha → 'Massas' (NÃO 'Padaria')\n")
	builder.WriteString("  • Pão, Baguete → 'Padaria' (NÃO 'Massas')\n")
	builder.WriteString("  • Chocolate, Bala, Sorvete → 'Doces e Sobremesas' (NÃO 'Salgadinhos e Snacks')\n")
	builder.WriteString("  • Chips, Amendoim, Pipoca → 'Salgadinhos e Snacks' (NÃO 'Doces e Sobremesas')\n")
	builder.WriteString("  • Azeite, Sal, Molho → 'Condimentos e Temperos' (NÃO 'Enlatados')\n")
	builder.WriteString("  • Milho em lata, Atum em lata → 'Enlatados e Conservas' (NÃO 'Condimentos')\n")
	builder.WriteString("  • Shampoo, Sabonete → 'Higiene Pessoal' (NÃO 'Limpeza Doméstica')\n")
	builder.WriteString("  • Detergente, Desinfetante → 'Limpeza Doméstica' (NÃO 'Higiene Pessoal')\n")
	builder.WriteString("  • Papel Higiênico, Guardanapo → 'Papel e Descartáveis' (NÃO 'Limpeza' ou 'Higiene')\n")
	builder.WriteString("  • Pizza congelada, Vegetais congelados → 'Congelados' (NÃO 'Doces' mesmo que seja sorvete)\n")
	builder.WriteString("\n")
	builder.WriteString("  * Se ainda houver dúvida, escolha a categoria que descreve MELHOR o produto principal\n")
	builder.WriteString("  * Use 'Outros' APENAS para produtos verdadeiramente únicos/raros que não se encaixam\n")
	builder.WriteString("  * Seja CONSISTENTE: produtos iguais devem SEMPRE estar na mesma categoria\n")
	builder.WriteString("\n")

	return builder.String()
}

// Funções auxiliares serão implementadas aqui
func callGeminiForRecategorization(prompt string) (map[string]interface{}, error) {
	// TODO: Implementar chamada real ao Gemini
	// Por enquanto retorna um mock
	return map[string]interface{}{
		"categorizations": []map[string]interface{}{},
	}, nil
}

func applyRecategorization(items []schemas.ReceiptItem, response map[string]interface{}, categories []schemas.Category) ([]ItemRecategorizationResult, int) {
	// TODO: Implementar aplicação das categorizações
	return []ItemRecategorizationResult{}, 0
}
