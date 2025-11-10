package handler

import (
	"net/http"
	"sort"
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
// @Description Create a new expense category for organizing receipt items
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCategoryRequest true "Category data (name is required, description/icon/color are optional)"
// @Success 201 {object} map[string]interface{} "Category created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /category [post]
func CreateCategoryHandler(ctx *gin.Context) {
	var request CreateCategoryRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Verifica se existe uma categoria deletada com o mesmo nome
	var existingCategory schemas.Category
	err := db.Unscoped().Where("name = ?", request.Name).First(&existingCategory).Error

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
				sendError(ctx, http.StatusInternalServerError, "Error reactivating category")
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
			sendError(ctx, http.StatusConflict, "Category with this name already exists")
			return
		}
	}

	// Categoria não existe, cria uma nova
	category := schemas.Category{
		Name:        request.Name,
		Description: request.Description,
		Icon:        request.Icon,
		Color:       request.Color,
	}

	if err := db.Create(&category).Error; err != nil {
		logger.ErrorF("error creating category: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error creating category")
		return
	}

	logger.InfoF("Category created with ID: %d", category.ID)
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Category created successfully",
		"data":    category.ToResponse(),
	})
}

// @Summary List all categories
// @Description Get all expense categories sorted by name. Returns: 1-Alimentação, 2-Transporte, 3-Saúde, 4-Lazer, 5-Educação, 6-Moradia, 7-Vestuário, 8-Outros
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of categories with count"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /categories [get]
func ListCategoriesHandler(ctx *gin.Context) {
	var categories []schemas.Category
	if err := db.Order("name ASC").Find(&categories).Error; err != nil {
		logger.ErrorF("error listing categories: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error listing categories")
		return
	}

	// Converte para response
	var responses []schemas.CategoryResponse
	for _, category := range categories {
		responses = append(responses, category.ToResponse())
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Categories retrieved successfully",
		"data":    responses,
		"count":   len(responses),
	})
}

// @Summary Get category details
// @Description Get details of a specific category by ID including all items that belong to this category
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID" example(1)
// @Success 200 {object} map[string]interface{} "Category details with items"
// @Failure 404 {object} ErrorResponse "Category not found"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Router /category/{id} [get]
func GetCategoryHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Busca a categoria
	var category schemas.Category
	if err := db.First(&category, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Category not found")
		return
	}

	// Busca todos os itens dessa categoria com informações do recibo usando GORM
	var receiptItems []schemas.ReceiptItem
	err := db.Preload("Product").
		Preload("Receipt").
		Where("category_id = ?", id).
		Order("receipt_id DESC").
		Find(&receiptItems).Error

	if err != nil {
		logger.ErrorF("error getting category items: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error getting category items")
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

	// Monta resposta
	response := CategoryWithItemsResponse{
		CategoryResponse: category.ToResponse(),
		Items:            items,
		ItemCount:        len(items),
		TotalValue:       totalValue,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Category retrieved successfully",
		"data":    response,
	})
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
// @Failure 400 {object} ErrorResponse "Invalid request or no fields to update"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} ErrorResponse "Category not found"
// @Router /category/{id} [patch]
func UpdateCategoryHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Category ID is required")
		return
	}

	var request UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var category schemas.Category
	if err := db.First(&category, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Category not found")
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
		sendError(ctx, http.StatusBadRequest, "No fields to update")
		return
	}

	if err := db.Save(&category).Error; err != nil {
		logger.ErrorF("error updating category: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error updating category")
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
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse "Cannot delete 'Não categorizado' category"
// @Router /category/{id} [delete]
func DeleteCategoryHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, "Category ID is required")
		return
	}

	var category schemas.Category
	if err := db.First(&category, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Category not found")
		return
	}

	// Não permite deletar a categoria "Não categorizado"
	if category.Name == "Não categorizado" {
		sendError(ctx, http.StatusBadRequest, "Cannot delete 'Não categorizado' category")
		return
	}

	// Busca a categoria "Não categorizado"
	var uncategorized schemas.Category
	if err := db.Where("name = ?", "Não categorizado").First(&uncategorized).Error; err != nil {
		logger.ErrorF("'Não categorizado' category not found: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "System category 'Não categorizado' not found")
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
		sendError(ctx, http.StatusInternalServerError, "Error moving items to uncategorized")
		return
	}

	itemsMoved := result.RowsAffected
	logger.InfoF("Moved %d items from category %s to 'Não categorizado'", itemsMoved, category.Name)

	// Deleta a categoria
	if err := tx.Delete(&category).Error; err != nil {
		tx.Rollback()
		logger.ErrorF("error deleting category: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error deleting category")
		return
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		logger.ErrorF("error committing transaction: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error committing deletion")
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
// @Failure 400 {object} ErrorResponse "Invalid date format"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
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
			sendError(ctx, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
			return
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			sendError(ctx, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
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
		sendError(ctx, http.StatusInternalServerError, "Error getting receipts")
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
			sendError(ctx, http.StatusInternalServerError, "Error getting receipt items")
			return
		}
	}

	// 4. Buscar todas as categorias usando GORM
	var categories []schemas.Category
	if err := db.Find(&categories).Error; err != nil {
		logger.ErrorF("error finding categories: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error getting categories")
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
