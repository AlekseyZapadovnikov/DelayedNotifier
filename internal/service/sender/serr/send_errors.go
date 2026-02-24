package serr

import "fmt"

var (
	ErrUnsupportedSendType = fmt.Errorf("unsupported send type")
)