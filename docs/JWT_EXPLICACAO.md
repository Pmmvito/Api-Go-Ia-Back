# 🔐 JWT e Sessões - Guia Completo

## 📊 **Sistema ATUAL (Token JWT de 7 dias)**

### **Como Funciona Agora:**

```
┌─────────────────────────────────────────────────────────────────────┐
│                    FLUXO ATUAL (Token 7 dias)                        │
└─────────────────────────────────────────────────────────────────────┘

1️⃣ USUÁRIO FAZ LOGIN
   ↓
   POST /api/v1/login
   { "email": "joao@example.com", "password": "senha123" }
   
   ↓
   
2️⃣ BACKEND GERA TOKEN JWT (válido por 7 DIAS)
   {
     "user_id": 123,
     "exp": 1731000000,  // Expira em 7 dias
     "iat": 1730395200   // Criado agora
   }
   
   ↓
   
3️⃣ FRONTEND RECEBE E ARMAZENA
   Response: {
     "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
     "user": { "id": 123, "name": "João", "email": "joao@example.com" }
   }
   
   Frontend salva no localStorage/sessionStorage:
   localStorage.setItem('auth_token', token)
   
   ↓
   
4️⃣ TODAS AS REQUISIÇÕES USAM ESSE TOKEN (por 7 dias)
   GET /api/v1/me
   Headers: { "Authorization": "Bearer eyJhbGciOiJIUzI1..." }
   
   ↓
   
5️⃣ BACKEND VALIDA O TOKEN
   - Verifica assinatura (secret key)
   - Verifica se não expirou (exp)
   - Verifica se não está na blacklist (logout)
   - Extrai user_id e coloca no contexto
```

### **Código Frontend (React/Vue/Angular) - Atual:**

```javascript
// ============================================
// 📁 frontend/src/services/api.js
// ============================================

import axios from 'axios';

const API_URL = 'http://localhost:8080/api/v1';

// Criar instância do axios
const api = axios.create({
  baseURL: API_URL,
  timeout: 10000
});

// Interceptor: Adicionar token em TODAS as requisições
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Interceptor: Tratar erros (token expirado)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expirou ou inválido - redirecionar para login
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// ============================================
// 📁 frontend/src/services/auth.js
// ============================================

export const authService = {
  // Login
  async login(email, password) {
    const response = await api.post('/login', { email, password });
    const { token, user } = response.data;
    
    // Salvar token e usuário
    localStorage.setItem('auth_token', token);
    localStorage.setItem('user', JSON.stringify(user));
    
    return { token, user };
  },
  
  // Logout
  async logout() {
    try {
      await api.post('/logout'); // Invalida token no backend
    } finally {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
  },
  
  // Verificar se está autenticado
  isAuthenticated() {
    return !!localStorage.getItem('auth_token');
  },
  
  // Pegar usuário logado
  getCurrentUser() {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  }
};

export default api;

// ============================================
// 📁 frontend/src/App.jsx (React exemplo)
// ============================================

import React, { useEffect, useState } from 'react';
import { authService } from './services/auth';
import api from './services/api';

function App() {
  const [user, setUser] = useState(null);
  const [receipts, setReceipts] = useState([]);

  useEffect(() => {
    // Verificar se está autenticado ao carregar
    if (authService.isAuthenticated()) {
      setUser(authService.getCurrentUser());
      loadReceipts();
    }
  }, []);

  const loadReceipts = async () => {
    try {
      const response = await api.get('/receipts');
      setReceipts(response.data.receipts);
    } catch (error) {
      console.error('Erro ao carregar recibos:', error);
    }
  };

  const handleLogin = async (email, password) => {
    try {
      const { user } = await authService.login(email, password);
      setUser(user);
      loadReceipts();
    } catch (error) {
      alert('Erro no login: ' + error.response?.data?.message);
    }
  };

  const handleLogout = async () => {
    await authService.logout();
    setUser(null);
    setReceipts([]);
  };

  return (
    <div>
      {user ? (
        <div>
          <h1>Bem-vindo, {user.name}!</h1>
          <button onClick={handleLogout}>Logout</button>
          <div>
            <h2>Seus Recibos:</h2>
            {receipts.map(receipt => (
              <div key={receipt.id}>{receipt.storeName} - R$ {receipt.total}</div>
            ))}
          </div>
        </div>
      ) : (
        <LoginForm onLogin={handleLogin} />
      )}
    </div>
  );
}
```

---

## ⚠️ **PROBLEMAS do Sistema Atual (Token 7 dias)**

### **1. 🔓 Janela de Ataque Longa**

