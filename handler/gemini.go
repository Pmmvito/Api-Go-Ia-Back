package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
)

// GeminiRequest define a estrutura do corpo da requisição para a API do Gemini.
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent representa o conteúdo da requisição, contendo múltiplas partes.
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart pode conter texto ou dados de imagem.
type GeminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *GeminiInlineData `json:"inlineData,omitempty"`
}

// GeminiInlineData contém os dados da imagem em base64.
type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// GeminiResponse define a estrutura da resposta da API do Gemini,
// focando em extrair o texto dos candidatos.
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata *GeminiUsageMetadata `json:"usageMetadata,omitempty"`
}

// GeminiUsageMetadata contém informações sobre o uso de tokens
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GeminiReceiptData é a estrutura que a IA do Gemini deve retornar após analisar um recibo.
type GeminiReceiptData struct {
	StoreName  string              `json:"storeName"`
	Date       string              `json:"date"`
	Items      []GeminiReceiptItem `json:"items"`
	Subtotal   float64             `json:"subtotal"`
	Discount   float64             `json:"discount"`
	Total      float64             `json:"total"`
	Currency   string              `json:"currency"`
	Confidence float64             `json:"confidence"`
	Notes      string              `json:"notes"`
}

// GeminiReceiptItem representa um item de recibo simplificado, conforme retornado pela IA.
type GeminiReceiptItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unitPrice"`
	Total       float64 `json:"total"`
	CategoryID  uint    `json:"categoryId"` // A IA retorna apenas o ID da categoria.
}

// AnalyzeReceiptWithGemini analisa uma ou múltiplas imagens de nota fiscal usando o Gemini
func AnalyzeReceiptWithGemini(imagesBase64 []string, currency string, locale string, amountHint *float64) (*GeminiReceiptData, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY não configurada")
	}

	// Busca categorias disponíveis
	var categories []schemas.Category
	db.Order("name ASC").Find(&categories)

	// Constrói o prompt com categorias
	prompt := buildReceiptPrompt(currency, locale, amountHint, categories, len(imagesBase64))

	// Prepara as partes da mensagem (prompt + todas as imagens)
	parts := []GeminiPart{
		{
			Text: prompt,
		},
	}

	// Adiciona cada imagem ao request
	for i, imageBase64 := range imagesBase64 {
		// Remove prefixo data:image se existir
		if strings.Contains(imageBase64, ",") {
			imageParts := strings.Split(imageBase64, ",")
			if len(imageParts) > 1 {
				imageBase64 = imageParts[1]
			}
		}

		// Adiciona imagem ao array de parts
		parts = append(parts, GeminiPart{
			InlineData: &GeminiInlineData{
				MimeType: "image/jpeg",
				Data:     imageBase64,
			},
		})

		logger.InfoF("Added image %d/%d to Gemini request", i+1, len(imagesBase64))
	}

	// Monta a requisição para o Gemini
	geminiReq := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: parts,
			},
		},
	}

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %v", err)
	}

	// Pega o modelo do .env ou usa padrão
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-1.5-flash"
	}

	// Modelos preview/experimentais usam v1beta, modelos estáveis usam v1
	apiVersion := "v1"
	if strings.Contains(model, "preview") || strings.Contains(model, "exp-") || strings.Contains(model, "2.5") {
		apiVersion = "v1beta"
	}

	// URL da API do Gemini
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/%s/models/%s:generateContent?key=%s", apiVersion, model, apiKey)

	// Faz a requisição HTTP
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar API do Gemini: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API do Gemini (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse da resposta do Gemini
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse da resposta do Gemini: %v", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("resposta vazia do Gemini")
	}

	// Extrai o texto JSON da resposta
	jsonText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Remove markdown code blocks se existirem
	jsonText = strings.TrimPrefix(jsonText, "```json")
	jsonText = strings.TrimPrefix(jsonText, "```")
	jsonText = strings.TrimSuffix(jsonText, "```")
	jsonText = strings.TrimSpace(jsonText)

	// Tenta corrigir erros comuns no JSON (vírgula antes de fechar objeto/array)
	jsonText = strings.ReplaceAll(jsonText, ",\n    }", "\n    }")
	jsonText = strings.ReplaceAll(jsonText, ", }", " }")
	jsonText = strings.ReplaceAll(jsonText, ",\n  }", "\n  }")
	jsonText = strings.ReplaceAll(jsonText, ", ]", " ]")
	jsonText = strings.ReplaceAll(jsonText, ",\n]", "\n]")

	// Parse do JSON retornado pela IA
	var receiptData GeminiReceiptData
	if err := json.Unmarshal([]byte(jsonText), &receiptData); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do JSON da IA: %v\nJSON recebido: %s", err, jsonText)
	}

	return &receiptData, nil
}

