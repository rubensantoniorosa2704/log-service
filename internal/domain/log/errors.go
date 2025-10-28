package log

import "errors"

var (
	ErrMessageRequired      = errors.New("log message cannot be empty")
	ErrLevelRequired        = errors.New("log level is required and must be valid")
	ErrApplicationIDInvalid = errors.New("application ID is required and must be a valid UUID")
	ErrUserIDInvalid        = errors.New("user ID is required and must be a valid UUID")
	ErrLogNotFound          = errors.New("log not found")
	ErrInvalidDateRange     = errors.New("invalid date range")
	ErrInvalidPagination    = errors.New("invalid pagination parameters")
)
