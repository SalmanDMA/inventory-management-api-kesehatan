package documents

import (
	"bytes"
	"fmt"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/jung-kurt/gofpdf"
)

func pstr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func GeneratePurchaseOrderPDF(po *models.PurchaseOrder) (string, []byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Reusable row writer
	row := func(label, val string) {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(45, 7, label)
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(0, 7, val)
		pdf.Ln(7)
	}

	// Header
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, "PURCHASE ORDER")
	pdf.Ln(12)

	// PO meta
	pdf.SetFont("Arial", "", 11)
	row("PO Number:", po.PONumber)
	row("PO Date:", po.PODate.Format("02 January 2006"))
	if po.EstimatedArrival != nil {
		row("Expected Delivery:", po.EstimatedArrival.Format("02 January 2006"))
	}
	row("Payment Terms:", po.TermOfPayment)
	row("Status:", po.POStatus)

	pdf.Ln(4)

	// Supplier
	if po.Supplier.ID.String() != "00000000-0000-0000-0000-000000000000" {
		pdf.SetFont("Arial", "B", 13)
		pdf.Cell(0, 8, "Supplier")
		pdf.Ln(9)

		row("Name:", po.Supplier.Name)

		if v := (po.Supplier.Code); v != "" {
			row("Code:", v)
		}
		if v := pstr(po.Supplier.Email); v != "" {
			row("Email:", v)
		}
		if v := pstr(po.Supplier.Phone); v != "" {
			row("Phone:", v)
		}
		if v := pstr(po.Supplier.Address); v != "" {
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(45, 7, "Address:")
			pdf.SetFont("Arial", "", 11)
			pdf.MultiCell(0, 6, v, "", "", false)
		}
		pdf.Ln(2)
	}

	// Items
	if len(po.PurchaseOrderItems) > 0 {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 13)
		pdf.Cell(0, 8, "Items")
		pdf.Ln(10)

		// header
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(10, 8, "No", "1", 0, "C", true, 0, "")
		pdf.CellFormat(70, 8, "Item Name", "1", 0, "C", true, 0, "")
		pdf.CellFormat(18, 8, "Qty", "1", 0, "C", true, 0, "")
		pdf.CellFormat(32, 8, "Unit Price", "1", 0, "R", true, 0, "")
		pdf.CellFormat(32, 8, "Total Price", "1", 0, "R", true, 0, "")
		pdf.CellFormat(18, 8, "Status", "1", 0, "C", true, 0, "")
		pdf.Ln(8)

		// rows
		pdf.SetFont("Arial", "", 9)
		for i, it := range po.PurchaseOrderItems {
			name := it.Item.Name
			if name == "" {
				name = "Unknown Item"
			}
			pdf.CellFormat(10, 8, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
			pdf.CellFormat(70, 8, name, "1", 0, "L", false, 0, "")
			pdf.CellFormat(18, 8, fmt.Sprintf("%d", it.Quantity), "1", 0, "C", false, 0, "")
			pdf.CellFormat(32, 8, formatRupiah(it.UnitPrice), "1", 0, "R", false, 0, "")
			pdf.CellFormat(32, 8, formatRupiah(it.TotalPrice), "1", 0, "R", false, 0, "")
			pdf.CellFormat(18, 8, it.Status, "1", 0, "C", false, 0, "")
			pdf.Ln(8)
		}

		// total
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(100, 8, "", "0", 0, "R", false, 0, "")
		pdf.CellFormat(30, 8, "TOTAL", "1", 0, "R", true, 0, "")
		pdf.CellFormat(32, 8, formatRupiah(po.TotalAmount), "1", 0, "R", true, 0, "")
		pdf.CellFormat(18, 8, "", "1", 0, "C", true, 0, "")
		pdf.Ln(12)
	}

	// Payment
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 8, "Payment")
	pdf.Ln(9)

	pdf.SetFont("Arial", "", 11)
	row("Total Amount:", formatRupiah(po.TotalAmount))
	row("Paid Amount:", formatRupiah(po.PaidAmount))
	row("Remaining:", formatRupiah(po.TotalAmount-po.PaidAmount))
	row("Payment Status:", po.PaymentStatus)

	if len(po.Payments) > 0 {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 7, "Payment History")
		pdf.Ln(9)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(30, 8, "Date", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 8, "Type", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 8, "Amount", "1", 0, "R", true, 0, "")
		pdf.CellFormat(30, 8, "Method", "1", 0, "C", true, 0, "")
		pdf.CellFormat(45, 8, "Reference", "1", 0, "L", true, 0, "")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 8)
		for _, p := range po.Payments {
			pdf.CellFormat(30, 8, p.PaymentDate.Format("02/01/2006"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 8, p.PaymentType, "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 8, formatRupiah(p.Amount), "1", 0, "R", false, 0, "")
			pdf.CellFormat(30, 8, p.PaymentMethod, "1", 0, "C", false, 0, "")
			pdf.CellFormat(45, 8, p.ReferenceNumber, "1", 0, "L", false, 0, "")
			pdf.Ln(8)
		}
	}

	// Notes
	if po.Notes != "" {
		pdf.Ln(6)
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 7, "Notes")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 6, po.Notes, "", "", false)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Generated at %s", time.Now().Format("02 January 2006 15:04:05")))

	// output to bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.pdf", po.PONumber, time.Now())
	return filename, buf.Bytes(), nil
}

func formatRupiah(amount int) string {
	switch {
	case amount >= 1_000_000_000:
		return fmt.Sprintf("Rp %.1fB", float64(amount)/1_000_000_000)
	case amount >= 1_000_000:
		return fmt.Sprintf("Rp %.1fM", float64(amount)/1_000_000)
	case amount >= 1_000:
		return fmt.Sprintf("Rp %.1fK", float64(amount)/1_000)
	default:
		return fmt.Sprintf("Rp %d", amount)
	}
}
