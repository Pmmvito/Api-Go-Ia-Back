# 📝 Guia Completo: Como Atualizar um Item de Nota Fiscal

**Endpoint:** `PATCH /api/v1/item/:id`  
**Autenticação:** ✅ Requer JWT Token (Bearer)  
**Descrição:** Atualiza campos específicos de um item de nota fiscal

---

## 🎯 **Como Usar**

### **1. Requisição HTTP**

```http
PATCH /api/v1/item/123 HTTP/1.1
Host: localhost:8080
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "categoryId": 5,
  "quantity": 3.0,
  "unitPrice": 12.50
}
```

### **2. Usando cURL**

```bash
curl -X PATCH http://localhost:8080/api/v1/item/123 \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "categoryId": 5,
    "quantity": 3.0,
    "unitPrice": 12.50
  }'
```

### **3. Usando JavaScript/Axios**

```javascript
const axios = require('axios');

// Atualizar item
async function atualizarItem(itemId, dados) {
  const token = localStorage.getItem('accessToken');
  
  try {
    const response = await axios.patch(
      `http://localhost:8080/api/v1/item/${itemId}`,
      dados,
      {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      }
    );
    
    console.log('✅ Item atualizado:', response.data);
    return response.data;
  } catch (error) {
    console.error('❌ Erro ao atualizar:', error.response?.data);
    throw error;
  }
}

// Exemplo de uso:
atualizarItem(123, {
  categoryId: 5,      // Muda categoria
  quantity: 3.0,      // Altera quantidade
  unitPrice: 12.50    // Novo preço unitário
});
```

### **4. Usando Python/Requests**

```python
import requests

def atualizar_item(item_id, dados, token):
    url = f"http://localhost:8080/api/v1/item/{item_id}"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    response = requests.patch(url, json=dados, headers=headers)
    
    if response.status_code == 200:
        print("✅ Item atualizado com sucesso!")
        return response.json()
    else:
        print(f"❌ Erro: {response.status_code}")
        print(response.json())
        return None

