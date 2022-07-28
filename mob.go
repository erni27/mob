package mob

import (
	"errors"
)

var (
	ErrHandlerNotFound = errors.New("handler not found")
	ErrInvalidHandler  = errors.New("invalid handler")
)
