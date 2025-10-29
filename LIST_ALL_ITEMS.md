# 📦 GET All Items - Listar Todos os Itens

## 🎯 Novo Endpoint

```
GET /api/v1/items
```

Lista **todos os itens** de **todos os recibos** do usuário autenticado em uma única chamada.

## 🔐 Autenticação

Requer Bearer Token no header:
```
Authorization: Bearer {seu_token_jwt}
```

## ✅ Resposta de Sucesso (200)

```json
{
  "message": "All items retrieved successfully",
  "totalItems": 15,
  "totalReceipts": 3,
  "data": [
    {
      "receiptId": 10,
      "storeName": "Supermercado ABC",
      "date": "2025-10-24",
      "itemIndex": 0,
      "currency": "BRL",
      "item": {
        "description": "Coca-Cola PET 2L",
        "quantity": 2,
        "unit": "un",
        "unitPrice": 6.99,
        "total": 13.98
      }
    },
    {
      "receiptId": 10,
      "storeName": "Supermercado ABC",
      "date": "2025-10-24",
      "itemIndex": 1,
      "currency": "BRL",
      "item": {
        "description": "Banana Prata",
        "quantity": 1.450,
        "unit": "kg",
        "unitPrice": 4.99,
        "total": 7.24
      }
    },
    {
      "receiptId": 11,
      "storeName": "Feira do Bairro",
      "date": "2025-10-23",
      "itemIndex": 0,
      "currency": "BRL",
      "item": {
        "description": "Tomate",
        "quantity": 0.850,
        "unit": "kg",
        "unitPrice": 7.99,
        "total": 6.79
      }
    }
  ]
}
```

## 📊 Estrutura da Resposta

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `message` | string | Mensagem de sucesso |
| `totalItems` | integer | Total de itens em todos os recibos |
| `totalReceipts` | integer | Total de recibos do usuário |
| `data` | array | Lista de todos os itens |

### Estrutura de Cada Item:

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `receiptId` | integer | ID do recibo de origem |
| `storeName` | string | Nome do estabelecimento |
| `date` | string | Data da compra (YYYY-MM-DD) |
| `itemIndex` | integer | Índice do item dentro do recibo |
| `currency` | string | Moeda (BRL, USD, etc) |
| `item` | object | Dados do item |
| `item.description` | string | Nome do produto |
| `item.quantity` | number | Quantidade ou peso |
| `item.unit` | string | Unidade (un, kg, g, l, ml) |
| `item.unitPrice` | number | Preço unitário |
| `item.total` | number | Total do item |

## ❌ Respostas de Erro

### 401 - Unauthorized
```json
{
  "errorCode": 401,
  "message": "User not authenticated"
}
```

### 500 - Internal Server Error
```json
{
  "errorCode": 500,
  "message": "Error listing receipts"
}
```

## 💡 Casos de Uso

### 1. Dashboard de Gastos
```javascript
// Buscar todos os itens
const response = await fetch('/api/v1/items', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { data, totalItems } = await response.json();

// Agrupar por categoria
const categories = {};
data.forEach(({ item }) => {
  const category = categorize(item.description);
  categories[category] = (categories[category] || 0) + item.total;
});

console.log('Gastos por categoria:', categories);
```

### 2. Busca Global de Produtos
```javascript
// Buscar todos os itens e filtrar por produto
const response = await fetch('/api/v1/items', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { data } = await response.json();

// Filtrar por produto específico
const bananas = data.filter(({ item }) => 
  item.description.toLowerCase().includes('banana')
);

console.log(`Encontradas ${bananas.length} compras de banana`);
```

### 3. Análise de Preços
```javascript
// Comparar preços do mesmo produto em diferentes compras
const response = await fetch('/api/v1/items', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { data } = await response.json();

// Agrupar por produto
const products = {};
data.forEach(({ item, storeName, date }) => {
  const key = item.description.toLowerCase();
  if (!products[key]) products[key] = [];
  products[key].push({
    store: storeName,
    date,
    price: item.unitPrice,
    unit: item.unit
  });
});

// Encontrar melhores preços
Object.entries(products).forEach(([product, purchases]) => {
  const cheapest = purchases.sort((a, b) => a.price - b.price)[0];
  console.log(`${product}: melhor preço R$ ${cheapest.price}/${cheapest.unit} em ${cheapest.store}`);
});
```

### 4. Relatório por Período
```javascript
// Filtrar itens por período
const response = await fetch('/api/v1/items', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { data } = await response.json();

const startDate = '2025-10-01';
const endDate = '2025-10-31';

const itemsInPeriod = data.filter(({ date }) => 
  date >= startDate && date <= endDate
);

const total = itemsInPeriod.reduce((sum, { item }) => sum + item.total, 0);
console.log(`Total gasto em outubro: R$ ${total.toFixed(2)}`);
```

