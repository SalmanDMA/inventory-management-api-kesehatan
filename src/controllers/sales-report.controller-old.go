package controllers

// import (
// 	"strconv"
// 	"time"

// 	"github.com/SalmanDMA/inventory-app/backend/src/configs"
// 	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
// 	"github.com/SalmanDMA/inventory-app/backend/src/models"
// 	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
// 	"github.com/SalmanDMA/inventory-app/backend/src/services"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/google/uuid"
// )

// // ---------------- GET SALES REPORT ----------------
// func GetSalesReport(ctx *fiber.Ctx) error {
// 	userInfo, ok := ctx.Locals("userInfo").(*models.User)
// 	if !ok {
// 		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
// 	}

// 	filters := models.SalesReportFilters{
// 		Period: ctx.Query("period", "month"),
// 		Page:   1,
// 		Limit:  50,
// 	}

// 	// Parse page & limit
// 	if pageStr := ctx.Query("page"); pageStr != "" {
// 		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
// 			filters.Page = page
// 		}
// 	}
// 	if limitStr := ctx.Query("limit"); limitStr != "" {
// 		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
// 			filters.Limit = limit
// 		}
// 	}

// 	// Parse dates
// 	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
// 		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
// 			filters.StartDate = &startDate
// 		}
// 	}
// 	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
// 		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
// 			filters.EndDate = &endDate
// 		}
// 	}

// 	// Parse UUIDs
// 	if salesPersonIDStr := ctx.Query("sales_person_id"); salesPersonIDStr != uuid.Nil.String() {
// 		if salesPersonID, err := uuid.Parse(salesPersonIDStr); err == nil {
// 			filters.SalesPersonID = &salesPersonID
// 		}
// 	}
// 	if areaIDStr := ctx.Query("area_id"); areaIDStr != uuid.Nil.String() {
// 		if areaID, err := uuid.Parse(areaIDStr); err == nil {
// 			filters.AreaID = &areaID
// 		}
// 	}
// 	if facilityIDStr := ctx.Query("facility_id"); facilityIDStr != uuid.Nil.String() {
// 		if facilityID, err := uuid.Parse(facilityIDStr); err == nil {
// 			filters.FacilityID = &facilityID
// 		}
// 	}

// 	// Parse status
// 	if soStatus := ctx.Query("so_status"); soStatus != "" {
// 		filters.SOStatus = &soStatus
// 	}
// 	if paymentStatus := ctx.Query("payment_status"); paymentStatus != "" {
// 		filters.PaymentStatus = &paymentStatus
// 	}

// 	salesReportrRepo := repositories.NewSalesReportRepository(configs.DB)
// 	salesPeronRepo := repositories.NewSalesPersonRepository(configs.DB)
// 	srService := services.NewSalesReportService(salesReportrRepo, salesPeronRepo)

// 	// Role-based filtering
// 	if userInfo.Role.Name == "sales" {
// 		salesPersonID, err := srService.GetSalesPersonIDByUserID(userInfo.ID)
// 		if err != nil {
// 			return helpers.Response(ctx, fiber.StatusForbidden, "Sales person not found", nil)
// 		}
// 		filters.SalesPersonID = &salesPersonID
// 	}

// 	// Validate filters
// 	if err := helpers.ValidateStruct(filters); err != nil {
// 		errorMessage := helpers.ExtractErrorMessages(err)
// 		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
// 	}

// 	// Get sales report data
// 	reportData, err := srService.GetSalesReport(filters)
// 	if err != nil {
// 		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
// 	}

// 	return helpers.Response(ctx, fiber.StatusOK, "Sales report retrieved successfully", reportData)
// }

// // ---------------- EXPORT SALES REPORT ----------------
// func ExportSalesReport(ctx *fiber.Ctx) error {
// 	userInfo, ok := ctx.Locals("userInfo").(*models.User)
// 	if !ok {
// 		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
// 	}

