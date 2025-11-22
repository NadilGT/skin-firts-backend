package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Operation string `json:"operation"`
	Error     string `json:"error,omitempty"`
}

// SendErrorResponse sends a response with an error message
func SendErrorResponse(c *fiber.Ctx, statusCode int, errorMessage string) error {
	response := Response{
		Operation: "Failed",
		Error:     errorMessage,
	}
	return c.Status(statusCode).JSON(response)
}

// SendSuccessResponse sends a response for a successful operation
func SendSuccessResponse(c *fiber.Ctx) error {
	response := Response{
		Operation: "Success",
		Error:     "",
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

// SendSuccessResponseWithData sends a response for a successful operation with data
func SendSuccessResponseWithData(c *fiber.Ctx, data interface{}) error {
	response := map[string]interface{}{
		"operation": "Success",
		"data":      data,
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

type CustomError struct {
	Code    int    `json:"-"  `      // HTTP status code to send to the client (ignored during JSON marshaling)
	Message string `json:"message" ` // User-friendly message to send to the client
	Err     error  `json:"-" `       // The actual underlying error (e.g., database error), not exposed in JSON
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *CustomError) Unwrap() error {
	return e.Err
}

func NewCustomError(code int, message string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
