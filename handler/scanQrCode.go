package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/PuerkitoBio/goquery"
)

// NFCeData representa os dados extraídos da NFC-e via scraping
type NFCeData struct {
	StoreName  string
	Date       string
	Items      []NFCeItem
	Subtotal   float64
	Discount   float64
	Total      float64
	ItemsCount int
	AccessKey  string
	Number     string
}

// NFCeItem representa um item da NFC-e
type NFCeItem struct {
	ItemNumber  int
	Code        string
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	Total       float64
}

// scrapeNFCe faz scraping da página da NFC-e e extrai os dados
func scrapeNFCe(url string) (*NFCeData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NFC-e page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch NFC-e page: status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	data := &NFCeData{
		Items: []NFCeItem{},
	}

	// Tenta extrair o nome da loja do elemento específico
	data.StoreName = strings.TrimSpace(doc.Find("#u20.txtTopo").First().Text())
	if data.StoreName == "" {
		// Fallback: tenta outros seletores
		data.StoreName = strings.TrimSpace(doc.Find(".txtCenter .text").First().Text())
	}
	if data.StoreName == "" {
		data.StoreName = strings.TrimSpace(doc.Find("#infos .text").First().Text())
	}

	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		if strings.Contains(text, "Número:") && data.Number == "" {
			re := regexp.MustCompile(`Número:\s*(\d+)`)
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				data.Number = matches[1]
			}
		}

		if strings.Contains(text, "Emissão:") && data.Date == "" {
			re := regexp.MustCompile(`Emissão:\s*(\d{2}/\d{2}/\d{4})`)
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				dateParts := strings.Split(matches[1], "/")
				if len(dateParts) == 3 {
					data.Date = fmt.Sprintf("%s-%s-%s", dateParts[2], dateParts[1], dateParts[0])
				}
			}
		}
	})

	data.AccessKey = strings.ReplaceAll(strings.TrimSpace(doc.Find(".chave").Text()), " ", "")

	itemNum := 0
	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		fullText := s.Text()
		if !strings.Contains(fullText, "Código:") {
			return
		}

		itemNum++

		description := ""
		if idx := strings.Index(fullText, " (Código:"); idx > 0 {
			description = strings.TrimSpace(fullText[:idx])
		}

		code := ""
		re := regexp.MustCompile(`\(Código:\s*([^\)]+)\)`)
		if matches := re.FindStringSubmatch(fullText); len(matches) > 1 {
			code = strings.TrimSpace(matches[1])
		}

		quantity := 0.0
		unit := "UN"
		re = regexp.MustCompile(`Qtde\.:\s*([0-9,\.]+)\s+UN:\s*([A-Z]+)`)
		if matches := re.FindStringSubmatch(fullText); len(matches) > 2 {
			quantityStr := strings.ReplaceAll(matches[1], ",", ".")
			quantity, _ = strconv.ParseFloat(quantityStr, 64)
			unit = matches[2]
		}

		total := 0.0
		totalText := strings.TrimSpace(s.Find("span.valor").First().Text())
		totalText = strings.ReplaceAll(totalText, "\u00a0", " ")
		if totalValue := extractNumericValue(totalText); totalValue != "" {
			total = parseFloat(totalValue)
		}

		unitPrice := 0.0
		unitPriceText := strings.TrimSpace(s.Find("span.RvlUnit").First().Text())
		unitPriceText = strings.ReplaceAll(unitPriceText, "\u00a0", " ")
		if unitPriceValue := extractNumericValue(unitPriceText); unitPriceValue != "" {
			unitPrice = parseFloat(unitPriceValue)
		} else if total > 0 && quantity > 0 {
			unitPrice = total / quantity
		}

		unitPrice = math.Round(unitPrice*100) / 100

		if description != "" {
			data.Items = append(data.Items, NFCeItem{
				ItemNumber:  itemNum,
				Code:        code,
				Description: description,
				Quantity:    quantity,
				Unit:        unit,
				UnitPrice:   unitPrice,
				Total:       total,
			})
		}
	})

	fullHTML := doc.Text()

	re := regexp.MustCompile(`Qtd\.\s*total\s*de\s*itens:\s*(\d+)`)
	if matches := re.FindStringSubmatch(fullHTML); len(matches) > 1 {
		data.ItemsCount, _ = strconv.Atoi(matches[1])
	}

	re = regexp.MustCompile(`Valor\s*total\s*R\$:\s*([0-9,.]+)`)
	if matches := re.FindStringSubmatch(fullHTML); len(matches) > 1 {
		data.Subtotal = parseFloat(matches[1])
	}

	re = regexp.MustCompile(`Descontos\s*R\$:\s*([0-9,.]+)`)
	if matches := re.FindStringSubmatch(fullHTML); len(matches) > 1 {
		data.Discount = parseFloat(matches[1])
	}

	re = regexp.MustCompile(`Valor\s*a\s*pagar\s*R\$:\s*([0-9,.]+)`)
	if matches := re.FindStringSubmatch(fullHTML); len(matches) > 1 {
		data.Total = parseFloat(matches[1])
	}

	if len(data.Items) == 0 {
		return nil, fmt.Errorf("no items found in NFC-e. Please check the QR Code URL")
	}

	logger.InfoF("✅ NFC-e scraped successfully: %s - %d items - Total: R$ %.2f", data.StoreName, len(data.Items), data.Total)

	return data, nil
}

