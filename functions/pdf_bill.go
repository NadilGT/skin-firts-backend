package functions

import (
	"bytes"
	"fmt"
	"lawyerSL-Backend/dto"

	"github.com/jung-kurt/gofpdf"
)

// GenerateBillPDF generates a patient-facing PDF bill.
// medicineNames maps medicineId -> human-readable name.
func GenerateBillPDF(bill dto.BillModel, medicineNames map[string]string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// ── Header ──────────────────────────────────────────────────────────────
	pdf.SetFillColor(30, 50, 100)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(0, 14, "Pharmacy Bill", "0", 1, "C", true, 0, "")
	pdf.Ln(4)

	// ── Bill Meta ───────────────────────────────────────────────────────────
	pdf.SetTextColor(30, 30, 30)

	addMetaRow := func(label, value string) {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetX(15)
		pdf.CellFormat(45, 7, label, "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(0, 7, value, "", 1, "L", false, 0, "")
	}

	addMetaRow("Bill ID:", bill.BillId)
	addMetaRow("Date:", bill.CreatedAt.Format("02 Jan 2006   15:04"))

	// Status with colour
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetX(15)
	pdf.CellFormat(45, 7, "Status:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	switch bill.Status {
	case "CONFIRMED":
		pdf.SetTextColor(0, 150, 0)
	case "PENDING":
		pdf.SetTextColor(200, 130, 0)
	default:
		pdf.SetTextColor(200, 0, 0)
	}
	pdf.CellFormat(0, 7, bill.Status, "", 1, "L", false, 0, "")
	pdf.SetTextColor(30, 30, 30)

	pdf.Ln(4)

	// ── Section divider ──────────────────────────────────────────────────────
	pdf.SetDrawColor(30, 50, 100)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(4)

	// ── Items Table Header ───────────────────────────────────────────────────
	// Columns: #  |  Medicine Name  |  Qty  |  Unit Price  |  Subtotal
	colW := []float64{10, 95, 20, 30, 30}
	headers := []string{"#", "Medicine", "Qty", "Unit Price", "Subtotal"}
	aligns := []string{"C", "L", "C", "R", "R"}

	pdf.SetFillColor(30, 50, 100)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 10)
	for i, h := range headers {
		pdf.CellFormat(colW[i], 8, h, "1", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(-1)

	// ── Items Table Rows ─────────────────────────────────────────────────────
	// Rows with the same medicineId are consolidated (same batch = same line)
	type consolidatedItem struct {
		name     string
		qty      int
		price    float64
	}
	// Preserve order and merge duplicates
	var order []string
	seen := map[string]*consolidatedItem{}
	for _, item := range bill.Items {
		name, ok := medicineNames[item.MedicineID]
		if !ok || name == "" {
			name = item.MedicineID // fallback to the short custom ID
		}
		if _, exists := seen[item.MedicineID]; !exists {
			order = append(order, item.MedicineID)
			seen[item.MedicineID] = &consolidatedItem{name: name, qty: 0, price: item.Price}
		}
		seen[item.MedicineID].qty += item.Quantity
	}

	pdf.SetTextColor(30, 30, 30)
	pdf.SetFont("Helvetica", "", 9)

	for idx, medID := range order {
		ci := seen[medID]
		if idx%2 == 0 {
			pdf.SetFillColor(248, 250, 255)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		subtotal := ci.price * float64(ci.qty)
		row := []string{
			fmt.Sprintf("%d", idx+1),
			ci.name,
			fmt.Sprintf("%d", ci.qty),
			fmt.Sprintf("%.2f", ci.price),
			fmt.Sprintf("%.2f", subtotal),
		}
		for i, cell := range row {
			pdf.CellFormat(colW[i], 7, cell, "1", 0, aligns[i], true, 0, "")
		}
		pdf.Ln(-1)
	}

	pdf.Ln(5)

	// ── Totals Block ─────────────────────────────────────────────────────────
	labelX := 120.0
	valueW := 75.0
	lineH := 7.0

	totals := []struct {
		label string
		value float64
		bold  bool
	}{
		{"Medicine Subtotal:", bill.TotalMedicinePrice, false},
		{"Additional Charges:", bill.AdditionalCharges, false},
		{"Grand Total:", bill.GrandTotal, true},
	}

	for _, t := range totals {
		if t.bold {
			pdf.SetFont("Helvetica", "B", 11)
			pdf.SetFillColor(30, 50, 100)
			pdf.SetTextColor(255, 255, 255)
		} else {
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetFillColor(240, 244, 255)
			pdf.SetTextColor(30, 30, 30)
		}
		pdf.SetX(labelX)
		pdf.CellFormat(valueW/2, lineH, t.label, "1", 0, "L", true, 0, "")
		pdf.CellFormat(valueW/2, lineH, fmt.Sprintf("%.2f", t.value), "1", 1, "R", true, 0, "")
	}

	pdf.SetTextColor(30, 30, 30)
	pdf.Ln(8)

	// ── Footer ───────────────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(2)
	pdf.CellFormat(0, 5, "This is a computer-generated bill and does not require a signature.", "0", 1, "C", false, 0, "")

	// ── Render ───────────────────────────────────────────────────────────────
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to render PDF: %w", err)
	}
	return buf.Bytes(), nil
}
