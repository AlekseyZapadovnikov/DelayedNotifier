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

	registerValidationWrap(v, "tg", validateTgWrap)
	registerValidationWrap(v, "port", validatePortWrap)
	registerValidationWrap(v, "valid_port", validatePortWrap)

	registerValidationWrap(v, "to_field", validateToField)
	registerValidationWrap(v, "from_field", validateFromField)

	return v
}

func registerValidationWrap(v *validator.Validate, tag string, fn validator.Func) {
	err := v.RegisterValidation(tag, fn)
	if err != nil {
		panic("validator wasn't set up correctly: " + err.Error())
	}
}
