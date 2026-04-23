package api

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateMedicine(c *fiber.Ctx) error {
	var medicine dto.MedicineModel
	if err := c.BodyParser(&medicine); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validation
	if medicine.Name == "" || medicine.Manufacturer == "" || medicine.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, manufacturer, and category are required",
		})
	}

	id, err := dao.GenerateId(context.Background(), "medicines", "MED")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	medicine.MedicineId = id
	medicine.CreatedAt = time.Now()
	medicine.UpdatedAt = time.Now()
	if medicine.Status == "" {
		medicine.Status = "Active"
	}

	if err := dao.DB_CreateMedicine(medicine); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create medicine: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Medicine created successfully",
		"data":    medicine,
	})
}

func SearchMedicines(c *fiber.Ctx) error {
	var query dto.SearchMedicineQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query parameters",
		})
	}

	// Set default pagination
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	medicines, total, err := dao.DB_SearchMedicines(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search medicines",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": medicines,
		"pagination": fiber.Map{
			"page":       query.Page,
			"limit":      query.Limit,
			"total":      total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

func GetMedicineByID(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid medicine ID",
		})
	}

	medicine, err := dao.DB_GetMedicineByID(objectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Medicine not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": medicine,
	})
}

func UpdateMedicine(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid medicine ID",
		})
	}

	var medicine dto.MedicineModel
	if err := c.BodyParser(&medicine); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	medicine.ID = objectID
	medicine.UpdatedAt = time.Now()

	if err := dao.DB_UpdateMedicine(medicine); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update medicine",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Medicine updated successfully",
		"data":    medicine,
	})
}

func DeleteMedicine(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid medicine ID",
		})
	}

	if err := dao.DB_DeleteMedicine(objectID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete medicine",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Medicine deleted successfully",
	})
}

func GetLowStockMedicines(c *fiber.Ctx) error {
	medicines, err := dao.DB_GetLowStockMedicines()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch low stock medicines",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": medicines,
	})
}

func GetMedicineByBarcode(c *fiber.Ctx) error {
	barcode := c.Query("barcode")
	if barcode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing barcode query parameter"})
	}
	medicine, err := dao.DB_GetMedicineByBarcode(barcode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Medicine not found for barcode: " + barcode})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": medicine})
}

// Debug endpoint to check what data is being sent
func DebugMedicineRequest(c *fiber.Ctx) error {
	body := c.Body()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"receivedBody": string(body),
		"contentType":  c.Get("Content-Type"),
	})
}

// ========== BATCH APIS ==========

func CreateMedicineBatch(c *fiber.Ctx) error {
	var batch dto.MedicineBatch
	if err := c.BodyParser(&batch); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate MedicineBatchId
	id, err := dao.GenerateId(context.Background(), "medicine_batches", "BAT")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	batch.BatchId = id

	batch.ID = primitive.NewObjectID()
	batch.CreatedAt = time.Now()

	if batch.Status == "" {
		batch.Status = "ACTIVE"
	}

	if err := dao.DB_CreateMedicineBatch(batch); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create medicine batch: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Medicine batch created successfully",
		"data":    batch,
	})
}

func GetBatchesByMedicineID(c *fiber.Ctx) error {
	id := c.Query("medicineId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing medicineId format",
		})
	}

	batches, err := dao.DB_GetBatchesByMedicineID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve batches",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": batches,
	})
}

func GetAvailableBatchesFEFO(c *fiber.Ctx) error {
	id := c.Query("medicineId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing medicineId format",
		})
	}

	branchId := c.Query("branchId")
	if branchId == "" {
		if val, ok := c.Locals("effectiveBranchId").(string); ok {
			branchId = val
		}
	}

	batches, err := dao.DB_GetAvailableBatchesFEFO(id, branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve available batches",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": batches,
	})
}

func GetActiveStockByMedicineID(c *fiber.Ctx) error {
	id := c.Query("medicineId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing medicineId in query",
		})
	}

	branchId := c.Query("branchId")
	if branchId == "" {
		if val, ok := c.Locals("effectiveBranchId").(string); ok {
			branchId = val
		}
	}

	totalStock, err := dao.DB_GetActiveStockByMedicineID(id, branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve stock",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"medicineId":  id,
		"activeStock": totalStock,
	})
}

