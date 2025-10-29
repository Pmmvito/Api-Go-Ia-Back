package handler

import (
	"net/http"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
)

// GetProductsHandler lista todos os produtos
// @Summary Listar todos os produtos
// @Description Lista todos os produtos cadastrados
// @Tags products
// @Produce json
// @Security BearerAuth
// @Success 200 {array} schemas.ProductResponse
// @Router /products [get]
func GetProductsHandler(ctx *gin.Context) {
	var products []schemas.Product

	db.Find(&products)
	ctx.JSON(http.StatusOK, products)
}

// GetProductByIDHandler busca produto por ID
// @Summary Buscar produto por ID
// @Description Busca produto pelo ID
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do produto"
// @Success 200 {object} schemas.ProductResponse
// @Failure 404 {object} map[string]string
// @Router /product/{id} [get]
func GetProductByIDHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	var product schemas.Product

	if err := db.Where("id = ?", id).First(&product).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, product)
}

// GetProductsByNameHandler busca produtos por nome (parcial)
// @Summary Buscar produtos por nome
// @Description Busca produtos pelo nome (parcial)
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param name path string true "Nome do produto (parcial)"
// @Success 200 {array} schemas.ProductResponse
// @Router /products/name/{name} [get]
func GetProductsByNameHandler(ctx *gin.Context) {
	name := ctx.Param("name")
	var products []schemas.Product

	db.Where("name ILIKE ?", "%"+name+"%").Find(&products)
	ctx.JSON(http.StatusOK, products)
}

// GetProductByIDAndNameHandler busca produto por ID e nome
// @Summary Buscar produto por ID e nome
// @Description Busca produto por ID e nome
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do produto"
// @Param name path string true "Nome do produto"
// @Success 200 {object} schemas.ProductResponse
// @Failure 404 {object} map[string]string
// @Router /product/{id}/name/{name} [get]
func GetProductByIDAndNameHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	name := ctx.Param("name")
	var product schemas.Product

	if err := db.Where("id = ? AND name = ?", id, name).First(&product).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, product)
}
