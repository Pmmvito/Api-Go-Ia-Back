# 🏷️ Sistema de Categorias - Documentação Completa

## 📋 Visão Geral

Sistema completo de categorização automática de itens com IA, incluindo CRUD de categorias e filtros avançados.

## 🎯 Funcionalidades

✅ **CRUD Completo de Categorias**
✅ **Categorização Automática pela IA**
✅ **15 Categorias Padrão Pré-carregadas**
✅ **Filtros por Categoria e Período**
✅ **Agrupamento Inteligente**
✅ **Relatórios por Categoria**

---

## 📦 Novos Endpoints

### **Categorias (CRUD)**

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/category` | Criar nova categoria |
| GET | `/categories` | Listar todas as categorias |
| GET | `/category/:id` | Buscar categoria por ID |
| PATCH | `/category/:id` | Atualizar categoria |
| DELETE | `/category/:id` | Deletar categoria |

### **Itens com Filtros**

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/items/filter` | Listar items com filtros (categoria + período) |

**Total de Endpoints Agora: 16** 🚀

---

## 🏷️ Categorias Padrão

Ao iniciar a aplicação, 15 categorias são criadas automaticamente:

| Categoria | Emoji | Cor | Descrição |
|-----------|-------|-----|-----------|
| Alimentação | 🍽️ | #FF6B6B | Alimentos em geral, comida preparada |
| Bebidas | 🥤 | #4ECDC4 | Refrigerantes, sucos, água |
| Frutas | 🍎 | #95E1D3 | Frutas frescas |
| Verduras e Legumes | 🥬 | #7FCD91 | Vegetais e verduras |
| Carnes e Peixes | 🥩 | #E74C3C | Carnes, frango, peixes |
| Laticínios | 🧀 | #F8E5B9 | Leite, queijo, iogurte |
| Padaria | 🍞 | #D4A574 | Pães, bolos, biscoitos |
| Higiene | 🧼 | #74B9FF | Higiene pessoal |
| Limpeza | 🧹 | #A29BFE | Produtos de limpeza |
| Frios | 🥓 | #FD79A8 | Presunto, mortadela |
| Congelados | 🧊 | #81ECEC | Alimentos congelados |
| Grãos e Cereais | 🌾 | #FDCB6E | Arroz, feijão, massas |
| Temperos | 🧂 | #FAB1A0 | Sal, açúcar, temperos |
| Doces | 🍰 | #FF7675 | Chocolates, sobremesas |
| Outros | 📦 | #B2BEC3 | Itens diversos |

---

## 🔧 Exemplos de Uso

### 1️⃣ Criar Nova Categoria

```bash
POST /api/v1/category
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Eletrônicos",
  "description": "Dispositivos eletrônicos e acessórios",
  "icon": "📱",
  "color": "#3498db"
}
```

**Resposta:**
```json
{
  "message": "Category created successfully",
  "data": {
    "id": 16,
    "createdAt": "2025-10-25T10:30:00Z",
    "updatedAt": "2025-10-25T10:30:00Z",
    "name": "Eletrônicos",
    "description": "Dispositivos eletrônicos e acessórios",
    "icon": "📱",
    "color": "#3498db"
  }
}
```

### 2️⃣ Listar Todas as Categorias

```bash
GET /api/v1/categories
Authorization: Bearer {token}
```

**Resposta:**
```json
{
  "message": "Categories retrieved successfully",
  "count": 15,
  "data": [
    {
      "id": 1,
      "name": "Alimentação",
      "description": "Alimentos em geral, comida preparada",
      "icon": "🍽️",
      "color": "#FF6B6B"
    },
    {
      "id": 2,
      "name": "Bebidas",
      "description": "Refrigerantes, sucos, água, bebidas alcoólicas",
      "icon": "🥤",
      "color": "#4ECDC4"
    }
    // ... outras categorias
  ]
}
```

### 3️⃣ Atualizar Categoria

```bash
PATCH /api/v1/category/5
Authorization: Bearer {token}
Content-Type: application/json

{
  "icon": "🍗",
  "description": "Carnes, frangos, peixes e frutos do mar"
}
```

### 4️⃣ Deletar Categoria

```bash
DELETE /api/v1/category/16
Authorization: Bearer {token}
```

---

## 🤖 Categorização Automática pela IA

Quando você escaneia uma nota fiscal, a IA automaticamente:

