package functions

import (
	"bytes"
	"fmt"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// GeneratePurchaseOrderPDF creates a styled PDF for the given PurchaseOrderModel
// and returns its raw bytes.
func GeneratePurchaseOrderPDF(po dto.PurchaseOrderModel) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pageW := 180.0 // usable width (210 - 15*2)

	// ─────────────────────────────────────────────
	// HEADER BAND  (deep teal background)
	// ─────────────────────────────────────────────
	pdf.SetFillColor(15, 98, 112)
	pdf.SetDrawColor(15, 98, 112)
	pdf.Rect(15, 15, pageW, 28, "F")

	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(15, 18)
	pdf.CellFormat(pageW, 9, "SKIN FIRST MEDICAL CENTER", "0", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(204, 236, 242)
	pdf.SetX(15)
	pdf.CellFormat(pageW, 5, "123 Health Street, Medical District, City", "0", 1, "C", false, 0, "")
	pdf.SetX(15)
	pdf.CellFormat(pageW, 5, "Tel: +1 (555) 123-4567   |   info@skinfirst.com", "0", 1, "C", false, 0, "")

	pdf.Ln(6)

	// ─────────────────────────────────────────────
	// "PURCHASE ORDER" label strip
	// ─────────────────────────────────────────────
	pdf.SetFillColor(232, 248, 251)
	pdf.SetDrawColor(15, 98, 112)
	pdf.Rect(15, pdf.GetY(), pageW, 10, "FD")

	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(15, 98, 112)
	pdf.SetX(15)
	pdf.CellFormat(pageW, 10, "PURCHASE ORDER", "0", 1, "C", false, 0, "")

	pdf.Ln(6)

	// ─────────────────────────────────────────────
	// PO META  (two-column layout)
	// ─────────────────────────────────────────────
	leftLabel := func(txt string) {
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(80, 80, 80)
		pdf.Cell(38, 6, txt)
	}
	leftValue := func(txt string) {
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(30, 30, 30)
		pdf.Cell(52, 6, txt)
	}
	rightLabel := func(txt string) {
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(80, 80, 80)
		pdf.Cell(38, 6, txt)
	}
	rightValue := func(txt string) {
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(30, 30, 30)
		pdf.Cell(52, 6, txt)
	}

	// Row 1 — PO ID / Status
	leftLabel("PO ID:")
	leftValue(po.PoId)
	rightLabel("Status:")
	// Colour-code the status text
	switch po.Status {
	case "APPROVED", "RECEIVED":
		pdf.SetTextColor(0, 140, 0)
	case "CANCELLED":
		pdf.SetTextColor(200, 0, 0)
	case "DRAFT":
		pdf.SetTextColor(130, 130, 0)
	default:
		pdf.SetTextColor(30, 30, 30)
	}
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(52, 6, po.Status)
	pdf.SetTextColor(30, 30, 30)
	pdf.Ln(6)

	// Row 2 — Date / Expected Date
	leftLabel("Date:")
	leftValue(po.CreatedAt.Format(time.RFC822))
	rightLabel("Expected Date:")
	expDate := "—"
	if !po.ExpectedDate.IsZero() {
		expDate = po.ExpectedDate.Format("02 Jan 2006")
	}
	rightValue(expDate)
	pdf.Ln(8)

	// Thin separator
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(15, pdf.GetY(), 15+pageW, pdf.GetY())
	pdf.Ln(4)

	// Row 3 — Supplier / Branch
	leftLabel("Supplier Name:")
	leftValue(po.SupplierName)
	rightLabel("Branch ID:")
	rightValue(po.BranchId)
	pdf.Ln(6)

	leftLabel("Supplier ID:")
	leftValue(po.SupplierId)
	rightLabel("Created By:")
	rightValue(po.CreatedBy)
	pdf.Ln(6)

	if po.Notes != "" {
		leftLabel("Notes:")
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(30, 30, 30)
		pdf.Cell(pageW-38, 6, po.Notes)
		pdf.Ln(6)
	}

	pdf.Ln(4)

	// ─────────────────────────────────────────────
	// ITEMS TABLE
	// ─────────────────────────────────────────────
	colW := [5]float64{8, 82, 20, 35, 35}
	headers := [5]string{"#", "Medicine Name", "Qty", "Unit Cost", "Total Cost"}
	aligns := [5]string{"C", "L", "C", "R", "R"}

	// Header row
	pdf.SetFillColor(15, 98, 112)
	pdf.SetDrawColor(15, 98, 112)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	for i, h := range headers {
		pdf.CellFormat(colW[i], 9, h, "1", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(-1)

	// Item rows with alternating shading
	pdf.SetFont("Arial", "", 9)
	for idx, item := range po.Items {
		if idx%2 == 0 {
			pdf.SetFillColor(245, 252, 253)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(30, 30, 30)
		pdf.SetDrawColor(180, 210, 215)

		pdf.CellFormat(colW[0], 8, fmt.Sprintf("%d", idx+1), "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW[1], 8, item.MedicineName, "1", 0, "L", true, 0, "")
		pdf.CellFormat(colW[2], 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW[3], 8, fmt.Sprintf("Rs. %.2f", item.UnitCost), "1", 0, "R", true, 0, "")
		pdf.CellFormat(colW[4], 8, fmt.Sprintf("Rs. %.2f", item.TotalCost), "1", 1, "R", true, 0, "")
	}

	// ─────────────────────────────────────────────
	// GRAND TOTAL
	// ─────────────────────────────────────────────
	pdf.Ln(4)

	labelW := colW[0] + colW[1] + colW[2] + colW[3]
	valueW := colW[4]

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(labelW, 7, "Grand Total:", "0", 0, "R", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(valueW, 7, fmt.Sprintf("Rs. %.2f", po.TotalAmount), "0", 1, "R", false, 0, "")

	pdf.Ln(2)

	// Grand total highlighted box
	totalBoxX := 15 + labelW
	pdf.SetFillColor(15, 98, 112)
	pdf.SetDrawColor(15, 98, 112)
	pdf.Rect(totalBoxX, pdf.GetY(), valueW, 10, "F")

	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(15, 98, 112)
	pdf.SetX(15)
	pdf.CellFormat(labelW, 10, "Amount Due:", "0", 0, "R", false, 0, "")

	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(valueW, 10, fmt.Sprintf("Rs. %.2f", po.TotalAmount), "0", 1, "R", false, 0, "")

	// ─────────────────────────────────────────────
	// FOOTER
	// ─────────────────────────────────────────────
	pdf.Ln(20)
	pdf.SetDrawColor(15, 98, 112)
	pdf.Line(15, pdf.GetY(), 15+pageW, pdf.GetY())
	pdf.Ln(4)

	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(pageW, 5, "This is a computer-generated Purchase Order and does not require a physical signature.", "0", 1, "C", false, 0, "")
	pdf.CellFormat(pageW, 5, "Thank you for your partnership with Skin First Medical Center.", "0", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PO PDF: %w", err)
	}
	return buf.Bytes(), nil
}
