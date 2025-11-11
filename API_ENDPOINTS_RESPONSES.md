# 📚 Documentação Completa - Respostas dos Endpoints da API

**Última Atualização:** 11/11/2025  
**Base URL:** `/api/v1`

---

## 📑 Índice

1. [Autenticação](#1-autenticação)
2. [Usuário](#2-usuário)
3. [Categorias](#3-categorias)
4. [Recibos](#4-recibos)
5. [Itens](#5-itens)
6. [Produtos](#6-produtos)
7. [Scan QR Code](#7-scan-qr-code)
8. [Uso de IA](#8-uso-de-ia)

---

## 1. Autenticação

### 🔓 POST /register
**Descrição:** Registrar novo usuário

**Request Body:**
```json
{
  "name": "João Silva",
  "email": "joao@example.com",
  "password": "senha123"
}
```

**Response (201 Created):**
```json
{
  "message": "Usuário criado com sucesso! Token de autenticação gerado",
  "data": {
    "id": 1,
    "createdAt": "2025-11-11T10:30:00Z",
    "updatedAt": "2025-11-11T10:30:00Z",
    "name": "João Silva",
    "email": "joao@example.com"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE3MzE0MDc0MDB9.abc123..."
}
```

---

### 🔑 POST /login
**Descrição:** Fazer login

**Request Body:**
```json
{
  "email": "joao@example.com",
  "password": "senha123"
}
```

**Response (200 OK):**
```json
{
  "message": "Login realizado com sucesso!",
  "data": {
    "id": 1,
    "createdAt": "2025-11-11T10:30:00Z",
    "updatedAt": "2025-11-11T10:30:00Z",
    "name": "João Silva",
    "email": "joao@example.com"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE3MzE0MDc0MDB9.abc123..."
}
```

---

### 🚪 POST /logout
**Descrição:** Fazer logout (invalida token)

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "✅ Logout realizado com sucesso! Token foi invalidado"
}
```

---

## 2. Usuário

### 👤 GET /me
**Descrição:** Buscar dados do usuário logado

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "Usuário autenticado encontrado",
  "data": {
    "id": 1,
    "createdAt": "2025-11-11T10:30:00Z",
    "updatedAt": "2025-11-11T10:30:00Z",
    "name": "João Silva",
    "email": "joao@example.com"
  }
}
```

---

### 🗑️ DELETE /user
**Descrição:** Deletar conta do usuário (soft delete)

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "✅ Conta deletada com sucesso! Todos os seus dados foram removidos (recibos, itens, categorias, listas de compras)"
}
```

---

## 3. Categorias

### 📁 GET /categories
**Descrição:** Listar todas as categorias com itemCount (COMPLETO)

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "Categories retrieved successfully",
  "data": [
    {
      "id": 1,
      "createdAt": "2025-11-10T08:00:00Z",
      "updatedAt": "2025-11-10T08:00:00Z",
      "name": "Grãos e Cereais",
      "description": "Arroz, feijão, lentilha, grão de bico, aveia",
      "icon": "🌾",
      "color": "#8B4513",
      "itemCount": 15
    },
    {
      "id": 2,
      "createdAt": "2025-11-10T08:00:00Z",
      "updatedAt": "2025-11-10T08:00:00Z",
      "name": "Massas",
      "description": "Macarrão, lasanha, nhoque",
      "icon": "🍝",
      "color": "#FFD700",
      "itemCount": 8
    },
    {
      "id": 3,
      "createdAt": "2025-11-10T08:00:00Z",
      "updatedAt": "2025-11-10T08:00:00Z",
      "name": "Padaria",
      "description": "Pão, baguete, pão de forma, brioche",
      "icon": "🍞",
      "color": "#D2691E",
      "itemCount": 12
    },
    {
      "id": 4,
      "createdAt": "2025-11-10T08:00:00Z",
      "updatedAt": "2025-11-10T08:00:00Z",
      "name": "Carnes e Proteínas",
      "description": "Carne bovina, frango, peixe, ovos",
      "icon": "🥩",
      "color": "#8B0000",
      "itemCount": 0
    }
  ],
  "count": 22
}
```

**Características:**
- ✅ Inclui timestamps (createdAt, updatedAt)
- ✅ Inclui itemCount para cada categoria
- ✅ Categorias ordenadas por nome (A-Z)

---

### ⚡ GET /categories/summary
**Descrição:** Listar categorias (ULTRA-LEVE - sem timestamps)

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "Categories summary retrieved successfully",
  "categories": [
    {
      "id": 1,
      "name": "Grãos e Cereais",
      "description": "Arroz, feijão, lentilha, grão de bico, aveia",
      "icon": "🌾",
      "color": "#8B4513",
      "itemCount": 15
    },
    {
      "id": 2,
      "name": "Massas",
      "description": "Macarrão, lasanha, nhoque",
      "icon": "🍝",
      "color": "#FFD700",
      "itemCount": 8
    },
    {
      "id": 3,
      "name": "Padaria",
      "description": "Pão, baguete, pão de forma, brioche",
      "icon": "🍞",
      "color": "#D2691E",
      "itemCount": 12
    }
  ],
  "total": 22
}
```

**Características:**
- ❌ SEM timestamps (mais leve)
- ✅ SEMPRE inclui itemCount
- ✅ Payload 40% menor que /categories
- ⚡ Ideal para listas e dropdowns

**Diferença de /categories:**
- 40% menos dados
- Resposta mais rápida
- Sem createdAt/updatedAt

---

### 🔍 GET /category/:id
**Descrição:** Buscar detalhes de uma categoria específica com todos os itens

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "Category retrieved successfully",
  "data": {
    "id": 1,
    "createdAt": "2025-11-10T08:00:00Z",
    "updatedAt": "2025-11-10T08:00:00Z",
    "name": "Grãos e Cereais",
    "description": "Arroz, feijão, lentilha, grão de bico, aveia",
    "icon": "🌾",
    "color": "#8B4513"
  },
  "items": [
    {
      "id": 101,
      "createdAt": "2025-11-10T14:20:00Z",
      "updatedAt": "2025-11-10T14:20:00Z",
      "receiptId": 5,
      "categoryId": 1,
      "productId": 50,
      "quantity": 5.0,
      "unitPrice": 8.50,
      "total": 42.50,
      "category": {
        "id": 1,
        "name": "Grãos e Cereais"
      },
      "product": {
        "id": 50,
        "name": "Arroz Integral",
        "unity": "kg"
      }
    },
    {
      "id": 102,
      "createdAt": "2025-11-10T14:20:00Z",
      "updatedAt": "2025-11-10T14:20:00Z",
      "receiptId": 5,
      "categoryId": 1,
      "productId": 51,
      "quantity": 2.0,
      "unitPrice": 6.90,
      "total": 13.80,
      "category": {
        "id": 1,
        "name": "Grãos e Cereais"
      },
      "product": {
        "id": 51,
        "name": "Feijão Preto",
        "unity": "kg"
      }
    }
  ],
  "itemCount": 15
}
```

**Características:**
- ✅ Retorna TODOS os itens da categoria
- ✅ Cada item inclui produto e categoria
- ⚠️ Pode ser pesado se categoria tiver muitos itens

---

### ➕ POST /category
**Descrição:** Criar nova categoria

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "name": "Eletrônicos",
  "description": "Celular, computador, fones",
  "icon": "💻",
  "color": "#4A90E2"
}
```

**Response (201 Created):**
```json
{
  "message": "✅ Categoria criada com sucesso!",
  "data": {
    "id": 25,
    "createdAt": "2025-11-11T11:00:00Z",
    "updatedAt": "2025-11-11T11:00:00Z",
    "name": "Eletrônicos",
    "description": "Celular, computador, fones",
    "icon": "💻",
    "color": "#4A90E2"
  }
}
```

---

### ✏️ PATCH /category/:id
**Descrição:** Atualizar categoria existente

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body (todos os campos são opcionais):**
```json
{
  "name": "Eletrônicos e Gadgets",
  "description": "Celular, computador, fones, tablets",
  "icon": "📱",
  "color": "#5A90F2"
}
```

**Response (200 OK):**
```json
{
  "message": "✅ Categoria atualizada com sucesso!",
  "data": {
    "id": 25,
    "createdAt": "2025-11-11T11:00:00Z",
    "updatedAt": "2025-11-11T11:05:00Z",
    "name": "Eletrônicos e Gadgets",
    "description": "Celular, computador, fones, tablets",
    "icon": "📱",
    "color": "#5A90F2"
  }
}
```

---

### 🗑️ DELETE /category/:id
**Descrição:** Deletar categoria (move itens para "Não categorizado")

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "✅ Categoria deletada com sucesso! 8 itens foram movidos para 'Não categorizado'"
}
```

---

### 📊 GET /categories/graph
**Descrição:** Gráfico de gastos por categoria (com filtro de período)

**Headers:**
```
Authorization: Bearer {token}
```

**Query Params:**
```
start_date=2025-11-01  (opcional, formato YYYY-MM-DD)
end_date=2025-11-30    (opcional, formato YYYY-MM-DD)
```

**Response (200 OK):**
```json
{
  "message": "Category graph data retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Grãos e Cereais",
      "description": "Arroz, feijão, lentilha",
      "icon": "🌾",
      "color": "#8B4513",
      "itemCount": 15,
      "totalSpent": 245.80
    },
    {
      "id": 2,
      "name": "Massas",
      "description": "Macarrão, lasanha, nhoque",
      "icon": "🍝",
      "color": "#FFD700",
      "itemCount": 8,
      "totalSpent": 156.40
    },
    {
      "id": 3,
      "name": "Padaria",
      "description": "Pão, baguete, pão de forma",
      "icon": "🍞",
      "color": "#D2691E",
      "itemCount": 12,
      "totalSpent": 89.50
    }
  ],
  "total": 491.70,
  "period": {
    "startDate": "2025-11-01",
    "endDate": "2025-11-30"
  }
}
```

**Características:**
- ✅ Filtra por período (opcional)
- ✅ Retorna totalSpent para cada categoria
- ✅ Útil para gráficos de pizza/barras

---

## 4. Recibos

### 📄 GET /receipts
**Descrição:** Listar todos os recibos do usuário

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "receipts": [
    {
      "id": 1,
      "storeName": "Supermercado Extra",
      "date": "2025-11-10",
      "items": [
        {
          "id": 101,
          "categoryId": 1,
          "category": {
            "id": 1,
            "name": "Grãos e Cereais"
          },
          "productId": 50,
          "product": {
            "id": 50,
            "name": "Arroz Integral",
            "unity": "kg"
          },
          "quantity": 5.0,
          "unitPrice": 8.50,
          "total": 42.50
        },
        {
          "id": 102,
          "categoryId": 1,
          "category": {
            "id": 1,
            "name": "Grãos e Cereais"
          },
          "productId": 51,
          "product": {
            "id": 51,
            "name": "Feijão Preto",
            "unity": "kg"
          },
          "quantity": 2.0,
          "unitPrice": 6.90,
          "total": 13.80
        }
      ],
      "total": 156.30,
      "currency": "BRL"
    },
    {
      "id": 2,
      "storeName": "Carrefour",
      "date": "2025-11-09",
      "items": [
        {
          "id": 103,
          "categoryId": 3,
          "category": {
            "id": 3,
            "name": "Padaria"
          },
          "productId": 52,
          "product": {
            "id": 52,
            "name": "Pão Francês",
            "unity": "kg"
          },
          "quantity": 1.5,
          "unitPrice": 12.90,
          "total": 19.35
        }
      ],
      "total": 89.70,
      "currency": "BRL"
    }
  ]
}
```

---

### 📄 GET /receipt/:id
**Descrição:** Buscar um recibo específico

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "id": 1,
  "createdAt": "2025-11-10T14:30:00Z",
  "updatedAt": "2025-11-10T14:30:00Z",
  "userId": 1,
  "storeName": "Supermercado Extra",
  "date": "2025-11-10",
  "items": [
    {
      "id": 101,
      "createdAt": "2025-11-10T14:30:00Z",
      "updatedAt": "2025-11-10T14:30:00Z",
      "receiptId": 1,
      "categoryId": 1,
      "category": {
        "id": 1,
        "name": "Grãos e Cereais"
      },
      "productId": 50,
      "product": {
        "id": 50,
        "name": "Arroz Integral",
        "unity": "kg"
      },
      "quantity": 5.0,
      "unitPrice": 8.50,
      "total": 42.50
    }
  ],
  "subtotal": 150.00,
  "discount": 10.00,
  "total": 140.00,
  "currency": "BRL",
  "confidence": 0.95,
  "notes": "NFC-e #123456 - Chave: 35201108427063000151550010001234561001234567"
}
```

---

### 📅 GET /receipts/date/:date
**Descrição:** Buscar recibos de uma data específica

**Headers:**
```
Authorization: Bearer {token}
```

**URL:**
```
GET /receipts/date/2025-11-10
```

**Response (200 OK):**
```json
{
  "receipts": [
    {
      "id": 1,
      "storeName": "Supermercado Extra",
      "date": "2025-11-10",
      "total": 140.00,
      "currency": "BRL"
    },
    {
      "id": 2,
      "storeName": "Carrefour",
      "date": "2025-11-10",
      "total": 89.70,
      "currency": "BRL"
    }
  ]
}
```

---

### 📆 GET /receipts/period
**Descrição:** Buscar recibos de um período

**Headers:**
```
Authorization: Bearer {token}
```

**Query Params:**
```
start_date=2025-11-01
end_date=2025-11-30
```

**Response (200 OK):**
```json
{
  "receipts": [
    {
      "id": 1,
      "storeName": "Supermercado Extra",
      "date": "2025-11-10",
      "total": 140.00,
      "currency": "BRL"
    },
    {
      "id": 2,
      "storeName": "Carrefour",
      "date": "2025-11-09",
      "total": 89.70,
      "currency": "BRL"
    }
  ],
  "totalSpent": 229.70,
  "count": 2
}
```

---

### 🧾 GET /receipts-basic
**Descrição:** Listar recibos (versão simplificada)

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "receipts": [
    {
      "id": 1,
      "storeName": "Supermercado Extra",
      "date": "2025-11-10",
      "itemCount": 15,
      "total": 140.00,
      "currency": "BRL"
    },
    {
      "id": 2,
      "storeName": "Carrefour",
      "date": "2025-11-09",
      "itemCount": 8,
      "total": 89.70,
      "currency": "BRL"
    }
  ]
}
```

**Características:**
- ✅ SEM lista de itens (mais leve)
- ✅ Inclui apenas itemCount
- ⚡ Ideal para listagens rápidas

---

### ✏️ PATCH /receipt/:id
**Descrição:** Atualizar recibo

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body (todos os campos são opcionais):**
```json
{
  "storeName": "Supermercado Extra - Unidade Centro",
  "date": "2025-11-10",
  "total": 145.00,
  "notes": "Nota atualizada"
}
```

**Response (200 OK):**
```json
{
  "message": "Receipt updated successfully",
  "data": {
    "id": 1,
    "storeName": "Supermercado Extra - Unidade Centro",
    "date": "2025-11-10",
    "total": 145.00,
    "notes": "Nota atualizada"
  }
}
```

---

### 🗑️ DELETE /receipt/:id
**Descrição:** Deletar recibo e todos os seus itens

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "Receipt deleted successfully"
}
```

---

## 5. Itens

### 📋 GET /items
**Descrição:** Listar todos os itens do usuário

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
[
  {
    "id": 101,
    "createdAt": "2025-11-10T14:30:00Z",
    "updatedAt": "2025-11-10T14:30:00Z",
    "receiptId": 1,
    "categoryId": 1,
    "productId": 50,
    "quantity": 5.0,
    "unitPrice": 8.50,
    "total": 42.50
  },
  {
    "id": 102,
    "createdAt": "2025-11-10T14:30:00Z",
    "updatedAt": "2025-11-10T14:30:00Z",
    "receiptId": 1,
    "categoryId": 1,
    "productId": 51,
    "quantity": 2.0,
    "unitPrice": 6.90,
    "total": 13.80
  }
]
```

---

### 🔍 GET /item/:id
**Descrição:** Buscar um item específico

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "id": 101,
  "createdAt": "2025-11-10T14:30:00Z",
  "updatedAt": "2025-11-10T14:30:00Z",
  "receiptId": 1,
  "categoryId": 1,
  "category": {
    "id": 1,
    "name": "Grãos e Cereais"
  },
  "productId": 50,
  "product": {
    "id": 50,
    "name": "Arroz Integral",
    "unity": "kg"
  },
  "quantity": 5.0,
  "unitPrice": 8.50,
  "total": 42.50
}
```

---

### 📅 GET /items/date/:date
**Descrição:** Buscar itens de uma data específica

**Headers:**
```
Authorization: Bearer {token}
```

**URL:**
```
GET /items/date/2025-11-10
```

**Response (200 OK):**
```json
[
  {
    "id": 101,
    "receiptId": 1,
    "categoryId": 1,
    "productId": 50,
    "quantity": 5.0,
    "unitPrice": 8.50,
    "total": 42.50
  }
]
```

---

### 📆 GET /items/period
**Descrição:** Buscar itens de um período

**Headers:**
```
Authorization: Bearer {token}
```

**Query Params:**
```
start=2025-11-01
end=2025-11-30
```

**Response (200 OK):**
```json
[
  {
    "id": 101,
    "receiptId": 1,
    "categoryId": 1,
    "productId": 50,
    "quantity": 5.0,
    "unitPrice": 8.50,
    "total": 42.50
  },
  {
    "id": 102,
    "receiptId": 1,
    "categoryId": 1,
    "productId": 51,
    "quantity": 2.0,
    "unitPrice": 6.90,
    "total": 13.80
  }
]
```

---

### ✏️ PATCH /item/:id
**Descrição:** Atualizar item

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body (todos os campos são opcionais):**
```json
{
  "categoryId": 2,
  "quantity": 6.0,
  "unitPrice": 8.00,
  "total": 48.00
}
```

**Response (200 OK):**
```json
{
  "id": 101,
  "createdAt": "2025-11-10T14:30:00Z",
  "updatedAt": "2025-11-11T10:00:00Z",
  "receiptId": 1,
  "categoryId": 2,
  "productId": 50,
  "quantity": 6.0,
  "unitPrice": 8.00,
  "total": 48.00
}
```

---

### 🗑️ DELETE /item/:id
**Descrição:** Deletar item

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "Item deleted successfully"
}
```

---

### 🤖 POST /items/recategorize
**Descrição:** Recategorizar itens usando IA

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "itemIds": [101, 102, 103]
}
```

**Response (200 OK):**
```json
{
  "message": "Items recategorized successfully",
  "itemsRecategorized": 3,
  "results": [
    {
      "itemId": 101,
      "productName": "Arroz Integral",
      "oldCategoryId": 1,
      "oldCategoryName": "Não categorizado",
      "newCategoryId": 1,
      "newCategoryName": "Grãos e Cereais",
      "changed": true
    },
    {
      "itemId": 102,
      "productName": "Feijão Preto",
      "oldCategoryId": 1,
      "oldCategoryName": "Não categorizado",
      "newCategoryId": 1,
      "newCategoryName": "Grãos e Cereais",
      "changed": true
    },
    {
      "itemId": 103,
      "productName": "Macarrão",
      "oldCategoryId": 1,
      "oldCategoryName": "Não categorizado",
      "newCategoryId": 2,
      "newCategoryName": "Massas",
      "changed": true
    }
  ]
}
```

---

## 6. Produtos

### 🛒 GET /products
**Descrição:** Listar todos os produtos do usuário

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
[
  {
    "id": 50,
    "createdAt": "2025-11-10T14:30:00Z",
    "updatedAt": "2025-11-10T14:30:00Z",
    "name": "Arroz Integral",
    "unity": "kg"
  },
  {
    "id": 51,
    "createdAt": "2025-11-10T14:30:00Z",
    "updatedAt": "2025-11-10T14:30:00Z",
    "name": "Feijão Preto",
    "unity": "kg"
  },
  {
    "id": 52,
    "createdAt": "2025-11-09T10:15:00Z",
    "updatedAt": "2025-11-09T10:15:00Z",
    "name": "Pão Francês",
    "unity": "kg"
  }
]
```

---

### 🔍 GET /products/:id
**Descrição:** Buscar um produto específico

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "id": 50,
  "createdAt": "2025-11-10T14:30:00Z",
  "updatedAt": "2025-11-10T14:30:00Z",
  "name": "Arroz Integral",
  "unity": "kg"
}
```

---

### 📅 GET /products/date/:date
**Descrição:** Buscar produtos de uma data específica

**Headers:**
```
Authorization: Bearer {token}
```

**URL:**
```
GET /products/date/2025-11-10
```

**Response (200 OK):**
```json
[
  {
    "id": 50,
    "name": "Arroz Integral",
    "unity": "kg"
  },
  {
    "id": 51,
    "name": "Feijão Preto",
    "unity": "kg"
  }
]
```

---

### 📆 GET /products/period
**Descrição:** Buscar produtos de um período

**Headers:**
```
Authorization: Bearer {token}
```

**Query Params:**
```
start=2025-11-01
end=2025-11-30
```

**Response (200 OK):**
```json
[
  {
    "id": 50,
    "name": "Arroz Integral",
    "unity": "kg"
  },
  {
    "id": 51,
    "name": "Feijão Preto",
    "unity": "kg"
  }
]
```

---

### ✏️ PATCH /products/:id
**Descrição:** Atualizar produto

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body (todos os campos são opcionais):**
```json
{
  "name": "Arroz Integral Orgânico",
  "unity": "kg"
}
```

**Response (200 OK):**
```json
{
  "id": 50,
  "createdAt": "2025-11-10T14:30:00Z",
  "updatedAt": "2025-11-11T10:30:00Z",
  "name": "Arroz Integral Orgânico",
  "unity": "kg"
}
```

---

### 🗑️ DELETE /products/:id
**Descrição:** Deletar produto e todos os seus itens

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "Product deleted successfully"
}
```

---

## 7. Scan QR Code

### 📸 POST /scan-qrcode/preview
**Descrição:** Preview da NFC-e (Etapa 1/2 - NÃO salva no banco)

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "qrCodeUrl": "http://www.fazenda.sp.gov.br/nfce/qrcode?p=35201108427063000151550010001234561001234567|2|1|1|ABC123"
}
```

