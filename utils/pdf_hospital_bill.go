package utils

import (
	"bytes"
	"fmt"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// color helpers
func setColor(pdf *gofpdf.Fpdf, r, g, b int) {
	pdf.SetTextColor(r, g, b)
}
func setFill(pdf *gofpdf.Fpdf, r, g, b int) {
	pdf.SetFillColor(r, g, b)
}
func setDraw(pdf *gofpdf.Fpdf, r, g, b int) {
	pdf.SetDrawColor(r, g, b)
}

// GenerateHospitalBillPDF creates a PDF for the hospital bill and returns its raw bytes.
func GenerateHospitalBillPDF(bill *dto.HospitalBillModel) ([]byte, error) {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pageW := 180.0 // usable width (210 - 15*2)

	// ─────────────────────────────────────────────
	// HEADER BAND  (deep teal background)
	// ─────────────────────────────────────────────
	setFill(pdf, 15, 98, 112)  // #0F6270
	setDraw(pdf, 15, 98, 112)
	pdf.Rect(15, 15, pageW, 28, "F")

	// Hospital name
	pdf.SetFont("Arial", "B", 18)
	setColor(pdf, 255, 255, 255)
	pdf.SetXY(15, 18)
	pdf.CellFormat(pageW, 9, "SKIN FIRST MEDICAL CENTER", "0", 1, "C", false, 0, "")

	// Sub-line
	pdf.SetFont("Arial", "", 9)
	setColor(pdf, 204, 236, 242) // light teal
	pdf.SetX(15)
	pdf.CellFormat(pageW, 5, "123 Health Street, Medical District, City", "0", 1, "C", false, 0, "")
	pdf.SetX(15)
	pdf.CellFormat(pageW, 5, "Tel: +1 (555) 123-4567   |   info@skinfirst.com", "0", 1, "C", false, 0, "")

	pdf.Ln(6)

	// ─────────────────────────────────────────────
	// "HOSPITAL BILL" label on a light accent strip
	// ─────────────────────────────────────────────
	setFill(pdf, 232, 248, 251) // very light teal
	setDraw(pdf, 15, 98, 112)
	pdf.Rect(15, pdf.GetY(), pageW, 10, "FD")

	pdf.SetFont("Arial", "B", 13)
	setColor(pdf, 15, 98, 112)
	pdf.SetX(15)
	pdf.CellFormat(pageW, 10, "HOSPITAL BILL", "0", 1, "C", false, 0, "")

	pdf.Ln(6)

	// ─────────────────────────────────────────────
	// BILL META  (two-column layout)
	// ─────────────────────────────────────────────
	leftLabel := func(txt string) {
		pdf.SetFont("Arial", "B", 9)
		setColor(pdf, 80, 80, 80)
		pdf.Cell(38, 6, txt)
	}
	leftValue := func(txt string) {
		pdf.SetFont("Arial", "", 9)
		setColor(pdf, 30, 30, 30)
		pdf.Cell(52, 6, txt)
	}
	rightLabel := func(txt string) {
		pdf.SetFont("Arial", "B", 9)
		setColor(pdf, 80, 80, 80)
		pdf.Cell(38, 6, txt)
	}
	rightValue := func(txt string) {
		pdf.SetFont("Arial", "", 9)
		setColor(pdf, 30, 30, 30)
		pdf.Cell(52, 6, txt)
	}

	// Row 1
	leftLabel("Bill ID:")
	leftValue(bill.HospitalBillId)
	pdf.Ln(6)

	// Row 2
	leftLabel("Date:")
	leftValue(bill.CreatedAt.Format(time.RFC822))
	pdf.Ln(8)

	// Thin separator
	setDraw(pdf, 200, 200, 200)
	pdf.Line(15, pdf.GetY(), 15+pageW, pdf.GetY())
	pdf.Ln(4)

	// Row 3 – Patient / Doctor
	leftLabel("Patient Name:")
	leftValue(bill.PatientName)
	rightLabel("Doctor Name:")
	rightValue(bill.DoctorName)
	pdf.Ln(6)

	leftLabel("Patient Phone:")
	leftValue(bill.PatientPhone)
	rightLabel("Visit Type:")
	rightValue(bill.VisitType)
	pdf.Ln(6)

	leftLabel("Patient Email:")
	leftValue(bill.PatientEmail)
	rightLabel("Visit Date:")
	rightValue(bill.VisitDate)
	pdf.Ln(6)

	leftLabel("Patient ID:")
	leftValue(bill.PatientID)
	rightLabel("Doctor ID:")
	rightValue(bill.DoctorID)
	pdf.Ln(8)

	// ─────────────────────────────────────────────
	// ITEMS TABLE
	// ─────────────────────────────────────────────
	colW := [4]float64{90, 30, 20, 40}

	// Table header
	setFill(pdf, 15, 98, 112)
	setDraw(pdf, 15, 98, 112)
	setColor(pdf, 255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	headers := [4]string{"Service Description", "Unit Price", "Qty", "Total"}
	aligns := [4]string{"L", "R", "C", "R"}
	for i, h := range headers {
		pdf.CellFormat(colW[i], 9, h, "1", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows with alternating shading
	pdf.SetFont("Arial", "", 9)
	for idx, item := range bill.Items {
		if idx%2 == 0 {
			setFill(pdf, 245, 252, 253)
		} else {
			setFill(pdf, 255, 255, 255)
		}
		setColor(pdf, 30, 30, 30)
		setDraw(pdf, 180, 210, 215)

		pdf.CellFormat(colW[0], 8, item.ServiceName, "1", 0, "L", true, 0, "")
		pdf.CellFormat(colW[1], 8, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", true, 0, "")
		pdf.CellFormat(colW[2], 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW[3], 8, fmt.Sprintf("%.2f", item.Total), "1", 1, "R", true, 0, "")
	}

	// ─────────────────────────────────────────────
	// TOTALS breakdown section
	// ─────────────────────────────────────────────
	pdf.Ln(4)
	subtotal := 0.0
	for _, item := range bill.Items {
		subtotal += item.Total
	}

	labelW := colW[0] + colW[1] + colW[2]
	valueW := colW[3]

	// Sub-calculated rows
	drawRow := func(label string, value float64, isNegative bool) {
		pdf.SetFont("Arial", "B", 9)
		setColor(pdf, 80, 80, 80)
		pdf.CellFormat(labelW, 7, label, "0", 0, "R", false, 0, "")
		
		pdf.SetFont("Arial", "", 9)
		setColor(pdf, 30, 30, 30)
		prefix := ""
		if isNegative {
			prefix = "-"
		}
		pdf.CellFormat(valueW, 7, fmt.Sprintf("%sRs. %.2f", prefix, value), "0", 1, "R", false, 0, "")
	}

	drawRow("Subtotal:", subtotal, false)
	if bill.Tax > 0 {
		drawRow("Tax:", bill.Tax, false)
	}
	if bill.Discount > 0 {
		drawRow("Discount:", bill.Discount, true)
	}

	pdf.Ln(2)

	// GRAND TOTAL box
	totalBoxX := 15 + labelW
	totalBoxW := valueW

	setFill(pdf, 15, 98, 112)
	setDraw(pdf, 15, 98, 112)
	pdf.Rect(totalBoxX, pdf.GetY(), totalBoxW, 10, "F")

	pdf.SetFont("Arial", "B", 10)
	setColor(pdf, 15, 98, 112)
	pdf.SetX(15)
	pdf.CellFormat(labelW, 10, "Grand Total:", "0", 0, "R", false, 0, "")

	setColor(pdf, 255, 255, 255)
	pdf.CellFormat(totalBoxW, 10, fmt.Sprintf("Rs. %.2f", bill.TotalAmount), "0", 1, "R", false, 0, "")

	// ─────────────────────────────────────────────
	// FOOTER
	// ─────────────────────────────────────────────
	pdf.Ln(20)
	setDraw(pdf, 15, 98, 112)
	pdf.Line(15, pdf.GetY(), 15+pageW, pdf.GetY())
	pdf.Ln(4)

	pdf.SetFont("Arial", "I", 8)
	setColor(pdf, 120, 120, 120)
	pdf.CellFormat(pageW, 5, "This is a computer generated bill and does not require a physical signature.", "0", 1, "C", false, 0, "")
	pdf.CellFormat(pageW, 5, "Thank you for choosing Skin First Medical Center.", "0", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF buffer: %w", err)
	}

	return buf.Bytes(), nil
}