package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/functions"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// GenerateBillPDF godoc
// GET /billing/pdf?billId=BIL-001
func GenerateBillPDF(c *fiber.Ctx) error {
	billId := c.Query("billId")
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if billId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "billId query param is required"})
	}

	// 1. Fetch the bill
	bill, err := dao.DB_GetBillByBillId(billId, branchId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Bill not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch bill"})
	}

	// 2. Collect unique medicine IDs from bill items
	seen := map[string]struct{}{}
	var medIDs []string
	for _, item := range bill.Items {
		if _, ok := seen[item.MedicineID]; !ok {
			seen[item.MedicineID] = struct{}{}
			medIDs = append(medIDs, item.MedicineID)
		}
	}

	// 3. Resolve medicine names (single batch DB query)
	medicineNames, err := dao.DB_GetMedicineNamesByIDs(medIDs)
	if err != nil {
		// Non-fatal: PDF will fall back to medicineId strings
		medicineNames = map[string]string{}
	}

	// 4. Generate PDF
	pdfBytes, err := functions.GenerateBillPDF(*bill, medicineNames)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate PDF: " + err.Error()})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "inline; filename=\""+billId+".pdf\"")
	return c.Status(fiber.StatusOK).Send(pdfBytes)
}
