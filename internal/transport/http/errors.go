package httptransport

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-081/internal/domain/shared"
	"github.com/wyw14/cry-081/internal/middleware"
)

type errorResponse struct {
	Code      string                  `json:"code"`
	Message   string                  `json:"message"`
	Fields    []shared.FieldViolation `json:"fields,omitempty"`
	RequestID string                  `json:"request_id"`
}

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	response := errorResponse{Code: "INTERNAL_ERROR", Message: "the request could not be completed", RequestID: middleware.RequestIDFrom(c)}
	var domainError *shared.DomainError
	if errors.As(err, &domainError) {
		response.Code = domainError.Code
		response.Message = domainError.Message
		response.Fields = domainError.Violations
	}
	switch {
	case errors.Is(err, shared.ErrValidation):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, shared.ErrUnauthenticated):
		status = http.StatusUnauthorized
	case errors.Is(err, shared.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, shared.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, shared.ErrConflict), errors.Is(err, shared.ErrInvalidState), errors.Is(err, shared.ErrVersionConflict):
		status = http.StatusConflict
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		status = http.StatusRequestTimeout
	}
	c.AbortWithStatusJSON(status, response)
}
