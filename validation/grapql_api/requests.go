package grapql_api

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/go-playground/validator/v10"
	"github.com/siper92/api-base/validation"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type ErrorHandler struct {
	ctx              context.Context
	CustomGetMessage func(err validator.FieldError) string
}

func NewErrorHandler(ctx context.Context) ErrorHandler {
	return ErrorHandler{
		ctx: ctx,
	}
}

func NewErrorHandlerWithCustomMessages(ctx context.Context, msgGetter func(err validator.FieldError) string) ErrorHandler {
	return ErrorHandler{
		ctx:              ctx,
		CustomGetMessage: msgGetter,
	}
}

func (e ErrorHandler) AddFieldErrorToResponse(field string, message string) error {
	graphql.AddError(e.ctx, &gqlerror.Error{
		Message: message,
		Extensions: map[string]any{
			"field": field,
		},
	})

	return gqlerror.Errorf("Validation errors")
}

func (e ErrorHandler) AddValidationsErrors(err error) error {
	switch errors := err.(type) {
	case validation.StructValidatorResult:
		for _, fieldErr := range errors {
			_ = e.AddFieldErrorToResponse(fieldErr.Field(), e.getMessage(fieldErr))
		}
	case validator.ValidationErrors:
		for _, fieldErr := range errors {
			_ = e.AddFieldErrorToResponse(fieldErr.Field(), e.getMessage(fieldErr))
		}
	default:
		return gqlerror.Errorf("Validation errors: %T:%s", err, err.Error())
	}

	return gqlerror.Errorf("Validation errors")
}

func (e ErrorHandler) getMessage(err validator.FieldError) string {
	if e.CustomGetMessage != nil {
		errMsg := e.CustomGetMessage(err)
		if errMsg != "" {
			return errMsg
		}
	}

	// default messages
	switch err.Tag() {
	case "required":
		return "Това поле е задължително"
	case "email":
		return "Невалиден имейл адрес"
	case "min":
		return "Полето трябва да е поне " + err.Param() + " символа"
	case "max":
		return "Полето трябва да е най-много " + err.Param() + " символа"
	default:
		return "Невалидна стойност"
	}
}
