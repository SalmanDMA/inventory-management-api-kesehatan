package helpers

import (
	"github.com/gofiber/fiber/v2"
)

type ResponseWithData struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    any `json:"data"`
}

type ResponseWithoutData struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func Response(c *fiber.Ctx, code int, message string, payload any) error {
	status := "success"
	if code >= 400 {
		status = "failed"
	}

	if payload != nil {
		return c.Status(code).JSON(&ResponseWithData{
			Status:  status,
			Message: message,
			Data:    payload,
		})
	}

	return c.Status(code).JSON(&ResponseWithoutData{
		Status:  status,
		Message: message,
	})
}
