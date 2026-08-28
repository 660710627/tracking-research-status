package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"mime"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/660710627/my-research/internal/domain"
	"github.com/660710627/my-research/internal/service"
	"github.com/gin-gonic/gin"
)

type HealthChecker interface{ Check(context.Context) error }
type ResearchDeleter interface {
	Delete(context.Context, int64) error
}
type ResearchLister interface {
	List(context.Context) ([]service.Research, error)
}
type ResearchStatusUpdater interface {
	UpdateStatus(context.Context, service.UpdateResearchStatusInput) (service.Research, error)
}
type ResearchProcessUpdater interface {
	UpdateProcess(context.Context, service.UpdateResearchProcessInput) (service.Research, error)
}
type Dependencies struct {
	Health            HealthChecker
	MultipartResearch *MultipartResearchHandler
	ResearchList      ResearchLister
	ResearchDelete    ResearchDeleter
	ResearchStatus    ResearchStatusUpdater
	ResearchProcess   ResearchProcessUpdater
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
			writeError(ctx, 500, "INTERNAL_ERROR", "An unexpected internal error occurred.")
			return
		}
		if err := dependencies.Health.Check(ctx.Request.Context()); err != nil {
			if errors.Is(err, service.ErrServiceUnavailable) {
				writeError(ctx, 503, "SERVICE_UNAVAILABLE", "Service is temporarily unavailable.")
			} else {
				writeError(ctx, 500, "INTERNAL_ERROR", "An unexpected internal error occurred.")
			}
			return
		}
		ctx.JSON(200, gin.H{"status": "ok"})
	})
	router.POST("/api/v1/researches", func(ctx *gin.Context) { handleMultipartCreate(ctx, dependencies.MultipartResearch) })
	router.GET("/api/v1/researches", func(ctx *gin.Context) { handleListResearches(ctx, dependencies.ResearchList) })
	router.DELETE("/api/v1/researches/:id", func(ctx *gin.Context) { handleDeleteResearch(ctx, dependencies.ResearchDelete) })
	router.PATCH("/api/v1/researches/:id/status", func(ctx *gin.Context) { handleUpdateStatus(ctx, dependencies.ResearchStatus) })
	router.PATCH("/api/v1/researches/:id/process", func(ctx *gin.Context) { handleUpdateProcess(ctx, dependencies.ResearchProcess) })
	router.NoRoute(func(ctx *gin.Context) { writeError(ctx, 404, "ROUTE_NOT_FOUND", "Route was not found.") })
	router.NoMethod(func(ctx *gin.Context) { writeError(ctx, 405, "METHOD_NOT_ALLOWED", "Method is not allowed.") })
	return router
}

func handleListResearches(ctx *gin.Context, lister ResearchLister) {
	if ctx.Request.URL.RawQuery != "" {
		writeError(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Query parameters are not supported.")
		return
	}
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 1))
	if err != nil || len(body) != 0 {
		writeError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "This operation does not accept a request body.")
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
	if researches == nil {
		researches = []service.Research{}
	}
	ctx.JSON(http.StatusOK, researches)
}

