package helpers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

type SuccessResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func Success(e *core.RequestEvent, message string, data interface{}) {
	var successResponse SuccessResponse
	successResponse.Status = true
	successResponse.Message = message
	successResponse.Data = data
	e.JSON(http.StatusOK, successResponse)

}

func Error(e *core.RequestEvent, message string) {
	var errorResponse ErrorResponse
	errorResponse.Status = false
	errorResponse.Message = message
	Logging("error", message)
	e.JSON(http.StatusOK, errorResponse)

}
