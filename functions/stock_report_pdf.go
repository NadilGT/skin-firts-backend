package functions

import (
	"bytes"
	"fmt"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// GenerateStockReportPDF generates a styled PDF for the stock level report.
func GenerateStockReportPDF(items []dto.StockReportItem, branchId string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pageW := 180.0

	// ── Header ───────────────────────────────────────────────────────────────
	pdf.SetFillColor(15, 32, 75)
	pdf.Rect(0, 0, 210, 34, "F")

	pdf.SetFillColor(0, 184, 169)
	pdf.Rect(0, 32, 210, 3, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetXY(15, 7)
	pdf.CellFormat(pageW, 12, "Stock Level Report", "0", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(160, 200, 230)
	pdf.SetX(15)
	pdf.CellFormat(pageW, 7, "Analytics Report  |  PharmacyOS", "0", 1, "C", false, 0, "")

	// ── Meta Info ────────────────────────────────────────────────────────────
	pdf.SetXY(15, 42)

	branchLabel := "All Branches"
	if branchId != "" {
		branchLabel = branchId
	}

	var lowStockCount int
	for _, item := range items {
		if item.IsLowStock {
			lowStockCount++
		}
	}

	addMeta := func(label, value string) {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(80, 80, 110)
		pdf.CellFormat(35, 6, label, "0", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(20, 20, 50)
		pdf.CellFormat(60, 6, value, "0", 1, "L", false, 0, "")
		pdf.SetX(15)
	}

	addMeta("Generated:", time.Now().Format("02 Jan 2006"))
	addMeta("Branch:", branchLabel)
	addMeta("Total SKUs:", fmt.Sprintf("%d medicines", len(items)))
	addMeta("Low Stock Alerts:", fmt.Sprintf("%d items", lowStockCount))

	pdf.Ln(4)

	// ── Divider ───────────────────────────────────────────────────────────────
	pdf.SetDrawColor(0, 184, 169)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)

	// ── Table Header ──────────────────────────────────────────────────────────
	colW := []float64{70, 35, 25, 25, 25}
	headers := []string{"Medicine Name", "Category", "Total Qty", "Reorder Lvl", "Batches"}
	aligns := []string{"L", "L", "C", "C", "C"}

	pdf.SetFillColor(15, 32, 75)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 9)
	for i, h := range headers {
		pdf.CellFormat(colW[i], 9, h, "0", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFillColor(0, 184, 169)
	pdf.Rect(15, pdf.GetY(), pageW, 0.5, "F")
	pdf.Ln(0.5)

	// ── Table Rows ────────────────────────────────────────────────────────────
	for idx, item := range items {
		rowY := pdf.GetY()

		if item.IsLowStock {
			pdf.SetFillColor(255, 240, 240)
		} else if idx%2 == 0 {
			pdf.SetFillColor(245, 250, 255)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.Rect(15, rowY, pageW, 8, "F")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(20, 20, 50)
		pdf.SetXY(15, rowY)
		pdf.CellFormat(colW[0], 8, item.MedicineName, "0", 0, "L", false, 0, "")

		pdf.SetTextColor(60, 60, 100)
		pdf.CellFormat(colW[1], 8, item.Category, "0", 0, "L", false, 0, "")

		// Qty — red if low stock
		if item.IsLowStock {
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetTextColor(180, 40, 40)
		} else {
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(15, 100, 80)
		}
		pdf.CellFormat(colW[2], 8, fmt.Sprintf("%d", item.TotalQty), "0", 0, "C", false, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(80, 80, 80)
		pdf.CellFormat(colW[3], 8, fmt.Sprintf("%d", item.ReorderLevel), "0", 0, "C", false, 0, "")
		pdf.CellFormat(colW[4], 8, fmt.Sprintf("%d", item.TotalBatches), "0", 1, "C", false, 0, "")

		pdf.SetDrawColor(220, 225, 235)
		pdf.SetLineWidth(0.2)
		pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	}

	pdf.Ln(6)

	// ── Legend ────────────────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(180, 40, 40)
	pdf.SetX(15)
	pdf.CellFormat(pageW, 6, "Red rows indicate stock at or below the reorder level.", "0", 1, "L", false, 0, "")

	// ── Footer ────────────────────────────────────────────────────────────────
	pdf.Ln(4)
	footerY := pdf.GetY()
	pdf.SetFillColor(0, 184, 169)
	pdf.Rect(0, footerY, 210, 0.5, "F")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(140, 140, 160)
	pdf.SetX(15)
	pdf.CellFormat(pageW/2, 5, fmt.Sprintf("Generated: %s", time.Now().Format("02 Jan 2006  15:04:05")), "0", 0, "L", false, 0, "")
	pdf.CellFormat(pageW/2, 5, "PharmacyOS  |  Confidential", "0", 0, "R", false, 0, "")

	// ── Render ────────────────────────────────────────────────────────────────
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to render PDF: %w", err)
	}
	return buf.Bytes(), nil
}
