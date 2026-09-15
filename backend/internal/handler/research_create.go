package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/660710627/my-research/internal/service"
	"github.com/gin-gonic/gin"
)

const maxMultipartBodyBytes int64 = 22_020_096

var createPartNames = map[string]struct{}{
	"title": {}, "continuationOfId": {}, "isSubsidized": {},
	"projectMembers": {}, "fundingType": {}, "fundingSourceName": {},
	"contractNumber": {}, "contractFile": {}, "projectType": {},
	"researchKind": {}, "responsibleProjectUnit": {},
	"responsibleBudgetUnit": {}, "startDate": {}, "endDate": {},
	"budgetAmount": {}, "thaiAbstract": {}, "englishAbstract": {},
	"objectives": {}, "keywords": {},
}

type createRequestError struct {
	status  int
	code    string
	message string
	field   string
}

type contractFileResponse struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

type createResearchResponse struct {
	ID                     int64                         `json:"id"`
	Title                  string                        `json:"title"`
	ContinuationOfID       *int64                        `json:"continuationOfId"`
	IsSubsidized           bool                          `json:"isSubsidized"`
	ProjectMembers         []service.ResearchMemberInput `json:"projectMembers"`
	FundingType            string                        `json:"fundingType"`
	FundingSourceName      string                        `json:"fundingSourceName"`
	ContractNumber         string                        `json:"contractNumber"`
	ContractFile           contractFileResponse          `json:"contractFile"`
	ProjectType            string                        `json:"projectType"`
	ResearchKind           string                        `json:"researchKind"`
	ResponsibleProjectUnit string                        `json:"responsibleProjectUnit"`
	ResponsibleBudgetUnit  string                        `json:"responsibleBudgetUnit"`
	StartDate              string                        `json:"startDate"`
	EndDate                string                        `json:"endDate"`
	BudgetAmount           float64                       `json:"budgetAmount"`
	ThaiAbstract           string                        `json:"thaiAbstract"`
	EnglishAbstract        string                        `json:"englishAbstract"`
	Objectives             string                        `json:"objectives"`
	Keywords               string                        `json:"keywords"`
	Status                 string                        `json:"status"`
	Process                string                        `json:"process"`
}

func createResearchHandler(creator ResearchCreator) gin.HandlerFunc {
	return func(c *gin.Context) {
		input, filename, fileSize, requestErr := decodeCreateRequest(c.Request)
		if requestErr != nil {
			writeCreateRequestError(c, requestErr)
			return
		}
		created, err := creator.Create(c.Request.Context(), input)
		if err != nil {
			writeCreateServiceError(c, err)
			return
		}

		members := make([]service.ResearchMemberInput, len(input.ProjectMembers))
		for index, member := range input.ProjectMembers {
			member.FullName = strings.TrimSpace(member.FullName)
			member.Email = strings.TrimSpace(member.Email)
			member.Affiliation = strings.TrimSpace(member.Affiliation)
			members[index] = member
		}
		budgetAmount, _ := strconv.ParseFloat(string(input.BudgetAmount), 64)

		c.JSON(http.StatusCreated, createResearchResponse{
			ID: created.ID, Title: created.Title,
			ContinuationOfID: input.ContinuationOfID,
			IsSubsidized:     *input.IsSubsidized, ProjectMembers: members,
			FundingType:       input.FundingType,
			FundingSourceName: strings.TrimSpace(input.FundingSourceName),
			ContractNumber:    strings.TrimSpace(input.ContractNumber),
			ContractFile: contractFileResponse{
				Filename: strings.TrimSpace(filename), ContentType: "application/pdf", SizeBytes: fileSize,
			},
			ProjectType: input.ProjectType, ResearchKind: input.ResearchKind,
			ResponsibleProjectUnit: strings.TrimSpace(input.ResponsibleProjectUnit),
			ResponsibleBudgetUnit:  strings.TrimSpace(input.ResponsibleBudgetUnit),
			StartDate:              created.StartDate, EndDate: created.EndDate,
			BudgetAmount:    budgetAmount,
			ThaiAbstract:    strings.TrimSpace(input.ThaiAbstract),
			EnglishAbstract: strings.TrimSpace(input.EnglishAbstract),
			Objectives:      strings.TrimSpace(input.Objectives),
			Keywords:        strings.TrimSpace(input.Keywords),
			Status:          created.Status, Process: created.Process,
		})
	}
}

