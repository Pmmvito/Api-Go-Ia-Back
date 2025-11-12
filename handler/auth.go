package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/config"
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
// @Failure 400 {object} ErrorResponse "Dados de registro inválidos: verifique se nome (mínimo 2 caracteres), email válido e senha (mínimo 6 caracteres) foram fornecidos corretamente | Este email já está cadastrado. Por favor, utilize outro email ou faça login | Este email foi utilizado em uma conta deletada e não pode ser reutilizado por questões de segurança"
// @Failure 500 {object} ErrorResponse "Erro ao processar a senha durante o cadastro. Por favor, tente novamente | Erro ao criar usuário no banco de dados. Por favor, tente novamente mais tarde | Usuário criado com sucesso, mas houve erro ao gerar o token de autenticação. Por favor, faça login"
// @Router /register [post]
func RegisterHandler(ctx *gin.Context) {
	var request RegisterRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados de registro inválidos: verifique se nome (mínimo 2 caracteres), email válido e senha (mínimo 6 caracteres) foram fornecidos corretamente")
		return
	}

	// Verifica se o email já existe (incluindo usuários deletados)
	// Usamos Unscoped() para buscar também usuários com deleted_at não null
	var existingUser schemas.User
	if err := db.Unscoped().Where("email = ?", request.Email).First(&existingUser).Error; err == nil {
		// Email encontrado - pode ser usuário ativo ou deletado
		if existingUser.DeletedAt.Valid {
			sendError(ctx, http.StatusBadRequest, "Este email foi utilizado em uma conta deletada e não pode ser reutilizado por questões de segurança")
		} else {
			sendError(ctx, http.StatusBadRequest, "Este email já está cadastrado. Por favor, utilize outro email ou faça login")
		}
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
		sendError(ctx, http.StatusInternalServerError, "Erro ao processar a senha durante o cadastro. Por favor, tente novamente")
		return
	}

	// Salva no banco
	if err := db.Create(&user).Error; err != nil {
		logger.ErrorF("error creating user: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao criar usuário no banco de dados. Por favor, tente novamente mais tarde")
		return
	}

	// Cria categorias padrão para o novo usuário
	if err := config.CreateDefaultCategoriesForUser(db, user.ID); err != nil {
		logger.ErrorF("error creating default categories for user: %v", err.Error())
		// Não falha o registro por isso, mas loga o erro
		// O usuário pode criar suas categorias manualmente depois
	}

	// Gera token JWT
	token, err := GenerateJWT(user.ID)
	if err != nil {
		logger.ErrorF("error generating token: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Usuário criado com sucesso, mas houve erro ao gerar o token de autenticação. Por favor, faça login")
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
// @Failure 400 {object} ErrorResponse "Dados de login inválidos: email e senha são obrigatórios"
// @Failure 401 {object} ErrorResponse "Email ou senha incorretos. Verifique suas credenciais e tente novamente"
// @Failure 500 {object} ErrorResponse "Erro ao gerar token de autenticação. Por favor, tente novamente"
// @Router /login [post]
func LoginHandler(ctx *gin.Context) {
	var request LoginRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados de login inválidos: email e senha são obrigatórios")
		return
	}

	// Busca usuário por email
	var user schemas.User
	if err := db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		sendError(ctx, http.StatusUnauthorized, "Email ou senha incorretos. Verifique suas credenciais e tente novamente")
		return
	}

	// Verifica senha
	if !user.CheckPassword(request.Password) {
		sendError(ctx, http.StatusUnauthorized, "Email ou senha incorretos. Verifique suas credenciais e tente novamente")
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
		sendError(ctx, http.StatusInternalServerError, "Erro ao gerar token de autenticação. Por favor, tente novamente")
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
// @Failure 401 {object} ErrorResponse "Usuário não encontrado no contexto de autenticação. Token pode estar inválido ou expirado"
// @Router /me [get]
func MeHandler(ctx *gin.Context) {
	// Pega o usuário do contexto (injetado pelo middleware)
	userInterface, exists := ctx.Get("user")
	if !exists {
		sendError(ctx, http.StatusUnauthorized, "Usuário não encontrado no contexto de autenticação. Token pode estar inválido ou expirado")
		return
	}

	user := userInterface.(schemas.User)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User retrieved successfully",
		"data":    user.ToResponse(),
	})
}

