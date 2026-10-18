package http

import (
	"errors"
	"net/http"

	"github.com/PavelLymin/diploma/pkg/apperrors"
)

type Details struct {
	Message string `json:"message,omitempty"`
	Field   string `json:"field,omitempty"`
}

type ErrorResponse struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Details   *Details `json:"details,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
}

func MapError(err error) (int, ErrorResponse) {
	var status int
	var response ErrorResponse

	switch {
	case errors.Is(err, apperrors.ErrUnauthorized):
		status = http.StatusUnauthorized
		response.Code = "unauthorized"
		response.Message = "authentication required"

	case errors.Is(err, apperrors.ErrForbidden):
		status = http.StatusForbidden
		response.Code = "forbidden"
		response.Message = "insufficient permissions"

	case errors.Is(err, apperrors.ErrNotFound):
		status = http.StatusNotFound
		response.Code = "not_found"
		response.Message = "resource not found"

	case errors.Is(err, apperrors.ErrInternalServer):
		status = http.StatusInternalServerError
		response.Code = "internal_server_error"
		response.Message = "internal server error"

	default:
		status = http.StatusInternalServerError
		response.Code = "internal_server_error"
		response.Message = "internal server error"
	}

	return status, response
}
