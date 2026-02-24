package valid

import (
	"net/mail"
	"regexp"
	"strings"
)

func ValidateEmail(value string) bool {
	// Проверка email через стандартную библиотеку
	if _, err := mail.ParseAddress(value); err == nil {
		return true
	}
	return false
}

// Проверка Telegram username:
// - начинается с @
// - длина от 5 до 32 символов (включая @)
// - содержит только буквы, цифры и подчёркивания
func validateTg(userName string) bool {

	if strings.HasPrefix(userName, "@") {
		username := userName[1:]
		if len(username) < 4 || len(username) > 32 {
			return false
		}
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
		return matched
	}

	return false
}

func ValidateEmailOrTg(val string) bool {
	return ValidateEmail(val) || validateTg(val)
}
