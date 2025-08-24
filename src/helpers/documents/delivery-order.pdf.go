package documents

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/jung-kurt/gofpdf"
)

func GenerateDeliveryOrderPDF(so *models.SalesOrder) (string, []byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Format angka Indonesia: 1.234.567
	formatIDR := func(n int) string {
		s := strconv.FormatInt(int64(n), 10)
		neg := false
		if strings.HasPrefix(s, "-") {
			neg = true
			s = s[1:]
		}
		var chunks []string
		for len(s) > 3 {
			chunks = append([]string{s[len(s)-3:]}, chunks...)
			s = s[:len(s)-3]
		}
		if s != "" {
			chunks = append([]string{s}, chunks...)
		}
		out := strings.Join(chunks, ".")
		if neg {
			out = "-" + out
		}
		return out
	}

	// === Header ===
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, "DELIVERY ORDER")
	pdf.Ln(12)

	// Helper row writer
	row := func(label, val string) {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(45, 7, label)
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(0, 7, val)
		pdf.Ln(7)
	}

	// === Meta Info ===
	pdf.SetFont("Arial", "", 11)
	row("DO Number:", so.SONumber) // sementara pakai SO Number
	row("DO Date:", so.SODate.Format("02 January 2006"))
	row("Sales Order Ref:", so.SONumber)
	row("Status:", so.SOStatus)
	pdf.Ln(4)

	// === Customer Info ===
	if so.Customer.ID.String() != "00000000-0000-0000-0000-000000000000" {
		pdf.SetFont("Arial", "B", 13)
		pdf.Cell(0, 8, "Customer")
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

	// === Items (No | Item Name | Quantity | Unit Price | Subtotal) ===
	if len(so.SalesOrderItems) > 0 {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 13)
		pdf.Cell(0, 8, "Items Delivered")
		pdf.Ln(10)

		colNo := 10.0
		colName := 80.0
		colQty := 25.0
		colPrice := 35.0
		colSubtotal := 35.0

		// Header
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(colNo, 8, "No", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colName, 8, "Item Name", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colQty, 8, "Quantity", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colPrice, 8, "Unit Price", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colSubtotal, 8, "Subtotal", "1", 0, "C", true, 0, "")
		pdf.Ln(8)

		// Rows
		pdf.SetFont("Arial", "", 9)
		var grandTotal int
		for i, it := range so.SalesOrderItems {
			name := it.Item.Name
			if name == "" {
				name = "Unknown Item"
			}
			sub := it.TotalPrice
			if sub == 0 {
				sub = it.Quantity * it.UnitPrice
			}
			grandTotal += sub

			pdf.CellFormat(colNo, 8, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
			pdf.CellFormat(colName, 8, name, "1", 0, "L", false, 0, "")
			pdf.CellFormat(colQty, 8, fmt.Sprintf("%d", it.Quantity), "1", 0, "C", false, 0, "")
			pdf.CellFormat(colPrice, 8, "Rp "+formatIDR(it.UnitPrice), "1", 0, "R", false, 0, "")
			pdf.CellFormat(colSubtotal, 8, "Rp "+formatIDR(sub), "1", 0, "R", false, 0, "")
			pdf.Ln(8)
		}

		// Total summary row
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(colNo+colName+colQty+colPrice, 8, "Total", "1", 0, "R", true, 0, "")
		pdf.CellFormat(colSubtotal, 8, "Rp "+formatIDR(grandTotal), "1", 0, "R", true, 0, "")
		pdf.Ln(12)
	}

	// === Notes ===
	if so.Notes != "" {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 7, "Notes")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 6, so.Notes, "", "", false)
	}

	// === Signature ===
	pdf.Ln(15)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(80, 7, "Sender")
	pdf.Cell(80, 7, "Receiver")
	pdf.Ln(20)
	pdf.Cell(80, 7, "(..................)")
	pdf.Cell(80, 7, "(..................)")
	pdf.Ln(10)

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Generated at %s", time.Now().Format("02 January 2006 15:04:05")))

	// Output to bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	filename := fmt.Sprintf("DO_%s_%s.pdf", so.SONumber, time.Now().Format("20060102150405"))
	return filename, buf.Bytes(), nil
}
