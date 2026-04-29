package functions

import (
	"bytes"
	"fmt"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// GenerateTopSellingPDF generates a clean analytics PDF for top-selling medicines.
func GenerateTopSellingPDF(items []dto.TopSellingItem, query dto.AnalyticsQuery) ([]byte, error) {
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
	pdf.CellFormat(pageW, 12, "Top Selling Medicines", "0", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(160, 200, 230)
	pdf.SetX(15)
	pdf.CellFormat(pageW, 7, "Analytics Report  |  PharmacyOS", "0", 1, "C", false, 0, "")

	// ── Meta Info ────────────────────────────────────────────────────────────
	pdf.SetXY(15, 42)

	fromStr := "All Time"
	toStr := time.Now().Format("02 Jan 2006")
	if query.From != "" {
		if t, err := time.Parse("2006-01-02", query.From); err == nil {
			fromStr = t.Format("02 Jan 2006")
		} else {
			fromStr = query.From
		}
	}
	if query.To != "" {
		if t, err := time.Parse("2006-01-02", query.To); err == nil {
			toStr = t.Format("02 Jan 2006")
		} else {
			toStr = query.To
		}
	}

	branchLabel := "All Branches"
	if query.BranchId != "" {
		branchLabel = query.BranchId
	}

	addMeta := func(label, value string) {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(80, 80, 110)
		pdf.CellFormat(30, 6, label, "0", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(20, 20, 50)
		pdf.CellFormat(60, 6, value, "0", 1, "L", false, 0, "")
		pdf.SetX(15)
	}

	addMeta("Period:", fmt.Sprintf("%s  –  %s", fromStr, toStr))
	addMeta("Branch:", branchLabel)
	addMeta("Showing:", fmt.Sprintf("Top %d medicines", query.Limit))

	pdf.Ln(4)

	// ── Divider ───────────────────────────────────────────────────────────────
	pdf.SetDrawColor(0, 184, 169)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)

	// ── Table Header ──────────────────────────────────────────────────────────
	colW := []float64{14, 100, 32, 34}
	headers := []string{"#", "Medicine Name", "Units Sold", "Revenue (LKR)"}
	aligns := []string{"C", "L", "C", "R"}

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
	medalColors := [][3]int{
		{255, 215, 0},
		{192, 192, 192},
		{205, 127, 50},
	}

	var totalQty int
	var totalRevenue float64
	for _, item := range items {
		totalQty += item.TotalQtySold
		totalRevenue += item.TotalRevenue
	}

	for idx, item := range items {
		rowY := pdf.GetY()

		if idx%2 == 0 {
			pdf.SetFillColor(245, 250, 255)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.Rect(15, rowY, pageW, 8, "F")

		// Rank
		if idx < 3 {
			mc := medalColors[idx]
			pdf.SetFillColor(mc[0], mc[1], mc[2])
			pdf.Rect(15, rowY, colW[0], 8, "F")
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetTextColor(40, 40, 40)
		} else {
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(80, 80, 80)
		}
		pdf.SetXY(15, rowY)
		pdf.CellFormat(colW[0], 8, fmt.Sprintf("%d", idx+1), "0", 0, "C", false, 0, "")

		// Medicine name
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(20, 20, 50)
		pdf.CellFormat(colW[1], 8, item.MedicineName, "0", 0, "L", false, 0, "")

		// Units sold
		pdf.SetTextColor(50, 50, 80)
		pdf.CellFormat(colW[2], 8, fmt.Sprintf("%d", item.TotalQtySold), "0", 0, "C", false, 0, "")

		// Revenue
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(15, 100, 80)
		pdf.CellFormat(colW[3], 8, fmt.Sprintf("%.2f", item.TotalRevenue), "0", 0, "R", false, 0, "")

		pdf.Ln(8)

		pdf.SetDrawColor(220, 225, 235)
		pdf.SetLineWidth(0.2)
		pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	}

	pdf.Ln(2)

	// ── Totals Row ────────────────────────────────────────────────────────────
	totY := pdf.GetY()
	pdf.SetFillColor(15, 32, 75)
	pdf.Rect(15, totY, pageW, 10, "F")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(15, totY)
	pdf.CellFormat(colW[0]+colW[1], 10, "  TOTALS", "0", 0, "L", false, 0, "")
	pdf.CellFormat(colW[2], 10, fmt.Sprintf("%d", totalQty), "0", 0, "C", false, 0, "")
	pdf.CellFormat(colW[3], 10, fmt.Sprintf("%.2f", totalRevenue), "0", 1, "R", false, 0, "")

	// ── Footer ────────────────────────────────────────────────────────────────
	pdf.Ln(6)
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