package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/660710627/my-research/internal/service"
	"github.com/gin-gonic/gin"
)

type HealthChecker interface {
	Check(context.Context) error
}

type ResearchCreator interface {
	Create(context.Context, service.CreateResearchInput) (service.Research, error)
}

type ResearchLister interface {
	List(context.Context) ([]service.Research, error)
}

type ResearchUpdater interface {
	Update(context.Context, service.UpdateResearchInput) (service.Research, error)
}

type ResearchDeleter interface {
	Delete(context.Context, int64) error
}

type Dependencies struct {
	Health         HealthChecker
	Researches     ResearchCreator
	ResearchList   ResearchLister
	ResearchUpdate ResearchUpdater
	ResearchDelete ResearchDeleter
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
		if dependencies.Health == nil {
			writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
			return
		}
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

	router.POST("/api/v1/researches", func(ctx *gin.Context) {
		handleCreateResearch(ctx, dependencies.Researches)
	})
	router.GET("/api/v1/researches", func(ctx *gin.Context) {
		handleListResearches(ctx, dependencies.ResearchList)
	})
	router.PUT("/api/v1/researches/:id", func(ctx *gin.Context) {
		handleUpdateResearch(ctx, dependencies.ResearchUpdate)
	})
	router.DELETE("/api/v1/researches/:id", func(ctx *gin.Context) {
		handleDeleteResearch(ctx, dependencies.ResearchDelete)
	})

	router.NoRoute(func(ctx *gin.Context) {
		writeError(ctx, http.StatusNotFound, "ROUTE_NOT_FOUND", "Route was not found.")
	})
	router.NoMethod(func(ctx *gin.Context) {
		writeError(ctx, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method is not allowed.")
	})

	return router
}

func handleDeleteResearch(ctx *gin.Context, deleter ResearchDeleter) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Research ID must be a positive integer.")
		return
	}
	if ctx.Request.URL.RawQuery != "" {
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Query parameters are not supported.")
		return
	}
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 1))
	if err != nil || len(body) != 0 {
		writeError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Request body must be empty.")
		return
	}
	if deleter == nil {
		writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
		return
	}

	err = deleter.Delete(ctx.Request.Context(), id)
	switch {
	case err == nil:
		ctx.Status(http.StatusNoContent)
	case errors.Is(err, service.ErrResearchNotFound):
		writeError(ctx, http.StatusNotFound, "RESEARCH_NOT_FOUND", "Research was not found.")
	case errors.Is(err, service.ErrResearchHasContinuations):
		writeError(ctx, http.StatusConflict, "RESEARCH_HAS_CONTINUATIONS", "Research has continuations and cannot be deleted.")
	case errors.Is(err, service.ErrValidation):
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request values failed validation.")
	default:
		writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
	}
}

func handleListResearches(ctx *gin.Context, lister ResearchLister) {
	if ctx.Request.URL.RawQuery != "" {
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Query parameters are not supported.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 1))
	if err != nil || len(body) != 0 {
		writeError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Request body must be empty.")
		return
	}
	if lister == nil {
		writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
		return
	}

	researches, err := lister.List(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
		return
	}
	ctx.JSON(http.StatusOK, researches)
}

const maxCreateResearchBody = 64 * 1024

var errInvalidRequestBody = errors.New("invalid request body")

func handleCreateResearch(ctx *gin.Context, creator ResearchCreator) {
	mediaType, _, err := mime.ParseMediaType(ctx.GetHeader("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		writeError(ctx, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxCreateResearchBody+1))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "INVALID_JSON", "Request body must contain valid JSON.")
		return
	}
	if len(body) > maxCreateResearchBody {
		writeError(ctx, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Request body exceeds 64 KiB.")
		return
	}
	if !json.Valid(body) {
		writeError(ctx, http.StatusBadRequest, "INVALID_JSON", "Request body must contain exactly one valid JSON value.")
		return
	}

	input, err := decodeCreateResearchInput(body)
	if err != nil {
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body does not match the required schema.")
		return
	}
	if creator == nil {
		writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
		return
	}

	created, err := creator.Create(ctx.Request.Context(), input)
	switch {
	case err == nil:
		ctx.JSON(http.StatusCreated, created)
	case errors.Is(err, service.ErrContinuationNotFound):
		writeError(ctx, http.StatusNotFound, "CONTINUATION_NOT_FOUND", "Continuation research was not found.")
	case errors.Is(err, service.ErrTitleAlreadyExists):
		writeError(ctx, http.StatusConflict, "TITLE_ALREADY_EXISTS", "Research title already exists.")
	case errors.Is(err, service.ErrValidation):
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request values failed validation.")
	default:
		writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
	}
}

func handleUpdateResearch(ctx *gin.Context, updater ResearchUpdater) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Research ID must be a positive integer.")
		return
	}

	mediaType, _, err := mime.ParseMediaType(ctx.GetHeader("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		writeError(ctx, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxCreateResearchBody+1))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "INVALID_JSON", "Request body must contain valid JSON.")
		return
	}
	if len(body) > maxCreateResearchBody {
		writeError(ctx, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Request body exceeds 64 KiB.")
		return
	}
	if !json.Valid(body) {
		writeError(ctx, http.StatusBadRequest, "INVALID_JSON", "Request body must contain exactly one valid JSON value.")
		return
	}

	input, err := decodeUpdateResearchInput(body)
	if err != nil {
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body does not match the required schema.")
		return
	}
	input.ID = id
	if updater == nil {
		writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
		return
	}

	updated, err := updater.Update(ctx.Request.Context(), input)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, updated)
	case errors.Is(err, service.ErrResearchNotFound):
		writeError(ctx, http.StatusNotFound, "RESEARCH_NOT_FOUND", "Research was not found.")
	case errors.Is(err, service.ErrTitleAlreadyExists):
		writeError(ctx, http.StatusConflict, "TITLE_ALREADY_EXISTS", "Research title already exists.")
	case errors.Is(err, service.ErrValidation):
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request values failed validation.")
	default:
		writeError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred.")
	}
}

