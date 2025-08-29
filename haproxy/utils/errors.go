package utils

import "fmt"

// APIError represents an error from the HAProxy API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CustomError wraps API errors with additional context
type CustomError struct {
	Message  string
	APIError *APIError
}

func (e *CustomError) Error() string {
	if e.APIError != nil {
		return fmt.Sprintf("API Error %d: %s", e.APIError.Code, e.APIError.Message)
	}
	return e.Message
}

// NewCustomError creates a new CustomError
func NewCustomError(message string, apiError *APIError) *CustomError {
	return &CustomError{
		Message:  message,
		APIError: apiError,
	}
}

