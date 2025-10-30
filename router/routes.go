package router

import (
	docs "github.com/Pmmvito/Golang-Api-Exemple/docs"
	"github.com/Pmmvito/Golang-Api-Exemple/handler"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitializeRoutes configura todos os endpoints da API, dividindo-os em rotas públicas e protegidas.
func InitializeRoutes(router *gin.Engine) {

	//initialize Handler
	handler.InitializerHandler()
	basePatch := "/api/v1"
	docs.SwaggerInfo.BasePath = basePatch

	// Rotas públicas (sem autenticação)
	public := router.Group(basePatch)
	{
		public.POST("/register", handler.RegisterHandler)
		public.POST("/login", handler.LoginHandler)
	}

	// Rotas protegidas (requerem autenticação JWT)
	protected := router.Group(basePatch)
	protected.Use(AuthMiddleware())
	{
		protected.GET("/me", handler.MeHandler)

		// Rotas de categorias
		protected.POST("/category", handler.CreateCategoryHandler)
		protected.GET("/categories", handler.ListCategoriesHandler)
		protected.GET("/category/:id", handler.GetCategoryHandler)
		protected.PATCH("/category/:id", handler.UpdateCategoryHandler)
		protected.DELETE("/category/:id", handler.DeleteCategoryHandler)

		// Rotas de produtos
		protected.GET("/products", handler.GetProductsHandler)
		protected.GET("/products/:id", handler.GetProductByIDHandler)
		// Buscar todos os produtos de uma data específica (YYYY-MM-DD)
		protected.GET("/products/date/:date", handler.GetProductsByDateHandler)
		// Buscar todos os produtos dentro de um período (query params: start, end)
		protected.GET("/products/period", handler.GetProductsByPeriodHandler)

		// Rotas de recibos
		protected.GET("/receipts", handler.GetReceiptsHandler)
		protected.GET("/receipts/date/:date", handler.GetReceiptsByDateHandler)
		protected.GET("/receipts/period", handler.GetReceiptsByPeriodHandler)
		protected.GET("/receipt/:id", handler.GetReceiptByIDHandler)

		// Rotas de recibos básicos (ultra-simplificados para seleção)
		protected.GET("/receipts-basic", handler.GetReceiptsBasicHandler)
		protected.GET("/receipts-basic/date/:date", handler.GetReceiptsBasicByDateHandler)
		protected.GET("/receipts-basic/period", handler.GetReceiptsBasicByPeriodHandler)

		// Rotas de itens
		protected.GET("/items", handler.GetItemsHandler)
		protected.GET("/items/date/:date", handler.GetItemsByDateHandler)
		protected.GET("/item/:id", handler.GetItemByIDHandler)
		// Buscar itens por período (query params: start, end)
		protected.GET("/items/period", handler.GetItemsByPeriodHandler)

		// 🆕 QR Code Flow (2 etapas)
		protected.POST("/scan-qrcode/preview", handler.ScanQRCodePreviewHandler) // Etapa 1: Preview (não salva)
		protected.POST("/scan-qrcode/confirm", handler.ScanQRCodeConfirmHandler) // Etapa 2: Confirma e salva
	}

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
