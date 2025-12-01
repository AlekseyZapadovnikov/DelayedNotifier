package valid

import (
    "github.com/go-playground/validator/v10"
    "fmt"
    "strings"
)

// Функция превращает технические ошибки в читаемые сообщения
func RecordValidationDescription(err error) string {
    // Проверяем, действительно ли это ошибки валидации
    validationErrors, ok := err.(validator.ValidationErrors)
    if !ok {
        return err.Error()
    }

    var errMsgs []string
    for _, e := range validationErrors {
        switch e.Field() {
        case "SendChan":
            errMsgs = append(errMsgs, "sendChan must be either 'tg' or 'mail'")
        case "From":
            errMsgs = append(errMsgs, "field 'from' must be a valid email or telegram username")
        case "To":
            errMsgs = append(errMsgs, "field 'to' contains invalid addresses")
        case "Msg":
             errMsgs = append(errMsgs, "message cannot be empty")
        case "Date":
             errMsgs = append(errMsgs, "dateTime is required")
        default:
            errMsgs = append(errMsgs, fmt.Sprintf("field '%s' is invalid", e.Field()))
        }
    }

    return strings.Join(errMsgs, "; ")
}