// ForgotPasswordRequest define a estrutura para solicitar recuperação de senha
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"joao@example.com"`
}

// ResetPasswordRequest define a estrutura para redefinir a senha
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email" example:"joao@example.com"`
	Token       string `json:"token" binding:"required,len=6" example:"123456"`
	NewPassword string `json:"newPassword" binding:"required,min=6" example:"novaSenha123"`
}

// @Summary Request password reset
// @Description Send a 6-digit code to user's email for password recovery. Code expires in 15 minutes.
// @Tags 🔐 Authentication
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "User email"
// @Success 200 {object} map[string]interface{} "Reset code sent successfully"
// @Failure 400 {object} ErrorResponse "Dados inválidos: email é obrigatório e deve ser válido"
// @Failure 404 {object} ErrorResponse "Usuário não encontrado com este email"
// @Failure 500 {object} ErrorResponse "Erro ao gerar código de recuperação | Erro ao enviar email"
// @Router /auth/forgot-password [post]
func ForgotPasswordHandler(ctx *gin.Context) {
	var request ForgotPasswordRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados inválidos: email é obrigatório e deve ser válido")
		return
	}

	// Busca usuário por email
	var user schemas.User
	if err := db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		// Por segurança, não revela se o email existe ou não
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Se o email estiver cadastrado, você receberá um código de recuperação",
		})
		return
	}

	// Gera código de 6 dígitos
	code, err := GenerateRandomCode(6)
	if err != nil {
		logger.ErrorF("error generating reset code: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao gerar código de recuperação. Por favor, tente novamente")
		return
	}

	// Invalida tokens anteriores do usuário
	db.Model(&schemas.PasswordReset{}).
		Where("user_id = ? AND used = false", user.ID).
		Update("used", true)

	// Cria novo token de recuperação
	passwordReset := schemas.PasswordReset{
		UserID:    user.ID,
		Token:     code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	if err := db.Create(&passwordReset).Error; err != nil {
		logger.ErrorF("error creating password reset: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao criar token de recuperação. Por favor, tente novamente")
		return
	}

	// Envia email
	emailService := config.NewEmailService()
	if !emailService.IsConfigured() {
		logger.ErrorF("email service not configured")
		sendError(ctx, http.StatusInternalServerError, "Serviço de email não configurado. Configure as variáveis SMTP_EMAIL e SMTP_PASSWORD")
		return
	}

	if err := emailService.SendPasswordResetEmail(user.Email, user.Name, code); err != nil {
		logger.ErrorF("error sending email: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao enviar email. Por favor, tente novamente")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Código de recuperação enviado para seu email. Válido por 15 minutos.",
	})
}

