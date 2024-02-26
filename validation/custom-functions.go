package validation

import (
	"github.com/go-playground/validator/v10"
	core_utils "github.com/siper92/core-utils"
)

func IsValidEmail(email string) bool {
	return core_utils.SimpleStructValidation(struct {
		Email string `validate:"required,email"`
	}{
		Email: email,
	}).HasFieldErrorFor("Email") == false
}

func ValidatePassword(fl validator.FieldLevel) bool {
	passwordValue := fl.Field().String()
	minVal := 6
	if len(passwordValue) < minVal {
		return false
	}

	if len(passwordValue) != len([]rune(passwordValue)) {
		// not ascii
		return false
	}

	var (
		hasUpperCase bool
		hasDigit     bool
	)

	for _, char := range passwordValue {
		if char >= 'A' && char <= 'Z' {
			hasUpperCase = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	return hasUpperCase && hasDigit
}
