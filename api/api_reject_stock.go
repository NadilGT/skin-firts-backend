package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateRejectStock creates a new reject stock request in PENDING status.
//
// POST /reject-stock
// Body: { batchId, medicineId, branchId, type, quantity, reason, notes }
func CreateRejectStock(c *fiber.Ctx) error {
	var r dto.RejectStockModel
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if r.BatchId == "" || r.MedicineId == "" || r.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "BatchId, medicineId, and quantity > 0 are required"})
	}
	if r.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Type is required (EXPIRED | DAMAGED | RETURN_TO_SUPPLIER)"})
	}

	if err := EnforceBranchId(&r.BranchId, c); err != nil {
		return err
	}
	createdBy, _ := c.Locals("uid").(string)

	rejectId, err := dao.GenerateId(context.Background(), "reject_stock", "REJ")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	r.RejectId = rejectId
	r.Status = "PENDING"
	r.CreatedBy = createdBy
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()

	if err := dao.DB_CreateRejectStock(r); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create reject stock: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Reject stock request created (PENDING)", 
		"data": r,
		"effectiveBranchId": r.BranchId,
	})
}

// GetRejectStocks returns paginated reject stock records with optional filters.
//
// GET /reject-stock?branchId=&status=&type=&from=&to=&page=&limit=
func GetRejectStocks(c *fiber.Ctx) error {
	var query dto.SearchRejectQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	records, total, err := dao.DB_SearchRejectStock(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch reject stock: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": records,
		"pagination": fiber.Map{
			"page": query.Page, "limit": query.Limit, "total": total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

// GetRejectStockByID returns a single reject stock record.
//
// GET /reject-stock/:id
func GetRejectStockByID(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid reject stock ID"})
	}
	r, err := dao.DB_GetRejectStockByID(objectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reject stock record not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": r})
}

// ApproveRejectStock transitions a reject stock request from PENDING → APPROVED.
// The approver's identity is extracted from the JWT.
//
// PATCH /reject-stock/:id/approve
// Body (optional): { "notes": "..." }
func ApproveRejectStock(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid reject stock ID"})
	}
	var body struct {
		Notes string `json:"notes"`
	}
	_ = c.BodyParser(&body)
	approvedBy, _ := c.Locals("uid").(string)

	if err := dao.DB_ApproveRejectStock(objectID, approvedBy, body.Notes); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to approve: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Reject stock approved"})
}

// ExecuteRejectStock transitions an APPROVED reject stock to COMPLETED:
// deducts the batch quantity and writes a REJECT StockMovement.
//
// PATCH /reject-stock/:id/execute
func ExecuteRejectStock(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid reject stock ID"})
	}
	executedBy, _ := c.Locals("uid").(string)

	if err := dao.DB_ExecuteRejectStock(objectID, executedBy); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to execute reject: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Reject stock executed — stock deducted and movement recorded"})
}
