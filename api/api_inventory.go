package api

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ──────────────────────────────────────────────
//  Stock Valuation
// ──────────────────────────────────────────────

func GetStockValuation(c *fiber.Ctx) error {
	branchId := c.Query("branchId")
	result, err := dao.DB_GetStockValuation(branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stock valuation: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
}

// ──────────────────────────────────────────────
//  Expiry Alerts
// ──────────────────────────────────────────────

func GetExpiryAlerts(c *fiber.Ctx) error {
	branchId := c.Query("branchId")
	days := c.QueryInt("days", 90)
	alerts, err := dao.DB_GetExpiryAlerts(branchId, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get expiry alerts: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  alerts,
		"count": len(alerts),
		"days":  days,
	})
}

// ──────────────────────────────────────────────
//  Stock Report
// ──────────────────────────────────────────────

func GetInventoryStockReport(c *fiber.Ctx) error {
	branchId := c.Query("branchId")
	report, err := dao.DB_GetStockReport(branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stock report: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": report, "count": len(report)})
}

// ──────────────────────────────────────────────
//  Stock Transfer
// ──────────────────────────────────────────────

func CreateStockTransfer(c *fiber.Ctx) error {
	var transfer dto.StockTransferModel
	if err := c.BodyParser(&transfer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if transfer.FromBranchId == "" || transfer.ToBranchId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "FromBranchId and ToBranchId are required"})
	}
	if len(transfer.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one item is required"})
	}

	// ── Reserve Stock First ──
	var successfullyReserved []dto.TransferItem
	for _, item := range transfer.Items {
		if err := dao.DB_ReserveSpecificStock(item.StockId, transfer.FromBranchId, item.Quantity); err != nil {
			// Revert all previously reserved items in this request
			dao.DB_RevertTransferStockReservation(successfullyReserved, transfer.FromBranchId)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to reserve stock for %s: %v", item.MedicineName, err),
			})
		}
		successfullyReserved = append(successfullyReserved, item)
	}

	id, err := dao.GenerateId(context.Background(), "stock_transfers", "TRF")
	if err != nil {
		dao.DB_RevertTransferStockReservation(successfullyReserved, transfer.FromBranchId)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	transfer.TransferId = id
	transfer.Status = "PENDING"
	transfer.CreatedAt = time.Now()
	transfer.UpdatedAt = time.Now()

	if err := dao.DB_CreateStockTransfer(transfer); err != nil {
		dao.DB_RevertTransferStockReservation(successfullyReserved, transfer.FromBranchId)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create stock transfer: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Stock transfer created (PENDING)", "data": transfer})
}

func GetStockTransfers(c *fiber.Ctx) error {
	var query dto.SearchTransferQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	transfers, total, err := dao.DB_SearchStockTransfers(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch stock transfers"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": transfers,
		"pagination": fiber.Map{
			"page": query.Page, "limit": query.Limit, "total": total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

func GetStockTransferByTransferID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing transfer ID"})
	}
	transfer, err := dao.DB_GetStockTransferByTransferID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Stock transfer not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": transfer})
}

func CompleteStockTransfer(c *fiber.Ctx) error {
	id := c.Query("id")
	if err := dao.DB_CompleteStockTransfer(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to complete transfer: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stock transfer completed successfully"})
}

func CancelStockTransfer(c *fiber.Ctx) error {
	id := c.Query("id")
	if err := dao.DB_CancelStockTransfer(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to cancel transfer"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stock transfer cancelled"})
}

// ApproveStockTransfer transitions a transfer from PENDING → APPROVED.
// Must be called before CompleteStockTransfer.
//
// PATCH /stock-transfers/:id/approve
func ApproveStockTransfer(c *fiber.Ctx) error {
	id := c.Query("id")
	approvedBy, _ := c.Locals("uid").(string)
	if err := dao.DB_ApproveStockTransfer(id, approvedBy); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to approve transfer: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stock transfer approved — can now be completed"})
}

