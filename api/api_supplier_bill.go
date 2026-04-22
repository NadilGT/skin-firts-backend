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

// CreateSupplierBill creates a new supplier invoice record.
// Standard flow: PO → GRN → SupplierBill
//
// POST /supplier-bills
// Body: { supplierId, supplierName, purchaseOrderId?, grnId?, branchId?, items[], totalAmount, dueDate? }
func CreateSupplierBill(c *fiber.Ctx) error {
	var bill dto.SupplierBillModel
	if err := c.BodyParser(&bill); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if bill.SupplierId == "" || len(bill.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "SupplierId and at least one item are required"})
	}

	// Auto-inject branchId from JWT
	if branchId, ok := c.Locals("effectiveBranchId").(string); ok && branchId != "" {
		bill.BranchId = branchId
	}
	bill.CreatedBy, _ = c.Locals("uid").(string)

	// Compute totals from items
	var computedTotal float64
	for i := range bill.Items {
		bill.Items[i].TotalCost = float64(bill.Items[i].Quantity) * bill.Items[i].UnitCost
		computedTotal += bill.Items[i].TotalCost
	}
	if bill.TotalAmount <= 0 {
		bill.TotalAmount = computedTotal
	}
	bill.DueAmount = bill.TotalAmount - bill.PaidAmount

	if bill.PaymentStatus == "" {
		switch {
		case bill.PaidAmount >= bill.TotalAmount:
			bill.PaymentStatus = "PAID"
		case bill.PaidAmount > 0:
			bill.PaymentStatus = "PARTIAL"
		default:
			bill.PaymentStatus = "UNPAID"
		}
	}

	billId, err := dao.GenerateId(context.Background(), "supplier_bills", "SBL")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	bill.BillId = billId
	bill.CreatedAt = time.Now()
	bill.UpdatedAt = time.Now()

	if err := dao.DB_CreateSupplierBill(bill); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create supplier bill: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Supplier bill created successfully", "data": bill})
}

// GetSupplierBills returns paginated supplier bills with optional filters.
//
// GET /supplier-bills?supplierId=&branchId=&paymentStatus=&from=&to=&page=&limit=
func GetSupplierBills(c *fiber.Ctx) error {
	var query dto.SearchSupplierBillQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query"})
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	bills, total, err := dao.DB_SearchSupplierBills(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch supplier bills: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": bills,
		"pagination": fiber.Map{
			"page": query.Page, "limit": query.Limit, "total": total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

// GetSupplierBillByID returns a single supplier bill by its MongoDB ObjectID.
//
// GET /supplier-bills/:id
func GetSupplierBillByID(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid supplier bill ID"})
	}
	bill, err := dao.DB_GetSupplierBillByID(objectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier bill not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": bill})
}

// UpdateSupplierBillPayment records a payment against a supplier bill.
// The payment is accumulated (cumulative), so paidAmount in the request is the
// new payment amount — not the total paid to date.
//
// PATCH /supplier-bills/:id/payment
// Body: { paidAmount, paymentMethod, notes }
func UpdateSupplierBillPayment(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid supplier bill ID"})
	}
	var req dto.UpdateSupplierBillPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.PaidAmount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "PaidAmount must be greater than 0"})
	}
	if err := dao.DB_UpdateSupplierBillPayment(objectID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update payment: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Payment recorded successfully"})
}
