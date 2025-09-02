package utils

import "github.com/go-playground/validator/v10"

func CustomValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return len(value) >= 3
}