## 🔍 Filtros e Análises Possíveis

### Por Estabelecimento
```javascript
const itemsByStore = data.reduce((acc, item) => {
  const store = item.storeName;
  if (!acc[store]) acc[store] = [];
  acc[store].push(item);
  return acc;
}, {});
```

### Por Unidade de Medida
```javascript
const itemsByWeight = data.filter(({ item }) => item.unit === 'kg');
const itemsByUnit = data.filter(({ item }) => item.unit === 'un');

console.log(`${itemsByWeight.length} itens por peso`);
console.log(`${itemsByUnit.length} itens por unidade`);
```

### Produtos Mais Comprados
```javascript
const productCount = {};
data.forEach(({ item }) => {
  const product = item.description;
  productCount[product] = (productCount[product] || 0) + 1;
});

const topProducts = Object.entries(productCount)
  .sort((a, b) => b[1] - a[1])
  .slice(0, 10);

console.log('Top 10 produtos mais comprados:', topProducts);
```

## 📈 Exemplo de Resposta Real

```json
{
  "message": "All items retrieved successfully",
  "totalItems": 8,
  "totalReceipts": 2,
  "data": [
    {
      "receiptId": 15,
      "storeName": "Supermercado Pão de Açúcar",
      "date": "2025-10-24",
      "itemIndex": 0,
      "currency": "BRL",
      "item": {
        "description": "Água de Coco Sococo 200ml",
        "quantity": 1,
        "unit": "un",
        "unitPrice": 1.85,
        "total": 1.85
      }
    },
    {
      "receiptId": 15,
      "storeName": "Supermercado Pão de Açúcar",
      "date": "2025-10-24",
      "itemIndex": 1,
      "currency": "BRL",
      "item": {
        "description": "Coca-Cola PET 2,5L",
        "quantity": 1,
        "unit": "un",
        "unitPrice": 4.65,
        "total": 4.65
      }
    },
    {
      "receiptId": 15,
      "storeName": "Supermercado Pão de Açúcar",
      "date": "2025-10-24",
      "itemIndex": 2,
      "currency": "BRL",
      "item": {
        "description": "Azeite Galo 500ml",
        "quantity": 1,
        "unit": "un",
        "unitPrice": 14.95,
        "total": 14.95
      }
    },
    {
      "receiptId": 14,
      "storeName": "Feira Livre Centro",
      "date": "2025-10-22",
      "itemIndex": 0,
      "currency": "BRL",
      "item": {
        "description": "Banana Prata",
        "quantity": 2.5,
        "unit": "kg",
        "unitPrice": 4.99,
        "total": 12.48
      }
    }
  ]
}
```

## 🎯 Diferenças entre Endpoints

| Endpoint | Retorna | Uso |
|----------|---------|-----|
| `GET /receipts` | Lista de recibos completos | Ver histórico de compras |
| `GET /receipt/:id` | Um recibo completo | Ver detalhes de uma compra |
| **`GET /items`** | **Todos os itens de todos os recibos** | **Análises, relatórios, buscas** |
| `GET /receipt/:id/item/:index` | Um item específico | Ver detalhes de um item |

## ⚡ Performance

- ✅ **Uma única chamada**: Evita múltiplas requisições
- ✅ **Ordenação**: Recibos ordenados por data (mais recente primeiro)
- ✅ **Contexto completo**: Cada item vem com informações do recibo
- ✅ **Filtro no cliente**: Dados vêm completos, você filtra como quiser

## 🔗 Fluxo Típico

```mermaid
GET /items
    ↓
Retorna TODOS os itens
    ↓
Filtra no frontend (por data, loja, produto, etc)
    ↓
Exibe dashboard/relatórios
```

## 📝 Notas Importantes

- ⚠️ Endpoint pode retornar **muitos dados** se houver muitos recibos
- ✅ Considere adicionar **paginação** se necessário no futuro
- ✅ Itens vêm com `receiptId` e `itemIndex` para referência
- ✅ Útil para **análises** e **relatórios** gerais
- ✅ Filtragem deve ser feita no **frontend** ou considere adicionar **query params**

## 🚀 Próximas Melhorias Possíveis

1. **Paginação**: `GET /items?page=1&limit=50`
2. **Filtros**: `GET /items?store=ABC&startDate=2025-10-01`
3. **Ordenação**: `GET /items?sortBy=total&order=desc`
4. **Busca**: `GET /items?search=banana`

---

**Total de Endpoints**: 10 funcionais! 🎉
