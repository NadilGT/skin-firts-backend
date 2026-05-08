package api

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/functions"
	"lawyerSL-Backend/utils"
	"log"
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

	// ── Auto-fill UnitCost from SupplierMedicinePrice catalogue ─────────────
	var total float64
	for i := range po.Items {
		priceRecord, err := dao.DB_GetSupplierMedicinePriceBySupplierAndMedicine(
			po.SupplierId,
			po.Items[i].MedicineID,
		)
		if err != nil || priceRecord == nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": fmt.Sprintf(
					"No active price configured for supplier '%s' and medicine '%s' (%s). "+
						"Please add a price via POST /supplier-medicine-price first.",
					po.SupplierId,
					po.Items[i].MedicineName,
					po.Items[i].MedicineID,
				),
			})
		}
		po.Items[i].UnitCost = priceRecord.UnitPrice
		po.Items[i].TotalCost = float64(po.Items[i].Quantity) * po.Items[i].UnitCost
		total += po.Items[i].TotalCost
	}
	po.TotalAmount = total
	// ─────────────────────────────────────────────────────────────────────────


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

	// ── Fire-and-forget: generate PDF and email it to the supplier ──────────
	go func(po dto.PurchaseOrderModel) {
		// 1. Generate the PDF
		pdfBytes, err := functions.GeneratePurchaseOrderPDF(po)
		if err != nil {
			log.Printf("[PO EMAIL] Failed to generate PDF for %s: %v", po.PoId, err)
			return
		}

		// 2. Look up supplier email
		supplier, err := dao.DB_GetSupplierByID(po.SupplierId)
		if err != nil || supplier == nil {
			log.Printf("[PO EMAIL] Could not fetch supplier %s: %v", po.SupplierId, err)
			return
		}
		if supplier.Email == "" {
			log.Printf("[PO EMAIL] Supplier %s has no email — skipping", po.SupplierId)
			return
		}

		// 3. Build email body
		subject := fmt.Sprintf("Purchase Order %s — Skin First Medical Center", po.PoId)
		body := fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:600px;margin:auto">
  <div style="background:#0F6270;padding:20px;text-align:center">
    <h2 style="color:#fff;margin:0">Skin First Medical Center</h2>
    <p style="color:#CCF0F6;margin:4px 0 0">Purchase Order Notification</p>
  </div>
  <div style="padding:24px;border:1px solid #e0e0e0">
    <p>Dear <strong>%s</strong>,</p>
    <p>Please find attached the Purchase Order <strong>%s</strong> raised on <strong>%s</strong>.</p>
    <table style="width:100%%;border-collapse:collapse;margin:16px 0">
      <tr><td style="padding:6px;font-weight:bold;color:#555">PO ID:</td><td style="padding:6px">%s</td></tr>
      <tr style="background:#f5fcfd"><td style="padding:6px;font-weight:bold;color:#555">Status:</td><td style="padding:6px">%s</td></tr>
      <tr><td style="padding:6px;font-weight:bold;color:#555">Total Amount:</td><td style="padding:6px">Rs. %.2f</td></tr>
    </table>
    <p style="color:#555">Kindly acknowledge receipt of this order at your earliest convenience.</p>
    <p style="color:#888;font-size:12px;margin-top:32px">This is an automated email from Skin First Medical Center. Please do not reply directly to this message.</p>
  </div>
</div>`,
			supplier.Name,
			po.PoId,
			po.CreatedAt.Format("02 Jan 2006"),
			po.PoId,
			po.Status,
			po.TotalAmount,
		)

		// 4. Send email with PDF attached
		filename := fmt.Sprintf("%s.pdf", po.PoId)
		if err := utils.SendEmailWithAttachment([]string{supplier.Email}, subject, body, pdfBytes, filename); err != nil {
			log.Printf("[PO EMAIL] Failed to send email for %s to %s: %v", po.PoId, supplier.Email, err)
			return
		}
		log.Printf("[PO EMAIL] ✅ PO %s emailed to supplier %s (%s)", po.PoId, supplier.Name, supplier.Email)
	}(po)
	// ───────────────────────────────────────────────────────────────────────

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Purchase order created successfully",
		"data":    po,
	})
}

func GetPurchaseOrders(c *fiber.Ctx) error {
	var query dto.SearchPOQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}

	branchId, err := ResolveBranchId(c, query.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	query.BranchId = branchId

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

func GetPurchaseOrdersByStatus(c *fiber.Ctx) error {
	branchIdInput := c.Query("branchId")
	status := c.Query("status")

	branchId, err := ResolveBranchId(c, branchIdInput)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	query := dto.SearchPOQuery{
		BranchId: branchId,
		Status:   status,
		Page:     1,
		Limit:    1000, // Large limit for filtered list
	}

	pos, total, err := dao.DB_SearchPurchaseOrders(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch purchase orders"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  pos,
		"total": total,
	})
}

func GetPurchaseOrderByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing purchase order ID"})
	}

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	po, err := dao.DB_GetPurchaseOrderByID(id, branchId)
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

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := dao.DB_UpdatePOStatus(id, branchId, req, approvedBy); err != nil {
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
	})
}

func GetGRNs(c *fiber.Ctx) error {
	var query dto.SearchGRNQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}

	branchId, err := ResolveBranchId(c, query.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	query.BranchId = branchId

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
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing GRN ID"})
	}

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	grn, err := dao.DB_GetGRNByID(id, branchId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "GRN not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": grn})
}
