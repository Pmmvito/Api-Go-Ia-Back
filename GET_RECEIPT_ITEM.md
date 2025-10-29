# 🔍 GET Receipt Item - Buscar Item Específico

## 🎯 Novo Endpoint

```
GET /api/v1/receipt/{id}/item/{itemIndex}
```

Busca um item específico de um recibo pelo seu índice.

## 🔐 Autenticação

Requer Bearer Token no header:
```
Authorization: Bearer {seu_token_jwt}
```

## 📋 Parâmetros

| Parâmetro | Tipo | Localização | Descrição |
|-----------|------|-------------|-----------|
| `id` | integer | Path | ID do recibo |
| `itemIndex` | integer | Path | Índice do item (começa em 0) |

## ✅ Resposta de Sucesso (200)

```json
{
  "message": "Item 2 retrieved successfully",
  "receiptId": 15,
  "itemIndex": 2,
  "totalItems": 5,
  "item": {
    "description": "Banana Prata",
    "quantity": 1.450,
    "unit": "kg",
    "unitPrice": 4.99,
    "total": 7.24
  }
}
```

## ❌ Respostas de Erro

### 400 - Bad Request
```json
{
  "errorCode": 400,
  "message": "Item index out of range (0-4)"
}
```

### 401 - Unauthorized
```json
{
  "errorCode": 401,
  "message": "User not authenticated"
}
```

### 404 - Not Found
```json
{
  "errorCode": 404,
  "message": "Receipt not found"
}
```

## 📊 Exemplos de Uso

### Exemplo 1: Buscar primeiro item
```bash
GET /api/v1/receipt/15/item/0
Authorization: Bearer eyJhbGc...
```

**Resposta:**
```json
{
  "message": "Item 0 retrieved successfully",
  "receiptId": 15,
  "itemIndex": 0,
  "totalItems": 4,
  "item": {
    "description": "Coca-Cola PET 2L",
    "quantity": 2,
    "unit": "un",
    "unitPrice": 6.99,
    "total": 13.98
  }
}
```

### Exemplo 2: Buscar item por peso
```bash
GET /api/v1/receipt/15/item/2
Authorization: Bearer eyJhbGc...
```

**Resposta:**
```json
{
  "message": "Item 2 retrieved successfully",
  "receiptId": 15,
  "itemIndex": 2,
  "totalItems": 4,
  "item": {
    "description": "Tomate",
    "quantity": 0.850,
    "unit": "kg",
    "unitPrice": 7.99,
    "total": 6.79
  }
}
```

### Exemplo 3: Índice inválido
```bash
GET /api/v1/receipt/15/item/10
Authorization: Bearer eyJhbGc...
```

**Resposta:**
```json
{
  "errorCode": 400,
  "message": "Item index out of range (0-3)"
}
```

## 🔄 Fluxo Típico

1. **Listar todos os recibos**: `GET /receipts`
2. **Obter recibo completo**: `GET /receipt/{id}`
3. **Ver detalhes de item específico**: `GET /receipt/{id}/item/{itemIndex}` ⭐ NOVO
4. **Editar item**: `PATCH /receipt/{id}/item/{itemIndex}`

## 💡 Casos de Uso

### 1. Verificar detalhes antes de editar
```javascript
// 1. Buscar item atual
const response = await fetch('/api/v1/receipt/15/item/2', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { item } = await response.json();

// 2. Modificar e atualizar
const updated = { ...item, quantity: 1.0 };
await fetch('/api/v1/receipt/15/item/2', {
  method: 'PATCH',
  headers: { 
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(updated)
});
```

### 2. Validar dados de um item
```bash
# Buscar item para validar
curl -X GET "http://localhost:8080/api/v1/receipt/15/item/0" \
  -H "Authorization: Bearer eyJhbGc..."

# Resposta mostra totalItems para saber quantos existem
{
  "totalItems": 4,  // ← Saber o range válido
  "itemIndex": 0,
  "item": { ... }
}
```

### 3. Interface de busca
```typescript
// Função helper para buscar item
async function getReceiptItem(receiptId: number, itemIndex: number) {
  const url = `/api/v1/receipt/${receiptId}/item/${itemIndex}`;
  const response = await fetch(url, {
    headers: { 'Authorization': `Bearer ${getToken()}` }
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }
  
  return response.json();
}

// Uso
try {
  const data = await getReceiptItem(15, 2);
  console.log(`Item: ${data.item.description}`);
  console.log(`Total de items: ${data.totalItems}`);
} catch (error) {
  console.error('Erro ao buscar item:', error.message);
}
```

## 📈 Informações Retornadas

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `message` | string | Mensagem de sucesso |
| `receiptId` | integer | ID do recibo |
| `itemIndex` | integer | Índice do item buscado |
| `totalItems` | integer | Total de items no recibo |
| `item` | object | Dados completos do item |
| `item.description` | string | Nome do produto |
| `item.quantity` | number | Quantidade ou peso |
| `item.unit` | string | Unidade (un, kg, g, l, ml) |
| `item.unitPrice` | number | Preço unitário |
| `item.total` | number | Total do item |

## 🎯 Vantagens

✅ **Precisão**: Busca exata de um item sem trazer todo o recibo  
✅ **Performance**: Menos dados trafegados  
✅ **Validação**: Retorna `totalItems` para validar índice  
✅ **Contexto**: Mostra receiptId e itemIndex na resposta  
✅ **Segurança**: Valida ownership (apenas recibos do usuário)  

## 🔗 Endpoints Relacionados

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/scan-receipt` | Escanear nova nota fiscal |
| GET | `/receipts` | Listar todos os recibos |
| GET | `/receipt/{id}` | Buscar recibo completo |
| PATCH | `/receipt/{id}` | Editar dados gerais do recibo |
| **GET** | `/receipt/{id}/item/{itemIndex}` | **Buscar item específico** ⭐ |
| PATCH | `/receipt/{id}/item/{itemIndex}` | Editar item específico |

## 📝 Notas

- Índice começa em **0** (zero-based)
- Apenas o dono do recibo pode acessar seus items
- Se o índice for inválido, retorna erro com o range válido
- Não modifica dados, apenas consulta

---

**Swagger**: Documentação disponível em `/swagger/index.html` após executar `swag init`
