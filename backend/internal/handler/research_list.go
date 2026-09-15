package handler

import (
	"context"
	"github.com/660710627/my-research/internal/service"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type ResearchLister interface {
	List(context.Context) ([]service.Research, error)
}

func WithResearchLister(lister ResearchLister) Option {
	return func(config *routerConfig) { config.researchLister = lister }
}
func listResearchHandler(lister ResearchLister) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			var probe [1]byte
			n, err := io.ReadFull(c.Request.Body, probe[:])
			if n > 0 || (err != nil && err != io.EOF) {
				writeError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "This operation does not accept a request body.")
				return
			}
		}
		if c.Request.URL.RawQuery != "" {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, errorResponse{Error: errorBody{Code: "VALIDATION_ERROR", Message: "Request validation failed.", FieldErrors: []fieldErrorBody{{Field: "query", Message: "This operation does not accept query parameters."}}}})
			return
		}
		items, err := lister.List(c.Request.Context())
		if err != nil {
			writeError(c, http.StatusInternalServerError, codeInternalError, "An unexpected internal error occurred.")
			return
		}
		if items == nil {
			items = make([]service.Research, 0)
		}
		c.JSON(http.StatusOK, items)
	}
}
