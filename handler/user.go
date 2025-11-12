package handler

import (
	"net/http"
	"time"

	"github.com/Pmmvito/Golang-Api-Exemple/config"
	"github.com/Pmmvito/Golang-Api-Exemple/schemas"
	"github.com/gin-gonic/gin"
)

// @Summary Delete user account
// @Description Soft delete user account and all associated data (receipts, items, products, shopping lists, tokens). Email cannot be reused even after deletion.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 401 {object} ErrorResponse "ID do usuário não encontrado no contexto de autenticação. Por favor, faça login novamente"
// @Failure 404 {object} ErrorResponse "Usuário não encontrado no banco de dados. Pode ter sido deletado anteriormente"
// @Failure 500 {object} ErrorResponse "Erro ao buscar notas fiscais do usuário durante a exclusão | Erro ao deletar itens das notas fiscais. Operação cancelada | Erro ao deletar notas fiscais. Operação cancelada | Erro ao buscar listas de compras durante a exclusão | Erro ao deletar itens das listas de compras. Operação cancelada | Erro ao deletar listas de compras. Operação cancelada | Erro ao deletar tokens da blacklist. Operação cancelada | Erro ao deletar registros de uso da IA. Operação cancelada | Erro ao deletar usuário. Operação cancelada | Erro ao confirmar a exclusão no banco de dados. Por favor, tente novamente"
// @Router /user [delete]
func DeleteUserHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		sendError(ctx, http.StatusUnauthorized, "ID do usuário não encontrado no contexto de autenticação. Por favor, faça login novamente")
		return
	}

	// Busca o usuário
	var user schemas.User
	if err := db.First(&user, userID).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Usuário não encontrado no banco de dados. Pode ter sido deletado anteriormente")
		return
	}

	// Inicia transação para garantir atomicidade
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Contador de itens deletados
	var receiptsDeleted, itemsDeleted, shoppingListsDeleted, listItemsDeleted int64
	var tokenBlacklistDeleted, aiTokenUsageDeleted int64

	// 1. Buscar todos os receipts do usuário
	var receipts []schemas.Receipt
	if err := tx.Where("user_id = ?", user.ID).Find(&receipts).Error; err != nil {
		tx.Rollback()
		config.GetLogger("handler").ErrorF("error finding receipts: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao buscar notas fiscais do usuário durante a exclusão")
		return
	}

	// 2. Para cada receipt, deletar apenas os items (NÃO deletamos produtos pois podem ser compartilhados)
	for _, receipt := range receipts {
		// Soft delete dos receipt items
		result := tx.Where("receipt_id = ?", receipt.ID).Delete(&schemas.ReceiptItem{})
		if result.Error != nil {
			tx.Rollback()
			config.GetLogger("handler").ErrorF("error deleting receipt items: %v", result.Error.Error())
			sendError(ctx, http.StatusInternalServerError, "Erro ao deletar itens das notas fiscais. Operação cancelada")
			return
		}
		itemsDeleted += result.RowsAffected
	}

	// 3. Soft delete dos receipts
	if len(receipts) > 0 {
		result := tx.Where("user_id = ?", user.ID).Delete(&schemas.Receipt{})
		if result.Error != nil {
			tx.Rollback()
			config.GetLogger("handler").ErrorF("error deleting receipts: %v", result.Error.Error())
			sendError(ctx, http.StatusInternalServerError, "Erro ao deletar notas fiscais. Operação cancelada")
			return
		}
		receiptsDeleted = result.RowsAffected
	}

	// 4. Buscar e deletar shopping lists
	var shoppingLists []schemas.ShoppingList
	if err := tx.Where("user_id = ?", user.ID).Find(&shoppingLists).Error; err != nil {
		tx.Rollback()
		config.GetLogger("handler").ErrorF("error finding shopping lists: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao buscar listas de compras durante a exclusão")
		return
	}

	// 5. Deletar list items de cada shopping list
	for _, list := range shoppingLists {
		result := tx.Where("list_id = ?", list.ID).Delete(&schemas.ListItem{})
		if result.Error != nil {
			tx.Rollback()
			config.GetLogger("handler").ErrorF("error deleting list items: %v", result.Error.Error())
			sendError(ctx, http.StatusInternalServerError, "Erro ao deletar itens das listas de compras. Operação cancelada")
			return
		}
		listItemsDeleted += result.RowsAffected
	}

	// 6. Soft delete das shopping lists
	if len(shoppingLists) > 0 {
		result := tx.Where("user_id = ?", user.ID).Delete(&schemas.ShoppingList{})
		if result.Error != nil {
			tx.Rollback()
			config.GetLogger("handler").ErrorF("error deleting shopping lists: %v", result.Error.Error())
			sendError(ctx, http.StatusInternalServerError, "Erro ao deletar listas de compras. Operação cancelada")
			return
		}
		shoppingListsDeleted = result.RowsAffected
	}

	// 7. Deletar token blacklist entries do usuário (hard delete - tabela auxiliar)
	result := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&schemas.TokenBlacklist{})
	if result.Error != nil {
		tx.Rollback()
		config.GetLogger("handler").ErrorF("error deleting token blacklist: %v", result.Error.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao deletar tokens da blacklist. Operação cancelada")
		return
	}
	tokenBlacklistDeleted = result.RowsAffected

	// 8. Deletar AI token usage do usuário (hard delete - tabela de logs)
	result = tx.Unscoped().Where("user_id = ?", user.ID).Delete(&schemas.AITokenUsage{})
	if result.Error != nil {
		tx.Rollback()
		config.GetLogger("handler").ErrorF("error deleting AI token usage: %v", result.Error.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao deletar registros de uso da IA. Operação cancelada")
		return
	}
	aiTokenUsageDeleted = result.RowsAffected

	// 9. Soft delete do usuário
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		config.GetLogger("handler").ErrorF("error deleting user: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao deletar usuário. Operação cancelada")
		return
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		config.GetLogger("handler").ErrorF("error committing transaction: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao confirmar a exclusão no banco de dados. Por favor, tente novamente")
		return
	}

	config.GetLogger("handler").InfoF("User %d deleted successfully with all associated data", user.ID)
	config.GetLogger("handler").InfoF("Deleted: %d receipts, %d items, %d shopping lists, %d list items, %d token blacklist entries, %d AI token usage records",
		receiptsDeleted, itemsDeleted, shoppingListsDeleted, listItemsDeleted, tokenBlacklistDeleted, aiTokenUsageDeleted)

	ctx.JSON(http.StatusOK, gin.H{
		"message":              "User account and all associated data deleted successfully",
		"receiptsDeleted":      receiptsDeleted,
		"itemsDeleted":         itemsDeleted,
		"shoppingListsDeleted": shoppingListsDeleted,
		"listItemsDeleted":     listItemsDeleted,
		"note":                 "Your email cannot be used to create a new account. Products are preserved if used by other users.",
	})
}

