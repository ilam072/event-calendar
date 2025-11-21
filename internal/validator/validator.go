package validator

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{validate: validator.New()}
}
func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

func (v *Validator) FormatValidationErrors(err error) map[string]string {
	errorsMap := make(map[string]string)

	if err == nil {
		return errorsMap
	}

	var ve validator.ValidationErrors
	if ok := errors.As(err, &ve); !ok {
		errorsMap["error"] = err.Error()
		return errorsMap
	}

	for _, fe := range ve {
		field := fe.Field()
		var msg string

		switch fe.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is required", field)
		case "min":
			msg = fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		case "max":
			msg = fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
		case "email":
			msg = fmt.Sprintf("%s must be a valid email", field)
		default:
			msg = fmt.Sprintf("%s is invalid", field)
		}

		errorsMap[field] = msg
	}

	return errorsMap
}
