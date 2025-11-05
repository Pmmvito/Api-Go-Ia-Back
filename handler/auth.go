package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RegisterRequest define a estrutura de dados para o registro de um novo usuário.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2" example:"João Silva"`
	Email    string `json:"email" binding:"required,email" example:"joao@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"senha123"`
}

// LoginRequest define a estrutura de dados para o login de um usuário.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"joao@example.com"`
	Password string `json:"password" binding:"required" example:"senha123"`
}

// AuthResponse define a estrutura da resposta de autenticação, contendo o token JWT e os dados do usuário.
type AuthResponse struct {
	Message string               `json:"message"`
	Token   string               `json:"token"`
	User    schemas.UserResponse `json:"user"`
}

// GenerateJWT gera um token JWT para o usuário com validade de 7 dias.
func GenerateJWT(userID uint) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key-change-in-production"
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // Token válido por 7 dias
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// @Summary Register new user
// @Description Create a new user account. After registration, use the login endpoint to get your JWT token.
// @Tags 🔐 Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data (name, email, password)"
// @Success 201 {object} AuthResponse "User created successfully with JWT token"
// @Failure 400 {object} ErrorResponse "Invalid request or email already registered"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /register [post]
func RegisterHandler(ctx *gin.Context) {
	var request RegisterRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Verifica se o email já existe
	var existingUser schemas.User
	if err := db.Where("email = ?", request.Email).First(&existingUser).Error; err == nil {
		sendError(ctx, http.StatusBadRequest, "Email already registered")
		return
	}

	// Cria novo usuário
	user := schemas.User{
		Name:  request.Name,
		Email: request.Email,
	}

	// Hash da senha
	if err := user.HashPassword(request.Password); err != nil {
		logger.ErrorF("error hashing password: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error creating user")
		return
	}

	// Salva no banco
	if err := db.Create(&user).Error; err != nil {
		logger.ErrorF("error creating user: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error creating user")
		return
	}

	// Gera token JWT
	token, err := GenerateJWT(user.ID)
	if err != nil {
		logger.ErrorF("error generating token: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error generating authentication token")
		return
	}

	// 🔒 NOVO: Salva token no usuário
	user.ActiveToken = &token
	if err := db.Save(&user).Error; err != nil {
		logger.ErrorF("error saving active token: %v", err.Error())
		// Não falha o registro por isso, apenas loga
	}

	// Retorna resposta
	ctx.JSON(http.StatusCreated, AuthResponse{
		Message: "User registered successfully",
		Token:   token,
		User:    user.ToResponse(),
	})
}

// @Summary Login
// @Description Authenticate user with email and password. Returns a JWT token valid for 7 days. Use this token in the Authorization header as "Bearer {token}" for all protected endpoints.
// @Tags 🔐 Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User credentials (email and password)"
// @Success 200 {object} AuthResponse "Login successful with JWT token"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Invalid email or password"
// @Router /login [post]
func LoginHandler(ctx *gin.Context) {
	var request LoginRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Busca usuário por email
	var user schemas.User
	if err := db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		sendError(ctx, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Verifica senha
	if !user.CheckPassword(request.Password) {
		sendError(ctx, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// 🔒 NOVO: Invalida token anterior se existir
	if user.ActiveToken != nil && *user.ActiveToken != "" {
		logger.InfoF("Invalidating previous token for user %d", user.ID)

		// Adiciona token anterior à blacklist
		expiresAt := time.Now().Add(time.Hour * 24 * 7) // Mesmo TTL do token
		db.Create(&schemas.TokenBlacklist{
			UserID:    user.ID,
			Token:     *user.ActiveToken,
			ExpiresAt: expiresAt,
		})
	}

	// Gera novo token JWT
	token, err := GenerateJWT(user.ID)
	if err != nil {
		logger.ErrorF("error generating token: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Error generating authentication token")
		return
	}

	// 🔒 NOVO: Salva novo token no usuário
	user.ActiveToken = &token
	if err := db.Save(&user).Error; err != nil {
		logger.ErrorF("error saving active token: %v", err.Error())
		// Não falha o login por isso, apenas loga
	}

	// Retorna resposta
	ctx.JSON(http.StatusOK, AuthResponse{
		Message: "Login successful",
		Token:   token,
		User:    user.ToResponse(),
	})
}

// @Summary Get current user profile
// @Description Get information about the authenticated user (requires JWT token)
// @Tags 🔐 Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} schemas.UserResponse "User information retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Router /me [get]
func MeHandler(ctx *gin.Context) {
	// Pega o usuário do contexto (injetado pelo middleware)
	userInterface, exists := ctx.Get("user")
	if !exists {
		sendError(ctx, http.StatusUnauthorized, "User not found")
		return
	}

	user := userInterface.(schemas.User)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User retrieved successfully",
		"data":    user.ToResponse(),
	})
}
