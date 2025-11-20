package handler


import (
	"github.com/Pmmvito/Golang-Api-Exemple/config"
	"gorm.io/gorm"
)

var (
	logger *config.Logger
	db     *gorm.DB
)

// InitializerHandler inicializa o logger e a instância do banco de dados para o pacote handler.
func InitializerHandler() {
	logger = config.GetLogger("handler")
	db = config.GetPostgreSQL()
}
