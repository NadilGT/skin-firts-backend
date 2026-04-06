package utils

import (
	"bytes"
	"fmt"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// GenerateHospitalBillPDF creates a PDF for the hospital bill and returns its raw bytes.
func GenerateHospitalBillPDF(bill *dto.HospitalBillModel) ([]byte, error) {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Hospital Header
	pdf.CellFormat(190, 10, "SKIN FIRST MEDICAL CENTER", "0", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(190, 5, "123 Health Street, Medical District, City", "0", 1, "C", false, 0, "")
	pdf.CellFormat(190, 5, "Contact: +1 (555) 123-4567 | Email: info@skinfirst.com", "0", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Bill Details Section
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "HOSPITAL BILL", "0", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Bill ID:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(150, 7, bill.HospitalBillId)
	pdf.Ln(-1)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(150, 7, bill.CreatedAt.Format(time.RFC822))
	pdf.Ln(10)

	// Patient & Doctor Info
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Patient Name:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(55, 7, bill.PatientName)
	
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Doctor Name:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(55, 7, bill.DoctorName)
	pdf.Ln(-1)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Patient ID:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(55, 7, bill.PatientID)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Doctor ID:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(55, 7, bill.DoctorID)
	pdf.Ln(15)

	// Service Breakdown Table Header
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(90, 10, "Service Description", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Unit Price", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 10, "Qty", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, "Total", "1", 1, "C", false, 0, "")

	// Table Content
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(90, 10, bill.ServiceName, "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("%.2f", bill.UnitPrice), "1", 0, "R", false, 0, "")
	pdf.CellFormat(20, 10, fmt.Sprintf("%d", bill.Quantity), "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("%.2f", bill.TotalAmount), "1", 1, "R", false, 0, "")

	// Total Amount
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(140, 10, "Grand Total:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("Rs. %.2f", bill.TotalAmount), "0", 1, "R", false, 0, "")

	// Footer
	pdf.Ln(30)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(190, 5, "This is a computer generated bill and does not require a physical signature.", "0", 1, "C", false, 0, "")
	pdf.CellFormat(190, 5, "Thank you for choosing Skin First Medical Center.", "0", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF buffer: %w", err)
	}

	return buf.Bytes(), nil
}