```
Cenário de Ataque:
┌────────────────────────────────────────────────────────────────┐
│ DIA 1: Usuário faz login em café público                      │
│        Token armazenado: "eyJhbGciOiJIUzI1..."                │
│                                                                │
│ DIA 2: Hacker intercepta rede WiFi do café                    │
│        Rouba o token do localStorage (XSS ou network sniff)   │
│                                                                │
│ DIA 3-7: Hacker tem 5 DIAS para usar o token roubado!        │
│          Pode acessar TODOS os dados do usuário               │
│          Usuário nem percebe que foi hackeado                 │
│                                                                │
│ ❌ PROBLEMA: Token válido por muito tempo = risco alto        │
└────────────────────────────────────────────────────────────────┘
```

### **2. 🚫 Logout Não Revoga Imediatamente**

```javascript
// Frontend: Usuário clica em "Logout"
await authService.logout();
localStorage.removeItem('auth_token'); // Remove do navegador

// ❌ MAS: Se alguém copiou o token antes, ainda funciona!
// O token só é invalidado quando chega no backend (blacklist)
// Se hacker já copiou, pode usar até expirar (7 dias)
```

### **3. 💾 Impossível Revogar Sem Blacklist Global**

```
Se você quer "deslogar todos os dispositivos":
┌─────────────────────────────────────────────────────────────┐
│ Sistema atual:                                              │
│ - Precisa manter blacklist de TODOS os tokens              │
│ - Blacklist cresce infinitamente                           │
│ - Performance degrada com milhões de tokens                 │
│                                                             │
│ ❌ PROBLEMA: Não escala bem                                 │
└─────────────────────────────────────────────────────────────┘
```

---

## ✅ **SOLUÇÃO: Access Token Curto (15min) + Refresh Token (7 dias)**

### **Como Funcionaria com Tokens Curtos:**

```
┌───────────────────────────────────────────────────────────────────┐
│          FLUXO MELHORADO (Access 15min + Refresh 7 dias)          │
└───────────────────────────────────────────────────────────────────┘

1️⃣ USUÁRIO FAZ LOGIN
   POST /api/v1/login
   { "email": "joao@example.com", "password": "senha123" }
   
   ↓
   
2️⃣ BACKEND GERA 2 TOKENS
   
   Access Token (curto - 15 minutos):
   {
     "user_id": 123,
     "type": "access",
     "exp": 1730396100,  // Expira em 15 minutos
     "iat": 1730395200
   }
   
   Refresh Token (longo - 7 dias):
   {
     "user_id": 123,
     "type": "refresh",
     "exp": 1731000000,  // Expira em 7 dias
     "iat": 1730395200
   }
   
   ↓
   
3️⃣ FRONTEND RECEBE E ARMAZENA
   Response: {
     "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.access...",
     "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh...",
     "expiresIn": 900,  // 15 minutos em segundos
     "user": { "id": 123, "name": "João" }
   }
   
   localStorage.setItem('access_token', accessToken)
   localStorage.setItem('refresh_token', refreshToken)
   
   ↓
   
4️⃣ REQUISIÇÕES USAM ACCESS TOKEN
   GET /api/v1/receipts
   Headers: { "Authorization": "Bearer <access_token>" }
   
   ↓
   
5️⃣ APÓS 15 MINUTOS: Access Token Expira
   GET /api/v1/receipts
   Response: 401 Unauthorized { "message": "Token expirado" }
   
   ↓
   
6️⃣ FRONTEND RENOVA AUTOMATICAMENTE
   POST /api/v1/auth/refresh
   Headers: { "Authorization": "Bearer <refresh_token>" }
   
   Response: {
     "accessToken": "novo_access_token...",
     "expiresIn": 900
   }
   
   ↓
   
7️⃣ REPETE REQUISIÇÃO ORIGINAL COM NOVO TOKEN
   GET /api/v1/receipts
   Headers: { "Authorization": "Bearer <novo_access_token>" }
   
   ✅ Usuário nem percebe! Transparente!
```

### **Código Frontend com Refresh Token:**

```javascript
// ============================================
// 📁 frontend/src/services/api.js (MELHORADO)
// ============================================

import axios from 'axios';

const API_URL = 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_URL,
  timeout: 10000
});

let isRefreshing = false;
let failedQueue = [];

const processQueue = (error, token = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

// Interceptor: Adicionar access token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Interceptor: Renovar token automaticamente
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // Se 401 e não é tentativa de refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      
      // Se já está renovando, adiciona à fila
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(token => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return api(originalRequest);
          })
          .catch(err => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      const refreshToken = localStorage.getItem('refresh_token');

      if (!refreshToken) {
        // Sem refresh token - redirecionar para login
        localStorage.clear();
        window.location.href = '/login';
        return Promise.reject(error);
      }

      try {
        // Renovar access token
        const response = await axios.post(`${API_URL}/auth/refresh`, {}, {
          headers: { Authorization: `Bearer ${refreshToken}` }
        });

        const { accessToken } = response.data;
        localStorage.setItem('access_token', accessToken);

        // Processar fila de requisições pendentes
        processQueue(null, accessToken);

        // Repetir requisição original com novo token
        originalRequest.headers.Authorization = `Bearer ${accessToken}`;
        return api(originalRequest);

      } catch (refreshError) {
        // Refresh falhou - logout
        processQueue(refreshError, null);
        localStorage.clear();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

export default api;

// ============================================
// 📁 frontend/src/services/auth.js (MELHORADO)
// ============================================

export const authService = {
  async login(email, password) {
    const response = await api.post('/login', { email, password });
    const { accessToken, refreshToken, user } = response.data;
    
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
    localStorage.setItem('user', JSON.stringify(user));
    
    return { user };
  },
  
  async logout() {
    try {
      // Invalida refresh token no backend
      await api.post('/logout');
    } finally {
      localStorage.clear();
      window.location.href = '/login';
    }
  },
  
  isAuthenticated() {
    return !!localStorage.getItem('refresh_token');
  },
  
  getCurrentUser() {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  }
};
```

