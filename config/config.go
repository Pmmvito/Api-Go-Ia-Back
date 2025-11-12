package config

import (
	"fmt"
	"os"
	"strconv"

	"gorm.io/gorm"
)

var (
	// db é um ponteiro para a instância do banco de dados GORM.
	// É inicializado pela função Init e usado em toda a aplicação
	// para interagir com o banco de dados PostgreSQL.
	db *gorm.DB
	// logger é um ponteiro para a instância do Logger customizado.
	// É usado para registrar mensagens com diferentes níveis de severidade em toda a aplicação.
	logger *Logger
)

// Init inicializa a configuração para a aplicação.
// Ele configura a conexão com o banco de dados chamando InitializePostgreSQL.
// Retorna um erro se a inicialização do banco de dados falhar.
func Init() error {
	var err error

	//initialize PostgreSQL
	db, err = InitializePostgreSQL()
	if err != nil {
		return fmt.Errorf("erro initializing postgresql %v: ", err)
	}

	// Inicializar AI Worker Pool
	maxWorkers := getEnvAsInt("MAX_AI_WORKERS", 3)       // 3 workers para Gemini 2.5 Flash Free (10 RPM)
	queueSize := getEnvAsInt("AI_QUEUE_SIZE", 50)         // Fila de 50 jobs

	InitAIWorkerPool(maxWorkers, queueSize)

	return nil
}

// GetPostgreSQL retorna a instância singleton do banco de dados GORM.
// Esta função deve ser chamada para obter uma referência ao banco de dados
// após ter sido inicializada pela função Init.
func GetPostgreSQL() *gorm.DB {
	return db
}

// GetLogger retorna uma nova instância de Logger para um pacote ou contexto específico.
// p é uma string que representa o nome do pacote ou contexto para o qual o logger é criado.
// Ajuda a identificar a origem das mensagens de log.
func GetLogger(p string) *Logger {
	//INITIALIZER LOGGER

	logger = NewLogger(p)
	return logger
}

// getEnvAsInt busca variável de ambiente como int com valor padrão
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