**Response (200 OK):**
```json
{
  "message": "✅ Preview ready! 15 items extracted. You can now edit, remove items, or confirm to save.",
  "data": {
    "storeName": "SUPERMERCADO EXTRA LTDA",
    "date": "2025-11-10",
    "items": [
      {
        "tempId": 1,
        "description": "ARROZ INTEGRAL 5KG",
        "quantity": 1.0,
        "unit": "un",
        "unitPrice": 42.50,
        "total": 42.50,
        "deleted": false
      },
      {
        "tempId": 2,
        "description": "FEIJAO PRETO 1KG",
        "quantity": 2.0,
        "unit": "kg",
        "unitPrice": 6.90,
        "total": 13.80,
        "deleted": false
      },
      {
        "tempId": 3,
        "description": "MACARRAO ESPAGUETE 500G",
        "quantity": 3.0,
        "unit": "un",
        "unitPrice": 4.50,
        "total": 13.50,
        "deleted": false
      }
    ],
    "itemsCount": 15,
    "subtotal": 150.00,
    "discount": 10.00,
    "total": 140.00,
    "accessKey": "35201108427063000151550010001234561001234567",
    "number": "123456",
    "qrCodeUrl": "http://www.fazenda.sp.gov.br/nfce/qrcode?p=35201108427063000151550010001234561001234567|2|1|1|ABC123"
  }
}
```

