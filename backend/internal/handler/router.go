package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/660710627/my-research/internal/service"
	"github.com/gin-gonic/gin"
)

type HealthChecker interface {
	Check(context.Context) error
}

type Dependencies struct {
	Health HealthChecker
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRouter(dependencies Dependencies) *gin.Engine {
	router := gin.New()
	router.HandleMethodNotAllowed = true

	router.GET("/health", func(ctx *gin.Context) {
		err := dependencies.Health.Check(ctx.Request.Context())
		switch {
		case err == nil:
			ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
		case errors.Is(err, service.ErrServiceUnavailable):
			writeError(ctx, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is temporarily unavailable.")
		default:
			writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
		}
	})

	router.NoRoute(func(ctx *gin.Context) {
		writeError(ctx, http.StatusNotFound, "ROUTE_NOT_FOUND", "Route was not found.")
	})
	router.NoMethod(func(ctx *gin.Context) {
		writeError(ctx, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method is not allowed.")
	})

	return router
}

func writeError(ctx *gin.Context, status int, code, message string) {
	ctx.AbortWithStatusJSON(status, errorBody{
		Error: errorDetail{Code: code, Message: message},
	})
}
