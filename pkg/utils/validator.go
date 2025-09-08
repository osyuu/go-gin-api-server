package utils

import (
	"go-gin-api-server/internal/model"
	"regexp"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/go-playground/validator/v10"
)

// UsernameValidator validate username format
// allow: letter, number, underscore(_), hyphen(-)
// dont allow: start with number or special character
func UsernameValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	pattern := `^[a-zA-Z][a-zA-Z0-9_-]*$`

	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return false
	}

	return matched
}

// UsernameOrEmailValidator validate that at least one of username or email is provided
func UsernameOrEmailValidator(sl validator.StructLevel) {
	username := ""
	email := ""

	if usernameField := sl.Current().FieldByName("Username"); usernameField.IsValid() {
		username = usernameField.String()
	}
	if emailField := sl.Current().FieldByName("Email"); emailField.IsValid() {
		email = emailField.String()
	}

	if username == "" && email == "" {
		sl.ReportError(sl.Current().FieldByName("Username"), "username", "Username", "username_or_email", "")
		sl.ReportError(sl.Current().FieldByName("Email"), "email", "Email", "username_or_email", "")
	}
}

// RegisterCustomValidators 註冊自定義驗證器
func RegisterCustomValidators(v *validator.Validate) {
	err := v.RegisterValidation("username", UsernameValidator)
	if err != nil {
		logger.Fatalf("Failed to register username validator: %v", err)
	}
	v.RegisterStructValidation(UsernameOrEmailValidator, model.LoginRequest{})
	v.RegisterStructValidation(UsernameOrEmailValidator, model.RegisterRequest{})
}