func decodeCreateResearchInput(body []byte) (service.CreateResearchInput, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	first, err := decoder.Token()
	if err != nil || first != json.Delim('{') {
		return service.CreateResearchInput{}, errInvalidRequestBody
	}

	fields := make(map[string]json.RawMessage, 3)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return service.CreateResearchInput{}, errInvalidRequestBody
		}
		key, ok := keyToken.(string)
		if !ok {
			return service.CreateResearchInput{}, errInvalidRequestBody
		}
		if _, duplicate := fields[key]; duplicate {
			return service.CreateResearchInput{}, errInvalidRequestBody
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return service.CreateResearchInput{}, errInvalidRequestBody
		}
		fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return service.CreateResearchInput{}, errInvalidRequestBody
	}

	if len(fields) != 3 {
		return service.CreateResearchInput{}, errInvalidRequestBody
	}
	titleJSON, titleFound := fields["title"]
	descriptionJSON, descriptionFound := fields["description"]
	continuationJSON, continuationFound := fields["continuationOfId"]
	if !titleFound || !descriptionFound || !continuationFound {
		return service.CreateResearchInput{}, errInvalidRequestBody
	}

	var input service.CreateResearchInput
	if err := json.Unmarshal(titleJSON, &input.Title); err != nil || bytes.Equal(titleJSON, []byte("null")) {
		return service.CreateResearchInput{}, errInvalidRequestBody
	}
	if err := json.Unmarshal(descriptionJSON, &input.Description); err != nil || bytes.Equal(descriptionJSON, []byte("null")) {
		return service.CreateResearchInput{}, errInvalidRequestBody
	}
	if !bytes.Equal(bytes.TrimSpace(continuationJSON), []byte("null")) {
		var continuation int64
		if err := json.Unmarshal(continuationJSON, &continuation); err != nil {
			return service.CreateResearchInput{}, errInvalidRequestBody
		}
		input.ContinuationOfID = &continuation
	}
	return input, nil
}

func decodeUpdateResearchInput(body []byte) (service.UpdateResearchInput, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	first, err := decoder.Token()
	if err != nil || first != json.Delim('{') {
		return service.UpdateResearchInput{}, errInvalidRequestBody
	}

	fields := make(map[string]json.RawMessage, 2)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return service.UpdateResearchInput{}, errInvalidRequestBody
		}
		key, ok := keyToken.(string)
		if !ok {
			return service.UpdateResearchInput{}, errInvalidRequestBody
		}
		if _, duplicate := fields[key]; duplicate {
			return service.UpdateResearchInput{}, errInvalidRequestBody
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return service.UpdateResearchInput{}, errInvalidRequestBody
		}
		fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return service.UpdateResearchInput{}, errInvalidRequestBody
	}
	if len(fields) != 2 {
		return service.UpdateResearchInput{}, errInvalidRequestBody
	}

	titleJSON, titleFound := fields["title"]
	descriptionJSON, descriptionFound := fields["description"]
	if !titleFound || !descriptionFound {
		return service.UpdateResearchInput{}, errInvalidRequestBody
	}

	var input service.UpdateResearchInput
	if err := json.Unmarshal(titleJSON, &input.Title); err != nil || bytes.Equal(bytes.TrimSpace(titleJSON), []byte("null")) {
		return service.UpdateResearchInput{}, errInvalidRequestBody
	}
	if err := json.Unmarshal(descriptionJSON, &input.Description); err != nil || bytes.Equal(bytes.TrimSpace(descriptionJSON), []byte("null")) {
		return service.UpdateResearchInput{}, errInvalidRequestBody
	}
	return input, nil
}

func writeError(ctx *gin.Context, status int, code, message string) {
	ctx.AbortWithStatusJSON(status, errorBody{
		Error: errorDetail{Code: code, Message: message},
	})
}
