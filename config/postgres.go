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
	// Categorias padrão reformuladas para serem DISTINTAS e não confundir a IA
	// Cada categoria tem um foco ÚNICO e específico
	defaultCategories := []schemas.Category{
		{Name: "Grãos e Cereais", Description: "Arroz, feijão, lentilha, aveia, granola, cereais matinais", Icon: "🌾", Color: "#F4A261"},
		{Name: "Massas", Description: "Macarrão, lasanha, nhoque, massas secas e frescas", Icon: "🍝", Color: "#E9C46A"},
		{Name: "Padaria", Description: "Pães, baguetes, brioche, croissant, pão de forma", Icon: "🍞", Color: "#D4A574"},
		{Name: "Carnes e Proteínas", Description: "Carne bovina, suína, frango, peixe, frutos do mar, ovos", Icon: "🥩", Color: "#E74C3C"},
		{Name: "Frios e Embutidos", Description: "Presunto, mortadela, salame, peito de peru, salsicha, linguiça", Icon: "🥓", Color: "#C0392B"},
		{Name: "Laticínios", Description: "Leite, queijos, requeijão, creme de leite, iogurtes, manteiga", Icon: "🧀", Color: "#F1C40F"},
		{Name: "Frutas e Vegetais", Description: "Frutas frescas, verduras, legumes, saladas, ervas", Icon: "🥬", Color: "#27AE60"},
		{Name: "Bebidas", Description: "Refrigerante, suco, água, isotônico, energético (NÃO álcool, NÃO café)", Icon: "🥤", Color: "#3498DB"},
		{Name: "Bebidas Alcoólicas", Description: "Cerveja, vinho, destilados, drinks (APENAS bebidas com álcool)", Icon: "🍺", Color: "#8E44AD"},
		{Name: "Café e Chá", Description: "Café em pó, café expresso, chás, infusões, mate (APENAS estas bebidas)", Icon: "☕", Color: "#6F4E37"},
		{Name: "Congelados", Description: "Alimentos congelados, pizzas congeladas, vegetais congelados, pratos prontos congelados", Icon: "🧊", Color: "#81ECEC"},
		{Name: "Doces e Sobremesas", Description: "Chocolates, bombons, balas, gomas, pudim, gelatina, sorvetes", Icon: "🍫", Color: "#FF7675"},
		{Name: "Salgadinhos e Snacks", Description: "Chips, batata frita, amendoim, pipoca, biscoitos salgados", Icon: "🥨", Color: "#FD79A8"},
		{Name: "Condimentos e Temperos", Description: "Sal, açúcar, especiarias, molhos prontos, vinagre, azeite, óleo", Icon: "🧂", Color: "#E67E22"},
		{Name: "Enlatados e Conservas", Description: "Milho, ervilha, atum, sardinha, palmito, azeitona em lata/vidro", Icon: "🥫", Color: "#95A5A6"},
		{Name: "Higiene Pessoal", Description: "Sabonete, shampoo, condicionador, desodorante, creme dental, escova", Icon: "🧼", Color: "#A29BFE"},
		{Name: "Limpeza Doméstica", Description: "Detergente, desinfetante, água sanitária, amaciante, esponja, vassoura", Icon: "🧹", Color: "#0984E3"},
		{Name: "Papel e Descartáveis", Description: "Papel higiênico, papel toalha, guardanapo, copos e pratos descartáveis", Icon: "�", Color: "#74B9FF"},
		{Name: "Bebê e Infantil", Description: "Fraldas, lenços umedecidos, papinhas, leite em pó infantil", Icon: "👶", Color: "#FFA07A"},
		{Name: "Pet Shop", Description: "Ração para cães e gatos, petiscos, areia sanitária para pets", Icon: "🐾", Color: "#FF6348"},
		{Name: "Outros", Description: "Produtos não enquadrados em nenhuma categoria acima", Icon: "📦", Color: "#B2BEC3"},
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
