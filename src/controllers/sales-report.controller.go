package controllers

import (
	"bytes"
	"fmt"
	"io"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// ---------------- GET SALES REPORT SUMMARY ----------------
// GetSalesReportSummary
// @Summary Get sales report summary
// @Description Retrieve sales KPIs summary for a given period/date range.
// @Tags SalesReport
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param period query string false "Period preset (day|week|month|year|custom)" default(month)
// @Param start_date query string false "Start date (YYYY-MM-DD). Required if period=custom."
// @Param end_date query string false "End date (YYYY-MM-DD). Required if period=custom."
// @Param sales_person_id query string false "Filter by Sales Person ID (UUID)"
// @Param area_id query string false "Filter by Area ID (UUID)"
// @Param customer_id query string false "Filter by Customer ID (UUID)"
// @Param so_status query string false "Filter by Sales Order status"
// @Param payment_status query string false "Filter by payment status"
// @Success 200 {object} models.SalesReportSummary "Sales report summary retrieved successfully"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/sales-report/summary [get]
func GetSalesReportSummary(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	paginationReq := &models.PaginationRequest{}
	salesReportRepo := repositories.NewSalesReportRepository(configs.DB)
	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesOrderRepo := repositories.NewSalesOrderRepository(configs.DB)
	srService := services.NewSalesReportService(salesReportRepo, salesPersonRepo, salesOrderRepo)

	// Role-based filtering
	if userInfo.Role.Name == "sales" {
		salesPersonID, err := srService.GetSalesPersonIDByUserID(userInfo.ID)
		if err != nil {
			return helpers.Response(ctx, fiber.StatusForbidden, "Sales person not found", nil)
		}
		paginationReq.SalesPersonID = salesPersonID.String()
	}

	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	summary, err := srService.GetSalesReportSummary(paginationReq)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales report summary retrieved successfully", summary)
}

// ---------------- GET SALES REPORT CHARTS ----------------
// GetSalesReportCharts
// @Summary Get sales report charts data
// @Description Retrieve datasets for charts (trend, status distribution, by person, payment comparison, heatmap).
// @Tags SalesReport
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param period query string false "Period preset (day|week|month|year|custom)" default(month)
// @Param start_date query string false "Start date (YYYY-MM-DD). Required if period=custom."
// @Param end_date query string false "End date (YYYY-MM-DD). Required if period=custom."
// @Param sales_person_id query string false "Filter by Sales Person ID (UUID)"
// @Param area_id query string false "Filter by Area ID (UUID)"
// @Param customer_id query string false "Filter by Customer ID (UUID)"
// @Param so_status query string false "Filter by Sales Order status"
// @Param payment_status query string false "Filter by payment status"
// @Success 200 {object} models.SalesReportChartData "Sales report charts retrieved successfully"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/sales-report/charts [get]
func GetSalesReportCharts(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	paginationReq := &models.PaginationRequest{}
	salesReportRepo := repositories.NewSalesReportRepository(configs.DB)
	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesOrderRepo := repositories.NewSalesOrderRepository(configs.DB)
	srService := services.NewSalesReportService(salesReportRepo, salesPersonRepo, salesOrderRepo)

	if userInfo.Role.Name == "sales" {
		salesPersonID, err := srService.GetSalesPersonIDByUserID(userInfo.ID)
		if err != nil {
			return helpers.Response(ctx, fiber.StatusForbidden, "Sales person not found", nil)
		}
		paginationReq.SalesPersonID = salesPersonID.String()
	}

	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	charts, err := srService.GetSalesReportCharts(paginationReq)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales report charts retrieved successfully", charts)
}

// ---------------- GET SALES REPORT DETAILS ----------------
// GetSalesReportDetails
// @Summary Get sales report details (paginated)
// @Description Retrieve sales report detailed rows with pagination.
// @Tags SalesReport
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size (max 1000)" default(50)
// @Param period query string false "Period preset (day|week|month|year|custom)" default(month)
// @Param start_date query string false "Start date (YYYY-MM-DD). Required if period=custom."
// @Param end_date query string false "End date (YYYY-MM-DD). Required if period=custom."
// @Param sales_person_id query string false "Filter by Sales Person ID (UUID)"
// @Param area_id query string false "Filter by Area ID (UUID)"
// @Param customer_id query string false "Filter by Customer ID (UUID)"
// @Param so_status query string false "Filter by Sales Order status"
// @Param payment_status query string false "Filter by payment status"
// @Success 200 {object} map[string]interface{} "Details data + pagination"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/sales-report/details [get]
func GetSalesReportDetails(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	paginationReq := &models.PaginationRequest{}
	salesReportRepo := repositories.NewSalesReportRepository(configs.DB)
	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesOrderRepo := repositories.NewSalesOrderRepository(configs.DB)
	srService := services.NewSalesReportService(salesReportRepo, salesPersonRepo, salesOrderRepo)

	if userInfo.Role.Name == "sales" {
		salesPersonID, err := srService.GetSalesPersonIDByUserID(userInfo.ID)
		if err != nil {
			return helpers.Response(ctx, fiber.StatusForbidden, "Sales person not found", nil)
		}
		paginationReq.SalesPersonID = salesPersonID.String()
	}

if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	result, err := srService.GetSalesReportDetails(paginationReq)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales report details retrieved successfully", result)
}

// ---------------- GET SALES REPORT INSIGHTS ----------------
// GetSalesReportInsights
// @Summary Get sales report insights
// @Description Retrieve insights (top items/customers, overdue invoices, performance).
// @Tags SalesReport
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param period query string false "Period preset (day|week|month|year|custom)" default(month)
// @Param start_date query string false "Start date (YYYY-MM-DD). Required if period=custom."
// @Param end_date query string false "End date (YYYY-MM-DD). Required if period=custom."
// @Param sales_person_id query string false "Filter by Sales Person ID (UUID)"
// @Param area_id query string false "Filter by Area ID (UUID)"
// @Param customer_id query string false "Filter by Customer ID (UUID)"
// @Param so_status query string false "Filter by Sales Order status"
// @Param payment_status query string false "Filter by payment status"
// @Success 200 {object} models.SalesReportInsights "Sales report insights retrieved successfully"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/sales-report/insights [get]
func GetSalesReportInsights(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	paginationReq := &models.PaginationRequest{}
	salesReportRepo := repositories.NewSalesReportRepository(configs.DB)
	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesOrderRepo := repositories.NewSalesOrderRepository(configs.DB)
	srService := services.NewSalesReportService(salesReportRepo, salesPersonRepo, salesOrderRepo)

	if userInfo.Role.Name == "sales" {
		salesPersonID, err := srService.GetSalesPersonIDByUserID(userInfo.ID)
		if err != nil {
			return helpers.Response(ctx, fiber.StatusForbidden, "Sales person not found", nil)
		}
		paginationReq.SalesPersonID = salesPersonID.String()
	}

	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	insights, err := srService.GetSalesReportInsights(paginationReq)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales report insights retrieved successfully", insights)
}

// ExportSalesReportExcel
// @Summary Export sales report to Excel
// @Description Exports filtered sales orders into an Excel file via query params.
// @Tags SalesReport
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param start_date query string false "RFC3339 date-time (local) start"
// @Param end_date query string false "RFC3339 date-time (local) end"
// @Param period query string false "day|week|month|year|custom"
// @Param sales_person_id query string false
// @Param area_id query string false
// @Param customer_id query string false
// @Param so_status query string false
// @Param payment_status query string false
// @Success 200 {file} file "Excel file"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/sales-report/excel [get]
func ExportSalesReportExcel(ctx *fiber.Ctx) error {
	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	salesReportRepo := repositories.NewSalesReportRepository(configs.DB)
	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesOrderRepo := repositories.NewSalesOrderRepository(configs.DB)
	svc := services.NewSalesReportService(salesReportRepo, salesPersonRepo, salesOrderRepo)

	filename, fileExcel, err := svc.GenerateSalesReportExcel(paginationReq)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	pr, pw := io.Pipe()
	go func() {
		_, werr := fileExcel.WriteTo(pw) 
		_ = fileExcel.Close()            
		_ = pw.CloseWithError(werr)
	}()

	return ctx.SendStream(pr, -1)
}

// Generate PDF Sales Report 
// @Summary Generate pdf
// @Description Stream sales report PDF directly.
// @Tags SalesReport
// @Accept json
// @Produce application/pdf
// @Param id path string true "Sales Report ID"
// @Success 200 {file} file "PDF stream"
// @Failure 404 {string} string "Sales report not found"
// @Failure 500 {string} string "Failed to generate pdf document"
// @Router /api/v1/sales-report/{id}/pdf [get]
func ExportSalesReportPDF(ctx *fiber.Ctx) error {
	soId := ctx.Params("id")

	salesReportRepo := repositories.NewSalesReportRepository(configs.DB)
	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesOrderRepo := repositories.NewSalesOrderRepository(configs.DB)
	srService := services.NewSalesReportService(salesReportRepo, salesPersonRepo, salesOrderRepo)

	filename, pdfBytes, err := srService.GenerateSalesReportPDF(soId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to generate pdf document", err.Error())
	}

	ctx.Set("Content-Type", "application/pdf")
	ctx.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	return ctx.SendStream(bytes.NewReader(pdfBytes))
}
