package handler

import (
	"net/http"

	apperrors "github.com/eneskaya/insider-messaging/pkg/errors"
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		statusCode := getHTTPStatusCode(appErr.Code)
		c.JSON(statusCode, ErrorResponse{
			Error: appErr.Message,
			Code:  string(appErr.Code),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "internal server error",
		Code:  string(apperrors.ErrorCodeInternal),
	})
}

func getHTTPStatusCode(code apperrors.ErrorCode) int {
	switch code {
	case apperrors.ErrorCodeValidation:
		return http.StatusBadRequest
	case apperrors.ErrorCodeNotFound:
		return http.StatusNotFound
	case apperrors.ErrorCodeAlreadyExists:
		return http.StatusConflict
	case apperrors.ErrorCodeTimeout:
		return http.StatusRequestTimeout
	case apperrors.ErrorCodeRateLimit:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
