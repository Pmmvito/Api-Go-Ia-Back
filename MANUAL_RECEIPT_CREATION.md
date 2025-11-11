# 📝 Criação Manual de Nota Fiscal

## POST /api/v1/receipt

Endpoint para criar notas fiscais manualmente, sem necessidade de escanear QR Code.

---

## 🔑 Autenticação
Requer token JWT no header:
```
Authorization: Bearer <seu-token-jwt>
```

---

## 📥 Request Body

```json
{
  "storeName": "Supermercado Silva",
  "date": "2024-11-11",
  "items": [
    {
      "productName": "Arroz Integral",
      "productUnit": "kg",
      "categoryId": 1,
      "quantity": 2.5,
      "unitPrice": 15.90,
      "total": 39.75
    },
    {
      "productName": "Feijão Preto",
      "productUnit": "kg",
      "categoryId": 1,
      "quantity": 1.0,
      "unitPrice": 8.50,
      "total": 8.50
    },
    {
      "productName": "Detergente",
      "productUnit": "un",
      "categoryId": 2,
      "quantity": 3,
      "unitPrice": 2.99,
      "total": 8.97
    }
  ],
  "subtotal": 60.00,
  "discount": 2.78,
  "total": 57.22,
  "currency": "BRL",
  "notes": "Compra mensal - promoção de arroz"
}
```

---

## 📋 Campos

### **Obrigatórios:**
| Campo | Tipo | Descrição | Exemplo |
|-------|------|-----------|---------|
| `storeName` | string | Nome da loja | "Supermercado Silva" |
| `date` | string | Data da compra (YYYY-MM-DD) | "2024-11-11" |
| `items` | array | Lista de itens (mínimo 1) | Ver estrutura abaixo |
| `total` | number | Valor total da nota | 57.22 |

### **Opcionais:**
| Campo | Tipo | Descrição | Padrão |
|-------|------|-----------|--------|
| `subtotal` | number | Subtotal antes do desconto | 0 |
| `discount` | number | Valor do desconto | 0 |
| `currency` | string | Moeda (ISO 4217) | "BRL" |
| `notes` | string | Observações | "" |

### **Estrutura do Item (obrigatórios):**
| Campo | Tipo | Descrição | Exemplo |
|-------|------|-----------|---------|
| `productName` | string | Nome do produto | "Arroz Integral" |
| `productUnit` | string | Unidade (kg, un, l, g, ml) | "kg" |
| `categoryId` | number | ID da categoria (do usuário) | 1 |
| `quantity` | number | Quantidade | 2.5 |
| `unitPrice` | number | Preço unitário | 15.90 |
| `total` | number | Total do item | 39.75 |

---

## ✅ Resposta de Sucesso (201 Created)

```json
{
  "message": "Nota fiscal criada com sucesso",
  "data": {
    "id": 42,
    "createdAt": "2024-11-11T20:30:00Z",
    "updatedAt": "2024-11-11T20:30:00Z",
    "userId": 5,
    "storeName": "Supermercado Silva",
    "date": "2024-11-11",
    "items": [
      {
        "id": 101,
        "createdAt": "2024-11-11T20:30:00Z",
        "updatedAt": "2024-11-11T20:30:00Z",
        "receiptId": 42,
        "categoryId": 1,
        "category": {
          "id": 1,
          "name": "Alimentação"
        },
        "productId": 234,
        "product": {
          "id": 234,
          "name": "Arroz Integral",
          "unity": "kg"
        },
        "quantity": 2.5,
        "unitPrice": 15.90,
        "total": 39.75
      },
      {
        "id": 102,
        "createdAt": "2024-11-11T20:30:00Z",
        "updatedAt": "2024-11-11T20:30:00Z",
        "receiptId": 42,
        "categoryId": 1,
        "category": {
          "id": 1,
          "name": "Alimentação"
        },
        "productId": 235,
        "product": {
          "id": 235,
          "name": "Feijão Preto",
          "unity": "kg"
        },
        "quantity": 1.0,
        "unitPrice": 8.50,
        "total": 8.50
      },
      {
        "id": 103,
        "createdAt": "2024-11-11T20:30:00Z",
        "updatedAt": "2024-11-11T20:30:00Z",
        "receiptId": 42,
        "categoryId": 2,
        "category": {
          "id": 2,
          "name": "Limpeza"
        },
        "productId": 89,
        "product": {
          "id": 89,
          "name": "Detergente",
          "unity": "un"
        },
        "quantity": 3,
        "unitPrice": 2.99,
        "total": 8.97
      }
    ],
    "subtotal": 60.00,
    "discount": 2.78,
    "total": 57.22,
    "currency": "BRL",
    "confidence": 1.0,
    "notes": "Compra mensal - promoção de arroz"
  }
}
```

