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

// CreateStockAdjustment creates an adjustment request (PENDING)
func CreateStockAdjustment(c *fiber.Ctx) error {
	var r dto.StockAdjustmentModel
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if r.BatchId == "" || r.Quantity <= 0 || r.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "BatchId, Reason, and quantity > 0 are required"})
	}
	if r.Type != "ADJUSTMENT_IN" && r.Type != "ADJUSTMENT_OUT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Type must be ADJUSTMENT_IN or ADJUSTMENT_OUT"})
	}

	branchId, err := ResolveBranchId(c, r.BranchId)
	if err != nil {
		return err
	}
	r.BranchId = branchId
	r.CreatedBy, _ = c.Locals("uid").(string)

	adjId, err := dao.GenerateId(context.Background(), "stock_adjustments", "ADJ")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	r.AdjustmentId = adjId
	r.Status = "PENDING"
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()

	if err := dao.DB_CreateStockAdjustment(r); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create adjustment request: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Adjustment created (PENDING)", 
		"data": r,
		"effectiveBranchId": r.BranchId,
	})
}

// GetStockAdjustments returns paginated adjustments
func GetStockAdjustments(c *fiber.Ctx) error {
	var query dto.SearchAdjustmentQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}

	records, total, err := dao.DB_SearchStockAdjustments(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": records,
		"pagination": fiber.Map{
			"total": total,
		},
	})
}

// ApproveStockAdjustment moves request from PENDING -> APPROVED
func ApproveStockAdjustment(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	var body struct {
		Notes string `json:"notes"`
	}
	_ = c.BodyParser(&body)
	approvedBy, _ := c.Locals("uid").(string)

	if err := dao.DB_ApproveStockAdjustment(objectID, approvedBy, body.Notes); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Approval failed: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stock adjustment approved"})
}

// ExecuteStockAdjustment formally enacts the adjustment and changes batch quantity
func ExecuteStockAdjustment(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	executedBy, _ := c.Locals("uid").(string)

	if err := dao.DB_ExecuteStockAdjustment(objectID, executedBy); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Execution failed: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stock adjustment executed successfully"})
}
