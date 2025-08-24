package documents

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/jung-kurt/gofpdf"
)

func GenerateReceiptPDF(so *models.SalesOrder) (string, []byte, error) {
	if len(so.Payments) == 0 {
		return "", nil, fmt.Errorf("no payments found for this sales order")
	}

	// pilih pembayaran terakhir (by PaymentDate)
	pays := make([]models.Payment, 0, len(so.Payments))
	for _, p := range so.Payments {
		pays = append(pays, p)
	}
	sort.Slice(pays, func(i, j int) bool {
		return pays[i].PaymentDate.After(pays[j].PaymentDate)
	})
	last := pays[0]

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// === Header ===
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, "RECEIPT / KWITANSI")
	pdf.Ln(12)

	// helper row
	row := func(label, val string) {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(45, 7, label)
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(0, 7, val)
		pdf.Ln(7)
	}

	// Meta
	pdf.SetFont("Arial", "", 11)
	receiptNo := fmt.Sprintf("RCP-%s-%s", so.SONumber, last.PaymentDate.Format("20060102150405"))
	row("Receipt Number:", receiptNo)
	row("Receipt Date:", last.PaymentDate.Format("02 January 2006"))
	row("SO Reference:", so.SONumber)
	row("Payment Type:", last.PaymentType)
	row("Payment Method:", last.PaymentMethod)
	if last.ReferenceNumber != "" {
		row("Reference:", last.ReferenceNumber)
	}
	pdf.Ln(2)

	// Received From (Customer)
	if so.Customer.ID.String() != "00000000-0000-0000-0000-000000000000" {
		pdf.SetFont("Arial", "B", 13)
		pdf.Cell(0, 8, "Received From")
		pdf.Ln(9)

		row("Name:", so.Customer.Name)
		if v := so.Customer.Email; v != nil {
			row("Email:", *v)
		}
		if v := so.Customer.Phone; v != nil {
			row("Phone:", *v)
		}
		if v := so.Customer.Address; v != nil {
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(45, 7, "Address:")
			pdf.SetFont("Arial", "", 11)
			pdf.MultiCell(0, 6, *v, "", "", false)
		}
		pdf.Ln(2)
	}

	// Amount
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(45, 8, "Amount:")
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Rp "+formatIDR(last.Amount))
	pdf.Ln(10)

	// For / Description
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, "For")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	desc := fmt.Sprintf("Payment for Sales Order %s", so.SONumber)
	if last.Notes != "" {
		desc = desc + "\n" + last.Notes
	}
	pdf.MultiCell(0, 6, desc, "", "", false)
	pdf.Ln(4)

	// Payment progress summary
	paidToDate := so.PaidAmount
	remaining := so.TotalAmount - paidToDate
	if remaining < 0 {
		remaining = 0
	}

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 8, "Payment Summary")
	pdf.Ln(10)

	colL := 60.0
	colR := 50.0

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(colL, 7, "Total SO Amount")
	pdf.CellFormat(colR, 7, "Rp "+formatIDR(so.TotalAmount), "1", 0, "R", false, 0, "")
	pdf.Ln(7)

	pdf.Cell(colL, 7, "Paid to Date (before this)")
	pdf.CellFormat(colR, 7, "Rp "+formatIDR(paidToDate-last.Amount), "1", 0, "R", false, 0, "")
	pdf.Ln(7)

	pdf.Cell(colL, 7, "This Payment")
	pdf.CellFormat(colR, 7, "Rp "+formatIDR(last.Amount), "1", 0, "R", false, 0, "")
	pdf.Ln(7)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(colL, 8, "Remaining (after this)")
	pdf.CellFormat(colR, 8, "Rp "+formatIDR(remaining), "1", 0, "R", false, 0, "")
	pdf.Ln(12)

	// Signature
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(80, 7, "Received By")
	pdf.Cell(80, 7, "Customer")
	pdf.Ln(20)
	pdf.Cell(80, 7, "(..................)")
	pdf.Cell(80, 7, "(..................)")
	pdf.Ln(10)

	// Footer
	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Generated at %s", time.Now().Format("02 January 2006 15:04:05")))

	// Output
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", nil, fmt.Errorf("failed to generate Receipt PDF: %w", err)
	}
	filename := fmt.Sprintf("RCP_%s_%s.pdf", so.SONumber, time.Now().Format("20060102150405"))
	return filename, buf.Bytes(), nil
}