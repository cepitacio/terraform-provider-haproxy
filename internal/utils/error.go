package utils

import (
	"fmt"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CustomError struct {
	StatusCode int
	APIError   *APIError
	RawMessage string
}

func (e *CustomError) Error() string {
	if e.APIError != nil {
		return fmt.Sprintf("status: %d, code: %d, message: %s", e.StatusCode, e.APIError.Code, e.APIError.Message)
	}
	return e.RawMessage
}
