package api

import (
	"lawyerSL-Backend/dao"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ─────────────────────────────────────────────────────────────
// GET /api/notifications?userId=...&lastId=...&limit=20
//
// Cursor-based (WhatsApp-style) pagination:
//   - First page: omit lastId
//   - Next page:  pass the _id of the oldest item from the
//     previous page as lastId
// ─────────────────────────────────────────────────────────────
func GetNotifications(c *fiber.Ctx) error {
	userID := c.Query("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required query parameter: userId",
		})
	}

	lastID := c.Query("lastId", "")

	limit := int64(20)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	notifications, err := dao.DB_GetNotificationsByUserID(userID, lastID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve notifications: " + err.Error(),
		})
	}

	// Provide the nextCursor for the client (the _id of the oldest item returned)
	var nextCursor string
	if len(notifications) == int(limit) {
		nextCursor = notifications[len(notifications)-1].ID.Hex()
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"notifications": notifications,
		"nextCursor":    nextCursor, // empty string means no more pages
	})
}

// ─────────────────────────────────────────────────────────────
// PATCH /api/notifications/:notificationId/read
// ─────────────────────────────────────────────────────────────
func MarkNotificationRead(c *fiber.Ctx) error {
	notificationID := c.Params("notificationId")
	if notificationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing notificationId param",
		})
	}

	if err := dao.DB_MarkNotificationRead(notificationID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Notification marked as read",
	})
}

// ─────────────────────────────────────────────────────────────
// PATCH /api/notifications/read-all?userId=...
// ─────────────────────────────────────────────────────────────
func MarkAllNotificationsRead(c *fiber.Ctx) error {
	userID := c.Query("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required query parameter: userId",
		})
	}

	if err := dao.DB_MarkAllNotificationsRead(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark all notifications as read: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "All notifications marked as read",
	})
}
