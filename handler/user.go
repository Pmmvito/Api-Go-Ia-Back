package handler

import (
	"net/http"

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