// UpdateProfileRequest define a estrutura para atualizar perfil do usuário
type UpdateProfileRequest struct {
	Name  *string `json:"name,omitempty" example:"João Silva"`
	Email *string `json:"email,omitempty" example:"novo@example.com"`
}

// VerifyEmailRequest define a estrutura para solicitar verificação de email
type VerifyEmailRequest struct {
	NewEmail string `json:"newEmail" binding:"required,email" example:"novo@example.com"`
}

// ConfirmEmailRequest define a estrutura para confirmar novo email
type ConfirmEmailRequest struct {
	NewEmail string `json:"newEmail" binding:"required,email" example:"novo@example.com"`
	Token    string `json:"token" binding:"required,len=6" example:"123456"`
}

// EmailVerification armazena códigos de verificação de email
type EmailVerification struct {
	UserID    uint
	NewEmail  string
	Token     string
	ExpiresAt time.Time
	Used      bool
}

// Mapa temporário para armazenar verificações de email (em produção, use banco de dados)
var emailVerifications = make(map[uint]*EmailVerification)

// @Summary Update user profile
// @Description Update user name. Email changes require verification code.
// @Tags 👤 User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile data to update"
// @Success 200 {object} map[string]interface{} "Profile updated successfully"
// @Failure 400 {object} ErrorResponse "Dados inválidos | Nenhum campo para atualizar"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Erro ao atualizar perfil"
// @Router /user/profile [patch]
func UpdateProfileHandler(ctx *gin.Context) {
	var request UpdateProfileRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados inválidos. Verifique os campos enviados")
		return
	}

	// Pega usuário do contexto
	userInterface, _ := ctx.Get("user")
	user := userInterface.(schemas.User)

	updated := false

	// Atualiza nome se fornecido
	if request.Name != nil && *request.Name != "" {
		user.Name = *request.Name
		updated = true
	}

	// Para email, precisamos de verificação - apenas retorna instrução
	if request.Email != nil && *request.Email != "" {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Para alterar o email, use o endpoint POST /user/request-email-change",
			"info":    "Alteração de email requer verificação por código enviado ao novo email",
		})
		return
	}

	if !updated {
		sendError(ctx, http.StatusBadRequest, "Nenhum campo válido foi fornecido para atualização")
		return
	}

	// Salva alterações
	if err := db.Save(&user).Error; err != nil {
		logger.ErrorF("error updating profile: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao atualizar perfil. Por favor, tente novamente")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Perfil atualizado com sucesso",
		"user":    user.ToResponse(),
	})
}