# Exemplo de uso:
token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
atualizar_item(123, {
    "categoryId": 5,
    "quantity": 3.0,
    "unitPrice": 12.50
}, token)
```

---

## 📋 **Campos Disponíveis para Atualização**

**Todos os campos são OPCIONAIS** - você só envia o que quer alterar:

| Campo | Tipo | Descrição | Exemplo |
|-------|------|-----------|---------|
| `categoryId` | number | ID da categoria do item | `5` |
| `productId` | number | ID do produto | `42` |
| `quantity` | number | Quantidade (pode ter decimais) | `3.5` |
| `unitPrice` | number | Preço unitário (R$) | `12.50` |
| `total` | number | Total do item (R$) | `43.75` |

### ⚠️ **Importante:**
- Você **NÃO precisa** enviar todos os campos
- Envie **apenas** os campos que quer alterar
- Os campos não enviados **permanecem inalterados**

---

## 📤 **Exemplos de Requisição**

### **Exemplo 1: Alterar apenas a categoria**

```json
{
  "categoryId": 8
}
```

**Resultado:** Apenas a categoria muda. Quantidade, preço, etc. ficam iguais.

---

### **Exemplo 2: Alterar quantidade e preço**

```json
{
  "quantity": 5.0,
  "unitPrice": 8.99
}
```

**Resultado:** Quantidade vira 5.0 e preço 8.99. Categoria e produto não mudam.

---

### **Exemplo 3: Alterar categoria e produto**

```json
{
  "categoryId": 3,
  "productId": 150
}
```

**Resultado:** Item muda de categoria e produto. Preços/quantidades inalterados.

---

### **Exemplo 4: Recalcular total manualmente**

```json
{
  "quantity": 4.0,
  "unitPrice": 10.50,
  "total": 42.00
}
```

**Resultado:** Todos os valores numéricos atualizados.

---

## 📥 **Resposta do Servidor**

### **✅ Sucesso (200 OK)**

```json
{
  "id": 123,
  "createdAt": "2025-11-10T14:30:00Z",
  "updatedAt": "2025-11-13T10:15:00Z",
  "receiptId": 45,
  "categoryId": 5,
  "productId": 42,
  "quantity": 3.0,
  "unitPrice": 12.50,
  "total": 37.50
}
```

**Campos retornados:**
- `id`: ID do item
- `createdAt`: Data de criação (não muda)
- `updatedAt`: Data da última atualização (atualizada agora!)
- `receiptId`: ID da nota fiscal a que pertence
- `categoryId`: Categoria atual
- `productId`: Produto atual
- `quantity`: Quantidade atual
- `unitPrice`: Preço unitário atual
- `total`: Total atual

---

### **❌ Erros Possíveis**

#### **400 Bad Request - Campos inválidos**

```json
{
  "status": 400,
  "message": "Key: 'UpdateItemRequest.Quantity' Error:Field validation for 'Quantity' failed on the 'min' tag"
}
```

**Causa:** Valor negativo ou formato inválido  
**Solução:** Verificar tipos dos campos (numbers, não strings)

---

#### **401 Unauthorized - Token inválido/expirado**

```json
{
  "message": "Invalid or expired token",
  "errorCode": 401
}
```

**Causa:** Token JWT expirado (15 minutos) ou inválido  
**Solução:** Renovar token usando `/auth/refresh` ou fazer login novamente

---

#### **404 Not Found - Item não existe**

```json
{
  "status": 404,
  "message": "Item não encontrado ou não pertence ao usuário autenticado"
}
```

**Causas possíveis:**
- ID do item não existe
- Item pertence a outro usuário
- Item foi deletado (soft delete)

**Solução:** Verificar ID correto com `GET /items`

---

#### **500 Internal Server Error**

```json
{
  "status": 500,
  "message": "Erro ao atualizar item no banco de dados. Por favor, tente novamente"
}
```

**Causa:** Erro no servidor (banco de dados, etc.)  
**Solução:** Tentar novamente ou contactar suporte

---

## 🔒 **Segurança**

### **✅ O que o endpoint PROTEGE:**

1. **Autenticação Obrigatória:** Sem token = 401 Unauthorized
2. **Isolamento de Usuário:** Você só atualiza **seus** items
3. **Validação de Propriedade:** Item precisa pertencer a uma nota SUA
4. **Rate Limiting:** Máximo 100 requisições/segundo (global)

### **🔍 Como funciona a validação:**

```go
// Backend verifica:
db.Joins("INNER JOIN receipts ON receipts.id = receipt_items.receipt_id").
   Where("receipt_items.id = ? AND receipts.user_id = ?", itemId, userID).
   First(&item)
```

**Tradução:** "Só atualiza o item se ele pertencer a uma nota fiscal do usuário logado"

---

## 🧪 **Testando o Endpoint**

### **Passo 1: Obter Token**

```bash
# Login
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu@email.com",
    "password": "suasenha"
  }'

# Resposta (copie o accessToken):
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "a1b2c3...",
  "expiresIn": 900,
  "user": { ... }
}
```

---

### **Passo 2: Listar seus items**

```bash
curl -X GET http://localhost:8080/api/v1/items \
  -H "Authorization: Bearer SEU_ACCESS_TOKEN"

# Resposta (veja o ID do item que quer atualizar):
[
  {
    "id": 123,
    "receiptId": 45,
    "categoryId": 1,
    "productId": 42,
    "quantity": 2.0,
    "unitPrice": 10.00,
    "total": 20.00
  }
]
```

---

### **Passo 3: Atualizar o item**

```bash
curl -X PATCH http://localhost:8080/api/v1/item/123 \
  -H "Authorization: Bearer SEU_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "categoryId": 5,
    "quantity": 3.0
  }'

# Resposta:
{
  "id": 123,
  "receiptId": 45,
  "categoryId": 5,        // ✅ Mudou de 1 para 5
  "productId": 42,        // ⚪ Não mudou
  "quantity": 3.0,        // ✅ Mudou de 2.0 para 3.0
  "unitPrice": 10.00,     // ⚪ Não mudou
  "total": 20.00,         // ⚪ Não mudou (precisa atualizar manualmente)
  "updatedAt": "2025-11-13T10:15:00Z"
}
```

---

### **Passo 4: Verificar mudança**

```bash
curl -X GET http://localhost:8080/api/v1/item/123 \
  -H "Authorization: Bearer SEU_ACCESS_TOKEN"