// parseFloat converte string brasileira (1.683,25) para float64
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func extractNumericValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	raw = strings.ReplaceAll(raw, "\u00a0", " ")
	raw = strings.ReplaceAll(raw, "R$", "")
	clean := regexp.MustCompile(`\d[\d\.,]*`).FindString(raw)
	return strings.TrimSpace(clean)
}

// normalizeUnit padroniza as unidades de medida
func normalizeUnit(unit string) string {
	// Remove espaços e converte para minúsculas
	unit = strings.TrimSpace(unit)
	unit = strings.ToLower(unit)

	// Padroniza unidades comuns
	switch unit {
	case "un", "unid", "unidade", "und", "pç", "peça":
		return "un"
	case "kg", "kilo", "quilograma", "quilogramas":
		return "kg"
	case "g", "gr", "grama", "gramas":
		return "g"
	case "l", "lt", "litro", "litros":
		return "l"
	case "ml", "millilitro", "millilitros":
		return "ml"
	case "cx", "caixa":
		return "cx"
	case "pc", "pct", "pacote":
		return "pct"
	case "dz", "duzia", "dúzia":
		return "dz"
	default:
		// Retorna em minúsculas se não reconhecer
		return unit
	}
}

// CategorizedItem representa um item com sua categoria identificada pela IA
type CategorizedItem struct {
	Description string
	CategoryID  uint
}