func DeductStockFEFO(c *fiber.Ctx) error {
	var req dto.DeductStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Quantity must be greater than zero",
		})
	}

	if req.MedicineID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing medicine ID",
		})
	}

	branchId, _ := c.Locals("effectiveBranchId").(string)
	billItems, err := dao.DB_DeductStockFEFO(req.MedicineID, req.Quantity, "", branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to deduct stock: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Stock deducted successfully",
		"billItems": billItems,
	})
}
func CreateBill(c *fiber.Ctx) error {
	var req dto.CreateBillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	additionalChargesStr := c.Query("additionalCharges", "0")
	additionalCharges, _ := strconv.ParseFloat(additionalChargesStr, 64)
	patientID := c.Query("patientId")

	branchId, err := ResolveBranchId(c, req.BranchId)
	if err != nil {
		return err
	}
	req.BranchId = branchId

	var allBillItems []dto.BillItem
	var totalMedicinePrice float64

	for _, item := range req.Items {
		// Use DB_ReserveStockFEFO to apply a hard soft-lock (reservedQuantity)
		billItems, err := dao.DB_ReserveStockFEFO(item.MedicineID, branchId, item.Quantity)
		if err != nil {
			// If we fail on item N, we MUST rollback reservations for items 1 to N-1
			if len(allBillItems) > 0 {
				dao.DB_RevertStockReservation(allBillItems)
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to prepare bill for medicine %s: %s", item.MedicineID, err.Error()),
			})
		}
		for _, bi := range billItems {
			allBillItems = append(allBillItems, bi)
			totalMedicinePrice += bi.Price * float64(bi.Quantity)
		}
	}

	// Calculate totals
	grandTotal := totalMedicinePrice + additionalCharges
	discount := req.Discount
	tax := req.Tax
	netTotal := grandTotal - discount + tax
	if netTotal < 0 {
		netTotal = 0
	}

	// Payment calculations
	paidAmount := req.PaidAmount
	if paidAmount < 0 {
		paidAmount = 0
	}
	dueAmount := netTotal - paidAmount
	if dueAmount < 0 {
		dueAmount = 0
	}
	paymentStatus := "PENDING"
	if paidAmount >= netTotal {
		paymentStatus = "PAID"
	} else if paidAmount > 0 {
		paymentStatus = "PARTIAL"
	}

	billId, err := dao.GenerateId(context.Background(), "bills", "BIL")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate bill ID: "+err.Error())
	}

	bill := dto.BillModel{
		ID:                 primitive.NewObjectID(),
		BillId:             billId,
		PatientID:          patientID,
		Items:              allBillItems,
		TotalMedicinePrice: totalMedicinePrice,
		AdditionalCharges:  additionalCharges,
		GrandTotal:         grandTotal,
		Discount:           discount,
		Tax:                tax,
		NetTotal:           netTotal,
		PaidAmount:         paidAmount,
		DueAmount:          dueAmount,
		PaymentStatus:      paymentStatus,
		PaymentMethod:      req.PaymentMethod,
		CustomerName:       req.CustomerName,
		CustomerPhone:      req.CustomerPhone,
		BranchId:           req.BranchId,
		Notes:              req.Notes,
		CreatedBy:          req.CreatedBy,
		Status:             "PENDING",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := dao.DB_CreateBill(bill); err != nil {
		dao.DB_RevertStockReservation(allBillItems)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save pending bill: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Bill created successfully (Pending Confirmation)",
		"data":    bill,
		"effectiveBranchId": bill.BranchId,
	})
}

// CancelBill manually cancels a pending bill and releases its reserved stock immediately.
func CancelBill(c *fiber.Ctx) error {
	billId := c.Query("billId")
	bill, err := dao.DB_GetBillByBillId(billId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Bill not found"})
	}
	if bill.Status != "PENDING" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Only PENDING bills can be cancelled"})
	}

	// Release stock reservations
	dao.DB_RevertStockReservation(bill.Items)

	// Update bill status to CANCELLED
	if err := dao.DB_UpdateBillStatus(billId, "CANCELLED"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to cancel bill"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Bill cancelled and stock reservations released"})
}

func ConfirmBill(c *fiber.Ctx) error {
	billId := c.Query("billId")

	bill, err := dao.DB_GetBillByBillId(billId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Bill not found",
		})
	}

	if bill.Status != "PENDING" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bill is already " + bill.Status,
		})
	}

	var successfullyDeducted []dto.BillItem

	// Deduct the exact batches matched during CreateBill
	for _, item := range bill.Items {
		_, err = dao.DB_DeductFromBatchAtomic(item.BatchID, bill.BranchId, item.Quantity)
		if err != nil {
			// Someone bought this exact batch between creation and confirmation
			dao.DB_RevertStockDeduction(successfullyDeducted)
			dao.DB_UpdateBillStatus(billId, "FAILED")
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to confirm bill: Stock for medicine %s is no longer available. Please create a new bill.", item.MedicineID),
			})
		}

		successfullyDeducted = append(successfullyDeducted, item)

		// Write SALE StockMovement for this confirmed deduction
		_ = dao.DB_WriteSaleMovement(item, billId, bill.BranchId)
	}

	// All stock deducted successfully, mark bill as confirmed
	err = dao.DB_UpdateBillStatus(billId, "CONFIRMED")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update bill status, but stock was deducted",
		})
	}

	bill.Status = "CONFIRMED"
	bill.UpdatedAt = time.Now()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Bill confirmed and stock deducted successfully",
		"data":    bill,
	})
}