// @Summary Request email change
// @Description Send verification code to new email address
// @Tags 👤 User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body VerifyEmailRequest true "New email address"
// @Success 200 {object} map[string]interface{} "Verification code sent"
// @Failure 400 {object} ErrorResponse "Dados inválidos | Email já em uso"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Erro ao enviar código"
// @Router /user/request-email-change [post]
func RequestEmailChangeHandler(ctx *gin.Context) {
	var request VerifyEmailRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados inválidos: email é obrigatório e deve ser válido")
		return
	}

	// Pega usuário do contexto
	userInterface, _ := ctx.Get("user")
	user := userInterface.(schemas.User)

	// Verifica se o novo email já existe
	var existingUser schemas.User
	if err := db.Where("email = ? AND id != ?", request.NewEmail, user.ID).First(&existingUser).Error; err == nil {
		sendError(ctx, http.StatusBadRequest, "Este email já está em uso por outra conta")
		return
	}

	// Gera código de verificação
	code, err := GenerateRandomCode(6)
	if err != nil {
		logger.ErrorF("error generating verification code: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao gerar código de verificação")
		return
	}

	// Armazena verificação (em produção, use banco de dados)
	emailVerifications[user.ID] = &EmailVerification{
		UserID:    user.ID,
		NewEmail:  request.NewEmail,
		Token:     code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	// Envia email
	emailService := config.NewEmailService()
	if !emailService.IsConfigured() {
		logger.ErrorF("email service not configured")
		sendError(ctx, http.StatusInternalServerError, "Serviço de email não configurado")
		return
	}

	if err := emailService.SendEmailVerificationCode(request.NewEmail, user.Name, code); err != nil {
		logger.ErrorF("error sending email: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao enviar email de verificação")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Código de verificação enviado para o novo email. Válido por 15 minutos.",
	})
}

// @Summary Confirm email change
// @Description Confirm email change with verification code
// @Tags 👤 User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ConfirmEmailRequest true "New email and verification code"
// @Success 200 {object} map[string]interface{} "Email updated successfully"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 401 {object} ErrorResponse "Código inválido ou expirado | Unauthorized"
// @Failure 500 {object} ErrorResponse "Erro ao atualizar email"
// @Router /user/confirm-email-change [post]
func ConfirmEmailChangeHandler(ctx *gin.Context) {
	var request ConfirmEmailRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.ErrorF("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, "Dados inválidos: email e código (6 dígitos) são obrigatórios")
		return
	}

	// Pega usuário do contexto
	userInterface, _ := ctx.Get("user")
	user := userInterface.(schemas.User)

	// Verifica se existe verificação pendente
	verification, exists := emailVerifications[user.ID]
	if !exists {
		sendError(ctx, http.StatusUnauthorized, "Nenhuma verificação de email pendente. Solicite um novo código")
		return
	}

	// Valida código
	if verification.Used {
		sendError(ctx, http.StatusUnauthorized, "Código já utilizado. Solicite um novo código")
		return
	}

	if time.Now().After(verification.ExpiresAt) {
		sendError(ctx, http.StatusUnauthorized, "Código expirado. Solicite um novo código")
		return
	}

	if verification.Token != request.Token {
		sendError(ctx, http.StatusUnauthorized, "Código inválido")
		return
	}

	if verification.NewEmail != request.NewEmail {
		sendError(ctx, http.StatusBadRequest, "Email não corresponde ao da verificação")
		return
	}

	// Verifica novamente se o email não foi usado por outra conta
	var existingUser schemas.User
	if err := db.Where("email = ? AND id != ?", request.NewEmail, user.ID).First(&existingUser).Error; err == nil {
		sendError(ctx, http.StatusBadRequest, "Este email já está em uso por outra conta")
		return
	}

	// Atualiza email
	user.Email = request.NewEmail
	if err := db.Save(&user).Error; err != nil {
		logger.ErrorF("error updating email: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "Erro ao atualizar email. Por favor, tente novamente")
		return
	}

	// Marca verificação como usada
	verification.Used = true

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Email atualizado com sucesso!",
		"user":    user.ToResponse(),
	})
}
