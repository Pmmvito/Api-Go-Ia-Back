package router

import (
	"os"

	"github.com/gin-gonic/gin"
)

// Initialize inicializa o roteador Gin, configura as rotas da API e inicia o servidor.
func Initialize() {
	//Initialize Router
	router := gin.Default()

	// 🔒 Middlewares de segurança
	router.Use(SecureMiddleware()) // HTTPS + Security Headers
	router.Use(CORSMiddleware())    // CORS seguro

	//Initialize routes
	InitializeRoutes(router)

	// 🔒 Iniciar servidor com TLS em produção
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("ENV") == "production" {
		certFile := os.Getenv("TLS_CERT_FILE")
		keyFile := os.Getenv("TLS_KEY_FILE")

		if certFile != "" && keyFile != "" {
			logger.InfoF("🔒 Iniciando servidor HTTPS na porta %s", port)
			if err := router.RunTLS(":"+port, certFile, keyFile); err != nil {
				logger.ErrorF("Erro ao iniciar servidor HTTPS: %v", err)
				panic(err)
			}
		} else {
			logger.WarnF("⚠️  Produção sem TLS! Configure TLS_CERT_FILE e TLS_KEY_FILE")
			router.Run(":" + port)
		}
	} else {
		logger.InfoF("🚀 Iniciando servidor HTTP (desenvolvimento) na porta %s", port)
		router.Run(":" + port)
	}
}