---

## ❌ Possíveis Erros

### 400 Bad Request - Dados inválidos
```json
{
  "message": "Dados inválidos. Verifique os campos obrigatórios: storeName, date, items (com productName, productUnit, categoryId, quantity, unitPrice, total) e total",
  "errorCode": "400"
}
```

**Causas:**
- Campos obrigatórios faltando
- Formato de data inválido (use YYYY-MM-DD)
- Array `items` vazio
- Valores numéricos <= 0 (quantity, unitPrice, total)
- Tipos de dados incorretos

---

### 400 Bad Request - Categoria inválida
```json
{
  "message": "Uma ou mais categorias não foram encontradas ou não pertencem ao usuário autenticado",
  "errorCode": "400"
}
```

**Causas:**
- `categoryId` não existe
- `categoryId` pertence a outro usuário

**Solução:** Use `GET /api/v1/categories/summary` para obter suas categorias válidas.

---

### 401 Unauthorized
```json
{
  "message": "Unauthorized - Invalid or missing token",
  "errorCode": "401"
}
```

**Causas:**
- Token JWT ausente no header
- Token expirado
- Token inválido

---

### 500 Internal Server Error
```json
{
  "message": "Erro ao criar nota fiscal. Por favor, tente novamente",
  "errorCode": "500"
}
```

**Causas possíveis:**
- Erro no banco de dados
- Falha na transação
- Erro ao criar produto

---

## 🔍 Comportamento Importante

### **1. Criação Automática de Produtos**
- Se o produto não existir (nome + unidade), ele será criado automaticamente
- Produtos são reutilizados se já existirem com mesmo nome e unidade

**Exemplo:**
```json
{
  "productName": "Leite Integral",
  "productUnit": "l"
}
```
Se "Leite Integral" (em litros) já existir → reutiliza  
Se não existir → cria novo produto

---

### **2. Validação de Categorias**
- Todas as categorias devem pertencer ao usuário autenticado
- Use `GET /api/v1/categories/summary` para listar suas categorias

---

### **3. Transação Atômica**
- Se houver erro em qualquer etapa, NADA é salvo (rollback automático)
- Garante consistência: ou tudo é criado, ou nada

---

### **4. Confidence Score**
- Notas criadas manualmente sempre têm `confidence: 1.0` (100%)
- Notas escaneadas por QR Code têm score variável da IA

---

## 📝 Exemplo de Uso com cURL

```bash
curl -X POST https://api.example.com/api/v1/receipt \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer seu-token-jwt-aqui" \
  -d '{
    "storeName": "Supermercado Silva",
    "date": "2024-11-11",
    "items": [
      {
        "productName": "Arroz Integral",
        "productUnit": "kg",
        "categoryId": 1,
        "quantity": 2.5,
        "unitPrice": 15.90,
        "total": 39.75
      }
    ],
    "total": 39.75,
    "currency": "BRL"
  }'
```

---

## 🎯 Casos de Uso

### **Cenário 1: Nota Fiscal em Papel**
Quando você tem uma nota física mas não tem QR Code para escanear.

### **Cenário 2: Compra Online**
Para registrar compras de e-commerce que não têm nota fiscal com QR Code.

### **Cenário 3: Correção de Dados**
Quando o scanner de QR Code falhou ou trouxe dados incorretos.

### **Cenário 4: Migração de Dados**
Para importar notas fiscais de outros sistemas.

---

## 🔗 Endpoints Relacionados

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/api/v1/categories/summary` | Listar suas categorias |
| GET | `/api/v1/receipts` | Listar todas as notas |
| GET | `/api/v1/receipt/:id` | Ver detalhes de uma nota |
| PATCH | `/api/v1/receipt/:id` | Editar uma nota |
| DELETE | `/api/v1/receipt/:id` | Deletar uma nota |
| POST | `/api/v1/scan-qrcode/preview` | Escanear QR Code (alternativa) |

---

## 💡 Dicas

1. **Sempre valide as categorias antes:** Use `GET /categories/summary` para garantir que os IDs existem.

2. **Formato de data:** Use sempre `YYYY-MM-DD` (ex: 2024-11-11).

3. **Unidades padronizadas:** Use unidades simples:
   - `kg`, `g` (peso)
   - `l`, `ml` (volume)
   - `un` (unidade)

4. **Currency padrão:** Se não informar, será usado `BRL` automaticamente.

5. **Cálculos:** Verifique se `subtotal - discount = total` para evitar inconsistências.

---

## 🚀 Status do Endpoint

✅ **Implementado e testado**  
✅ **Documentação Swagger atualizada**  
✅ **Transação atômica (ACID)**  
✅ **Validação de segurança (user_id)**  
✅ **Criação automática de produtos**
