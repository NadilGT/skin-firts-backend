package api

import (
	"fmt"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/functions"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetTopSellingMedicinesPDF(c *fiber.Ctx) error {
	var query dto.AnalyticsQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}
 
	branchId, err := ResolveBranchId(c, query.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	query.BranchId = branchId
 
	if query.Limit == 0 {
		query.Limit = 10
	}
 
	// 1. Fetch data
	items, err := dao.DB_GetTopSellingMedicines(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to get top-selling medicines: " + err.Error()},
		)
	}
 
	// 2. Generate PDF
	pdfBytes, err := functions.GenerateTopSellingPDF(items, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to generate PDF: " + err.Error()},
		)
	}
 
	// 3. Stream as a downloadable PDF
	filename := fmt.Sprintf("top_selling_%s.pdf", time.Now().Format("20060102_1504"))
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	return c.Status(fiber.StatusOK).Send(pdfBytes)
}

func GetSalesReport(c *fiber.Ctx) error {
	var query dto.AnalyticsQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}

	branchId, err := ResolveBranchId(c, query.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	query.BranchId = branchId

	if query.Period == "" {
		query.Period = "daily"
	}

	// 1. Fetch data
	items, err := dao.DB_GetSalesReport(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to get sales report: " + err.Error()},
		)
	}

	// 2. Generate PDF
	pdfBytes, err := functions.GenerateSalesReportPDF(items, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to generate PDF: " + err.Error()},
		)
	}

	// 3. Stream as downloadable PDF
	filename := fmt.Sprintf("sales_report_%s.pdf", time.Now().Format("20060102_1504"))
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	return c.Status(fiber.StatusOK).Send(pdfBytes)
}

func GetProfitMarginReport(c *fiber.Ctx) error {
	var query dto.AnalyticsQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}

	branchId, err := ResolveBranchId(c, query.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	query.BranchId = branchId

	// 1. Fetch data
	items, err := dao.DB_GetProfitMarginReport(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to get profit margin report: " + err.Error()},
		)
	}

	// 2. Generate PDF
	pdfBytes, err := functions.GenerateProfitMarginPDF(items, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to generate PDF: " + err.Error()},
		)
	}

	// 3. Stream as downloadable PDF
	filename := fmt.Sprintf("profit_margin_%s.pdf", time.Now().Format("20060102_1504"))
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	return c.Status(fiber.StatusOK).Send(pdfBytes)
}

func GetExpiryReport(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	days := c.QueryInt("days", 90)

	// 1. Fetch data
	alerts, err := dao.DB_GetExpiryAlerts(branchId, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to get expiry report: " + err.Error()},
		)
	}

	// 2. Generate PDF
	pdfBytes, err := functions.GenerateExpiryReportPDF(alerts, branchId, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to generate PDF: " + err.Error()},
		)
	}

	// 3. Stream as downloadable PDF
	filename := fmt.Sprintf("expiry_report_%s.pdf", time.Now().Format("20060102_1504"))
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	return c.Status(fiber.StatusOK).Send(pdfBytes)
}

func GetStockReportAnalytics(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// 1. Fetch data
	report, err := dao.DB_GetStockReport(branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to get stock report: " + err.Error()},
		)
	}

	// 2. Generate PDF
	pdfBytes, err := functions.GenerateStockReportPDF(report, branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to generate PDF: " + err.Error()},
		)
	}

	// 3. Stream as downloadable PDF
	filename := fmt.Sprintf("stock_report_%s.pdf", time.Now().Format("20060102_1504"))
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	return c.Status(fiber.StatusOK).Send(pdfBytes)
}
