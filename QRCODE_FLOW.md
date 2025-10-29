# 🆕 Novo Fluxo de QR Code em 2 Etapas

## 📋 Visão Geral

O novo fluxo permite ao usuário **visualizar, editar ou excluir items** antes de salvar no banco de dados.

### Fluxo Anterior (1 etapa - Deprecated)
```
POST /scan-qrcode → Scraping → IA Categoriza → Salva no DB → Retorna
```

### ✨ Novo Fluxo (2 etapas)
```
ETAPA 1: POST /scan-qrcode/preview → Scraping → Retorna JSON (NÃO SALVA)
         ↓
   Usuário visualiza/edita
         ↓
ETAPA 2: POST /scan-qrcode/confirm → IA Categoriza → Salva no DB → Retorna
```

---

## 🔄 Endpoints

### 1️⃣ Etapa 1: Preview (Não salva)

**Endpoint:** `POST /api/v1/scan-qrcode/preview`

**Headers:**
```json
{
  "Authorization": "Bearer YOUR_JWT_TOKEN"
}
```

**Request Body:**
```json
{
  "qrCodeUrl": "https://www.fazenda.pr.gov.br/nfce/qrcode?p=41251012236122000160650040002968961931695243|2|1|1|..."
}
```

**Response (200 OK):**
```json
{
  "message": "✅ Preview ready! 15 items extracted. You can now edit, remove items, or confirm to save.",
  "data": {
    "storeName": "Supermercado XYZ",
    "date": "2025-01-15",
    "items": [
      {
        "tempId": 1,
        "description": "ARROZ BRANCO TIPO 1 5KG",
        "quantity": 1.0,
        "unit": "UN",
        "unitPrice": 25.90,
        "total": 25.90
      },
      {
        "tempId": 2,
        "description": "FEIJAO PRETO 1KG",
        "quantity": 2.0,
        "unit": "UN",
        "unitPrice": 8.50,
        "total": 17.00
      }
    ],
    "itemsCount": 15,
    "subtotal": 250.00,
    "discount": 10.00,
    "total": 240.00,
    "accessKey": "41251012236122000160650040002968961931695243",
    "number": "002968",
    "qrCodeUrl": "https://www.fazenda.pr.gov.br/nfce/qrcode?p=..."
  }
}
```

**Características:**
- ⚡ **Rápido**: Apenas scraping (2-5 segundos)
- 💰 **Gratuito**: Não usa IA nesta etapa
- 📋 **Editável**: Retorna dados para edição
- 🚫 **Não salva**: Nenhum dado no banco ainda

---

### 2️⃣ Etapa 2: Confirmar e Salvar

**Endpoint:** `POST /api/v1/scan-qrcode/confirm`

**Headers:**
```json
{
  "Authorization": "Bearer YOUR_JWT_TOKEN"
}
```

**Request Body:**
```json
{
  "storeName": "Supermercado XYZ",
  "date": "2025-01-15",
  "items": [
    {
      "tempId": 1,
      "description": "ARROZ BRANCO TIPO 1 5KG",  // ✏️ Pode editar
      "quantity": 1.0,                            // ✏️ Pode editar
      "unit": "UN",                               // ✏️ Pode editar
      "unitPrice": 25.90,                         // ✏️ Pode editar
      "total": 25.90                              // ✏️ Pode editar
    },
    {
      "tempId": 2,
      "description": "FEIJAO PRETO 1KG",
      "quantity": 2.0,
      "unit": "UN",
      "unitPrice": 8.50,
      "total": 17.00,
      "deleted": true  // ❌ Item será ignorado
    }
  ],
  "subtotal": 250.00,
  "discount": 10.00,
  "total": 240.00,
  "accessKey": "41251012236122000160650040002968961931695243",
  "number": "002968",
  "qrCodeUrl": "https://www.fazenda.pr.gov.br/nfce/qrcode?p=..."
}
```

**Response (201 Created):**
```json
{
  "message": "✅ Receipt saved successfully! 14 items categorized with AI and saved to database.",
  "data": {
    "id": 123,
    "userId": 1,
    "storeName": "Supermercado XYZ",
    "date": "2025-01-15",
    "total": 240.00,
    "items": [
      {
        "id": 456,
        "receiptId": 123,
        "description": "ARROZ BRANCO TIPO 1 5KG",
        "quantity": 1.0,
        "unit": "UN",
        "unitPrice": 25.90,
        "total": 25.90,
        "categoryId": 5,
        "category": {
          "id": 5,
          "name": "Grãos e Cereais"
        }
      }
    ],
    "createdAt": "2025-01-15T10:30:00Z"
  }
}
```

**Características:**
- 🤖 **IA**: Categoriza items automaticamente
- 💾 **Salva**: Persiste no banco de dados
- ✅ **Validações**: Valida items editados
- 🗑️ **Filtra**: Remove items marcados como `deleted: true`

---

## 🔀 Comparação: Antes vs Depois

| Aspecto | Antes (1 etapa) | Depois (2 etapas) |
|---------|----------------|-------------------|
| **Preview** | ❌ Não | ✅ Sim |
| **Edição** | ❌ Não | ✅ Sim (descrição, quantidade, preço) |
| **Exclusão** | ❌ Não | ✅ Sim (`deleted: true`) |
| **Cancelar** | ❌ Salva sempre | ✅ Pode cancelar (não chamar etapa 2) |
| **IA** | ⚡ Imediata | ⏱️ Apenas na confirmação |
| **Velocidade Preview** | 🐢 Lenta (IA+Save) | ⚡ Rápida (só scraping) |
| **Flexibilidade** | 📦 Nenhuma | 🎨 Total |