1. **Recebe a lista de categorias** disponíveis
2. **Analisa cada item** da nota
3. **Atribui a categoria mais apropriada** 

### Exemplo de Response com Categorias:

```json
{
  "message": "Receipt analyzed successfully",
  "data": {
    "items": [
      {
        "description": "Coca-Cola PET 2L",
        "quantity": 2,
        "unit": "un",
        "unitPrice": 6.99,
        "total": 13.98,
        "categoryName": "Bebidas"  // ⬅️ Categoria atribuída pela IA
      },
      {
        "description": "Banana Prata",
        "quantity": 1.450,
        "unit": "kg",
        "unitPrice": 4.99,
        "total": 7.24,
        "categoryName": "Frutas"  // ⬅️ Categoria atribuída pela IA
      },
      {
        "description": "Sabonete Dove",
        "quantity": 3,
        "unit": "un",
        "unitPrice": 2.50,
        "total": 7.50,
        "categoryName": "Higiene"  // ⬅️ Categoria atribuída pela IA
      }
    ]
  }
}
```

---

## 🔍 Filtrando Items por Categoria e Período

### Endpoint: `GET /items/filter`

**Query Parameters:**
- `category` (opcional): Nome da categoria
- `startDate` (opcional): Data inicial (YYYY-MM-DD)
- `endDate` (opcional): Data final (YYYY-MM-DD)

### Exemplo 1: Todos os Itens Agrupados por Categoria

```bash
GET /api/v1/items/filter
Authorization: Bearer {token}
```

**Resposta:**
```json
{
  "message": "Items retrieved successfully",
  "totalItems": 45,
  "totalReceipts": 8,
  "filters": {
    "category": "",
    "startDate": "",
    "endDate": ""
  },
  "groupedByCategory": [
    {
      "categoryName": "Bebidas",
      "itemCount": 12,
      "totalAmount": 85.50,
      "items": [
        {
          "description": "Coca-Cola PET 2L",
          "quantity": 2,
          "unit": "un",
          "unitPrice": 6.99,
          "total": 13.98,
          "categoryName": "Bebidas"
        }
        // ... outros items de bebidas
      ]
    },
    {
      "categoryName": "Frutas",
      "itemCount": 8,
      "totalAmount": 42.30,
      "items": [...]
    }
    // ... outras categorias
  ],
  "data": [
    // Lista completa de todos os items
  ]
}
```

### Exemplo 2: Filtrar Apenas "Bebidas"

```bash
GET /api/v1/items/filter?category=Bebidas
Authorization: Bearer {token}
```

**Resposta:**
```json
{
  "message": "Items retrieved successfully",
  "totalItems": 12,
  "totalReceipts": 5,
  "filters": {
    "category": "Bebidas",
    "startDate": "",
    "endDate": ""
  },
  "data": [
    {
      "description": "Coca-Cola PET 2L",
      "quantity": 2,
      "unit": "un",
      "unitPrice": 6.99,
      "total": 13.98,
      "categoryName": "Bebidas"
    },
    {
      "description": "Suco de Laranja Natural 1L",
      "quantity": 1,
      "unit": "l",
      "unitPrice": 8.50,
      "total": 8.50,
      "categoryName": "Bebidas"
    }
    // ... outros items de bebidas
  ]
}
```

### Exemplo 3: Filtrar por Período

```bash
GET /api/v1/items/filter?startDate=2025-10-01&endDate=2025-10-31
Authorization: Bearer {token}
```

### Exemplo 4: Combinar Filtros

```bash
GET /api/v1/items/filter?category=Frutas&startDate=2025-10-01&endDate=2025-10-15
Authorization: Bearer {token}
```

**Resultado:** Apenas frutas compradas entre 01 e 15 de outubro

---

## 📊 Casos de Uso Práticos

### 1. Dashboard de Gastos por Categoria

```javascript
// Buscar todos os items agrupados
const response = await fetch('/api/v1/items/filter', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { groupedByCategory } = await response.json();

// Criar gráfico de pizza
groupedByCategory.forEach(category => {
  console.log(`${category.categoryName}: R$ ${category.totalAmount.toFixed(2)}`);
});

// Output:
// Bebidas: R$ 85.50
// Frutas: R$ 42.30
// Carnes e Peixes: R$ 120.00
// ...
```

