package main

import (
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/config"
	"github.com/Pmmvito/Golang-Api-Exemple/router"
	"github.com/joho/godotenv"
)

var (
	logger *config.Logger
)

// @title API de Gestão de Notas Fiscais com IA
// @version 2.0
// @description API REST para autenticação JWT, análise inteligente de notas fiscais de supermercado e gestão de categorias
// @description
// @description ## 🚀 Funcionalidades Principais:
// @description - 🔐 **Autenticação JWT** com tokens de 7 dias
// @description - 🤖 **Análise de Notas com IA** usando Google Gemini (2-3s)
// @description - 📊 **Categorização Automática** de produtos de supermercado
// @description - 📝 **CRUD Completo** de recibos e categorias
// @description - 🔍 **Filtros Avançados** por categoria, data e valor
// @description - ⚡ **Respostas Otimizadas** (55% menores)
// @description
// @description ## 📖 Como Usar:
// @description 1. Registre-se em `/register`
// @description 2. Faça login em `/login` para obter o token JWT
// @description 3. Clique em **Authorize** 🔓 (cadeado verde) e cole: `Bearer SEU_TOKEN`
// @description 4. Agora você pode testar todos os endpoints protegidos!
// @description
// @description ## 🛒 Categorias de Supermercado (45+ categorias):
// @description **Alimentos Básicos:** Grãos e Cereais, Massas, Farinhas
// @description **Proteínas:** Carnes Vermelhas, Aves, Peixes, Frios e Embutidos
// @description **Laticínios:** Leite, Queijos, Iogurtes, Manteiga
// @description **Hortifruti:** Frutas, Verduras, Legumes
// @description **Padaria:** Pães, Bolos e Tortas, Biscoitos
// @description **Bebidas:** Refrigerantes, Sucos, Água, Bebidas Alcoólicas, Cafés e Chás
// @description **Congelados:** Congelados, Sorvetes
// @description **Despensa:** Óleos, Temperos, Molhos, Enlatados
// @description **Doces:** Chocolates, Doces e Balas, Sobremesas
// @description **Snacks:** Salgadinhos, Lanches Rápidos
// @description **Higiene:** Higiene Bucal, Corporal, Papel Higiênico, Fraldas
// @description **Limpeza:** Limpeza Geral, Limpeza de Roupas, Descartáveis
// @description **Outros:** Pet Shop, Alimentação Infantil, Utilidades Domésticas
// @description
// @description Use `/categories` para ver todas as categorias com IDs!
// @termsOfService http://swagger.io/terms/

// @contact.name Equipe de Desenvolvimento
// @contact.email suporte@exemplo.com
// @contact.url https://github.com/Pmmvito/Golang-Api-Exemple

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host 147.185.221.212:61489
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Digite "Bearer" seguido do seu token JWT (obtido no login). Exemplo: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

// main é a função de entrada da aplicação.
// Ela inicializa o logger, carrega as variáveis de ambiente, configura o banco de dados,
// define o timezone e inicia o roteador da API.
func main() {
	logger = config.GetLogger("main")

	// Carrega as variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		logger.ErrorF("Erro ao carregar o arquivo .env: %v", err)
		return
	}

	// Inicializa as configurações do projeto, como o banco de dados.
	err = config.Init()
	if err != nil {
		logger.ErrorF("config initialization erro: %v", err)
		return
	}

	// Define o timezone do projeto para America/Sao_Paulo para consistência de datas.
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		logger.WarnF("Não foi possível carregar a timezone America/Sao_Paulo: %v", err)
	} else {
		time.Local = loc
		logger.InfoF("Timezone definida para %s", loc)
	}

	// Inicializa e inicia o roteador da API.
	router.Initialize()
}
