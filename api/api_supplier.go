package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ──────────────────────────────────────────────
//  Supplier
// ──────────────────────────────────────────────

func CreateSupplier(c *fiber.Ctx) error {
	var supplier dto.SupplierModel
	if err := c.BodyParser(&supplier); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if supplier.Name == "" || supplier.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Supplier name and phone are required"})
	}
	id, err := dao.GenerateId(context.Background(), "suppliers", "SUP")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	supplier.SupplierId = id
	supplier.CreatedAt = time.Now()
	supplier.UpdatedAt = time.Now()
	if supplier.Status == "" {
		supplier.Status = "ACTIVE"
	}
	if err := dao.DB_CreateSupplier(supplier); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create supplier: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Supplier created successfully", "data": supplier})
}

func GetSuppliers(c *fiber.Ctx) error {
	var query dto.SearchSupplierQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	suppliers, total, err := dao.DB_SearchSuppliers(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch suppliers"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": suppliers,
		"pagination": fiber.Map{
			"page": query.Page, "limit": query.Limit, "total": total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

func GetSupplierByID(c *fiber.Ctx) error {
	id := c.Query("id")
	supplier, err := dao.DB_GetSupplierByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": supplier})
}

func UpdateSupplier(c *fiber.Ctx) error {
	id := c.Query("id")
	var supplier dto.SupplierModel
	if err := c.BodyParser(&supplier); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := dao.DB_UpdateSupplier(id, supplier); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update supplier"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Supplier updated successfully"})
}

func DeleteSupplier(c *fiber.Ctx) error {
	id := c.Query("id")
	if err := dao.DB_DeleteSupplier(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete supplier"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Supplier deleted successfully"})
}

// ──────────────────────────────────────────────
//  Purchase Orders
// ──────────────────────────────────────────────

func CreatePurchaseOrder(c *fiber.Ctx) error {
	var po dto.PurchaseOrderModel
	if err := c.BodyParser(&po); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if po.SupplierId == "" || len(po.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "SupplierId and at least one item are required"})
	}
	branchId, err := ResolveBranchId(c, po.BranchId)
	if err != nil {
		return err
	}
	po.BranchId = branchId
	// Calculate total
	var total float64
	for i := range po.Items {
		po.Items[i].TotalCost = float64(po.Items[i].Quantity) * po.Items[i].UnitCost
		total += po.Items[i].TotalCost
	}
	po.TotalAmount = total

	id, err := dao.GenerateId(context.Background(), "purchase_orders", "PO")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	po.PoId = id
	po.CreatedAt = time.Now()
	po.UpdatedAt = time.Now()
	if po.Status == "" {
		po.Status = "DRAFT"
	}
	if err := dao.DB_CreatePurchaseOrder(po); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create purchase order: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Purchase order created successfully", 
		"data": po,
		"effectiveBranchId": po.BranchId,
	})
}

func GetPurchaseOrders(c *fiber.Ctx) error {
	var query dto.SearchPOQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	pos, total, err := dao.DB_SearchPurchaseOrders(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch purchase orders"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": pos,
		"pagination": fiber.Map{
			"page": query.Page, "limit": query.Limit, "total": total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

func GetPurchaseOrderByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing purchase order ID"})
	}
	po, err := dao.DB_GetPurchaseOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Purchase order not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": po})
}

func UpdatePurchaseOrderStatus(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing purchase order ID"})
	}
	var req dto.UpdatePOStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Status is required"})
	}
	// Extract caller identity for approval audit trail
	approvedBy, _ := c.Locals("uid").(string)
	if err := dao.DB_UpdatePOStatus(id, req, approvedBy); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update PO status: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Purchase order status updated to " + req.Status})
}

// ──────────────────────────────────────────────
//  GRN
// ──────────────────────────────────────────────

func CreateGRN(c *fiber.Ctx) error {
	var grn dto.GRNModel
	if err := c.BodyParser(&grn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if grn.SupplierId == "" || len(grn.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "SupplierId and at least one item are required"})
	}
	branchId, err := ResolveBranchId(c, grn.BranchId)
	if err != nil {
		return err
	}
	grn.BranchId = branchId
	id, err := dao.GenerateId(context.Background(), "grn", "GRN")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	grn.GrnId = id
	if grn.ReceivedDate.IsZero() {
		grn.ReceivedDate = time.Now()
	}
	grn.CreatedAt = time.Now()

	// This also auto-creates medicine batches for each item
	if err := dao.DB_CreateGRN(grn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create GRN: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "GRN created and stock updated successfully",
		"data":    grn,
		"effectiveBranchId": grn.BranchId,
	})
}

func GetGRNs(c *fiber.Ctx) error {
	var query dto.SearchGRNQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	grns, total, err := dao.DB_SearchGRNs(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch GRNs"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": grns,
		"pagination": fiber.Map{
			"page": query.Page, "limit": query.Limit, "total": total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

func GetGRNByID(c *fiber.Ctx) error {
	id := c.Query("id")
	grn, err := dao.DB_GetGRNByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "GRN not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": grn})
}
