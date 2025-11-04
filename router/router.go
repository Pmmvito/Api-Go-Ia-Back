package router

import (
	"github.com/gin-gonic/gin"
)

// Initialize inicializa o roteador Gin, configura as rotas da API e inicia o servidor.
func Initialize() {
	//Initialize Router
	router := gin.Default()
	//Initialize routes
	InitializeRoutes(router)

	router.Run(":8080")
}
