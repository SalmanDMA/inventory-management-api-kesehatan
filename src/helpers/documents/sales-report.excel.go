package documents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/xuri/excelize/v2"
)

// GenerateSalesReportExcel menyimpan file .xlsx ke dir, mengembalikan fullPath & filename
// Kolom: SO Number | Date | Sales Person | Customer | Area | Status | Payment | Amount
func GenerateSalesReportExcel(items []models.SalesReportDetailItem, dir string) (string, string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir %s: %w", dir, err)
	}

	f := excelize.NewFile()
	const sheet = "Sales Report"
	// ganti sheet default
	f.SetSheetName("Sheet1", sheet)

	// Header
	headers := []string{
		"SO Number", "Date", "Sales Person", "Customer",
		"Area", "Status", "Payment", "Amount",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	// Styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#2980B9"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "DDDDDD", Style: 1},
			{Type: "right", Color: "DDDDDD", Style: 1},
			{Type: "top", Color: "DDDDDD", Style: 1},
			{Type: "bottom", Color: "DDDDDD", Style: 1},
		},
	})
	rowStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "DDDDDD", Style: 1},
			{Type: "right", Color: "DDDDDD", Style: 1},
			{Type: "top", Color: "DDDDDD", Style: 1},
			{Type: "bottom", Color: "DDDDDD", Style: 1},
		},
	})
	amountStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt:    3, // #,##0
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "DDDDDD", Style: 1},
			{Type: "right", Color: "DDDDDD", Style: 1},
			{Type: "top", Color: "DDDDDD", Style: 1},
			{Type: "bottom", Color: "DDDDDD", Style: 1},
		},
	})

	_ = f.SetCellStyle(sheet, "A1", "H1", headerStyle)

	// Data
	for r, it := range items {
		row := r + 2
		getArea := func() string {
			if it.Customer.Area.Name != "" {
				return it.Customer.Area.Name
			}
			return ""
		}

		values := []interface{}{
			it.SONumber,
			it.SODate.Format("02 Jan 2006"),
			it.SalesPerson.Name,
			it.Customer.Name,
			getArea(),
			strings.ToUpper(it.SOStatus),
			strings.ToUpper(it.PaymentStatus),
			it.TotalAmount,
		}
		for c, v := range values {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			_ = f.SetCellValue(sheet, cell, v)
		}
		// Apply styles for the row
		left, _ := excelize.CoordinatesToCellName(1, row)
		right, _ := excelize.CoordinatesToCellName(8, row)
		_ = f.SetCellStyle(sheet, left, right, rowStyle)
		// Right align & number format for Amount
		amtCell, _ := excelize.CoordinatesToCellName(8, row)
		_ = f.SetCellStyle(sheet, amtCell, amtCell, amountStyle)
	}

	// Column widths
	_ = f.SetColWidth(sheet, "A", "A", 15) // SO Number
	_ = f.SetColWidth(sheet, "B", "B", 13) // Date
	_ = f.SetColWidth(sheet, "C", "C", 22) // Sales Person
	_ = f.SetColWidth(sheet, "D", "D", 28) // Customer
	_ = f.SetColWidth(sheet, "E", "E", 18) // Area
	_ = f.SetColWidth(sheet, "F", "F", 12) // Status
	_ = f.SetColWidth(sheet, "G", "G", 12) // Payment
	_ = f.SetColWidth(sheet, "H", "H", 15) // Amount

	filename := fmt.Sprintf("sales_report_%s.xlsx", time.Now().Format("20060102_150405"))
	fullPath := filepath.Join(dir, filename)

	if err := f.SaveAs(fullPath); err != nil {
		return "", "", fmt.Errorf("save excel: %w", err)
	}
	return fullPath, filename, nil
}
