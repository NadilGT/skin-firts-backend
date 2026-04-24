package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

// ──────────────────────────────────────────────
//  Payment Management
// ──────────────────────────────────────────────

func GetPharmacyBills(c *fiber.Ctx) error {
	var query dto.SearchBillQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	bills, total, err := dao.DB_SearchPharmacyBills(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch bills: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": bills,
		"pagination": fiber.Map{
			"page": query.Page, "limit": query.Limit, "total": total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

func GetPharmacyBillByID(c *fiber.Ctx) error {
	billId := c.Params("billId")
	if billId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing billId"})
	}

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	bill, err := dao.DB_GetBillByBillId(billId, branchId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Bill not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": bill})
}

func UpdatePharmacyBillPayment(c *fiber.Ctx) error {
	billId := c.Params("billId")
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if billId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing billId"})
	}
	var req dto.UpdateBillPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if req.PaidAmount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "PaidAmount must be greater than 0"})
	}
	if err := dao.DB_UpdateBillPayment(billId, branchId, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update payment: " + err.Error()})
	}
	// Return updated bill
	bill, _ := dao.DB_GetBillByBillId(billId, branchId)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Payment updated successfully", "data": bill})
}

func GetDailySalesSummary(c *fiber.Ctx) error {
	branchId := c.Query("branchId")
	date := c.Query("date") // YYYY-MM-DD, optional — defaults to today
	summary, err := dao.DB_GetDailySalesSummary(branchId, date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get daily summary: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": summary})
}

func GetRevenueSummary(c *fiber.Ctx) error {
	branchId := c.Query("branchId")
	from := c.Query("from")
	to := c.Query("to")
	result, err := dao.DB_GetRevenueSummary(branchId, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get revenue summary: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
}

func GetPendingPayments(c *fiber.Ctx) error {
	branchId := c.Query("branchId")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	bills, total, err := dao.DB_GetPendingPayments(branchId, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get pending payments"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  bills,
		"total": total,
	})
}
