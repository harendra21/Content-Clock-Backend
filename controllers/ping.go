package controllers

import (
	"content-clock/helpers"

	"github.com/pocketbase/pocketbase/core"
)

// @Summary Health Check Endpoint
// @Schemes
// @Description Simple ping endpoint to check if the API is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} helpers.SuccessResponse "ping success"
// @Failure 400 {object} helpers.ErrorResponse "error"
// @Router /ping [get]
func Ping(e *core.RequestEvent) {
	helpers.Success(e, "Ping success", nil)
}
