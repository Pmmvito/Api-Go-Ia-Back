package config

import (
	"os"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
		&schemas.User{},         // 1. Usuários (independente)
		&schemas.Category{},     // 2. Categorias (independente)
		&schemas.Product{},      // 3. Produtos (depende de Category)
		&schemas.Receipt{},      // 4. Notas fiscais (depende de User)
		&schemas.ReceiptItem{},  // 5. Itens de nota (depende de Receipt e Product)
		&schemas.ShoppingList{}, // 6. Listas de compras (depende de User)
		&schemas.ListItem{},     // 7. Itens de lista (depende de ShoppingList e Product)
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

// createDefaultCategories cria categorias padrão no banco de dados
func createDefaultCategories(db *gorm.DB, logger *Logger) {
	defaultCategories := []schemas.Category{
		// Alimentos Básicos
		{Name: "Grãos e Cereais", Description: "Arroz, feijão, lentilha, grão-de-bico, aveia", Icon: "🌾", Color: "#F4A261"},
		{Name: "Massas", Description: "Macarrão, lasanha, nhoque, massas frescas", Icon: "🍝", Color: "#E9C46A"},
		{Name: "Farinhas e Fermentos", Description: "Farinha de trigo, fubá, fermento, polvilho", Icon: "🧺", Color: "#D4A574"},

		// Proteínas
		{Name: "Carnes Vermelhas", Description: "Bovina, suína, cordeiro", Icon: "🥩", Color: "#E74C3C"},
		{Name: "Aves", Description: "Frango, chester, peru, codorna", Icon: "�", Color: "#F39C12"},
		{Name: "Peixes e Frutos do Mar", Description: "Peixes, camarão, lula, polvo", Icon: "🐟", Color: "#3498DB"},
		{Name: "Frios e Embutidos", Description: "Presunto, mortadela, salame, salsicha", Icon: "🥓", Color: "#E67E22"},

		// Laticínios
		{Name: "Leite e Derivados", Description: "Leite integral, desnatado, sem lactose", Icon: "🥛", Color: "#ECF0F1"},
		{Name: "Queijos", Description: "Mussarela, prato, parmesão, gorgonzola", Icon: "🧀", Color: "#F1C40F"},
		{Name: "Iogurtes", Description: "Iogurte natural, grego, com frutas", Icon: "�", Color: "#F8E5B9"},
		{Name: "Manteiga e Margarina", Description: "Manteiga, margarina, creme de leite", Icon: "🧈", Color: "#FEF5E7"},

		// Hortifruti
		{Name: "Frutas", Description: "Maçã, banana, laranja, mamão, melancia", Icon: "🍎", Color: "#E74C3C"},
		{Name: "Verduras", Description: "Alface, rúcula, couve, espinafre", Icon: "🥬", Color: "#27AE60"},
		{Name: "Legumes", Description: "Tomate, cenoura, batata, cebola, pimentão", Icon: "🥕", Color: "#F39C12"},

		// Padaria e Confeitaria
		{Name: "Pães", Description: "Pão francês, integral, de forma, bisnaga", Icon: "🍞", Color: "#D4A574"},
		{Name: "Bolos e Tortas", Description: "Bolo caseiro, tortas doces e salgadas", Icon: "🍰", Color: "#FF7675"},
		{Name: "Biscoitos e Bolachas", Description: "Biscoitos doces, salgados, cream cracker", Icon: "🍪", Color: "#FDCB6E"},

		// Bebidas
		{Name: "Refrigerantes", Description: "Coca-Cola, Guaraná, Sprite, Fanta", Icon: "�", Color: "#E74C3C"},
		{Name: "Sucos", Description: "Sucos naturais, de caixinha, polpas", Icon: "🧃", Color: "#F39C12"},
		{Name: "Água", Description: "Água mineral, com gás, sem gás", Icon: "💧", Color: "#3498DB"},
		{Name: "Bebidas Alcoólicas", Description: "Cerveja, vinho, destilados", Icon: "🍺", Color: "#8E44AD"},
		{Name: "Cafés e Chás", Description: "Café em pó, cápsulas, chás diversos", Icon: "☕", Color: "#6C3483"},

		// Congelados
		{Name: "Congelados", Description: "Legumes congelados, pratos prontos, sorvetes", Icon: "🧊", Color: "#81ECEC"},
		{Name: "Sorvetes", Description: "Sorvetes de massa, picolés, açaí", Icon: "🍦", Color: "#DDA0DD"},

		// Despensa
		{Name: "Óleos e Azeites", Description: "Óleo de soja, azeite de oliva, óleo de coco", Icon: "🫒", Color: "#9ACD32"},
		{Name: "Temperos e Condimentos", Description: "Sal, açúcar, pimenta, alho, cebola em pó", Icon: "�", Color: "#FAB1A0"},
		{Name: "Molhos", Description: "Molho de tomate, ketchup, mostarda, maionese", Icon: "🍅", Color: "#E74C3C"},
		{Name: "Enlatados e Conservas", Description: "Milho, ervilha, atum, sardinha", Icon: "�", Color: "#95A5A6"},

		// Doces e Sobremesas
		{Name: "Chocolates", Description: "Barras de chocolate, bombons, ovos de páscoa", Icon: "🍫", Color: "#6C3483"},
		{Name: "Doces e Balas", Description: "Balas, pirulitos, chicletes, doces diversos", Icon: "🍬", Color: "#FF6B9D"},
		{Name: "Sobremesas Prontas", Description: "Pudins, gelatinas, mousses", Icon: "🍮", Color: "#FDCB6E"},

		// Snacks e Petiscos
		{Name: "Salgadinhos", Description: "Chips, pipoca, amendoim, castanhas", Icon: "🥨", Color: "#F39C12"},
		{Name: "Lanches Rápidos", Description: "Barras de cereal, granola, sanduíches naturais", Icon: "🥪", Color: "#E9C46A"},

		// Higiene Pessoal
		{Name: "Higiene Bucal", Description: "Pasta de dente, escova, fio dental, enxaguante", Icon: "🪥", Color: "#74B9FF"},
		{Name: "Higiene Corporal", Description: "Sabonete, shampoo, condicionador, desodorante", Icon: "🧼", Color: "#A29BFE"},
		{Name: "Papel Higiênico", Description: "Papel higiênico, lenços de papel", Icon: "🧻", Color: "#DFE6E9"},
		{Name: "Fraldas e Absorventes", Description: "Fraldas descartáveis, absorventes", Icon: "👶", Color: "#FD79A8"},

		// Limpeza
		{Name: "Limpeza Geral", Description: "Detergente, desinfetante, água sanitária", Icon: "�", Color: "#0984E3"},
		{Name: "Limpeza de Roupas", Description: "Sabão em pó, amaciante, alvejante", Icon: "🧺", Color: "#6C5CE7"},
		{Name: "Descartáveis", Description: "Guardanapos, copos, pratos descartáveis", Icon: "🥤", Color: "#B2BEC3"},

		// Pet
		{Name: "Pet Shop", Description: "Ração, petiscos, areia sanitária", Icon: "🐾", Color: "#FF6348"},

		// Bebê
		{Name: "Alimentação Infantil", Description: "Papinhas, leite em pó, mingau", Icon: "�", Color: "#FFA07A"},

		// Outros
		{Name: "Utilidades Domésticas", Description: "Pilhas, lâmpadas, velas, fósforos", Icon: "💡", Color: "#95A5A6"},
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
