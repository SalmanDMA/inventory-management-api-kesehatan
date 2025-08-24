package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all customers with pagination
// @Description Get all customers with pagination, filtering, and search.
// @Tags Customer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param customer_id query string false "Filter by customer ID"
// @Success 200 {array} models.ResponseGetCustomer
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving customers"
// @Router /api/v1/customer [get]
func CustomerControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	customerRepo := repositories.NewCustomerRepository(configs.DB)
	customerService := services.NewCustomerService(customerRepo)
	
	customersResponse, err := customerService.GetAllCustomersPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting customers", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all customers", customersResponse)
}

// @Summary Get By ID customer
// @Description Get By ID customer
// @Tags customers
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetCustomer
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting customer"
// @Router /api/v1/customer/{id} [get]
func CustomerControllerGetByID(c *fiber.Ctx) error {
	_, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Customer info not found", nil)
	}

	id := c.Params("id")
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	customerService := services.NewCustomerService(customerRepo)
	customerResponse, err := customerService.GetCustomerByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting customer", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Customer fetched successfully", customerResponse)
}

	// @Summary Create customer
	// @Description Create customer. For now only accessible by users with DEVELOPER or SUPERADMIN customers.
	// @Tags Customer
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param customerCreateRequest body models.CustomerCreateRequest true "Customer create request"
	// @Success 200 {object} models.ResponseGetCustomer
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create customers"
	// @Failure 409 {string} string "Customer already exists"
	// @Failure 500 {string} string "Error creating customer"
	// @Router /api/v1/customer [post]
	func CustomerControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		customer := new(models.CustomerCreateRequest)
	
		if err := ctx.BodyParser(customer); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		if err := helpers.ValidateStruct(customer); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		customerRepo := repositories.NewCustomerRepository(configs.DB)
		customerService := services.NewCustomerService(customerRepo)
		
		if _ , err := customerService.CreateCustomer(customer, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Customer created successfully", nil)
	}

	// @Summary Update customer
	// @Description Update customer. For now only accessible by users with DEVELOPER or SUPERADMIN customers.
	// @Tags Customer
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Customer ID"
	// @Param customerUpdateRequest body models.CustomerCreateRequest true "Customer update request"
	// @Success 200 {object} models.ResponseGetCustomer
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update customers"
	// @Failure 404 {string} string "Customer not found"
	// @Failure 500 {string} string "Error updating customer"
	// @Router /api/v1/customer/{id} [put]
	func CustomerControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		customer := new(models.CustomerCreateRequest)
	
		if err := ctx.BodyParser(customer); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		customerId := ctx.Params("id")
	
		if err := helpers.ValidateStruct(customer); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		customerRepo := repositories.NewCustomerRepository(configs.DB)
		customerService := services.NewCustomerService(customerRepo)
		if _ , err := customerService.UpdateCustomer(customerId, customer, ctx, userInfo); err != nil {
			if err == repositories.ErrCustomerNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Customer not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating customer", nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Customer updated successfully", nil)
	}

// CustomerControllerDelete adalah handler untuk endpoint customer
// @Summary Delete customer
// @Description Delete customer. For now only accessible by customers with DEVELOPER or SUPERADMIN customers. If the customer is hard deleted, the customer's avatar will be deleted as well. If the customer is soft deleted, the customer's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags Customer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.CustomerIsHardDeleteRequest true "Customer delete request body"
// @Success 200 {array} models.Customer
// @Failure 403 {string} string "Forbidden: You do not have access to delete customer"
// @Failure 404 {string} string "Customer not found"
// @Failure 500 {string} string "Error deleting customer"
// @Router /api/v1/customer/delete [delete]
func CustomerControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Customer info not found", nil)
		}

	customerRequest := new(models.CustomerIsHardDeleteRequest)
	if err := ctx.BodyParser(customerRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(customerRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	customerRepo := repositories.NewCustomerRepository(configs.DB)
	customerService := services.NewCustomerService(customerRepo)

	if err := customerService.DeleteCustomers(customerRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if customerRequest.IsHardDelete == "hardDelete"  {
		message = "Customers deleted successfully"
	} else {
		message = "Customers moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// CustomerControllerRestore restores a soft-deleted customer
// @Summary Restore customer
// @Description Restore customer. For now only accessible by customers with DEVELOPER or SUPERADMIN customers.
// @Tags Customer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.CustomerRestoreRequest true "Customer restore request body"
// @Success 200 {array} models.Customer
// @Failure 403 {string} string "Forbidden: You do not have access to restore customer"
// @Failure 404 {string} string "Customer not found"
// @Failure 500 {string} string "Error restoring customer"
// @Router /api/v1/customer/restore [put]
func CustomerControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Customer info not found", nil)
		}

	customerRequest := new(models.CustomerRestoreRequest)
	if err := ctx.BodyParser(customerRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(customerRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	customerRepo := repositories.NewCustomerRepository(configs.DB)
	customerService := services.NewCustomerService(customerRepo)

	restoredCustomers, err := customerService.RestoreCustomers(customerRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Customers restored successfully", restoredCustomers)
}