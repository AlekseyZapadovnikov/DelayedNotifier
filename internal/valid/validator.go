package valid

import (
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = setupValidator()
}

func setupValidator() *validator.Validate {
	v := validator.New()

	registerValidationWrap("tg", validateTgWrap)
	registerValidationWrap("port", validatePortWrap)

	registerValidationWrap("to_field", validateToField)
	registerValidationWrap("from_field", validateFromField)

	return v
}

func registerValidationWrap(tag string, fn validator.Func) {
	err := Validate.RegisterValidation(tag, fn)
	if err != nil {
		msg := "validator wasn`t setup correctly" + err.Error()
		panic(msg)
	}
}
