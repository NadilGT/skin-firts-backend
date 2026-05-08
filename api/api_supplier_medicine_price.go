package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// ──────────────────────────────────────────────
//
//	Supplier Medicine Price Handlers
//
// ──────────────────────────────────────────────

// CreateSupplierMedicinePrice godoc
// POST /supplier-medicine-price
func CreateSupplierMedicinePrice(c *fiber.Ctx) error {
	var p dto.SupplierMedicinePrice
	if err := c.BodyParser(&p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}

	// Validation
	if p.SupplierId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "supplierId is required"})
	}
	if p.MedicineId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "medicineId is required"})
	}
	if p.UnitPrice <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unitPrice must be greater than 0"})
	}

	// Verify supplier exists
	if _, err := dao.DB_GetSupplierByID(p.SupplierId); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Supplier not found: " + p.SupplierId})
	}

	priceId, err := dao.GenerateId(context.Background(), "supplier_medicine_prices", "SMP")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate price ID: "+err.Error())
	}
	p.PriceId = priceId
	p.IsActive = true
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	if err := dao.DB_CreateSupplierMedicinePrice(p); err != nil {
		if dao.IsDuplicateKeyError(err) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "A price entry for this supplier + medicine combination already exists. Use PUT to update it.",
			})
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create price: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Supplier medicine price created successfully",
		"data":    p,
	})
}

// GetSupplierMedicinePrices godoc
// GET /supplier-medicine-price?supplierId=&medicineId=&isActive=
func GetSupplierMedicinePrices(c *fiber.Ctx) error {
	var query dto.SearchSupplierMedicinePriceQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}

	prices, err := dao.DB_GetSupplierMedicinePrices(query)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch prices")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  prices,
		"total": len(prices),
	})
}

// GetSupplierMedicinePriceByID godoc
// GET /supplier-medicine-price/:id
func GetSupplierMedicinePriceByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id param is required"})
	}

	p, err := dao.DB_GetSupplierMedicinePriceByID(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Price record not found"})
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": p})
}

// UpdateSupplierMedicinePrice godoc
// PUT /supplier-medicine-price/:id
func UpdateSupplierMedicinePrice(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id param is required"})
	}

	var req dto.UpdateSupplierMedicinePriceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}

	if req.UnitPrice < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unitPrice cannot be negative"})
	}

	if err := dao.DB_UpdateSupplierMedicinePrice(id, req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update price: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Supplier medicine price updated successfully"})
}

// DeleteSupplierMedicinePrice godoc
// DELETE /supplier-medicine-price/:id  (soft delete — sets isActive = false)
func DeleteSupplierMedicinePrice(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id param is required"})
	}

	if err := dao.DB_DeleteSupplierMedicinePrice(id); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to deactivate price: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Supplier medicine price deactivated (soft deleted)"})
}