// 	var exportRequest models.SalesReportExportRequest
// 	if err := ctx.BodyParser(&exportRequest); err != nil {
// 		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
// 	}

// 	salesReportrRepo := repositories.NewSalesReportRepository(configs.DB)
// 	salesPeronRepo := repositories.NewSalesPersonRepository(configs.DB)
// 	srService := services.NewSalesReportService(salesReportrRepo, salesPeronRepo)

// 	if userInfo.Role.Name == "sales" {
// 		salesPersonID, err := srService.GetSalesPersonIDByUserID(userInfo.ID)
// 		if err != nil {
// 			return helpers.Response(ctx, fiber.StatusForbidden, "Sales person not found", nil)
// 		}
// 		exportRequest.Filters.SalesPersonID = &salesPersonID
// 	}

// 	if err := helpers.ValidateStruct(exportRequest); err != nil {
// 		errorMessage := helpers.ExtractErrorMessages(err)
// 		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
// 	}

// 	exportResponse, err := srService.ExportSalesReport(exportRequest)
// 	if err != nil {
// 		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
// 	}

// 	return helpers.Response(ctx, fiber.StatusOK, "Sales report exported successfully", exportResponse)
// }

// // ---------------- GET SALES REPORT SUMMARY ----------------
// func GetSalesReportSummary(ctx *fiber.Ctx) error {
// 	userInfo, ok := ctx.Locals("userInfo").(*models.User)
// 	if !ok {
// 		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
// 	}

// 	filters := models.SalesReportFilters{
// 		Period: ctx.Query("period", "month"),
// 	}

// 	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
// 		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
// 			filters.StartDate = &startDate
// 		}
// 	}
// 	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
// 		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
// 			filters.EndDate = &endDate
// 		}
// 	}

// 	salesReportrRepo := repositories.NewSalesReportRepository(configs.DB)
// 	salesPeronRepo := repositories.NewSalesPersonRepository(configs.DB)
// 	srService := services.NewSalesReportService(salesReportrRepo, salesPeronRepo)

// 	if userInfo.Role.Name == "sales" {
// 		salesPersonID, err := srService.GetSalesPersonIDByUserID(userInfo.ID)
// 		if err != nil {
// 			return helpers.Response(ctx, fiber.StatusForbidden, "Sales person not found", nil)
// 		}
// 		filters.SalesPersonID = &salesPersonID
// 	}

// 	summary, err := srService.GetSalesReportSummary(filters)
// 	if err != nil {
// 		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
// 	}

// 	return helpers.Response(ctx, fiber.StatusOK, "Sales report summary retrieved successfully", summary)
// }

// // ---------------- GET SALES REPORT CHARTS ----------------
// func GetSalesReportCharts(ctx *fiber.Ctx) error {
// 	userInfo, ok := ctx.Locals("userInfo").(*models.User)
// 	if !ok {
// 		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
// 	}

// 	filters := models.SalesReportFilters{
// 		Period: ctx.Query("period", "month"),
// 	}

// 	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
// 		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
// 			filters.StartDate = &startDate
// 		}
// 	}
// 	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
// 		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
// 			filters.EndDate = &endDate
// 		}
// 	}

// 	salesReportrRepo := repositories.NewSalesReportRepository(configs.DB)
// 	salesPeronRepo := repositories.NewSalesPersonRepository(configs.DB)
// 	srService := services.NewSalesReportService(salesReportrRepo, salesPeronRepo)

// 	if userInfo.Role.Name == "sales" {
// 		salesPersonID, err := srService.GetSalesPersonIDByUserID(userInfo.ID)
// 		if err != nil {
// 			return helpers.Response(ctx, fiber.StatusForbidden, "Sales person not found", nil)
// 		}
// 		filters.SalesPersonID = &salesPersonID
// 	}

// 	charts, err := srService.GetSalesReportCharts(filters)
// 	if err != nil {
// 		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
// 	}

// 	return helpers.Response(ctx, fiber.StatusOK, "Sales report charts retrieved successfully", charts)
// }