func decodeCreateRequest(request *http.Request) (service.CreateResearchInput, string, int64, *createRequestError) {
	if request.URL.RawQuery != "" {
		return service.CreateResearchInput{}, "", 0, validationRequestError("query", "Query parameters are not supported.")
	}
	mediaType, parameters, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "multipart/form-data") {
		return service.CreateResearchInput{}, "", 0, &createRequestError{status: http.StatusUnsupportedMediaType, code: "UNSUPPORTED_MEDIA_TYPE", message: "Content-Type is not supported for this operation."}
	}
	boundary := parameters["boundary"]
	if strings.TrimSpace(boundary) == "" {
		return service.CreateResearchInput{}, "", 0, validationRequestError("body", "Multipart boundary is required.")
	}
	if request.ContentLength > maxMultipartBodyBytes {
		return service.CreateResearchInput{}, "", 0, payloadTooLargeRequestError()
	}
	body, err := io.ReadAll(io.LimitReader(request.Body, maxMultipartBodyBytes+1))
	if err != nil {
		return service.CreateResearchInput{}, "", 0, validationRequestError("body", "Multipart body could not be read.")
	}
	if int64(len(body)) > maxMultipartBodyBytes {
		return service.CreateResearchInput{}, "", 0, payloadTooLargeRequestError()
	}

	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	values := make(map[string][]byte, len(createPartNames))
	filenames := make(map[string]string)
	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return service.CreateResearchInput{}, "", 0, validationRequestError("body", "Multipart body is malformed.")
		}
		name := part.FormName()
		if _, allowed := createPartNames[name]; !allowed {
			_ = part.Close()
			return service.CreateResearchInput{}, "", 0, validationRequestError(nameOrBody(name), "Unknown form part.")
		}
		if _, duplicate := values[name]; duplicate {
			_ = part.Close()
			return service.CreateResearchInput{}, "", 0, validationRequestError(name, "Form part must occur exactly once.")
		}
		if name == "projectMembers" && !partMediaTypeIs(part, "application/json") {
			_ = part.Close()
			return service.CreateResearchInput{}, "", 0, validationRequestError(name, "Form part must use application/json.")
		}
		if name == "contractFile" && !partMediaTypeIs(part, "application/pdf") {
			_ = part.Close()
			return service.CreateResearchInput{}, "", 0, validationRequestError(name, "Contract file must use application/pdf.")
		}
		value, readErr := io.ReadAll(part)
		_ = part.Close()
		if readErr != nil {
			return service.CreateResearchInput{}, "", 0, validationRequestError(name, "Form part could not be read.")
		}
		values[name] = value
		filenames[name] = part.FileName()
	}
	for name := range createPartNames {
		if _, present := values[name]; !present {
			return service.CreateResearchInput{}, "", 0, validationRequestError(name, "This field is required.")
		}
	}

	input, parseErr := parseCreateValues(values, filenames)
	if parseErr != nil {
		return service.CreateResearchInput{}, "", 0, parseErr
	}
	return input, filenames["contractFile"], int64(len(values["contractFile"])), nil
}

func parseCreateValues(values map[string][]byte, filenames map[string]string) (service.CreateResearchInput, *createRequestError) {
	input := service.CreateResearchInput{
		Title: string(values["title"]), FundingType: string(values["fundingType"]),
		FundingSourceName: string(values["fundingSourceName"]),
		ContractNumber:    string(values["contractNumber"]), ProjectType: string(values["projectType"]),
		ResearchKind:           string(values["researchKind"]),
		ResponsibleProjectUnit: string(values["responsibleProjectUnit"]),
		ResponsibleBudgetUnit:  string(values["responsibleBudgetUnit"]),
		StartDate:              string(values["startDate"]), EndDate: string(values["endDate"]),
		BudgetAmount: json.Number(string(values["budgetAmount"])),
		ThaiAbstract: string(values["thaiAbstract"]), EnglishAbstract: string(values["englishAbstract"]),
		Objectives: string(values["objectives"]), Keywords: string(values["keywords"]),
		ContinuationOfIDPresent: true,
	}

	switch string(values["isSubsidized"]) {
	case "true":
		value := true
		input.IsSubsidized = &value
	case "false":
		value := false
		input.IsSubsidized = &value
	default:
		return service.CreateResearchInput{}, validationRequestError("isSubsidized", "Value must be true or false.")
	}

	continuation := string(values["continuationOfId"])
	if continuation != "null" {
		value, err := strconv.ParseInt(continuation, 10, 64)
		if err != nil {
			return service.CreateResearchInput{}, validationRequestError("continuationOfId", "Value must be null or a positive integer.")
		}
		input.ContinuationOfID = &value
	}

	if err := decodeMemberJSON(values["projectMembers"], &input.ProjectMembers); err != nil {
		return service.CreateResearchInput{}, validationRequestError("projectMembers", "Value must be one valid project member array.")
	}
	input.ContractFile = &service.ContractUpload{
		Filename: filenames["contractFile"], Reader: bytes.NewReader(values["contractFile"]),
	}
	return input, nil
}

