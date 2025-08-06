package helpers

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidateStruct(data any) error {
	validate := validator.New()
	return validate.Struct(data)
}

func ExtractErrorMessages(err error) string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validationErrors {
			var message string
			switch err.Tag() {
			case "required":
				message = err.Field() + " is required"
			case "unique":
				message = err.Field() + " already exists"
			case "email":
				message = err.Field() + " must be a valid email"
			case "min":
				message = err.Field() + " must be at least " + err.Param() + " characters long"
			case "max":
				message = err.Field() + " must be at most " + err.Param() + " characters long"
			default:
				message = err.Field() + " is invalid"
			}
			errors = append(errors, message)
		}
	}
	return strings.Join(errors[:1], " ")
}