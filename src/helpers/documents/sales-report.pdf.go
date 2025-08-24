package documents

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/jung-kurt/gofpdf"
)

func GenerateSalesReportPDF(sr *models.SalesReportDetailItem) (string, []byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	// margin body
	pdf.SetMargins(15, 22, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AliasNbPages("")

	// Warna
	primaryBlue := []int{41, 128, 185}
	lightGray := []int{248, 249, 250}
	darkGray := []int{52, 58, 64}
	borderGray := []int{218, 223, 229}

	// Layout
	headerH := 32.0
	logoX, logoY := 8.0, 6.0
	logoW, logoH := 26.0, 20.0
	textX := logoX + logoW + 8

	// Inner width (biar gampang hitung lebar kolom)
	lm, _, rm, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	innerW := pageW - lm - rm // = 180mm untuk A4 dengan margin 15/15

	// Lebar kolom standar tabel info
	labelW := 60.0
	valueW := innerW - labelW // 120.0

	// ---------- Header ----------
	pdf.SetHeaderFunc(func() {
		pdf.SetFillColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
		pdf.Rect(0, 0, 210, headerH, "F")

		// Logo kiri
		pdf.SetFillColor(255, 255, 255)
		pdf.Rect(logoX, logoY, logoW, logoH, "F")
		pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
		pdf.SetFont("Arial", "B", 8)
		pdf.Text(logoX+7, logoY+9, "COMPANY")
		pdf.Text(logoX+9, logoY+13, "LOGO")

		// Company title + alamat
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 14)
		pdf.SetXY(textX, 8)
		pdf.CellFormat(0, 6, "INVENTORY SOLUTION", "", 1, "L", false, 0, "")

		pdf.SetFont("Arial", "", 9)
		pdf.SetX(textX)
		pdf.CellFormat(0, 4.5, "Jl. Teknologi No. 123, Jakarta Selatan 12345", "", 1, "L", false, 0, "")
		pdf.SetX(textX)
		pdf.CellFormat(0, 4.5, "Tel: (021) 1234-5678 | Email: info@inventory.co.id", "", 1, "L", false, 0, "")

		// Doc title kanan
		pdf.SetFont("Arial", "B", 12)
		pdf.SetXY(150, 9)
		pdf.CellFormat(0, 6, "SALES ORDER", "", 1, "R", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.SetX(150)
		pdf.CellFormat(0, 4.5, "SO Report", "", 1, "R", false, 0, "")

		pdf.SetTextColor(0, 0, 0)
	})

	// ---------- Footer ----------
	pdf.SetFooterFunc(func() {
		pdf.SetY(-13)
		pdf.SetDrawColor(borderGray[0], borderGray[1], borderGray[2])
		pdf.Line(lm, pdf.GetY(), pageW-rm, pdf.GetY())
		pdf.Ln(1.5)

		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])

		half := innerW / 2
		left := fmt.Sprintf("Generated on %s", time.Now().Format("Mon, 02 Jan 2006 15:04:05"))
		right := fmt.Sprintf("Page %d of {nb}", pdf.PageNo())

		pdf.CellFormat(half, 4, left, "", 0, "L", false, 0, "")
		pdf.CellFormat(half, 4, right, "", 0, "R", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	pdf.AddPage()
	pdf.SetY(headerH + 6)

	// ---------- Helpers ----------
	sectionHeader := func(title string) {
		pdf.Ln(3)
		pdf.SetFillColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(innerW, 8, " "+title, "", 1, "L", true, 0, "")
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(1.5)
	}

	infoRow := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])

		pdf.SetFont("Arial", "B", 9.5)
		pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
		pdf.CellFormat(labelW, 7, " "+label, "LTB", 0, "L", true, 0, "")

		pdf.SetFont("Arial", "", 9.5)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(valueW, 7, " "+value, "RTB", 1, "L", true, 0, "")
	}

	// Payment Summary table rapi (full width 180mm)
	paymentTable := func(total, paid, unpaid int) {
		// Header baris untuk "PAYMENT SUMMARY" sudah dibuat di sectionHeader
		// Sekarang kita render 3 baris + baris GRAND TOTAL
		h := 9.0
		border := "1"

		// Paid
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])
		pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
		pdf.CellFormat(labelW, h, "  Paid Amount", border, 0, "L", true, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(valueW, h, formatRupiahIDR(paid), border, 1, "R", false, 0, "")

		// Unpaid
		pdf.SetFont("Arial", "B", 10)
		pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
		pdf.CellFormat(labelW, h, "  Unpaid Amount", border, 0, "L", true, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(valueW, h, formatRupiahIDR(unpaid), border, 1, "R", false, 0, "")

		// Total
		pdf.SetFont("Arial", "B", 10)
		pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
		pdf.CellFormat(labelW, h, "  Total Amount", border, 0, "L", true, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(valueW, h, formatRupiahIDR(total), border, 1, "R", false, 0, "")

		// GRAND TOTAL
		pdf.SetFillColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(labelW, h+1, "  GRAND TOTAL", border, 0, "L", true, 0, "")
		pdf.CellFormat(valueW, h+1, formatRupiahIDR(total), border, 1, "R", true, 0, "")

		pdf.SetTextColor(0, 0, 0)
	}

	// ---------- Content ----------
	sectionHeader("ORDER INFORMATION")
	infoRow("SO Number:", sr.SONumber)
	infoRow("SO Date:", sr.SODate.Format("02 January 2006"))
	infoRow("Status:", strings.ToUpper(sr.SOStatus))
	infoRow("Payment Status:", strings.ToUpper(sr.PaymentStatus))
	if sr.DueDate != nil {
		infoRow("Due Date:", sr.DueDate.Format("02 January 2006"))
	}
	infoRow("Created At:", sr.CreatedAt.Format("02 January 2006 15:04"))
	infoRow("Updated At:", sr.UpdatedAt.Format("02 January 2006 15:04"))

	sectionHeader("SALES PERSON")
	infoRow("Name:", sr.SalesPerson.Name)
	if v := pstr(sr.SalesPerson.Email); v != "" {
		infoRow("Email:", v)
	}
	if v := pstr(sr.SalesPerson.Phone); v != "" {
		infoRow("Phone:", v)
	}

	sectionHeader("CUSTOMER INFORMATION")
	infoRow("Customer Name:", sr.Customer.Name)
	infoRow("Customer Nomor:", sr.Customer.Nomor)
	if v := pstr(sr.Customer.Address); v != "" {
		// Label
		pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])
		pdf.SetFont("Arial", "B", 9.5)
		pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
		pdf.CellFormat(labelW, 7, " Address:", "LTB", 0, "L", true, 0, "")

		// Value multiline
		pdf.SetFont("Arial", "", 9.5)
		pdf.SetTextColor(0, 0, 0)
		lines := pdf.SplitLines([]byte(v), valueW-2) // padding 2mm
		for i, line := range lines {
			if i == 0 {
				pdf.CellFormat(valueW, 7, " "+string(line), "RTB", 1, "L", true, 0, "")
			} else {
				pdf.CellFormat(labelW, 7, "", "LB", 0, "L", true, 0, "")
				pdf.CellFormat(valueW, 7, " "+string(line), "RB", 1, "L", true, 0, "")
			}
		}
	}
	if v := pstr(sr.Customer.Phone); v != "" {
		infoRow("Phone:", v)
	}

	sectionHeader("PAYMENT SUMMARY")
	pdf.Ln(1)
	paymentTable(sr.TotalAmount, sr.PaidAmount, sr.UnpaidAmount)

	// Separator & ucapan
	pdf.Ln(6)
	pdf.SetDrawColor(borderGray[0], borderGray[1], borderGray[2])
	pdf.Line(lm, pdf.GetY(), pageW-rm, pdf.GetY())
	pdf.Ln(3)

	pdf.SetFont("Arial", "I", 8.5)
	pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
	pdf.CellFormat(innerW, 4.5, "Thank you for your business! For any inquiries, please contact our customer service.", "", 1, "C", false, 0, "")

	// Output
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	filename := fmt.Sprintf("SO_%s_%s.pdf", sanitizeFilename(sr.SONumber), time.Now().Format("20060102_150405"))
	return filename, buf.Bytes(), nil
}

// Rupiah format
func formatRupiahIDR(amount int) string {
	if amount == 0 {
		return "Rp 0"
	}
	sign := ""
	n := amount
	if n < 0 {
		sign = "-"
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteString(".")
		}
		b.WriteRune(r)
	}
	return fmt.Sprintf("%sRp %s", sign, b.String())
}

// Sanitize filename
func sanitizeFilename(s string) string {
	if s == "" {
		return "UNKNOWN"
	}
	repl := map[string]string{
		"/": "_", "\\": "_", ":": "_", "*": "_",
		"?": "_", "\"": "_", "<": "_", ">": "_",
		"|": "_", " ": "_",
	}
	for k, v := range repl {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}