// buildReceiptPrompt constrói o prompt para o Gemini
func buildReceiptPrompt(currency string, locale string, amountHint *float64, categories []schemas.Category, imageCount int) string {
	var builder strings.Builder

	builder.WriteString("Você é um assistente de finanças que extrai dados estruturados de notas fiscais em imagem.\n")

	if imageCount > 1 {
		builder.WriteString(fmt.Sprintf("IMPORTANTE: Você receberá %d imagens da MESMA nota fiscal. Analise TODAS as imagens e combine as informações em UM ÚNICO JSON.\n", imageCount))
		builder.WriteString("As imagens podem conter partes diferentes da nota (topo, meio, rodapé, etc.). Junte todos os itens em uma única lista.\n")
	}

	builder.WriteString("IMPORTANTE: Retorne APENAS um JSON válido e bem formatado, sem comentários, texto adicional ou vírgulas extras.\n")
	builder.WriteString("IDIOMA: Todas as descrições e observações devem estar em PORTUGUÊS (PT-BR). Traduza nomes de produtos se necessário.\n")
	builder.WriteString("Formato esperado:\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"storeName\": \"string - nome do estabelecimento\",\n")
	builder.WriteString("  \"date\": \"YYYY-MM-DD - data da compra\",\n")
	builder.WriteString("  \"items\": [\n")
	builder.WriteString("    {\n")
	builder.WriteString("      \"description\": \"string - nome do produto corrigido e legível\",\n")
	builder.WriteString("      \"quantity\": number - quantidade ou peso,\n")
	builder.WriteString("      \"unit\": \"string - unidade de medida: 'un', 'kg', 'g', 'l', 'ml'\",\n")
	builder.WriteString("      \"unitPrice\": number - preço por unidade ou por kg,\n")
	builder.WriteString("      \"total\": number - total do item,\n")
	builder.WriteString("      \"categoryId\": number - ID da categoria (apenas o número, não o nome)\n")
	builder.WriteString("    }\n")
	builder.WriteString("  ],\n")
	builder.WriteString("  \"subtotal\": number,\n")
	builder.WriteString("  \"discount\": number,\n")
	builder.WriteString("  \"total\": number,\n")
	builder.WriteString(fmt.Sprintf("  \"currency\": \"%s\",\n", strings.ToUpper(currency)))
	builder.WriteString("  \"confidence\": number entre 0 e 1,\n")
	builder.WriteString("  \"notes\": \"string - observações relevantes\"\n")
	builder.WriteString("}\n\n")

	// Adiciona lista de categorias disponíveis COM IDs
	if len(categories) > 0 {
		builder.WriteString("CATEGORIAS DISPONÍVEIS (use o ID para categoryId):\n")
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
	}

	builder.WriteString("Regras importantes:\n")
	builder.WriteString("- NUNCA deixe vírgulas extras antes de fechar objetos } ou arrays ]\n")
	builder.WriteString("- Garanta que o JSON seja válido e possa ser parseado sem erros\n")
	builder.WriteString("- Para cada item, use categoryId com APENAS O NÚMERO do ID da categoria (ex: 1, 2, 3)\n")
	builder.WriteString("- NÃO use o nome da categoria, APENAS o ID numérico\n")
	builder.WriteString("\n")
	builder.WriteString("⚠️ CATEGORIZAÇÃO ÚNICA E PRECISA (REGRA CRÍTICA):\n")
	builder.WriteString("  * CADA item deve estar em APENAS UMA categoria - escolha a MAIS ESPECÍFICA\n")
	builder.WriteString("  * Analise o produto e identifique sua categoria PRINCIPAL e ÚNICA\n")
	builder.WriteString("  * NUNCA coloque o mesmo produto em 2 categorias diferentes\n")
	builder.WriteString("\n")
	builder.WriteString("  📋 GUIA DE CATEGORIZAÇÃO (use para decidir):\n")
	builder.WriteString("  • Cerveja, Vinho, Whisky → 'Bebidas Alcoólicas' (NÃO 'Bebidas')\n")
	builder.WriteString("  • Café, Chá, Mate → 'Café e Chá' (NÃO 'Bebidas')\n")
	builder.WriteString("  • Refrigerante, Suco, Água → 'Bebidas' (NÃO 'Café e Chá')\n")
	builder.WriteString("  • Presunto, Mortadela, Salsicha → 'Frios e Embutidos' (NÃO 'Carnes e Proteínas')\n")
	builder.WriteString("  • Frango, Carne Bovina, Peixe → 'Carnes e Proteínas' (NÃO 'Frios e Embutidos')\n")
	builder.WriteString("  • Macarrão, Lasanha → 'Massas' (NÃO 'Padaria')\n")
	builder.WriteString("  • Pão, Baguete → 'Padaria' (NÃO 'Massas')\n")
	builder.WriteString("  • Chocolate, Bala, Sorvete → 'Doces e Sobremesas' (NÃO 'Salgadinhos e Snacks')\n")
	builder.WriteString("  • Chips, Amendoim, Pipoca → 'Salgadinhos e Snacks' (NÃO 'Doces e Sobremesas')\n")
	builder.WriteString("  • Azeite, Sal, Molho → 'Condimentos e Temperos' (NÃO 'Enlatados')\n")
	builder.WriteString("  • Milho em lata, Atum em lata → 'Enlatados e Conservas' (NÃO 'Condimentos')\n")
	builder.WriteString("  • Shampoo, Sabonete → 'Higiene Pessoal' (NÃO 'Limpeza Doméstica')\n")
	builder.WriteString("  • Detergente, Desinfetante → 'Limpeza Doméstica' (NÃO 'Higiene Pessoal')\n")
	builder.WriteString("  • Papel Higiênico, Guardanapo → 'Papel e Descartáveis' (NÃO 'Limpeza' ou 'Higiene')\n")
	builder.WriteString("  • Pizza congelada, Vegetais congelados → 'Congelados' (NÃO 'Doces' mesmo que seja sorvete)\n")
	builder.WriteString("\n")
	builder.WriteString("  * Se ainda houver dúvida, escolha a categoria que descreve MELHOR o produto principal\n")
	builder.WriteString("  * Use 'Outros' APENAS para produtos verdadeiramente únicos/raros que não se encaixam\n")
	builder.WriteString("  * Seja CONSISTENTE: produtos iguais devem SEMPRE estar na mesma categoria\n")
	builder.WriteString("\n")
	builder.WriteString("- Identifique corretamente a unidade de medida:\n")
	builder.WriteString("  * Use 'un' para itens vendidos por unidade (ex: refrigerante, sorvete)\n")
	builder.WriteString("  * Use 'kg' para itens vendidos por peso em quilogramas (ex: frutas, carnes, queijos)\n")
	builder.WriteString("  * Use 'g' para itens vendidos em gramas\n")
	builder.WriteString("  * Use 'l' para líquidos em litros\n")
	builder.WriteString("  * Use 'ml' para líquidos em mililitros\n")
	builder.WriteString("- Quando o item for por peso, a quantity será o peso (ex: 0.350 kg)\n")
	builder.WriteString("- Quando o item for por unidade, a quantity será o número de unidades (ex: 2 un)\n")
	builder.WriteString("- O unitPrice deve ser o preço POR unidade/kg, não o preço total\n")
	builder.WriteString("- Se algum valor não estiver presente, use null para números ou string vazia para textos.\n")
	builder.WriteString("- Use ponto como separador decimal.\n")
	builder.WriteString("- Se discount não for visível, use 0.\n")
	builder.WriteString("- MOEDA: SEMPRE use BRL (Real Brasileiro) no campo currency. Todos os valores estão em Reais (R$).\n")
	builder.WriteString("- Interprete todos os valores monetários em BRL (R$).\n")
	builder.WriteString("- Utilize o formato de data brasileiro (dd/mm/aaaa) e converta para YYYY-MM-DD.\n")
	builder.WriteString("- Corrija e traduza nomes de produtos para português brasileiro (ex: 'Apple' -> 'Maçã').\n")
	builder.WriteString("- Nomes de produtos devem estar abreviados ou com erros corrigidos e em português.\n")
	builder.WriteString("- A soma dos totais dos items deve bater com o subtotal.\n")
	builder.WriteString("- Total = Subtotal - Discount.\n")

	if amountHint != nil && *amountHint > 0 {
		builder.WriteString(fmt.Sprintf("- O total esperado aproximado é %.2f %s. Use isso apenas como referência para validar.\n", *amountHint, currency))
	}

	builder.WriteString("\nAnalise a imagem da nota fiscal e retorne apenas o JSON com todos os textos em português brasileiro.\n")

	return builder.String()
}
