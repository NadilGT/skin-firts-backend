package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetApprovals returns paginated approval records.
// Supports filtering by referenceType (PO|TRANSFER|REJECT), referenceId, status.
//
// GET /approvals?referenceType=&referenceId=&status=&page=&limit=
func GetApprovals(c *fiber.Ctx) error {
	var query dto.SearchApprovalQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	approvals, total, err := dao.DB_SearchApprovals(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch approvals: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": approvals,
		"pagination": fiber.Map{
			"page": query.Page, "limit": query.Limit, "total": total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

// ApproveRecord approves a pending approval record by its MongoDB ObjectID.
// The approver's UID is extracted from the JWT.
//
// PATCH /approvals/:id/approve
// Body (optional): { "notes": "..." }
func ApproveRecord(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid approval ID"})
	}
	var body dto.ApprovalActionRequest
	_ = c.BodyParser(&body)

	approvedBy, _ := c.Locals("uid").(string)
	if err := dao.DB_UpdateApprovalStatus(objectID, dto.ApprovalApproved, approvedBy, body.Notes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to approve: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Record approved successfully"})
}

// RejectRecord rejects a pending approval record by its MongoDB ObjectID.
//
// PATCH /approvals/:id/reject
// Body (optional): { "notes": "reason for rejection" }
func RejectRecord(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid approval ID"})
	}
	var body dto.ApprovalActionRequest
	_ = c.BodyParser(&body)

	approvedBy, _ := c.Locals("uid").(string)
	if err := dao.DB_UpdateApprovalStatus(objectID, dto.ApprovalRejected, approvedBy, body.Notes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reject: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Record rejected"})
}