---

## 🎯 **Comparação: Atual vs Melhorado**

| Aspecto | **Sistema Atual (Token 7 dias)** | **Sistema Melhorado (Access 15min + Refresh)** |
|---------|-----------------------------------|------------------------------------------------|
| **Segurança** | ⚠️ Média - Token válido por muito tempo | ✅ Alta - Token curto limita janela de ataque |
| **UX** | ✅ Simples - Usuário não precisa relogar | ✅ Transparente - Renovação automática |
| **Revogação** | ❌ Difícil - Precisa blacklist gigante | ✅ Fácil - Invalida apenas refresh token |
| **Complexidade Frontend** | ✅ Simples - 1 token | ⚠️ Média - Precisa lógica de renovação |
| **Complexidade Backend** | ✅ Simples - 1 endpoint | ⚠️ Média - Endpoint de refresh + validação |
| **Logout Todos Dispositivos** | ❌ Impossível (precisaria blacklist de tudo) | ✅ Possível (invalida refresh token no banco) |
| **Performance** | ✅ Rápida - JWT stateless | ✅ Rápida - Access token stateless, refresh no banco |
| **Risco se Token Vazar** | 🔴 ALTO - 7 dias de acesso | 🟡 BAIXO - 15 minutos de acesso |

---

## 🤔 **Quando Usar Cada Abordagem?**

### **Token Longo (7 dias) - Seu Sistema Atual**

✅ **BOM PARA:**
- Aplicações internas (menor risco)
- MVPs e protótipos
- Apps mobile (renovação constante drena bateria)
- Quando simplicidade > segurança máxima

❌ **RUIM PARA:**
- Aplicações públicas com dados sensíveis
- Fintech, healthcare, dados pessoais
- Apps que precisam "deslogar todos os dispositivos"

### **Token Curto + Refresh**

✅ **BOM PARA:**
- Aplicações públicas
- Dados sensíveis (financeiros, saúde, PII)
- Apps que precisam controle fino de sessões
- Compliance (LGPD, GDPR, PCI-DSS)

❌ **RUIM PARA:**
- Apps mobile (muitas renovações = bateria)
- Quando simplicidade é prioridade
- Infraestrutura limitada

---

## 📝 **Recomendação para Seu Sistema**

### **Manter Token de 7 Dias SE:**

1. ✅ Seu app não lida com dados super sensíveis
2. ✅ É para uso pessoal/acadêmico (TCC)
3. ✅ Você quer focar em features, não em segurança avançada
4. ✅ Tem rate limiting (você já tem! ✅)
5. ✅ Tem HTTPS (você já tem! ✅)

### **Migrar para Access + Refresh SE:**

1. ✅ Vai para produção com usuários reais
2. ✅ Lida com dados financeiros (recibos = gastos = sensível)
3. ✅ Quer compliance com LGPD
4. ✅ Precisa de "deslogar todos os dispositivos"
5. ✅ Quer segurança máxima

---

## 🚀 **Conclusão**

### **Seu sistema ATUAL está OK para TCC porque:**

✅ Tem rate limiting (limita ataques de força bruta)  
✅ Tem HTTPS enforcement (previne man-in-the-middle)  
✅ Tem blacklist de logout (invalida tokens)  
✅ Tem email enumeration protection  
✅ Tem bcrypt cost 12 (senhas bem protegidas)  

### **A vulnerabilidade do token de 7 dias é GERENCIÁVEL porque:**

- Se alguém roubar o token, tem rate limiting impedindo abuso massivo
- HTTPS previne interceptação de rede
- Blacklist no logout funciona para casos normais
- Para TCC, o tradeoff simplicidade vs segurança máxima vale a pena

### **Se fosse produção real, eu recomendaria:**

- Access token de 15 minutos
- Refresh token de 7 dias
- Refresh tokens armazenados no banco (revogáveis)
- Endpoint para "deslogar todos os dispositivos"

---

**📌 TLDR:** Seu sistema ATUAL é seguro o suficiente para TCC. A implementação de refresh tokens seria ótima para produção, mas adiciona complexidade que pode não valer a pena para um projeto acadêmico. **Foque em features! 🚀**

