package api

import (
	"context"
	"encoding/base64"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateHospitalBill generates a bill for a service and creates a PDF.
func CreateHospitalBill(c *fiber.Ctx) error {
	var req dto.CreateHospitalBillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one service item is required",
		})
	}

	// 1. Fetch Service Details for all items and calculate total
	var totalAmount float64
	var billItems []dto.HospitalBillItem

	for _, item := range req.Items {
		if item.ServiceID == "" || item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Valid serviceId and quantity (>0) are required for all items",
			})
		}

		service, err := dao.DB_GetServiceByServiceId(item.ServiceID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Service not found: " + item.ServiceID,
			})
		}

		itemTotal := service.UnitPrice * float64(item.Quantity)
		totalAmount += itemTotal

		billItems = append(billItems, dto.HospitalBillItem{
			ServiceID:   service.ServiceID,
			ServiceName: service.Name,
			Quantity:    item.Quantity,
			UnitPrice:   service.UnitPrice,
			Total:       itemTotal,
		})
	}

	// 2. Generate a unique Hospital Bill ID
	hospitalBillId, err := dao.GenerateId(context.Background(), "hospital_bills", "HB")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate hospital bill ID: "+err.Error())
	}

	// Format Fallback Names
	pName := req.PatientName
	if pName == "" && req.PatientID != "" {
		pName = "ID: " + req.PatientID
	} else if pName == "" {
		pName = "Walk-in Patient"
	}

	dName := req.DoctorName
	if dName == "" && req.DoctorID != "" {
		dName = "ID: " + req.DoctorID
	} else if dName == "" {
		dName = "N/A"
	}

	// 3. Prepare the Bill Document
	bill := dto.HospitalBillModel{
		ID:             primitive.NewObjectID(),
		HospitalBillId: hospitalBillId,
		PatientID:      req.PatientID,
		PatientName:    pName,
		DoctorID:       req.DoctorID,
		DoctorName:     dName,
		Items:          billItems,
		TotalAmount:    totalAmount,
		Confirm:        false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 4. Generate the PDF buffer
	pdfBytes, err := utils.GenerateHospitalBillPDF(&bill)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate PDF: " + err.Error(),
		})
	}
	
	// Convert raw PDF bytes to Base64
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)

	// 5. Save to Database
	if err := dao.DB_CreateHospitalBill(&bill); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save hospital bill: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Hospital bill generated successfully",
		"data":      bill,
		"pdfBase64": pdfBase64,
	})
}

// DownloadHospitalBillPDF generates the PDF on the fly and streams it back.
func DownloadHospitalBillPDF(c *fiber.Ctx) error {
	hospitalBillId := c.Params("id")
	if hospitalBillId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bill ID is required",
		})
	}

	bill, err := dao.DB_GetHospitalBill(hospitalBillId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Hospital bill not found",
		})
	}

	pdfBytes, err := utils.GenerateHospitalBillPDF(bill)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate PDF: " + err.Error(),
		})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=\""+hospitalBillId+".pdf\"")
	return c.Status(fiber.StatusOK).Send(pdfBytes)
}

// ConfirmHospitalBill marks a hospital bill as confirmed.
func ConfirmHospitalBill(c *fiber.Ctx) error {
	hospitalBillId := c.Query("id")
	if hospitalBillId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bill ID is required",
		})
	}

	bill, err := dao.DB_GetHospitalBill(hospitalBillId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Hospital bill not found",
		})
	}

	if bill.Confirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bill is already confirmed",
		})
	}

	err = dao.DB_ConfirmHospitalBill(hospitalBillId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to confirm bill: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Hospital bill confirmed successfully",
	})
}