func decodeMemberJSON(data []byte, target *[]service.ResearchMemberInput) error {
	duplicateDecoder := json.NewDecoder(bytes.NewReader(data))
	duplicateDecoder.UseNumber()
	if err := consumeJSONValue(duplicateDecoder); err != nil {
		return err
	}
	if _, err := duplicateDecoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	var rawMembers []map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMembers); err != nil {
		return err
	}
	for _, member := range rawMembers {
		raw, present := member["contributionPercent"]
		if present && len(bytes.TrimSpace(raw)) > 0 && bytes.TrimSpace(raw)[0] == '"' {
			return errors.New("contributionPercent must be a JSON number")
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func consumeJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		keys := map[string]struct{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("invalid object key")
			}
			if _, duplicate := keys[key]; duplicate {
				return errors.New("duplicate object key")
			}
			keys[key] = struct{}{}
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return errors.New("invalid object")
		}
	case '[':
		for decoder.More() {
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return errors.New("invalid array")
		}
	default:
		return errors.New("unexpected delimiter")
	}
	return nil
}

func writeCreateRequestError(c *gin.Context, requestErr *createRequestError) {
	if requestErr.code == "VALIDATION_ERROR" {
		c.AbortWithStatusJSON(requestErr.status, errorResponse{Error: errorBody{
			Code: requestErr.code, Message: requestErr.message,
			FieldErrors: []fieldErrorBody{{Field: requestErr.field, Message: requestErr.message}},
		}})
		return
	}
	writeError(c, requestErr.status, requestErr.code, requestErr.message)
}

func writeCreateServiceError(c *gin.Context, err error) {
	coded, ok := err.(interface{ ErrorCode() string })
	if !ok {
		writeError(c, http.StatusInternalServerError, codeInternalError, "An unexpected internal error occurred.")
		return
	}
	code := coded.ErrorCode()
	switch code {
	case "VALIDATION_ERROR":
		fieldErrors := []fieldErrorBody{{Field: "body", Message: "Request validation failed."}}
		if detailed, ok := err.(interface{ FieldErrors() []service.FieldError }); ok {
			fields := detailed.FieldErrors()
			if len(fields) > 0 {
				fieldErrors = make([]fieldErrorBody, len(fields))
				for index, field := range fields {
					fieldErrors[index] = fieldErrorBody{Field: field.Field, Message: field.Message}
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, errorResponse{Error: errorBody{
			Code: code, Message: "Request validation failed.", FieldErrors: fieldErrors,
		}})
	case "CONTINUATION_NOT_FOUND":
		writeError(c, http.StatusNotFound, code, "Continuation research was not found.")
	case "TITLE_ALREADY_EXISTS":
		writeError(c, http.StatusConflict, code, "A research with this title already exists.")
	case "CONTRACT_NUMBER_ALREADY_EXISTS":
		writeError(c, http.StatusConflict, code, "A research with this contract number already exists.")
	case "PAYLOAD_TOO_LARGE":
		writeError(c, http.StatusRequestEntityTooLarge, code, "Request payload exceeds the allowed size.")
	default:
		writeError(c, http.StatusInternalServerError, codeInternalError, "An unexpected internal error occurred.")
	}
}

func validationRequestError(field, message string) *createRequestError {
	return &createRequestError{status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR", message: message, field: field}
}

func payloadTooLargeRequestError() *createRequestError {
	return &createRequestError{status: http.StatusRequestEntityTooLarge, code: "PAYLOAD_TOO_LARGE", message: "Request payload exceeds the allowed size."}
}

func nameOrBody(name string) string {
	if name == "" {
		return "body"
	}
	return name
}

func partMediaTypeIs(part *multipart.Part, expected string) bool {
	mediaType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
	return err == nil && strings.EqualFold(mediaType, expected)
}
