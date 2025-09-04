package utils

import (
	"regexp"

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

// LoginRequestValidator validate login request
// ensure at least one of username or email is provided
func LoginRequestValidator(sl validator.StructLevel) {
	username := ""
	email := ""

	if usernameField := sl.Current().FieldByName("Username"); usernameField.IsValid() {
		username = usernameField.String()
	}
	if emailField := sl.Current().FieldByName("Email"); emailField.IsValid() {
		email = emailField.String()
	}

	if username == "" && email == "" {
		sl.ReportError(sl.Current().FieldByName("Username"), "username", "Username", "login_validation", "")
		sl.ReportError(sl.Current().FieldByName("Email"), "email", "Email", "login_validation", "")
	}
}