# Deve retornar o item com os valores atualizados!
```

---

## 💡 **Casos de Uso Comuns**

### **1. Recategorizar item**

**Cenário:** IA categorizou errado, usuário quer corrigir manualmente

```javascript
// Mudar categoria de "Não categorizado" (1) para "Grãos e Cereais" (2)
atualizarItem(123, {
  categoryId: 2
});
```

---

### **2. Corrigir quantidade**

**Cenário:** Escaneou nota errada, quantidade está incorreta

```javascript
// Tinha 2 unidades, na verdade eram 5
atualizarItem(123, {
  quantity: 5.0
});
```

---

### **3. Atualizar preço**

**Cenário:** Preço mudou após promoção/desconto

```javascript
// Preço de R$ 12.50 caiu para R$ 9.90
atualizarItem(123, {
  unitPrice: 9.90,
  total: 29.70  // 3 unidades x R$ 9.90
});
```

---

### **4. Trocar produto**

**Cenário:** Item foi associado ao produto errado

```javascript
// Mudar de produto 42 para produto 88
atualizarItem(123, {
  productId: 88
});
```

---

## 🚫 **Limitações**

### **❌ O que você NÃO pode fazer:**

1. **Atualizar `receiptId`** - Item não pode mudar de nota fiscal
   - Solução: Deletar item e criar novo na outra nota

2. **Atualizar `id`** - ID é imutável
   - ID é gerado automaticamente pelo banco

3. **Atualizar `createdAt`** - Data de criação não muda
   - Apenas `updatedAt` é atualizado automaticamente

4. **Atualizar items de OUTRAS pessoas**
   - Você só pode atualizar seus próprios items
   - Backend valida automaticamente

---

## 🔄 **Fluxo Completo (Frontend)**

```javascript
// 1. Usuário clica em "Editar Item"
function editarItem(itemId) {
  // Buscar dados atuais
  const item = await axios.get(`/api/v1/item/${itemId}`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  
  // 2. Mostrar modal com valores atuais
  mostrarModalEdicao(item.data);
}

// 3. Usuário altera campos no modal
function salvarAlteracoes(itemId, novosValores) {
  // Enviar apenas campos alterados
  const response = await axios.patch(
    `/api/v1/item/${itemId}`,
    novosValores,
    { headers: { Authorization: `Bearer ${token}` } }
  );
  
  // 4. Atualizar UI com valores atualizados
  atualizarLista(response.data);
  
  // 5. Mostrar mensagem de sucesso
  mostrarNotificacao('✅ Item atualizado com sucesso!');
}
```

---

## 📊 **Comparação: Antes vs Depois**

### **ANTES da atualização:**

```json
{
  "id": 123,
  "categoryId": 1,
  "quantity": 2.0,
  "unitPrice": 10.00,
  "total": 20.00
}
```

### **REQUISIÇÃO:**

```json
{
  "categoryId": 5,
  "quantity": 3.5
}
```

### **DEPOIS da atualização:**

```json
{
  "id": 123,
  "categoryId": 5,      // ✅ Mudou
  "quantity": 3.5,      // ✅ Mudou
  "unitPrice": 10.00,   // ⚪ Inalterado
  "total": 20.00        // ⚪ Inalterado
}
```

**Nota:** O campo `total` não é recalculado automaticamente! Se você alterar `quantity` ou `unitPrice`, deve calcular e enviar o novo `total`.

---

## 🎯 **Resumo Rápido**

| Item | Informação |
|------|------------|
| **Endpoint** | `PATCH /api/v1/item/:id` |
| **Método** | PATCH |
| **Auth** | ✅ Bearer Token (obrigatório) |
| **Campos** | Todos opcionais - envie só o que quer mudar |
| **Resposta** | Objeto com item completo atualizado |
| **Código Sucesso** | 200 OK |
| **Segurança** | Só atualiza items do usuário logado |
| **Rate Limit** | 100 req/s (global) |

---

## 📚 **Links Relacionados**

- 📖 [Documentação Completa da API](../API_ENDPOINTS_RESPONSES.md)
- 🔐 [Guia de Autenticação JWT](../docs/JWT_EXPLICACAO.md)
- 🛡️ [Correções de Segurança](../docs/SECURITY_FIXES.md)
- 🔄 [Swagger UI](http://localhost:8080/swagger/index.html) - Teste interativo

---

**Dúvidas?** Consulte a documentação Swagger ou entre em contato com o desenvolvedor.