// categorizeItemsWithAI usa o Gemini para categorizar os itens extraídos do scraping
func categorizeItemsWithAI(items []NFCeItem) ([]CategorizedItem, error) {
	logger.InfoF("🤖 categorizeItemsWithAI called with %d items", len(items))

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		logger.ErrorF("❌ GEMINI_API_KEY não configurada")
		return nil, fmt.Errorf("GEMINI_API_KEY não configurada")
	}
	logger.InfoF("✅ GEMINI_API_KEY found (length: %d)", len(apiKey))

	// Pega modelo do ambiente ou usa padrão
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-1.5-flash"
	}
	logger.InfoF("📦 Using model: %s", model)

	// Modelos preview/experimentais usam v1beta, modelos estáveis usam v1
	apiVersion := "v1"
	if strings.Contains(model, "preview") || strings.Contains(model, "exp-") || strings.Contains(model, "2.5") {
		apiVersion = "v1beta"
	}
	logger.InfoF("🔧 Using API version: %s", apiVersion)

	// Busca categorias disponíveis
	var categories []schemas.Category
	db.Order("name ASC").Find(&categories)

	if len(categories) == 0 {
		logger.ErrorF("❌ No categories found in database")
		return nil, fmt.Errorf("no categories found in database")
	}
	logger.InfoF("✅ Found %d categories in database", len(categories))

	// Monta prompt para categorização
	logger.InfoF("📝 Building categorization prompt...")
	prompt := buildCategorizationPrompt(items, categories)
	logger.InfoF("📝 Prompt built (%d chars)", len(prompt))

	// Prepara request para Gemini
	parts := []GeminiPart{
		{Text: prompt},
	}

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{Parts: parts},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.ErrorF("❌ Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Faz request para Gemini
	logger.InfoF("🌐 Calling Gemini API (model: %s, apiVersion: %s)...", model, apiVersion)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/%s/models/%s:generateContent?key=%s", apiVersion, model, apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.ErrorF("❌ Failed to call Gemini API: %v", err)
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	logger.InfoF("✅ Gemini API responded with status: %d", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorF("❌ Failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		logger.ErrorF("❌ Gemini API error (status %d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	logger.InfoF("📄 Response body length: %d bytes", len(body))

	// Parse response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		logger.ErrorF("❌ Failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		logger.ErrorF("❌ No response from Gemini")
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Extrai JSON da resposta
	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	logger.InfoF("📝 Gemini response text length: %d chars", len(responseText))
	logger.InfoF("📝 First 200 chars: %s", responseText[:min(200, len(responseText))])

	// Remove markdown code blocks se existirem
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	// Parse JSON de categorização
	var categorizedItems []CategorizedItem
	if err := json.Unmarshal([]byte(responseText), &categorizedItems); err != nil {
		logger.ErrorF("❌ Failed to parse categorization JSON: %v", err)
		logger.ErrorF("❌ Response text: %s", responseText)
		return nil, fmt.Errorf("failed to parse categorization JSON: %w - Response: %s", err, responseText)
	}

	logger.InfoF("✅ Successfully categorized %d items", len(categorizedItems))
	// Log primeiro item como exemplo
	if len(categorizedItems) > 0 {
		logger.InfoF("📋 Example: '%s' -> Category ID %d", categorizedItems[0].Description, categorizedItems[0].CategoryID)
	}

	return categorizedItems, nil
}

// Helper function para min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildCategorizationPrompt constrói o prompt para categorização
func buildCategorizationPrompt(items []NFCeItem, categories []schemas.Category) string {
	var builder strings.Builder

	builder.WriteString("Você é um especialista em categorização de produtos de supermercado.\n\n")
	builder.WriteString("TAREFA: Analise os itens da lista abaixo e atribua a melhor categoria para cada um.\n\n")

	// Lista de categorias disponíveis
	builder.WriteString("CATEGORIAS DISPONÍVEIS:\n")
	for _, cat := range categories {
		builder.WriteString(fmt.Sprintf("ID %d: %s", cat.ID, cat.Name))
		if cat.Icon != "" {
			builder.WriteString(fmt.Sprintf(" %s", cat.Icon))
		}
		if cat.Description != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", cat.Description))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")

	// Lista de itens para categorizar
	builder.WriteString("ITENS PARA CATEGORIZAR:\n")
	for i, item := range items {
		builder.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, item.Description, item.Unit))
	}
	builder.WriteString("\n")

	builder.WriteString("INSTRUÇÕES:\n")
	builder.WriteString("1. Para cada item, escolha o ID da categoria mais adequada\n")
	builder.WriteString("2. Use o ID numérico da categoria (ex: 1, 2, 3...)\n")
	builder.WriteString("3. Se não tiver certeza, escolha a categoria mais próxima\n")
	builder.WriteString("4. Retorne APENAS um array JSON válido no formato:\n")
	builder.WriteString("[\n")
	builder.WriteString("  {\"description\": \"NOME DO ITEM\", \"categoryId\": 1},\n")
	builder.WriteString("  {\"description\": \"NOME DO ITEM 2\", \"categoryId\": 2}\n")
	builder.WriteString("]\n\n")
	builder.WriteString("IMPORTANTE:\n")
	builder.WriteString("- Retorne APENAS o JSON, sem texto adicional\n")
	builder.WriteString("- Não adicione comentários ou explicações\n")
	builder.WriteString("- Mantenha a mesma ordem dos itens\n")
	builder.WriteString("- Use apenas IDs de categorias que existem na lista acima\n\n")
	builder.WriteString("RETORNE O JSON AGORA:\n")

	return builder.String()
}

// findBestCategoryForItem tenta encontrar a melhor categoria para um item (FALLBACK)
func findBestCategoryForItem(description string) *uint {
	desc := strings.ToLower(description)

	// Mapa de palavras-chave para categorias (IDs das 45+ categorias)
	keywords := map[string][]uint{
		// Grãos e Cereais
		"arroz":    {1},
		"feijao":   {1},
		"feijão":   {1},
		"lentilha": {1},
		"grao":     {1},
		"grão":     {1},
		"cereal":   {1},
		// Massas
		"macarrao":  {2},
		"macarrão":  {2},
		"mac":       {2},
		"massa":     {2},
		"espaguete": {2},
		"penne":     {2},
		"lasanha":   {2},
		"parafuso":  {2},
		// Carnes
		"carne":   {4},
		"boi":     {4},
		"patinho": {4},
		"acem":    {4},
		"picanha": {4},
		// Aves
		"frango": {5},
		"peito":  {5},
		"moela":  {5},
		"asa":    {5},
		// Frios
		"queijo":    {7},
		"mussarela": {7},
		"presunto":  {7},
		"mortadela": {7},
		"mort":      {7},
		"salsicha":  {7},
		"salame":    {7},
		"linguica":  {7},
		"calabresa": {7},
		"bacon":     {7},
		// Laticínios
		"leite":     {8},
		"iogurte":   {10},
		"manteiga":  {11},
		"margarina": {11},
		"requeijao": {11},
		"requeijão": {11},
		// Frutas
		"banana":  {12},
		"maca":    {12},
		"maçã":    {12},
		"laranja": {12},
		"melao":   {12},
		"melão":   {12},
		"mamao":   {12},
		"mamão":   {12},
		// Pães
		"pao": {15},
		"pão": {15},
		// Biscoitos
		"biscoito": {17},
		"bisc":     {17},
		"bolacha":  {17},
		"wafer":    {17},
		// Bebidas
		"refrig":   {18},
		"refriger": {18},
		"coca":     {18},
		"pepsi":    {18},
		"fanta":    {18},
		"guarana":  {18},
		"guaraná":  {18},
		// Sucos
		"suco":     {19},
		"tang":     {19},
		"refresco": {19},
		// Água
		"agua": {20},
		"água": {20},
		// Bebidas Alcoólicas
		"cerveja":    {21},
		"vinho":      {21},
		"vodka":      {21},
		"whisky":     {21},
		"cachaça":    {21},
		"caipirinha": {21},
		// Cafés
		"cafe":    {22},
		"café":    {22},
		"nescafe": {22},
		"nescafé": {22},
		// Sorvetes
		"sorvete": {24},
		// Óleos
		"oleo":   {25},
		"óleo":   {25},
		"azeite": {25},
		// Temperos
		"sal":     {26},
		"pimenta": {26},
		"alho":    {26},
		"cebola":  {26},
		"oregano": {26},
		"tempero": {26},
		"sazon":   {26},
		"paprica": {26},
		// Molhos
		"molho":    {27},
		"ketchup":  {27},
		"maionese": {27},
		"mostarda": {27},
		// Enlatados
		"sardinha": {28},
		"atum":     {28},
		"milho":    {28},
		"ervilha":  {28},
		"azeitona": {28},
		// Chocolates
		"chocolate":    {29},
		"choc":         {29},
		"nescau":       {29},
		"achocolatado": {29},
		// Doces
		"doce":     {30},
		"bala":     {30},
		"chiclete": {30},
		// Salgadinhos
		"salgadinho": {32},
		"salg":       {32},
		"doritos":    {32},
		"chips":      {32},
		// Higiene Bucal
		"escova":       {35},
		"creme dental": {35},
		"enxaguante":   {35},
		"fio dental":   {35},
		// Higiene Corporal
		"sabonete":      {36},
		"shampoo":       {36},
		"shamp":         {36},
		"condicionador": {36},
		"desodorante":   {36},
		"desod":         {36},
		// Papel Higiênico
		"papel hig": {37},
		"toalha":    {37},
		// Limpeza
		"detergente":   {41},
		"deterg":       {41},
		"sabao":        {41},
		"sabão":        {41},
		"alvejante":    {41},
		"amaciante":    {41},
		"limpa":        {41},
		"desinfetante": {41},
		"veja":         {41},
		// Pet Shop
		"racao":    {44},
		"ração":    {44},
		"pet":      {44},
		"cachorro": {44},
		"gato":     {44},
		// Outros
		"filtro": {46},
		"pilha":  {46},
		"vela":   {46},
	}

	// Busca por palavras-chave
	for keyword, categoryIDs := range keywords {
		if strings.Contains(desc, keyword) {
			categoryID := categoryIDs[0]
			return &categoryID
		}
	}

	// Se não encontrou, retorna nil (sem categoria)
	return nil
}
