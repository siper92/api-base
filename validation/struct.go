package validation

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	core_utils "github.com/siper92/core-utils"
	"strings"
)

type StructValidatorResult []validator.FieldError

func (s StructValidatorResult) Error() string {
	buff := bytes.NewBufferString("")
	for i := 0; i < len(s); i++ {
		buff.WriteString(s[i].Error())
		buff.WriteString("\n")
	}

	if buff.Len() == 0 {
		return ""
	}

	return strings.TrimSpace(buff.String())
}

func (s StructValidatorResult) HasError(field string) bool {
	for _, err := range s {
		if err.Field() == field {
			return true
		}
	}

	return false
}

func (s StructValidatorResult) SingleError() error {
	var messages []string
	for _, err := range s {
		var val validator.FieldError
		if errors.As(err, &val) {
			messages = append(messages, fmt.Sprintf("%s: %s", val.Field(), val.Tag()))
		}
	}

	return fmt.Errorf(strings.Join(messages, "\n"))
}

func ValidateApiData(s interface{}) error {
	validate := validator.New()

	err := validate.RegisterValidation("password", ValidatePassword)
	if err != nil {
		core_utils.ErrorWarning(err)
		return nil
	}

	err = validate.Struct(s)

	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		if _errors, ok := err.(validator.ValidationErrors); ok {
			return StructValidatorResult(_errors)
		}
	}

	return nil
}

func ValidateGenericStruct(s interface{}) StructValidatorResult {
	validate := validator.New()
	err := validate.Struct(s)
	var _errors StructValidatorResult

	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			core_utils.ErrorWarning(err)
		}

		if fieldErrors, ok := err.(validator.ValidationErrors); ok {
			return StructValidatorResult(fieldErrors)
		}
	}

	return _errors
}
