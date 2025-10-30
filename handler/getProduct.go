package handler

import (
	"net/http"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
)

// GetProductsHandler lida com a requisição para listar todos os produtos cadastrados no sistema.
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

// GetProductByIDHandler lida com a requisição para buscar um produto pelo seu ID.
// @Summary Buscar produto por ID
// @Description Busca produto pelo ID
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do produto"
// @Success 200 {object} schemas.ProductResponse
// @Failure 404 {object} map[string]string
// @Router /products/{id} [get]
func GetProductByIDHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	var product schemas.Product

	if err := db.Where("id = ?", id).First(&product).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, product)
}

// GetProductsByDateHandler busca todos os produtos de uma data específica
// @Summary Buscar produtos por data
// @Description Retorna todos os produtos criados em uma data específica (YYYY-MM-DD)
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param date path string true "Data no formato YYYY-MM-DD"
// @Success 200 {array} schemas.ProductResponse
// @Router /products/date/{date} [get]
func GetProductsByDateHandler(ctx *gin.Context) {
	dateStr := ctx.Param("date")
	d, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use YYYY-MM-DD"})
		return
	}

	start := d
	end := d.Add(24 * time.Hour)

	var products []schemas.Product
	db.Where("created_at >= ? AND created_at < ?", start, end).Find(&products)
	ctx.JSON(http.StatusOK, products)
}

// GetProductsByPeriodHandler busca todos os produtos dentro de um período de tempo
// @Summary Buscar produtos por período
// @Description Retorna todos os produtos criados entre as query params `start` e `end` (RFC3339 ou YYYY-MM-DD). Ambos são obrigatórios.
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param start query string true "Data/hora inicial (RFC3339 ou YYYY-MM-DD)"
// @Param end query string true "Data/hora final (RFC3339 ou YYYY-MM-DD)"
// @Success 200 {array} schemas.ProductResponse
// @Router /products/period [get]
func GetProductsByPeriodHandler(ctx *gin.Context) {
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
		// para incluir o dia final até 23:59:59
		end = e.Add(24 * time.Hour)
	}

	var products []schemas.Product
	db.Where("created_at >= ? AND created_at < ?", start, end).Find(&products)
	ctx.JSON(http.StatusOK, products)
}
