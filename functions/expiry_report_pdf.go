package functions

import (
	"bytes"
	"fmt"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// GenerateExpiryReportPDF generates a styled PDF for the expiry alert report.
func GenerateExpiryReportPDF(items []dto.ExpiryAlertItem, branchId string, days int) ([]byte, error) {
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
	pdf.CellFormat(pageW, 12, "Expiry Alert Report", "0", 1, "C", false, 0, "")

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

	addMeta := func(label, value string) {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(80, 80, 110)
		pdf.CellFormat(30, 6, label, "0", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(20, 20, 50)
		pdf.CellFormat(60, 6, value, "0", 1, "L", false, 0, "")
		pdf.SetX(15)
	}

	addMeta("Generated:", time.Now().Format("02 Jan 2006"))
	addMeta("Branch:", branchLabel)
	addMeta("Alert Window:", fmt.Sprintf("Expiring within %d days", days))
	addMeta("Total Alerts:", fmt.Sprintf("%d items", len(items)))

	pdf.Ln(4)

	// ── Divider ───────────────────────────────────────────────────────────────
	pdf.SetDrawColor(0, 184, 169)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)

	// ── Table Header ──────────────────────────────────────────────────────────
	colW := []float64{60, 30, 25, 30, 25, 10}
	headers := []string{"Medicine Name", "Batch No.", "Qty", "Expiry Date", "Days Left", "!"}
	aligns := []string{"L", "C", "C", "C", "C", "C"}

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

		// Urgency background: red < 7 days, amber < 30 days, light blue otherwise
		var urgencyColor [3]int
		var urgencyBadge string
		switch {
		case item.DaysToExpiry <= 7:
			urgencyColor = [3]int{255, 235, 235}
			urgencyBadge = "!"
		case item.DaysToExpiry <= 30:
			urgencyColor = [3]int{255, 248, 225}
			urgencyBadge = "~"
		default:
			urgencyColor = [3]int{235, 250, 245}
			urgencyBadge = ""
		}

		if idx%2 == 0 {
			pdf.SetFillColor(urgencyColor[0], urgencyColor[1], urgencyColor[2])
		} else {
			pdf.SetFillColor(urgencyColor[0]-5, urgencyColor[1]-5, urgencyColor[2]-5)
		}
		pdf.Rect(15, rowY, pageW, 8, "F")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(20, 20, 50)
		pdf.SetXY(15, rowY)
		pdf.CellFormat(colW[0], 8, item.MedicineName, "0", 0, "L", false, 0, "")

		pdf.SetTextColor(50, 50, 80)
		pdf.CellFormat(colW[1], 8, item.BatchNumber, "0", 0, "C", false, 0, "")
		pdf.CellFormat(colW[2], 8, fmt.Sprintf("%d", item.Quantity), "0", 0, "C", false, 0, "")
		pdf.CellFormat(colW[3], 8, item.ExpiryDate.Format("02 Jan 2006"), "0", 0, "C", false, 0, "")

		// Days left — color red if <= 7
		if item.DaysToExpiry <= 7 {
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetTextColor(180, 40, 40)
		} else if item.DaysToExpiry <= 30 {
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetTextColor(180, 120, 0)
		} else {
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(50, 50, 80)
		}
		pdf.CellFormat(colW[4], 8, fmt.Sprintf("%d", item.DaysToExpiry), "0", 0, "C", false, 0, "")
		pdf.CellFormat(colW[5], 8, urgencyBadge, "0", 1, "C", false, 0, "")

		pdf.SetDrawColor(220, 225, 235)
		pdf.SetLineWidth(0.2)
		pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	}

	pdf.Ln(6)

	// ── Legend ────────────────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(80, 80, 110)
	pdf.SetX(15)
	pdf.CellFormat(pageW, 6, "Legend:  ! = Critical (≤7 days)     ~ = Warning (≤30 days)", "0", 1, "L", false, 0, "")

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
