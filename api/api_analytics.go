package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

func GetTopSellingMedicines(c *fiber.Ctx) error {
	var query dto.AnalyticsQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}
	if query.Limit == 0 {
		query.Limit = 10
	}
	items, err := dao.DB_GetTopSellingMedicines(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get top-selling medicines: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": items})
}

func GetSalesReport(c *fiber.Ctx) error {
	var query dto.AnalyticsQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}
	if query.Period == "" {
		query.Period = "daily"
	}
	items, err := dao.DB_GetSalesReport(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get sales report: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": items, "period": query.Period})
}

func GetProfitMarginReport(c *fiber.Ctx) error {
	var query dto.AnalyticsQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}
	items, err := dao.DB_GetProfitMarginReport(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get profit margin report: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": items})
}

func GetExpiryReport(c *fiber.Ctx) error {
	branchId := c.Query("branchId")
	days := c.QueryInt("days", 90)
	alerts, err := dao.DB_GetExpiryAlerts(branchId, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get expiry report: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": alerts, "count": len(alerts), "days": days})
}

func GetStockReportAnalytics(c *fiber.Ctx) error {
	branchId := c.Query("branchId")
	report, err := dao.DB_GetStockReport(branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stock report: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": report, "count": len(report)})
}