**Características:**
- ⚡ Rápido (2-5 segundos)
- ✅ Extrai dados da NFC-e
- ✅ Retorna items com tempId para edição
- ❌ NÃO salva no banco
- ✅ Frontend pode editar/remover items

---

### ✅ POST /scan-qrcode/confirm
**Descrição:** Confirmar e salvar NFC-e (Etapa 2/2 - SALVA no banco)

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "qrCodeUrl": "http://www.fazenda.sp.gov.br/nfce/qrcode?p=35201108427063000151550010001234561001234567|2|1|1|ABC123",
  "storeName": "SUPERMERCADO EXTRA LTDA",
  "date": "2025-11-10",
  "items": [
    {
      "tempId": 1,
      "description": "ARROZ INTEGRAL 5KG",
      "quantity": 1.0,
      "unit": "un",
      "unitPrice": 42.50,
      "total": 42.50,
      "deleted": false
    },
    {
      "tempId": 2,
      "description": "FEIJAO PRETO 1KG",
      "quantity": 2.0,
      "unit": "kg",
      "unitPrice": 6.90,
      "total": 13.80,
      "deleted": false
    },
    {
      "tempId": 3,
      "description": "Item removido pelo usuário",
      "quantity": 1.0,
      "unit": "un",
      "unitPrice": 10.00,
      "total": 10.00,
      "deleted": true
    }
  ],
  "subtotal": 150.00,
  "discount": 10.00,
  "total": 140.00,
  "accessKey": "35201108427063000151550010001234561001234567",
  "number": "123456"
}
```

**Response (200 OK):**
```json
{
  "message": "✅ NFC-e processada com sucesso! Items estão sendo categorizados pela IA em segundo plano",
  "preview": {
    "storeName": "SUPERMERCADO EXTRA LTDA",
    "date": "2025-11-10",
    "itemsCount": 2,
    "total": 56.30,
    "accessKey": "35201108427063000151550010001234561001234567"
  },
  "aiProcessing": {
    "status": "processing",
    "message": "A IA está categorizando os items automaticamente. Isso pode levar alguns segundos.",
    "estimatedTime": "5-10 segundos"
  }
}
```

**Características:**
- ⚡ Resposta instantânea ao cliente
- 🤖 Categorização com IA em background
- ✅ Items com `deleted: true` são ignorados
- 💾 Salva receipt e items no banco
- 📊 Registra uso de tokens da IA

---

## 8. Uso de IA

### 📊 GET /ai-usage
**Descrição:** Histórico de uso de tokens da IA

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "createdAt": "2025-11-10T14:35:00Z",
    "userId": 1,
    "promptTokens": 2500,
    "responseTokens": 800,
    "totalTokens": 3300,
    "model": "gemini-1.5-flash",
    "endpoint": "/scan-qrcode/confirm"
  },
  {
    "id": 2,
    "createdAt": "2025-11-09T10:20:00Z",
    "userId": 1,
    "promptTokens": 1800,
    "responseTokens": 600,
    "totalTokens": 2400,
    "model": "gemini-1.5-flash",
    "endpoint": "/items/recategorize"
  }
]
```

