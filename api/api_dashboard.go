package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

// ──────────────────────────────────────────────
//  GET /analytics/appointments
// ──────────────────────────────────────────────
//
// Query params:
//   branchId  string  required
//   days      int     optional, default = 7
//
// Response: []{ "date": "YYYY-MM-DD", "count": 12 }

func GetAppointmentsAnalytics(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if branchId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "branchId is required"})
	}

	days := c.QueryInt("days", 7)
	if days <= 0 {
		days = 7
	}

	data, err := dao.DB_GetAppointmentsTimeSeries(branchId, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to fetch appointments analytics: " + err.Error()},
		)
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

// ──────────────────────────────────────────────
//  GET /analytics/revenue
// ──────────────────────────────────────────────
//
// Query params:
//   branchId  string  required
//   days      int     optional, default = 7
//
// Response: []{ "date": "YYYY-MM-DD", "totalRevenue": 1200.00 }
// Only counts completed payments (pharmacy paymentStatus=PAID + hospital confirm=true).

func GetRevenueAnalytics(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if branchId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "branchId is required"})
	}

	days := c.QueryInt("days", 7)
	if days <= 0 {
		days = 7
	}

	data, err := dao.DB_GetRevenueTimeSeries(branchId, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to fetch revenue analytics: " + err.Error()},
		)
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

// ──────────────────────────────────────────────
//  GET /analytics/summary
// ──────────────────────────────────────────────
//
// Query params:
//   branchId  string  required
//   days      int     optional, default = 7
//
// Response:
//   {
//     "totalAppointments": 120,
//     "totalRevenue":      8400.00,
//     "growthRate":        12.5       // % change vs previous equal period
//   }

func GetDashboardSummary(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if branchId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "branchId is required"})
	}

	days := c.QueryInt("days", 7)
	if days <= 0 {
		days = 7
	}

	summary, err := dao.DB_GetDashboardSummary(branchId, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to fetch dashboard summary: " + err.Error()},
		)
	}
	return c.Status(fiber.StatusOK).JSON(summary)
}
