package config

import (
	"os"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitializePostgreSQL conecta ao banco de dados PostgreSQL, realiza a auto-migração para os schemas,
// e cria categorias padrão se elas não existirem. Retorna uma instância de DB GORM ou um erro.
func InitializePostgreSQL() (*gorm.DB, error) {
	logger := GetLogger("postgres")

	// Carrega a DSN da variável de ambiente
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		logger.ErrorF("Variável de ambiente DATABASE_DSN não definida")
		return nil, nil // Ou retorne um erro apropriado
	}

	// Conecta ao banco de dados
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.ErrorF("Erro ao conectar com o PostgreSQL: %v", err)
		return nil, err
	}

	// Migra o schema (ordem importa por causa das FKs)
	err = db.AutoMigrate(
		&schemas.User{},           // 1. Usuários (independente)
		&schemas.TokenBlacklist{}, // 2. Blacklist de tokens (depende de User)
		&schemas.AITokenUsage{},   // 3. Uso de tokens da IA (depende de User)
		&schemas.Category{},       // 4. Categorias (independente)
		&schemas.Product{},        // 5. Produtos (depende de Category)
		&schemas.Receipt{},        // 6. Notas fiscais (depende de User)
		&schemas.ReceiptItem{},    // 7. Itens de nota (depende de Receipt e Product)
		&schemas.ShoppingList{},   // 8. Listas de compras (depende de User)
		&schemas.ListItem{},       // 9. Itens de lista (depende de ShoppingList e Product)
	)
	if err != nil {
		logger.ErrorF("Erro na automigração do PostgreSQL: %v", err)
		return nil, err
	}

	// Cria categorias padrão se não existirem
	createDefaultCategories(db, logger)

	logger.Info("Conexão com o PostgreSQL estabelecida e migração bem-sucedida.")
	return db, nil
}

// createDefaultCategories verifica a existência de categorias padrão no banco de dados
// e as cria se não estiverem presentes. Isso garante que a aplicação tenha um conjunto base de categorias para trabalhar.
func createDefaultCategories(db *gorm.DB, logger *Logger) {
	// Categorias padrão simplificadas e não redundantes para evitar confusão na IA
	defaultCategories := []schemas.Category{
		{Name: "Grãos e Cereais", Description: "Arroz, feijão, aveia, cereais", Icon: "🌾", Color: "#F4A261"},
		{Name: "Massas e Padaria", Description: "Macarrão, pães, bolos e produtos de padaria", Icon: "�", Color: "#E9C46A"},
		{Name: "Pantry e Ingredientes", Description: "Óleos, enlatados, condimentos, farinhas e fermentos", Icon: "�", Color: "#D4A574"},
		{Name: "Proteínas", Description: "Carnes, aves, peixes, ovos e frios", Icon: "�", Color: "#E74C3C"},
		{Name: "Laticínios", Description: "Leite, queijos, iogurtes, manteiga", Icon: "🧀", Color: "#F1C40F"},
		{Name: "Hortifruti", Description: "Frutas e vegetais frescos", Icon: "🥬", Color: "#27AE60"},
		{Name: "Bebidas", Description: "Água, sucos, refrigerantes, cafés, chás e bebidas alcoólicas", Icon: "☕", Color: "#3498DB"},
		{Name: "Congelados e Frios", Description: "Produtos congelados e sorvetes", Icon: "🧊", Color: "#81ECEC"},
		{Name: "Snacks e Doces", Description: "Chocolates, doces, salgadinhos e sobremesas", Icon: "🍫", Color: "#FF7675"},
		{Name: "Higiene e Cuidados Pessoais", Description: "Produtos de higiene pessoal e cuidados", Icon: "🧼", Color: "#A29BFE"},
		{Name: "Limpeza e Utilidades", Description: "Produtos de limpeza, lavanderia e descartáveis", Icon: "🧺", Color: "#0984E3"},
		{Name: "Bebê e Infantil", Description: "Fraldas, papinhas e produtos infantis", Icon: "�", Color: "#FFA07A"},
		{Name: "Pet", Description: "Ração e itens para animais de estimação", Icon: "�", Color: "#FF6348"},
		{Name: "Outros", Description: "Itens diversos não categorizados", Icon: "📦", Color: "#B2BEC3"},
	}

	for _, category := range defaultCategories {
		var exists schemas.Category
		if err := db.Where("name = ?", category.Name).First(&exists).Error; err != nil {
			// Categoria não existe, cria
			if err := db.Create(&category).Error; err != nil {
				logger.WarnF("Erro ao criar categoria padrão '%s': %v", category.Name, err)
			} else {
				logger.InfoF("Categoria padrão criada: %s", category.Name)
			}
		}
	}
}
