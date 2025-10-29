# Exemplos de Requisições da API

Este arquivo contém exemplos de como testar os endpoints da API.

## 1. Registro de Usuário

### Request
```http
POST http://localhost:8080/api/v1/register
Content-Type: application/json

{
  "name": "João Silva",
  "email": "teste@exemplo.com",
  "password": "senha123"
}
```

### Response (201 Created)
```json
{
  "message": "User registered successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAzMjI4NjAsImlhdCI6MTcyOTcxODA2MCwidXNlcl9pZCI6MX0.xxx",
  "user": {
    "id": 1,
    "createdAt": "2025-10-24T17:34:20.123456Z",
    "updatedAt": "2025-10-24T17:34:20.123456Z",
    "name": "João Silva",
    "email": "teste@exemplo.com"
  }
}
```

## 2. Login

### Request
```http
POST http://localhost:8080/api/v1/login
Content-Type: application/json

{
  "email": "teste@exemplo.com",
  "password": "senha123"
}
```

### Response (200 OK)
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAzMjI4NjAsImlhdCI6MTcyOTcxODA2MCwidXNlcl9pZCI6MX0.xxx",
  "user": {
    "id": 1,
    "createdAt": "2025-10-24T17:34:20.123456Z",
    "updatedAt": "2025-10-24T17:34:20.123456Z",
    "name": "João Silva",
    "email": "teste@exemplo.com"
  }
}
```

## 3. Obter Dados do Usuário Autenticado

### Request (Requer Token)
```http
GET http://localhost:8080/api/v1/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAzMjI4NjAsImlhdCI6MTcyOTcxODA2MCwidXNlcl9pZCI6MX0.xxx
```

### Response (200 OK)
```json
{
  "message": "User retrieved successfully",
  "data": {
    "id": 1,
    "createdAt": "2025-10-24T17:34:20.123456Z",
    "updatedAt": "2025-10-24T17:34:20.123456Z",
    "name": "João Silva",
    "email": "teste@exemplo.com"
  }
}
```

## Erros Comuns

### 400 Bad Request - Email Inválido
```json
{
  "message": "Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag",
  "errorCode": 400
}
```

### 400 Bad Request - Senha Curta
```json
{
  "message": "Key: 'RegisterRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag",
  "errorCode": 400
}
```

### 400 Bad Request - Nome Curto
```json
{
  "message": "Key: 'RegisterRequest.Name' Error:Field validation for 'Name' failed on the 'min' tag",
  "errorCode": 400
}
```

### 400 Bad Request - Email Já Registrado
```json
{
  "message": "Email already registered",
  "errorCode": 400
}
```

### 401 Unauthorized - Credenciais Inválidas
```json
{
  "message": "Invalid email or password",
  "errorCode": 401
}
```

### 401 Unauthorized - Token Ausente
```json
{
  "message": "Authorization header is required",
  "errorCode": 401
}
```

### 401 Unauthorized - Token Inválido
```json
{
  "message": "Invalid or expired token",
  "errorCode": 401
}
```

## Testando com cURL

### Registrar
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"name":"João Silva","email":"teste@exemplo.com","password":"senha123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"teste@exemplo.com","password":"senha123"}'
```

### Me (Substitua YOUR_TOKEN pelo token recebido)
```bash
curl -X GET http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```
