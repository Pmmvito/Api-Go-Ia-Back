package handler

import (
	"net/http"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
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

	// Cria a categoria
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
// @Description Get details of a specific category by ID
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID" example(1)
// @Success 200 {object} map[string]interface{} "Category details"
// @Failure 404 {object} ErrorResponse "Category not found"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Router /category/{id} [get]
func GetCategoryHandler(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Category retrieved successfully",
		"data":    category.ToResponse(),
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
// @Description Delete a category permanently. WARNING: This will also remove the category association from all receipt items (sets categoryId to null).
// @Tags 📁 Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID" example(1)
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
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

	if err := db.Delete(&category).Error; err != nil {
		logger.ErrorF("error deleting category: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error deleting category")
		return
	}

	logger.InfoF("Category %s deleted successfully", id)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}
