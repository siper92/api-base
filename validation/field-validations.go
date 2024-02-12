package validation

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	core_utils "github.com/siper92/core-utils"
)

func IsValidEmail(email string) bool {
	return ValidateGenericStruct(struct {
		Email string `validate:"required,email"`
	}{
		Email: email,
	}).HasError("Email") == false
}

func ValidatePassword(fl validator.FieldLevel) bool {
	passConfirmField := fl.Param()

	if passConfirmField == "" {
		passConfirmField = fmt.Sprintf("%sConfirm", fl.FieldName())
	}

	passwordValue := fl.Field().String()
	minVal := 6
	if len(passwordValue) < minVal {
		return false
	}

	passwordFieldValue := fl.Parent().FieldByName(passConfirmField).String()
	if passwordFieldValue == "" {
		core_utils.Debug(fmt.Sprintf("Password field %s not found", passConfirmField))
	}

	return passwordFieldValue == passwordValue
}
