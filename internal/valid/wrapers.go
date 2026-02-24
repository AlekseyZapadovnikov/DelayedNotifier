package valid

import (
	"strconv"

	"github.com/go-playground/validator/v10"
)

func validateTgWrap(fl validator.FieldLevel) bool {
	userName := fl.Field().String()
	return validateTg(userName)
}

// Custom validation function for port strings
func validatePortWrap(fl validator.FieldLevel) bool {
	portStr := fl.Field().String()
	if portStr == "" {
		return false
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	return port > 0 && port <= 65535
}

// проверяет валидность адресов/userna`ов, на которые отправляются сооющения
func validateToField(fl validator.FieldLevel) bool {
	field := fl.Field()
	for i := 0; i < field.Len(); i++ {
		el := field.Index(i).String()
		if !ValidateEmailOrTg(el) {
			return false
		}
	}
	return true
}

// проверяет на валидность tg или email
func validateFromField(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return ValidateEmailOrTg(value)
}
