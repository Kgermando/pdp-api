package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/models"
	"github.com/kgermando/pdp-api/utils"
)

// GetNotifications godoc
// GET /api/v1/notifications
func GetNotifications(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	p := utils.GetPagination(c)

	var total int64
	database.DB.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total)

	var notifications []models.Notification
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(p.PerPage).
		Offset(p.Offset).
		Find(&notifications)

	return utils.SuccessWithMeta(c, notifications, utils.BuildMeta(total, p), "Notifications récupérées")
}

// MarkNotificationRead godoc
// PATCH /api/v1/notifications/:id/read
func MarkNotificationRead(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	result := database.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)

	if result.RowsAffected == 0 {
		return utils.NotFound(c, "Notification non trouvée")
	}
	return utils.Success(c, nil, "Notification marquée comme lue")
}

// MarkAllNotificationsRead godoc
// PATCH /api/v1/notifications/read-all
func MarkAllNotificationsRead(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true)
	return utils.Success(c, nil, "Toutes les notifications marquées comme lues")
}

// GetUnreadCount godoc
// GET /api/v1/notifications/unread-count
func GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	var count int64
	database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count)
	return utils.Success(c, fiber.Map{"count": count}, "Nombre de notifications non lues")
}
