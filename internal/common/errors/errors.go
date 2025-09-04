package errors

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func NewAPIError(code int, message, details string) *APIError {
    return &APIError{
        Code:    code,
        Message: message,
        Details: details,
    }
}

func (e *APIError) Error() string {
    return e.Message
}

var (
    ErrInternal     = NewAPIError(http.StatusInternalServerError, "Internal Server Error", "")
    ErrNotFound     = NewAPIError(http.StatusNotFound, "Resource Not Found", "")
    ErrUnauthorized = NewAPIError(http.StatusUnauthorized, "Unauthorized", "")
    ErrBadRequest   = NewAPIError(http.StatusBadRequest, "Bad Request", "")
    ErrForbidden    = NewAPIError(http.StatusForbidden, "Forbidden", "")
    ErrValidation   = NewAPIError(http.StatusBadRequest, "Validation Failed", "")
)

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            if apiErr, ok := err.(*APIError); ok {
                c.JSON(apiErr.Code, apiErr)
                return
            }
            c.JSON(http.StatusInternalServerError, ErrInternal)
        }
    }
}