// @Summary Reset password with code
// @Description Reset user password using the 6-digit code received by email
// @Tags 🔐 Authentication
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Email, token and new password"
// @Success 200 {object} map[string]interface{} "Password reset successfully"
// @Failure 400 {object} ErrorResponse "Dados inválidos | Código deve ter 6 dígitos | Senha deve ter no mínimo 6 caracteres"
// @Failure 401 {object} ErrorResponse "Código inválido ou expirado"
// @Failure 404 {object} ErrorResponse "Usuário não encontrado"
// @Failure 500 {object} ErrorResponse "Erro ao atualizar senha | Erro ao enviar email de confirmação"
// @Router /auth/reset-password [post]
func ResetPasswordHandler(ctx *gin.Context) {
	var request ResetPasswordRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados inválidos: email, código (6 dígitos) e nova senha (mínimo 6 caracteres) são obrigatórios")
		return
	}

	// Busca usuário
	var user schemas.User
	if err := db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Usuário não encontrado com este email")
		return
	}

	// Busca token válido
	var passwordReset schemas.PasswordReset
	if err := db.Where("user_id = ? AND token = ? AND used = false", user.ID, request.Token).
		First(&passwordReset).Error; err != nil {
		sendError(ctx, http.StatusUnauthorized, "Código inválido ou já utilizado")
		return
	}

	// Verifica se o token ainda é válido
	if !passwordReset.IsValid() {
		sendError(ctx, http.StatusUnauthorized, "Código expirado. Solicite um novo código de recuperação")
		return
	}

	// Atualiza senha
	if err := user.HashPassword(request.NewPassword); err != nil {
		logger.ErrorF("error hashing password: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao processar nova senha. Por favor, tente novamente")
		return
	}

	if err := db.Save(&user).Error; err != nil {
		logger.ErrorF("error updating password: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao atualizar senha. Por favor, tente novamente")
		return
	}

	// Marca token como usado
	if err := passwordReset.MarkAsUsed(db); err != nil {
		logger.ErrorF("error marking token as used: %v", err.Error())
		// Não falha a operação, apenas loga
	}

	// Invalida tokens JWT ativos
	if user.ActiveToken != nil {
		expiresAt := time.Now().Add(time.Hour * 24 * 7)
		db.Create(&schemas.TokenBlacklist{
			UserID:    user.ID,
			Token:     *user.ActiveToken,
			ExpiresAt: expiresAt,
		})
		user.ActiveToken = nil
		db.Save(&user)
	}

	// Envia email de confirmação
	emailService := config.NewEmailService()
	if emailService.IsConfigured() {
		if err := emailService.SendPasswordChangedEmail(user.Email, user.Name); err != nil {
			logger.ErrorF("error sending confirmation email: %v", err.Error())
			// Não falha a operação, apenas loga
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Senha alterada com sucesso! Faça login com sua nova senha.",
	})
}

// ChangePasswordRequest define a estrutura para trocar senha quando logado
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required,min=6" example:"senhaAtual123"`
	NewPassword     string `json:"newPassword" binding:"required,min=6" example:"novaSenha123"`
}

// @Summary Change password (authenticated)
// @Description Change password for authenticated user. Requires current password for security. User will remain logged in after password change.
// @Tags 🔐 Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Current and new password"
// @Success 200 {object} map[string]interface{} "Password changed successfully"
// @Failure 400 {object} ErrorResponse "Dados inválidos: senha atual e nova senha (mínimo 6 caracteres) são obrigatórios | Nova senha deve ser diferente da senha atual"
// @Failure 401 {object} ErrorResponse "Senha atual incorreta. Verifique e tente novamente | Unauthorized"
// @Failure 500 {object} ErrorResponse "Erro ao processar nova senha | Erro ao atualizar senha"
// @Router /auth/change-password [post]
func ChangePasswordHandler(ctx *gin.Context) {
	var request ChangePasswordRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados inválidos: senha atual e nova senha (mínimo 6 caracteres) são obrigatórios")
		return
	}

	// Pega usuário do contexto
	userInterface, exists := ctx.Get("user")
	if !exists {
		sendError(ctx, http.StatusUnauthorized, "Usuário não encontrado no contexto de autenticação")
		return
	}
	user := userInterface.(schemas.User)

	// Verifica se a senha atual está correta
	if !user.CheckPassword(request.CurrentPassword) {
		sendError(ctx, http.StatusUnauthorized, "Senha atual incorreta. Verifique e tente novamente")
		return
	}

	// Verifica se a nova senha é diferente da atual
	if request.CurrentPassword == request.NewPassword {
		sendError(ctx, http.StatusBadRequest, "Nova senha deve ser diferente da senha atual")
		return
	}

	// Atualiza para nova senha
	if err := user.HashPassword(request.NewPassword); err != nil {
		logger.ErrorF("error hashing password: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao processar nova senha. Por favor, tente novamente")
		return
	}

	if err := db.Save(&user).Error; err != nil {
		logger.ErrorF("error updating password: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao atualizar senha. Por favor, tente novamente")
		return
	}

	// Envia email de notificação
	emailService := config.NewEmailService()
	if emailService.IsConfigured() {
		if err := emailService.SendPasswordChangedEmail(user.Email, user.Name); err != nil {
			logger.ErrorF("error sending confirmation email: %v", err.Error())
			// Não falha a operação, apenas loga
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Senha alterada com sucesso! Você permanece logado.",
	})
}
