package controllers

// @Summary Get By ID file
// @Description Get By ID file
// @Tags files
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetFile
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting file"
// @Router /api/v1/file/{id} [get]
// func FileControllerGetByID(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	fileRepo := repositories.NewUploadRepository(configs.DB)
// 	fileService := services.NewFileService(fileRepo)
// 	fileResponse, err := fileService.GetFileByID(id)

// 	if err != nil {
// 		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting file", nil)
// 	}

// 	return helpers.Response(c, fiber.StatusOK, "File fetched successfully", fileResponse)
// }