func handleMultipartCreate(ctx *gin.Context, creator *MultipartResearchHandler) {
	mediaType, _, err := mime.ParseMediaType(ctx.GetHeader("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "multipart/form-data") {
		writeError(ctx, 415, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be multipart/form-data.")
		return
	}
	input, err := parseMultipartResearch(ctx.Request)
	if err != nil {
		if errors.Is(err, errPayloadTooLarge) {
			writeError(ctx, 413, "PAYLOAD_TOO_LARGE", "Contract PDF exceeds 20 MiB.")
		} else if errors.Is(err, errUnsupportedContract) {
			writeError(ctx, 415, "UNSUPPORTED_MEDIA_TYPE", "Contract file must be application/pdf.")
		} else {
			writeError(ctx, 422, "VALIDATION_ERROR", "Multipart form does not match the required schema.")
		}
		return
	}
	if !validMultipartInput(input) {
		writeError(ctx, 422, "VALIDATION_ERROR", "Multipart form does not match the required schema.")
		return
	}
	if creator == nil {
		writeError(ctx, 500, "INTERNAL_ERROR", "An unexpected internal error occurred.")
		return
	}
	created, err := creator.Create(ctx.Request.Context(), input)
	switch {
	case err == nil:
		ctx.JSON(201, created)
	case errors.Is(err, service.ErrContinuationNotFound):
		writeError(ctx, 404, "CONTINUATION_NOT_FOUND", "Continuation research was not found.")
	case errors.Is(err, service.ErrTitleAlreadyExists):
		writeError(ctx, 409, "TITLE_ALREADY_EXISTS", "Research title already exists.")
	case errors.Is(err, service.ErrValidation):
		writeError(ctx, 422, "VALIDATION_ERROR", "Request values failed validation.")
	case errors.Is(err, service.ErrContractStage), errors.Is(err, service.ErrContractPublish):
		writeError(ctx, 500, "INTERNAL_ERROR", "An unexpected internal error occurred.")
	default:
		writeError(ctx, 500, "INTERNAL_ERROR", "An unexpected internal error occurred.")
	}
}

var errPayloadTooLarge = errors.New("payload too large")
var errUnsupportedContract = errors.New("unsupported contract")

func parseMultipartResearch(request *http.Request) (domain.CreateResearchInput, error) {
	reader, err := request.MultipartReader()
	if err != nil {
		return domain.CreateResearchInput{}, err
	}
	fields := map[string]string{}
	var upload domain.ContractUpload
	allowed := map[string]bool{"title": true, "isSubsidized": true, "projectMembers": true, "fundingType": true, "fundingSourceName": true, "contractNumber": true, "contractFile": true, "projectType": true, "researchKind": true, "responsibleProjectUnit": true, "responsibleBudgetUnit": true, "startDate": true, "endDate": true, "budgetAmount": true, "thaiAbstract": true, "englishAbstract": true, "objectives": true, "keywords": true, "continuationOfId": true}
	for {
		part, nextErr := reader.NextPart()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return domain.CreateResearchInput{}, nextErr
		}
		name := part.FormName()
		if !allowed[name] {
			return domain.CreateResearchInput{}, errors.New("unknown part")
		}
		if name == "contractFile" {
			if upload.Content != nil || !strings.EqualFold(part.Header.Get("Content-Type"), "application/pdf") {
				if !strings.EqualFold(part.Header.Get("Content-Type"), "application/pdf") {
					return domain.CreateResearchInput{}, errUnsupportedContract
				}
				return domain.CreateResearchInput{}, errors.New("duplicate file")
			}
			content, readErr := io.ReadAll(io.LimitReader(part, 20*1024*1024+1))
			if readErr != nil {
				return domain.CreateResearchInput{}, readErr
			}
			if len(content) > 20*1024*1024 {
				return domain.CreateResearchInput{}, errPayloadTooLarge
			}
			upload = domain.ContractUpload{Filename: part.FileName(), ContentType: "application/pdf", SizeBytes: int64(len(content)), Content: bytes.NewReader(content)}
			continue
		}
		if _, found := fields[name]; found {
			return domain.CreateResearchInput{}, errors.New("duplicate part")
		}
		value, readErr := io.ReadAll(io.LimitReader(part, 1024*1024))
		if readErr != nil {
			return domain.CreateResearchInput{}, readErr
		}
		fields[name] = string(value)
	}
	if len(fields) != 18 || upload.Content == nil {
		return domain.CreateResearchInput{}, errors.New("missing part")
	}
	var members []domain.ProjectMember
	decoder := json.NewDecoder(strings.NewReader(fields["projectMembers"]))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&members); err != nil {
		return domain.CreateResearchInput{}, err
	}
	if decoder.More() {
		return domain.CreateResearchInput{}, errors.New("extra members json")
	}
	var subsidized bool
	if fields["isSubsidized"] == "true" {
		subsidized = true
	} else if fields["isSubsidized"] != "false" {
		return domain.CreateResearchInput{}, errors.New("invalid boolean")
	}
	budget, err := strconv.ParseFloat(fields["budgetAmount"], 64)
	if err != nil {
		return domain.CreateResearchInput{}, err
	}
	var continuation *int64
	if fields["continuationOfId"] != "null" {
		value, err := strconv.ParseInt(fields["continuationOfId"], 10, 64)
		if err != nil {
			return domain.CreateResearchInput{}, err
		}
		continuation = &value
	}
	return domain.CreateResearchInput{ResearchData: domain.ResearchData{Title: fields["title"], ContinuationOfID: continuation, IsSubsidized: subsidized, ProjectMembers: members, FundingType: domain.FundingType(fields["fundingType"]), FundingSourceName: fields["fundingSourceName"], ContractNumber: fields["contractNumber"], ProjectType: domain.ProjectType(fields["projectType"]), ResearchKind: domain.ResearchKind(fields["researchKind"]), ResponsibleProjectUnit: fields["responsibleProjectUnit"], ResponsibleBudgetUnit: fields["responsibleBudgetUnit"], StartDate: fields["startDate"], EndDate: fields["endDate"], BudgetAmount: budget, ThaiAbstract: fields["thaiAbstract"], EnglishAbstract: fields["englishAbstract"], Objectives: fields["objectives"], Keywords: fields["keywords"]}, Contract: upload}, nil
}