### 2. Encontrar Categoria com Maior Gasto

```javascript
const maxCategory = groupedByCategory.reduce((max, cat) => 
  cat.totalAmount > max.totalAmount ? cat : max
);
console.log(`Categoria com maior gasto: ${maxCategory.categoryName} - R$ ${maxCategory.totalAmount}`);
```

### 3. Relatório Mensal por Categoria

```javascript
const startDate = '2025-10-01';
const endDate = '2025-10-31';

const response = await fetch(
  `/api/v1/items/filter?startDate=${startDate}&endDate=${endDate}`,
  { headers: { 'Authorization': `Bearer ${token}` } }
);

const { groupedByCategory, totalItems } = await response.json();

console.log(`Relatório de Outubro:`);
console.log(`Total de items: ${totalItems}`);
groupedByCategory.forEach(cat => {
  const percentage = (cat.totalAmount / totalGasto * 100).toFixed(1);
  console.log(`${cat.icon} ${cat.categoryName}: R$ ${cat.totalAmount.toFixed(2)} (${percentage}%)`);
});
```

### 4. Comparar Gastos entre Categorias

```javascript
// Bebidas de outubro
const bebidas = await fetch('/api/v1/items/filter?category=Bebidas&startDate=2025-10-01&endDate=2025-10-31')
  .then(r => r.json());

// Higiene de outubro  
const higiene = await fetch('/api/v1/items/filter?category=Higiene&startDate=2025-10-01&endDate=2025-10-31')
  .then(r => r.json());

const totalBebidas = bebidas.data.reduce((sum, item) => sum + item.total, 0);
const totalHigiene = higiene.data.reduce((sum, item) => sum + item.total, 0);

console.log(`Bebidas: R$ ${totalBebidas.toFixed(2)}`);
console.log(`Higiene: R$ ${totalHigiene.toFixed(2)}`);
```

---

## 🎨 Esquema de Dados

### Category
```typescript
{
  id: number;
  createdAt: Date;
  updatedAt: Date;
  name: string;         // Nome único da categoria
  description: string;  // Descrição detalhada
  icon: string;         // Emoji ou ícone
  color: string;        // Cor em hexadecimal (#FF5733)
}
```

### ReceiptItem (atualizado)
```typescript
{
  description: string;
  quantity: number;
  unit: string;
  unitPrice: number;
  total: number;
  categoryId?: number;     // ID da categoria (opcional)
  categoryName: string;    // Nome da categoria (preenchido pela IA)
}
```

---

## 🔄 Fluxo Completo

```mermaid
1. Criar/Listar Categorias
    ↓
2. Escanear Nota Fiscal (POST /scan-receipt)
    ↓
3. IA Recebe Lista de Categorias
    ↓
4. IA Categoriza Cada Item Automaticamente
    ↓
5. Items Salvos com categoryName
    ↓
6. Filtrar/Agrupar por Categoria (GET /items/filter)
    ↓
7. Gerar Relatórios e Análises
```

---

## 📈 Benefícios do Sistema

✅ **Automático**: IA categoriza sem intervenção manual
✅ **Flexível**: Crie suas próprias categorias
✅ **Visual**: Emojis e cores para melhor UX
✅ **Inteligente**: Agrupa e calcula totais automaticamente
✅ **Filtros Poderosos**: Por categoria, período ou ambos
✅ **Relatórios**: Dados prontos para dashboards

---

## 🚀 Resumo dos Endpoints

| Tipo | Quantidade | Endpoints |
|------|------------|-----------|
| Autenticação | 3 | register, login, me |
| Categorias | 5 | POST, GET, GET/:id, PATCH, DELETE |
| Recibos | 5 | scan, list, get, update, update-item |
| Items | 3 | list-all, get-item, **filter** |

**TOTAL: 16 ENDPOINTS** 🎉

---

## 💡 Próximas Melhorias Sugeridas

1. **Subcategorias**: Ex: Bebidas → Refrigerantes, Sucos, Águas
2. **Categorias Personalizadas por Usuário**: Cada user suas categorias
3. **Sugestões da IA**: IA sugere criar novas categorias baseado em padrões
4. **Orçamento por Categoria**: Definir limite mensal por categoria
5. **Notificações**: Alertar quando gastar muito em uma categoria

---

Teste agora e aproveite o sistema completo de categorização! 🏷️✨