---

### 📈 GET /ai-usage/summary
**Descrição:** Resumo de uso de tokens da IA

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "totalPromptTokens": 15000,
  "totalResponseTokens": 5000,
  "totalTokens": 20000,
  "totalRequests": 8,
  "averageTokensPerRequest": 2500,
  "mostUsedModel": "gemini-1.5-flash",
  "period": {
    "firstUsage": "2025-11-01T08:00:00Z",
    "lastUsage": "2025-11-10T14:35:00Z"
  }
}
```

---

## 📝 Notas Importantes

### 🔐 Autenticação
- Todos os endpoints protegidos requerem header `Authorization: Bearer {token}`
- Token expira após 7 dias
- Após logout, o token é invalidado (blacklist)

### 📊 Paginação
- Atualmente não há paginação
- Se houver muitos registros, considere usar filtros de data

### 🎨 Formatação de Datas
- **Input:** YYYY-MM-DD (ex: 2025-11-10)
- **Output:** ISO 8601 (ex: 2025-11-10T14:30:00Z)

### 💰 Valores Monetários
- Sempre em formato decimal: 42.50 (não "42,50")
- Currency padrão: "BRL"

### 🗑️ Soft Delete
- Deletar categoria: Move items para "Não categorizado"
- Deletar recibo: Deleta todos os items
- Deletar produto: Deleta todos os items associados
- Items deletados não aparecem em contagens

### ⚡ Performance
- Use `/categories/summary` ao invés de `/categories` para listagens
- Use `/receipts-basic` ao invés de `/receipts` para listagens rápidas
- Filtros de data ajudam a reduzir payload

---

**Última Atualização:** 11/11/2025  
**Versão da API:** 1.0  
**Documentação Swagger:** http://localhost:8080/swagger/index.html