---

## 🎯 Casos de Uso

### Caso 1: Salvar sem editar
```javascript
// 1. Preview
const preview = await fetch('/scan-qrcode/preview', {
  method: 'POST',
  body: JSON.stringify({ qrCodeUrl: url })
});

// 2. Confirmar direto (sem editar)
const saved = await fetch('/scan-qrcode/confirm', {
  method: 'POST',
  body: JSON.stringify(preview.data)
});
```

### Caso 2: Editar antes de salvar
```javascript
// 1. Preview
const preview = await fetch('/scan-qrcode/preview', { ... });

// 2. Usuário edita
preview.data.items[0].quantity = 2; // Edita quantidade
preview.data.items[1].description = "Arroz Integral"; // Edita nome

// 3. Confirmar com edições
const saved = await fetch('/scan-qrcode/confirm', {
  method: 'POST',
  body: JSON.stringify(preview.data)
});
```

### Caso 3: Remover items antes de salvar
```javascript
// 1. Preview
const preview = await fetch('/scan-qrcode/preview', { ... });

// 2. Marcar items para deletar
preview.data.items[2].deleted = true;
preview.data.items[5].deleted = true;

// 3. Confirmar (items deletados são ignorados)
const saved = await fetch('/scan-qrcode/confirm', {
  method: 'POST',
  body: JSON.stringify(preview.data)
});
```

### Caso 4: Cancelar operação
```javascript
// 1. Preview
const preview = await fetch('/scan-qrcode/preview', { ... });

// 2. Usuário visualiza mas não confirma
// --> Nada é salvo no banco!
```

---

## 📊 Performance

### Etapa 1 (Preview)
- ⚡ **Scraping**: 2-5 segundos
- 💰 **Custo**: $0 (gratuito)
- 📦 **Tamanho**: ~5-10 KB JSON

### Etapa 2 (Confirm)
- 🤖 **IA Gemini**: 1-3 segundos
- 💾 **Save DB**: 0.5-1 segundo
- 💰 **Custo**: ~$0.0001 por nota
- ✅ **Total**: 2-5 segundos

---

## 🛠️ Implementação

### Backend
- ✅ `handler/scanQrCodePreview.go` - Etapa 1 (Preview)
- ✅ `handler/scanQrCodeConfirm.go` - Etapa 2 (Confirm)
- ✅ `router/routes.go` - Rotas configuradas

### Frontend (Exemplo React Native)
```jsx
import { useState } from 'react';

function ScanQRCodeFlow() {
  const [preview, setPreview] = useState(null);
  
  // 1. Escanear e obter preview
  const handleScan = async (qrCodeUrl) => {
    const response = await api.post('/scan-qrcode/preview', { qrCodeUrl });
    setPreview(response.data.data);
  };
  
  // 2. Confirmar e salvar
  const handleConfirm = async () => {
    await api.post('/scan-qrcode/confirm', preview);
    alert('✅ Nota salva com sucesso!');
  };
  
  // 3. Cancelar
  const handleCancel = () => {
    setPreview(null); // Descarta preview
  };
  
  return (
    <>
      {!preview ? (
        <QRScanner onScan={handleScan} />
      ) : (
        <PreviewEditor 
          data={preview}
          onChange={setPreview}
          onConfirm={handleConfirm}
          onCancel={handleCancel}
        />
      )}
    </>
  );
}
```

---

## ⚠️ Endpoint Legacy

O endpoint antigo ainda funciona, mas está **deprecated**:

```
POST /api/v1/scan-qrcode  (⚠️  Deprecated - usar novo fluxo)
```

**Recomendação**: Migrar para o novo fluxo em 2 etapas.

---

## ✅ Vantagens do Novo Fluxo

1. **✏️ Edição**: Usuário pode corrigir dados antes de salvar
2. **🗑️ Exclusão**: Remove items indesejados
3. **🚫 Cancelamento**: Pode cancelar sem poluir banco
4. **⚡ Preview Rápido**: Resposta instantânea (sem IA)
5. **💰 Economia**: IA só é chamada se confirmar
6. **🎨 UX Melhor**: Usuário tem controle total

---

## 🐛 Troubleshooting

### Erro: "All items were deleted"
- **Causa**: Todos os items foram marcados como `deleted: true`
- **Solução**: Manter pelo menos 1 item ativo

### Erro: "At least one item is required"
- **Causa**: Array `items` vazio no request
- **Solução**: Enviar pelo menos 1 item

### Erro: "AI categorization failed"
- **Causa**: GEMINI_API_KEY inválida ou erro na API
- **Solução**: Verificar variável de ambiente

---

## 📝 TODO

- [ ] Implementar busca/criação automática de `Product` (atualmente `productId: 0`)
- [ ] Adicionar cache de preview (opcional)
- [ ] Implementar timeout para preview expirado (segurança)
- [ ] Adicionar validações adicionais (preços negativos, etc)

---

**Desenvolvido com ❤️ usando Go + Gin + GORM + Gemini AI**
