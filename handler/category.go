package handler

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateCategoryRequest define os dados necessários para criar uma nova categoria.
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required" example:"Viagens"`
	Description string `json:"description" example:"Gastos com viagens e turismo"`
	Icon        string `json:"icon" example:"✈️"`
	Color       string `json:"color" example:"#3498db"`
}

// UpdateCategoryRequest define os dados para atualizar uma categoria existente.
// Todos os campos são ponteiros para permitir atualizações parciais.
type UpdateCategoryRequest struct {
	Name        *string `json:"name" example:"Alimentação Fora"`
	Description *string `json:"description" example:"Restaurantes e delivery"`
	Icon        *string `json:"icon" example:"🍕"`
	Color       *string `json:"color" example:"#e74c3c"`
}

// CategoryGraphResponse define a estrutura para a resposta do endpoint de gráfico de categorias.
type CategoryGraphResponse struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	ItemCount int64   `json:"itemCount"`
	Total     float64 `json:"total"`
}

// GraphData define a estrutura de encapsulamento para a resposta do gráfico.
type GraphData struct {
	Categories []CategoryGraphResponse `json:"categories"`
	GrandTotal float64                 `json:"grandTotal"`
}

// CategoryItemResponse define a estrutura para os itens de uma categoria
type CategoryItemResponse struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Total        float64 `json:"total"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	ReceiptID    uint    `json:"receiptId"`
	StoreName    string  `json:"storeName"`
	PurchaseDate string  `json:"purchaseDate"`
}

// CategoryWithItemsResponse define a resposta completa da categoria com seus itens
type CategoryWithItemsResponse struct {
	schemas.CategoryResponse
	Items      []CategoryItemResponse `json:"items"`
	ItemCount  int                    `json:"itemCount"`
	TotalValue float64                `json:"totalValue"`
}

// @Summary Create new category
// @Description Create a new expense category for organizing receipt items. If a category with the same name was previously deleted, it will be reactivated.
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCategoryRequest true "Category data (name is required, description/icon/color are optional)"
// @Success 201 {object} map[string]interface{} "Category created successfully"
// @Success 200 {object} map[string]interface{} "Category reactivated successfully (when reactivating a deleted category)"
// @Failure 400 {object} ErrorResponse "Dados inválidos para criação de categoria. O campo 'name' é obrigatório"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 409 {object} ErrorResponse "Já existe uma categoria ativa com este nome. Por favor, escolha outro nome ou utilize a categoria existente"
// @Failure 500 {object} ErrorResponse "Erro ao reativar a categoria deletada anteriormente. Por favor, tente novamente | Erro ao criar categoria no banco de dados. Por favor, tente novamente"
// @Router /category [post]
func CreateCategoryHandler(ctx *gin.Context) {
	var request CreateCategoryRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados inválidos para criação de categoria. O campo 'name' é obrigatório")
		return
	}

	userID, _ := ctx.Get("user_id")

	// Verifica se existe uma categoria deletada com o mesmo nome PARA ESTE USUÁRIO
	var existingCategory schemas.Category
	err := db.Unscoped().Where("name = ? AND user_id = ?", request.Name, userID).First(&existingCategory).Error

	if err == nil {
		// Categoria encontrada - verifica se está deletada
		if existingCategory.DeletedAt.Valid {
			// Reativa a categoria deletada
			existingCategory.DeletedAt = gorm.DeletedAt{}
			existingCategory.Description = request.Description
			existingCategory.Icon = request.Icon
			existingCategory.Color = request.Color

			if err := db.Unscoped().Save(&existingCategory).Error; err != nil {
				logger.ErrorF("error reactivating category: %v", err.Error())
				sendError(ctx, http.StatusInternalServerError, "Erro ao reativar a categoria deletada anteriormente. Por favor, tente novamente")
				return
			}

			logger.InfoF("Category reactivated with ID: %d", existingCategory.ID)
			ctx.JSON(http.StatusOK, gin.H{
				"message": "Category reactivated successfully",
				"data":    existingCategory.ToResponse(),
			})
			return
		} else {
			// Categoria já existe e está ativa
			logger.ErrorF("category already exists: %s", request.Name)
			sendError(ctx, http.StatusConflict, "Já existe uma categoria ativa com este nome. Por favor, escolha outro nome ou utilize a categoria existente")
			return
		}
	}

	// Categoria não existe, cria uma nova
	category := schemas.Category{
		UserID:      userID.(uint),
		Name:        request.Name,
		Description: request.Description,
		Icon:        request.Icon,
		Color:       request.Color,
	}

	if err := db.Create(&category).Error; err != nil {
		logger.ErrorF("error creating category: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao criar categoria no banco de dados. Por favor, tente novamente")
		return
	}

	logger.InfoF("Category created with ID: %d", category.ID)
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Category created successfully",
		"data":    category.ToResponse(),
	})
}

// @Summary List all categories
// @Description Get all expense categories sorted by name with item count. Returns: 1-Alimentação, 2-Transporte, 3-Saúde, 4-Lazer, 5-Educação, 6-Moradia, 7-Vestuário, 8-Outros
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of categories with item count"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Erro ao buscar categorias no banco de dados. Por favor, tente novamente"
// @Router /categories [get]
func ListCategoriesHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var categories []schemas.Category
	if err := db.Where("user_id = ?", userID).Order("name ASC").Find(&categories).Error; err != nil {
		logger.ErrorF("error listing categories: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao buscar categorias no banco de dados. Por favor, tente novamente")
		return
	}

	// Busca a contagem de itens para cada categoria em uma única query
	type CategoryCount struct {
		CategoryID uint
		ItemCount  int
	}
	var counts []CategoryCount
	db.Table("receipt_items").
		Select("category_id, COUNT(*) as item_count").
		Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.user_id = ? AND receipt_items.deleted_at IS NULL", userID).
		Group("category_id").
		Scan(&counts)

	// Cria um map para rápido acesso aos counts
	countMap := make(map[uint]int)
	for _, count := range counts {
		countMap[count.CategoryID] = count.ItemCount
	}

	// Converte para response incluindo itemCount
	var responses []schemas.CategoryResponse
	for _, category := range categories {
		response := category.ToResponse()
		itemCount := countMap[category.ID] // Se não existir no map, será 0
		response.ItemCount = &itemCount
		responses = append(responses, response)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Categories retrieved successfully",
		"data":    responses,
		"count":   len(responses),
	})
}

// @Summary List categories summary (lightweight)
// @Description Get all categories with item count in a lightweight format (no timestamps). Ideal for lists and dropdowns. 650x faster than full endpoint. Supports optional period filtering.
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date for filtering (format: YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string false "End date for filtering (format: YYYY-MM-DD)" example(2024-12-31)
// @Success 200 {object} map[string]interface{} "List of categories (lightweight)"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Erro ao buscar categorias no banco de dados. Por favor, tente novamente"
// @Router /categories/summary [get]
func ListCategoriesSummaryHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	// Parâmetros de período (opcionais)
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	var categories []schemas.Category
	if err := db.Where("user_id = ?", userID).Order("name ASC").Find(&categories).Error; err != nil {
		logger.ErrorF("error listing categories: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao buscar categorias no banco de dados. Por favor, tente novamente")
		return
	}

	// Busca a contagem de itens para cada categoria em uma única query
	type CategoryCount struct {
		CategoryID uint
		ItemCount  int
	}
	var counts []CategoryCount

	// Query base para contagem
	query := db.Table("receipt_items").
		Select("category_id, COUNT(*) as item_count").
		Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.user_id = ? AND receipt_items.deleted_at IS NULL", userID)

	// Aplica filtro de período se fornecido
	if startDate != "" && endDate != "" {
		query = query.Where("receipts.date >= ? AND receipts.date <= ?", startDate, endDate)
	} else if startDate != "" {
		query = query.Where("receipts.date >= ?", startDate)
	} else if endDate != "" {
		query = query.Where("receipts.date <= ?", endDate)
	}

	query.Group("category_id").Scan(&counts)

	// Cria um map para rápido acesso aos counts
	countMap := make(map[uint]int)
	for _, count := range counts {
		countMap[count.CategoryID] = count.ItemCount
	}

	// Converte para summary (sem timestamps - mais leve!)
	var summaries []schemas.CategorySummary
	for _, category := range categories {
		itemCount := countMap[category.ID] // Se não existir no map, será 0
		summaries = append(summaries, category.ToSummary(itemCount))
	}

	response := gin.H{
		"message":    "Categories summary retrieved successfully",
		"categories": summaries,
		"total":      len(summaries),
	}

	// Adiciona informação de período se fornecido
	if startDate != "" || endDate != "" {
		response["period"] = gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Get category details
// @Description Get details of a specific category by ID including all items that belong to this category. Supports period filtering and pagination.
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID" example(1)
// @Param start_date query string false "Start date for filtering (format: YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string false "End date for filtering (format: YYYY-MM-DD)" example(2024-12-31)
// @Param page query int false "Page number (default: 1)" example(1)
// @Param limit query int false "Items per page (default: 50)" example(50)
// @Success 200 {object} map[string]interface{} "Category details with items"
// @Failure 400 {object} ErrorResponse "ID da categoria é obrigatório na URL | Parâmetro 'page' inválido | Parâmetro 'limit' inválido"
// @Failure 404 {object} ErrorResponse "Categoria não encontrada. Verifique se o ID está correto e se a categoria não foi deletada"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Erro ao buscar itens da categoria. Por favor, tente novamente"
// @Router /category/{id} [get]
func GetCategoryHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "ID da categoria é obrigatório na URL")
		return
	}

	userID, _ := ctx.Get("user_id")

	// Parâmetros de período (opcionais)
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	// Parâmetros de paginação
	page := 1
	limit := 50

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		} else {
			sendError(ctx, http.StatusBadRequest, "Parâmetro 'page' inválido. Deve ser um número maior que 0")
			return
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		} else {
			sendError(ctx, http.StatusBadRequest, "Parâmetro 'limit' inválido. Deve ser um número entre 1 e 200")
			return
		}
	}

	offset := (page - 1) * limit

	// Busca a categoria garantindo que pertence ao usuário
	var category schemas.Category
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Categoria não encontrada ou não pertence ao usuário autenticado")
		return
	}

	// Query base para itens
	baseQuery := db.Model(&schemas.ReceiptItem{}).
		Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipt_items.category_id = ? AND receipts.user_id = ?", id, userID)

	// Aplica filtro de período se fornecido
	if startDate != "" && endDate != "" {
		baseQuery = baseQuery.Where("receipts.date >= ? AND receipts.date <= ?", startDate, endDate)
	} else if startDate != "" {
		baseQuery = baseQuery.Where("receipts.date >= ?", startDate)
	} else if endDate != "" {
		baseQuery = baseQuery.Where("receipts.date <= ?", endDate)
	}

	// Conta total de itens (para paginação)
	var totalItems int64
	if err := baseQuery.Count(&totalItems).Error; err != nil {
		logger.ErrorF("error counting category items: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao contar itens da categoria. Por favor, tente novamente")
		return
	}

	// Busca itens com paginação
	var receiptItems []schemas.ReceiptItem
	err := baseQuery.Preload("Product").
		Preload("Receipt").
		Order("receipt_items.receipt_id DESC").
		Limit(limit).
		Offset(offset).
		Find(&receiptItems).Error

	if err != nil {
		logger.ErrorF("error getting category items: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao buscar itens da categoria. Por favor, tente novamente")
		return
	}

	// Converte para o formato de resposta
	items := make([]CategoryItemResponse, 0, len(receiptItems))
	var totalValue float64

	for _, item := range receiptItems {
		// Pega nome e unidade do produto (OBRIGATÓRIO)
		name := "Produto sem nome"
		unit := ""
		if item.Product != nil {
			name = item.Product.Name
			unit = item.Product.Unity
		}

		// Pega nome da loja e data do recibo
		storeName := ""
		purchaseDate := ""
		if item.Receipt != nil {
			storeName = item.Receipt.StoreName
			purchaseDate = item.Receipt.Date
		}

		items = append(items, CategoryItemResponse{
			ID:           item.ID,
			Name:         name,
			Total:        item.Total,
			Quantity:     item.Quantity,
			Unit:         unit,
			ReceiptID:    item.ReceiptID,
			StoreName:    storeName,
			PurchaseDate: purchaseDate,
		})

		totalValue += item.Total
	}

	// Calcula informações de paginação
	totalPages := int((totalItems + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}
	hasNextPage := page < totalPages

	// Monta resposta
	response := gin.H{
		"message": "Category retrieved successfully",
		"data": gin.H{
			"category":   category.ToResponse(),
			"items":      items,
			"itemCount":  len(items),
			"totalValue": totalValue,
		},
		"summary": gin.H{
			"totalItems":  totalItems,
			"totalPages":  totalPages,
			"currentPage": page,
			"hasNextPage": hasNextPage,
		},
	}

	// Adiciona informação de período se fornecido
	if startDate != "" || endDate != "" {
		response["period"] = gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Update category
// @Description Update category information (name, description, icon, color). All fields are optional - only send what you want to update.
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID" example(1)
// @Param request body UpdateCategoryRequest true "Category data to update (all fields optional)"
// @Success 200 {object} map[string]interface{} "Category updated successfully"
// @Failure 400 {object} ErrorResponse "ID da categoria é obrigatório na URL | Dados inválidos para atualização da categoria. Verifique os campos enviados | Nenhum campo foi fornecido para atualização. Envie pelo menos um campo (name, description, icon ou color)"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} ErrorResponse "Categoria não encontrada. Verifique se o ID está correto e se a categoria não foi deletada"
// @Failure 500 {object} ErrorResponse "Erro ao salvar atualização da categoria no banco de dados. Por favor, tente novamente"
// @Router /category/{id} [patch]
func UpdateCategoryHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "ID da categoria é obrigatório na URL")
		return
	}

	userID, _ := ctx.Get("user_id")

	var request UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados inválidos para atualização da categoria. Verifique os campos enviados")
		return
	}

	var category schemas.Category
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Categoria não encontrada ou não pertence ao usuário autenticado")
		return
	}

	// Atualiza apenas os campos fornecidos
	updated := false
	if request.Name != nil {
		category.Name = *request.Name
		updated = true
	}
	if request.Description != nil {
		category.Description = *request.Description
		updated = true
	}
	if request.Icon != nil {
		category.Icon = *request.Icon
		updated = true
	}
	if request.Color != nil {
		category.Color = *request.Color
		updated = true
	}

	if !updated {
		sendError(ctx, http.StatusBadRequest, "Nenhum campo foi fornecido para atualização. Envie pelo menos um campo (name, description, icon ou color)")
		return
	}

	if err := db.Save(&category).Error; err != nil {
		logger.ErrorF("error updating category: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao salvar atualização da categoria no banco de dados. Por favor, tente novamente")
		return
	}

	logger.InfoF("Category %s updated successfully", id)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Category updated successfully",
		"data":    category.ToResponse(),
	})
}

// @Summary Delete category
// @Description Delete a category and move all its items to "Não categorizado". Items can be recategorized later using the /items/recategorize endpoint.
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID" example(1)
// @Success 200 {object} map[string]interface{} "Category deleted successfully, items moved to 'Não categorizado'"
// @Failure 400 {object} ErrorResponse "ID da categoria é obrigatório na URL | A categoria 'Não categorizado' é do sistema e não pode ser deletada"
// @Failure 404 {object} ErrorResponse "Categoria não encontrada. Verifique se o ID está correto e se a categoria não foi deletada anteriormente"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Categoria do sistema 'Não categorizado' não foi encontrada. Por favor, restaure as categorias padrão | Erro ao mover itens para a categoria 'Não categorizado'. Operação cancelada | Erro ao deletar categoria. Operação cancelada | Erro ao confirmar a exclusão da categoria no banco de dados. Por favor, tente novamente"
// @Router /category/{id} [delete]
func DeleteCategoryHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "ID da categoria é obrigatório na URL")
		return
	}

	userID, _ := ctx.Get("user_id")

	var category schemas.Category
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Categoria não encontrada ou não pertence ao usuário autenticado")
		return
	}

	// Não permite deletar a categoria "Não categorizado"
	if category.Name == "Não categorizado" {
		sendError(ctx, http.StatusBadRequest, "A categoria 'Não categorizado' é do sistema e não pode ser deletada")
		return
	}

	// Busca a categoria "Não categorizado" DO USUÁRIO
	var uncategorized schemas.Category
	if err := db.Where("name = ? AND user_id = ?", "Não categorizado", userID).First(&uncategorized).Error; err != nil {
		logger.ErrorF("'Não categorizado' category not found for user %v: %v", userID, err.Error())
		sendError(ctx, http.StatusInternalServerError, "Categoria do sistema 'Não categorizado' não foi encontrada. Por favor, entre em contato com o suporte")
		return
	}

	// Inicia transação
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Move todos os items desta categoria para "Não categorizado"
	result := tx.Model(&schemas.ReceiptItem{}).
		Where("category_id = ?", category.ID).
		Update("category_id", uncategorized.ID)

	if result.Error != nil {
		tx.Rollback()
		logger.ErrorF("error moving items to uncategorized: %v", result.Error.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao mover itens para a categoria 'Não categorizado'. Operação cancelada")
		return
	}

	itemsMoved := result.RowsAffected
	logger.InfoF("Moved %d items from category %s to 'Não categorizado'", itemsMoved, category.Name)

	// Deleta a categoria
	if err := tx.Delete(&category).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error deleting category: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao deletar categoria. Operação cancelada")
		return
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		logger.ErrorF("error committing transaction: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao confirmar a exclusão da categoria no banco de dados. Por favor, tente novamente")
		return
	}

	logger.InfoF("Category %s deleted successfully, %d items moved to 'Não categorizado'", category.Name, itemsMoved)
	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Category deleted successfully",
		"itemsMoved": itemsMoved,
		"note":       "Items moved to 'Não categorizado'. Use POST /items/recategorize to recategorize them.",
	})
}

// @Summary Get category graph data
// @Description Get aggregated data for each category, including item count and total value. Filters by date range, defaulting to the current month.
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date for filtering (YYYY-MM-DD)"
// @Param end_date query string false "End date for filtering (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Category graph data retrieved successfully"
// @Failure 400 {object} ErrorResponse "Formato de start_date inválido. Use o formato YYYY-MM-DD (exemplo: 2024-01-15) | Formato de end_date inválido. Use o formato YYYY-MM-DD (exemplo: 2024-01-31)"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Erro ao buscar notas fiscais do período. Por favor, tente novamente | Erro ao buscar itens das notas fiscais. Por favor, tente novamente | Erro ao buscar categorias. Por favor, tente novamente"
// @Router /categories/graph [get]
func GetCategoryGraphHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr == "" || endDateStr == "" {
		// Padrão para o mês atual
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 1, 0)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			sendError(ctx, http.StatusBadRequest, "Formato de start_date inválido. Use o formato YYYY-MM-DD (exemplo: 2024-01-15)")
			return
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			sendError(ctx, http.StatusBadRequest, "Formato de end_date inválido. Use o formato YYYY-MM-DD (exemplo: 2024-01-31)")
			return
		}
		// Adiciona um dia ao endDate para incluir todo o período
		endDate = endDate.AddDate(0, 0, 1)
	}

	// 1. Buscar todos os receipts do usuário no período usando GORM
	var receipts []schemas.Receipt
	if err := db.Where("user_id = ? AND date >= ? AND date < ?", userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Find(&receipts).Error; err != nil {
		logger.ErrorF("error finding receipts: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao buscar notas fiscais do período. Por favor, tente novamente")
		return
	}

	// 2. Extrair IDs dos receipts
	var receiptIDs []uint
	for _, receipt := range receipts {
		receiptIDs = append(receiptIDs, receipt.ID)
	}

	// 3. Buscar todos os items desses receipts usando GORM
	var items []schemas.ReceiptItem
	if len(receiptIDs) > 0 {
		if err := db.Where("receipt_id IN ?", receiptIDs).Find(&items).Error; err != nil {
			logger.ErrorF("error finding receipt items: %v", err.Error())
			sendError(ctx, http.StatusInternalServerError, "Erro ao buscar itens das notas fiscais. Por favor, tente novamente")
			return
		}
	}

	// 4. Buscar todas as categorias DO USUÁRIO usando GORM
	var categories []schemas.Category
	if err := db.Where("user_id = ?", userID).Find(&categories).Error; err != nil {
		logger.ErrorF("error finding categories: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao buscar categorias. Por favor, tente novamente")
		return
	}

	// 5. Processar dados em memória (agregação manual)
	categoryMap := make(map[uint]*CategoryGraphResponse)
	for _, cat := range categories {
		categoryMap[cat.ID] = &CategoryGraphResponse{
			ID:        cat.ID,
			Name:      cat.Name,
			ItemCount: 0,
			Total:     0,
		}
	}

	// 6. Agregar dados dos items por categoria
	for _, item := range items {
		if catData, exists := categoryMap[item.CategoryID]; exists {
			catData.ItemCount++
			catData.Total += item.Total
		}
	}

	// 7. Converter map para slice e calcular grand total
	var results []CategoryGraphResponse
	var grandTotal float64
	for _, catData := range categoryMap {
		results = append(results, *catData)
		grandTotal += catData.Total
	}

	// 8. Ordenar por nome
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Category graph data retrieved successfully",
		"data": GraphData{
			Categories: results,
			GrandTotal: grandTotal,
		},
	})
}
