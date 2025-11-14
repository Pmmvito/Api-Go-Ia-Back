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
		&schemas.RefreshToken{},   // 3. 🔒 Refresh tokens (depende de User)
		&schemas.AITokenUsage{},   // 4. Uso de tokens da IA (depende de User)
		&schemas.PasswordReset{},  // 5. Tokens de recuperação de senha (depende de User)
		&schemas.Category{},       // 6. Categorias (independente)
		&schemas.Product{},        // 7. Produtos (depende de Category)
		&schemas.Receipt{},        // 8. Notas fiscais (depende de User)
		&schemas.ReceiptItem{},    // 9. Itens de nota (depende de Receipt e Product)
		&schemas.ShoppingList{},   // 10. Listas de compras (depende de User)
		&schemas.ListItem{},       // 11. Itens de lista (depende de ShoppingList e Product)
	)
	if err != nil {
		logger.ErrorF("Erro na automigração do PostgreSQL: %v", err)
		return nil, err
	}

	// Não cria mais categorias padrão globais aqui
	// As categorias serão criadas individualmente para cada usuário no registro

	logger.Info("Conexão com o PostgreSQL estabelecida e migração bem-sucedida.")
	return db, nil
}

// CreateDefaultCategoriesForUser cria categorias padrão para um usuário específico
// Esta função deve ser chamada após o registro de um novo usuário
func CreateDefaultCategoriesForUser(db *gorm.DB, userID uint) error {
	logger := GetLogger("postgres")

	// Categorias padrão reformuladas para serem DISTINTAS e não confundir a IA
	// Cada categoria tem um foco ÚNICO e específico
	defaultCategories := []schemas.Category{
		{UserID: userID, Name: "Não categorizado", Description: "Itens aguardando categorização", Icon: "❓", Color: "#95A5A6"},
		{UserID: userID, Name: "Grãos e Cereais", Description: "Arroz, feijão, lentilha, aveia, granola, cereais matinais", Icon: "🌾", Color: "#F4A261"},
		{UserID: userID, Name: "Massas", Description: "Macarrão, lasanha, nhoque, massas secas e frescas", Icon: "🍝", Color: "#E9C46A"},
		{UserID: userID, Name: "Padaria", Description: "Pães, baguetes, brioche, croissant, pão de forma", Icon: "🍞", Color: "#D4A574"},
		{UserID: userID, Name: "Carnes e Proteínas", Description: "Carne bovina, suína, frango, peixe, frutos do mar, ovos", Icon: "🥩", Color: "#E74C3C"},
		{UserID: userID, Name: "Frios e Embutidos", Description: "Presunto, mortadela, salame, peito de peru, salsicha, linguiça", Icon: "🥓", Color: "#C0392B"},
		{UserID: userID, Name: "Laticínios", Description: "Leite, queijos, requeijão, creme de leite, iogurtes, manteiga", Icon: "🧀", Color: "#F1C40F"},
		{UserID: userID, Name: "Frutas e Vegetais", Description: "Frutas frescas, verduras, legumes, saladas, ervas", Icon: "🥬", Color: "#27AE60"},
		{UserID: userID, Name: "Bebidas", Description: "Refrigerante, suco, água, isotônico, energético (NÃO álcool, NÃO café)", Icon: "🥤", Color: "#3498DB"},
		{UserID: userID, Name: "Bebidas Alcoólicas", Description: "Cerveja, vinho, destilados, drinks (APENAS bebidas com álcool)", Icon: "🍺", Color: "#8E44AD"},
		{UserID: userID, Name: "Café e Chá", Description: "Café em pó, café expresso, chás, infusões, mate (APENAS estas bebidas)", Icon: "☕", Color: "#6F4E37"},
		{UserID: userID, Name: "Congelados", Description: "Alimentos congelados, pizzas congeladas, vegetais congelados, pratos prontos congelados", Icon: "🧊", Color: "#81ECEC"},
		{UserID: userID, Name: "Doces e Sobremesas", Description: "Chocolates, bombons, balas, gomas, pudim, gelatina, sorvetes", Icon: "🍫", Color: "#FF7675"},
		{UserID: userID, Name: "Salgadinhos e Snacks", Description: "Chips, batata frita, amendoim, pipoca, biscoitos salgados", Icon: "🥨", Color: "#FD79A8"},
		{UserID: userID, Name: "Condimentos e Temperos", Description: "Sal, açúcar, especiarias, molhos prontos, vinagre, azeite, óleo", Icon: "🧂", Color: "#E67E22"},
		{UserID: userID, Name: "Enlatados e Conservas", Description: "Milho, ervilha, atum, sardinha, palmito, azeitona em lata/vidro", Icon: "🥫", Color: "#95A5A6"},
		{UserID: userID, Name: "Higiene Pessoal", Description: "Sabonete, shampoo, condicionador, desodorante, creme dental, escova", Icon: "🧼", Color: "#A29BFE"},
		{UserID: userID, Name: "Limpeza Doméstica", Description: "Detergente, desinfetante, água sanitária, amaciante, esponja, vassoura", Icon: "🧹", Color: "#0984E3"},
		{UserID: userID, Name: "Papel e Descartáveis", Description: "Papel higiênico, papel toalha, guardanapo, copos e pratos descartáveis", Icon: "🧻", Color: "#74B9FF"},
		{UserID: userID, Name: "Bebê e Infantil", Description: "Fraldas, lenços umedecidos, papinhas, leite em pó infantil", Icon: "👶", Color: "#FFA07A"},
		{UserID: userID, Name: "Pet Shop", Description: "Ração para cães e gatos, petiscos, areia sanitária para pets", Icon: "🐾", Color: "#FF6348"},
		{UserID: userID, Name: "Outros", Description: "Produtos não enquadrados em nenhuma categoria acima", Icon: "📦", Color: "#B2BEC3"},
	}

	for _, category := range defaultCategories {
		if err := db.Create(&category).Error; err != nil {
			logger.WarnF("Erro ao criar categoria padrão '%s' para usuário %d: %v", category.Name, userID, err)
			return err
		}
	}

	logger.InfoF("Categorias padrão criadas para usuário %d", userID)
	return nil
}
