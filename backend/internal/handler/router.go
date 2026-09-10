package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	codeInternalError      = "INTERNAL_ERROR"
	codeMethodNotAllowed   = "METHOD_NOT_ALLOWED"
	codeRouteNotFound      = "ROUTE_NOT_FOUND"
	codeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

type HealthChecker interface {
	CheckHealth(context.Context) error
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRouter(health HealthChecker) http.Handler {
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(recoverAsJSON())

	router.GET("/health", func(c *gin.Context) {
		if err := health.CheckHealth(c.Request.Context()); err != nil {
			writeHealthError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.NoRoute(func(c *gin.Context) {
		writeError(c, http.StatusNotFound, codeRouteNotFound, "Route was not found.")
	})
	router.NoMethod(func(c *gin.Context) {
		writeError(c, http.StatusMethodNotAllowed, codeMethodNotAllowed, "Method is not allowed for this route.")
	})

	return router
}

func writeHealthError(c *gin.Context, err error) {
	var coded interface {
		ErrorCode() string
	}
	if errors.As(err, &coded) && coded.ErrorCode() == codeServiceUnavailable {
		writeError(c, http.StatusServiceUnavailable, codeServiceUnavailable, "Service is temporarily unavailable.")
		return
	}
	writeError(c, http.StatusInternalServerError, codeInternalError, "An unexpected internal error occurred.")
}

func writeError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

func recoverAsJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recover() != nil {
				writeError(c, http.StatusInternalServerError, codeInternalError, "An unexpected internal error occurred.")
			}
		}()
		c.Next()
	}
}