func validMultipartInput(input domain.CreateResearchInput) bool {
	if (input.FundingType != domain.FundingTypeInternal && input.FundingType != domain.FundingTypeExternal) || (input.ProjectType != domain.ProjectTypeResearch && input.ProjectType != domain.ProjectTypeAcademicService) || (input.ResearchKind != domain.ResearchKindBudget && input.ResearchKind != domain.ResearchKindContinuation) || input.BudgetAmount <= 0 || math.Abs(input.BudgetAmount*100-math.Round(input.BudgetAmount*100)) > 0.000001 {
		return false
	}
	if _, err := time.Parse("02/01/2006", input.StartDate); err != nil {
		return false
	}
	if _, err := time.Parse("02/01/2006", input.EndDate); err != nil {
		return false
	}
	if input.ResearchKind == domain.ResearchKindBudget && input.ContinuationOfID != nil || input.ResearchKind == domain.ResearchKindContinuation && (input.ContinuationOfID == nil || *input.ContinuationOfID <= 0) {
		return false
	}
	lead, co := false, false
	for _, member := range input.ProjectMembers {
		if _, err := mail.ParseAddress(member.Email); err != nil || member.ContributionPercent <= 0 || member.ContributionPercent > 100 || math.Abs(member.ContributionPercent*100-math.Round(member.ContributionPercent*100)) > 0.000001 {
			return false
		}
		if member.Role == domain.MemberRoleLead {
			lead = true
		} else if member.Role == domain.MemberRoleCoResearcher {
			co = true
		} else {
			return false
		}
	}
	return lead && co
}

func handleDeleteResearch(ctx *gin.Context, deleter ResearchDeleter) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 || ctx.Request.URL.RawQuery != "" {
		writeError(ctx, 422, "VALIDATION_ERROR", "Research ID must be a positive integer.")
		return
	}
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 1))
	if err != nil || len(body) != 0 {
		writeError(ctx, 400, "INVALID_REQUEST_BODY", "Request body must be empty.")
		return
	}
	if deleter == nil {
		writeError(ctx, 500, "INTERNAL_ERROR", "An unexpected internal error occurred.")
		return
	}
	switch err := deleter.Delete(ctx.Request.Context(), id); {
	case err == nil:
		ctx.Status(204)
	case errors.Is(err, service.ErrResearchNotFound):
		writeError(ctx, 404, "RESEARCH_NOT_FOUND", "Research was not found.")
	case errors.Is(err, service.ErrResearchHasContinuations):
		writeError(ctx, 409, "RESEARCH_HAS_CONTINUATIONS", "Research has continuations and cannot be deleted.")
	default:
		writeError(ctx, 500, "INTERNAL_ERROR", "An unexpected internal error occurred.")
	}
}
func handleUpdateStatus(ctx *gin.Context, updater ResearchStatusUpdater) {
	handlePatch(ctx, func(id int64, value string) (service.Research, error) {
		if updater == nil {
			return service.Research{}, service.ErrInternal
		}
		return updater.UpdateStatus(ctx.Request.Context(), service.UpdateResearchStatusInput{ID: id, Status: value})
	}, "status")
}
func handleUpdateProcess(ctx *gin.Context, updater ResearchProcessUpdater) {
	handlePatch(ctx, func(id int64, value string) (service.Research, error) {
		if updater == nil {
			return service.Research{}, service.ErrInternal
		}
		return updater.UpdateProcess(ctx.Request.Context(), service.UpdateResearchProcessInput{ID: id, Process: value})
	}, "process")
}
func handlePatch(ctx *gin.Context, update func(int64, string) (service.Research, error), field string) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(ctx, 422, "VALIDATION_ERROR", "Research ID must be a positive integer.")
		return
	}
	media, _, err := mime.ParseMediaType(ctx.GetHeader("Content-Type"))
	if err != nil || !strings.EqualFold(media, "application/json") {
		writeError(ctx, 415, "UNSUPPORTED_MEDIA_TYPE", "Content-Type is not supported.")
		return
	}
	if ctx.Request.ContentLength == 0 {
		writeError(ctx, 400, "INVALID_JSON", "Request body must contain valid JSON.")
		return
	}
	var body map[string]string
	if err := json.NewDecoder(ctx.Request.Body).Decode(&body); err != nil || len(body) != 1 || body[field] == "" || field == "process" && !supportedProcess(body[field]) {
		writeError(ctx, 422, "VALIDATION_ERROR", "Request body does not match the required schema.")
		return
	}
	result, err := update(id, body[field])
	if err == nil {
		ctx.JSON(200, result)
		return
	}
	if errors.Is(err, service.ErrResearchNotFound) {
		writeError(ctx, 404, "RESEARCH_NOT_FOUND", "Research was not found.")
	} else if errors.Is(err, service.ErrInvalidStatusTransition) || errors.Is(err, service.ErrInvalidProcessTransition) || errors.Is(err, service.ErrProjectAlreadyEnded) {
		writeError(ctx, 409, "INVALID_STATUS_TRANSITION", "The requested transition is invalid.")
	} else {
		writeError(ctx, 500, "INTERNAL_ERROR", "An unexpected internal error occurred.")
	}
}

func supportedProcess(value string) bool {
	switch value {
	case "สัญญาโครงการ", "บันทึกข้อตกลง", "เปิดบัญชีธนาคาร", "การเบิกจ่ายเงิน", "การจัดสรรค่าธรรมเนียม", "การติดตามส่งรายงาน", "รายงานสรุปการใช้เงิน", "การปิดบัญชีธนาคาร":
		return true
	default:
		return false
	}
}
func writeError(ctx *gin.Context, status int, code, message string) {
	ctx.AbortWithStatusJSON(status, errorBody{Error: errorDetail{Code: code, Message: message}})
}
