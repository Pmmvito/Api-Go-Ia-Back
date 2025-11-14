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

    ⏰ APÓS 7 DIAS (168 horas):
    - Token EXPIRA automaticamente
    - Qualquer requisição retorna: 401 Unauthorized
    - Frontend detecta erro 401 e redireciona para /login
    - ✅ SIM, você precisa fazer login novamente!
```

### **⏰ O Que Acontece Após 8 Dias (Token Expirado):**

```
DIA 1 (10:00): Usuário faz login
                    ↓
                    Token criado: válido até Dia 8 (10:00)

DIA 1-7:       ✅ Token funciona normalmente
                    Usuário acessa o app sem problemas

DIA 8 (10:01): ❌ Token EXPIROU!
                    ↓
                    GET /api/v1/receipts
                    Response: 401 Unauthorized
                    {
                      "error": "Token expirado",
                      "message": "Faça login novamente"
                    }
                    ↓
                    Frontend detecta 401 → redireciona para /login
                    ↓
                    🔄 Usuário precisa fazer login novamente

┌─────────────────────────────────────────────────────────────┐
│ ⚠️ IMPORTANTE: Após 7 dias, o token é INVÁLIDO!            │
│                                                             │
│ Não importa se você ainda tem o token salvo:               │
│ - localStorage ainda tem o token antigo                    │
│ - Mas o backend REJEITA porque exp < agora                 │
│ - Você DEVE fazer login novamente para obter novo token    │
└─────────────────────────────────────────────────────────────┘
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
        // ⏰ Token expirou (após 7 dias) ou inválido
        // Limpa tudo e força novo login
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user');
        window.location.href = '/login';
        
        // Pode mostrar mensagem amigável:
        alert('Sua sessão expirou. Por favor, faça login novamente.');
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
     
     // 💡 OPCIONAL: Salvar timestamp para mostrar "expira em X dias"
     const expiresAt = Date.now() + (7 * 24 * 60 * 60 * 1000); // +7 dias
     localStorage.setItem('token_expires_at', expiresAt);
     
     return { token, user };
  },
  
  // Logout
  async logout() {
     try {
        await api.post('/logout'); // Invalida token no backend
     } finally {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user');
        localStorage.removeItem('token_expires_at');
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
  },
  
  // 💡 OPCIONAL: Verificar quantos dias faltam para expirar
  getDaysUntilExpiration() {
     const expiresAt = localStorage.getItem('token_expires_at');
     if (!expiresAt) return null;
     
     const now = Date.now();
     const diff = expiresAt - now;
     const days = Math.ceil(diff / (1000 * 60 * 60 * 24));
     
     return days > 0 ? days : 0; // Retorna 0 se já expirou
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
  const [daysLeft, setDaysLeft] = useState(null);

  useEffect(() => {
     // Verificar se está autenticado ao carregar
     if (authService.isAuthenticated()) {
        setUser(authService.getCurrentUser());
        setDaysLeft(authService.getDaysUntilExpiration());
        loadReceipts();
     }
  }, []);

  const loadReceipts = async () => {
     try {
        const response = await api.get('/receipts');
        setReceipts(response.data.receipts);
     } catch (error) {
        console.error('Erro ao carregar recibos:', error);
        // Se erro 401, interceptor já redireciona para /login
     }
  };

  const handleLogin = async (email, password) => {
     try {
        const { user } = await authService.login(email, password);
        setUser(user);
        setDaysLeft(7); // Token novo = 7 dias
        loadReceipts();
     } catch (error) {
        alert('Erro no login: ' + error.response?.data?.message);
     }
  };

  const handleLogout = async () => {
     await authService.logout();
     setUser(null);
     setReceipts([]);
     setDaysLeft(null);
  };

  return (
     <div>
        {user ? (
          <div>
             <h1>Bem-vindo, {user.name}!</h1>
             
             {/* 💡 Aviso de expiração */}
             {daysLeft !== null && daysLeft <= 1 && (
                <div style={{backgroundColor: 'yellow', padding: '10px'}}>
                  ⚠️ Sua sessão expira em {daysLeft} dia(s)! 
                  Faça login novamente para renovar.
                </div>
             )}
             
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

### **4. 😤 Experiência Ruim Após 7 Dias**

```
┌─────────────────────────────────────────────────────────────┐
│ Usuário abre app no DIA 8:                                 │
│ ↓                                                           │
│ ❌ Todas as requisições falham (401)                        │
│ ↓                                                           │
│ 😤 Forçado a fazer login novamente                         │
│ ↓                                                           │
│ Perde contexto (estava editando algo? Perdeu!)             │
│                                                             │
│ ⚠️ PROBLEMA: Interrompe fluxo do usuário                   │
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

    ⏰ APÓS 7 DIAS (sem usar o app):
    - Refresh token EXPIRA
    - Próxima vez que abrir: POST /auth/refresh → 401
    - Frontend redireciona para /login
    - ✅ Precisa fazer login novamente (igual ao sistema atual)
    
    💡 DIFERENÇA: Se usar o app DENTRO dos 7 dias, renova automaticamente!
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
          // ⏰ Refresh falhou (refresh token expirou após 7 dias) - logout
          processQueue(refreshError, null);
          localStorage.clear();
          alert('Sua sessão expirou após 7 dias. Por favor, faça login novamente.');
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
| **Login após 8 dias** | ❌ SIM - Token expirou, precisa relogar | ❌ SIM - Mas se usou nos últimos 7 dias, NÃO precisa |
| **Experiência de uso** | 😤 Interrompe após 7 dias SEMPRE | ✅ Renova automático se usar regularmente |
| **Segurança** | ⚠️ Média - Token válido por muito tempo | ✅ Alta - Token curto limita janela de ataque |
| **UX** | ⚠️ Ruim - Força relogin após 7 dias | ✅ Excelente - Renovação transparente |
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
- **❌ UX ruim: usuário SEMPRE precisa relogar após 7 dias**

### **Token Curto + Refresh**

✅ **BOM PARA:**
- Aplicações públicas
- Dados sensíveis (financeiros, saúde, PII)
- Apps que precisam controle fino de sessões
- Compliance (LGPD, GDPR, PCI-DSS)
- **✅ UX melhor: renovação automática se usar regularmente**

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
6. ⚠️ Aceita que usuário precisa relogar SEMPRE após 7 dias

### **Migrar para Access + Refresh SE:**

1. ✅ Vai para produção com usuários reais
2. ✅ Lida com dados financeiros (recibos = gastos = sensível)
3. ✅ Quer compliance com LGPD
4. ✅ Precisa de "deslogar todos os dispositivos"
5. ✅ Quer segurança máxima
6. ✅ **Quer UX melhor: usuário não precisa relogar se usar regularmente**

---

## 🚀 **Conclusão**

### **Respondendo sua pergunta: "E se passar 8 dias, eu tenho que fazer login de novo?"**

**Sistema ATUAL (Token 7 dias):**
- ✅ **SIM, SEMPRE!** Após 7 dias o token expira e você DEVE fazer login novamente
- Não importa se você usou ontem - se passou 7 dias desde o login, precisa relogar
- ❌ **UX ruim:** Interrompe fluxo do usuário a cada 7 dias

**Sistema com Refresh Token:**
- ⚠️ **DEPENDE!**
  - Se você NÃO usou o app nos últimos 7 dias: SIM, precisa relogar
  - Se você USOU o app nos últimos 7 dias: NÃO precisa, renova automaticamente!
- ✅ **UX melhor:** Só força login se realmente ficar inativo por 7 dias

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
- **✅ Melhor UX: usuário não é forçado a relogar a cada 7 dias**

---

**📌 TLDR:** 

No seu sistema ATUAL: **SIM, após 8 dias você DEVE fazer login novamente** (o token expira em 7 dias).

Com Refresh Token: **Só precisa relogar se ficar 7 dias SEM usar** (se usar dentro dos 7 dias, renova automaticamente e mantém sessão ativa).

**Foque em features! 🚀** Para TCC, sistema atual é suficiente. Refresh tokens são ótimos para produção.
