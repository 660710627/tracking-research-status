package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/domain"
	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/service"
)

func TestCreateResearchMultipartReturnsCompletePersistedResearch(t *testing.T) {
	creator := multipartResearchCreatorStub{create: func(_ context.Context, input domain.CreateResearchInput) (domain.Research, error) {
		return domain.Research{ID: 71, ResearchData: input.ResearchData, Contract: domain.ContractMetadata{Filename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 4}, Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ"}, nil
	}}
	router := multipartCreateRouter(creator)
	response := performMultipartCreate(t, router, multipartRequestOptions{})
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response)
	var body struct {
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		Status   string `json:"status"`
		Process  string `json:"process"`
		Contract struct {
			Filename    string `json:"filename"`
			ContentType string `json:"contentType"`
			SizeBytes   int64  `json:"sizeBytes"`
		} `json:"contractFile"`
		Members []domain.ProjectMember `json:"projectMembers"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v; body = %s", err, response.Body.String())
	}
	if body.ID <= 0 || body.Title != "โครงการใหม่" || len(body.Members) != 2 || body.Contract.Filename != "contract.pdf" || body.Contract.ContentType != "application/pdf" || body.Contract.SizeBytes != 4 || body.Status != "กำลังดำเนินการ" || body.Process != "สัญญาโครงการ" {
		t.Fatalf("response = %#v, want complete new research without PDF binary", body)
	}
}

func TestCreateResearchMultipartRejectsMissingUnknownAndDuplicateParts(t *testing.T) {
	required := []string{"title", "isSubsidized", "projectMembers", "fundingType", "fundingSourceName", "contractNumber", "contractFile", "projectType", "researchKind", "responsibleProjectUnit", "responsibleBudgetUnit", "startDate", "endDate", "budgetAmount", "thaiAbstract", "englishAbstract", "objectives", "keywords", "continuationOfId"}
	for _, field := range required {
		t.Run("missing "+field, func(t *testing.T) {
			response := performMultipartCreate(t, multipartCreateRouter(multipartResearchCreatorStub{}), multipartRequestOptions{omit: map[string]bool{field: true}})
			assertMultipartCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		})
		t.Run("duplicate "+field, func(t *testing.T) {
			response := performMultipartCreate(t, multipartCreateRouter(multipartResearchCreatorStub{}), multipartRequestOptions{duplicate: field})
			assertMultipartCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		})
	}
	t.Run("unknown form part", func(t *testing.T) {
		response := performMultipartCreate(t, multipartCreateRouter(multipartResearchCreatorStub{}), multipartRequestOptions{unknown: true})
		assertMultipartCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	})
}

func TestCreateResearchMultipartRejectsInvalidFieldsMembersAndContract(t *testing.T) {
	tests := []struct {
		name    string
		options multipartRequestOptions
		status  int
		code    string
	}{
		{name: "invalid members JSON", options: multipartRequestOptions{override: map[string]string{"projectMembers": "["}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "member without lead", options: multipartRequestOptions{override: map[string]string{"projectMembers": `[{"fullName":"A","email":"a@example.com","affiliation":"X","contributionPercent":50,"role":"CO_RESEARCHER"},{"fullName":"B","email":"b@example.com","affiliation":"X","contributionPercent":50,"role":"CO_RESEARCHER"}]`}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "member without co researcher", options: multipartRequestOptions{override: map[string]string{"projectMembers": `[{"fullName":"A","email":"a@example.com","affiliation":"X","contributionPercent":50,"role":"LEAD"},{"fullName":"B","email":"b@example.com","affiliation":"X","contributionPercent":50,"role":"LEAD"}]`}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "invalid member email", options: multipartRequestOptions{override: map[string]string{"projectMembers": `[{"fullName":"A","email":"bad","affiliation":"X","contributionPercent":50,"role":"LEAD"},{"fullName":"B","email":"b@example.com","affiliation":"X","contributionPercent":50,"role":"CO_RESEARCHER"}]`}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "invalid contribution percent", options: multipartRequestOptions{override: map[string]string{"projectMembers": `[{"fullName":"A","email":"a@example.com","affiliation":"X","contributionPercent":100.001,"role":"LEAD"},{"fullName":"B","email":"b@example.com","affiliation":"X","contributionPercent":50,"role":"CO_RESEARCHER"}]`}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "invalid enum", options: multipartRequestOptions{override: map[string]string{"fundingType": "OTHER"}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "invalid Buddhist date", options: multipartRequestOptions{override: map[string]string{"startDate": "2569-01-01"}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "negative budget", options: multipartRequestOptions{override: map[string]string{"budgetAmount": "-1"}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "over precision budget", options: multipartRequestOptions{override: map[string]string{"budgetAmount": "1.001"}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "budget research with continuation", options: multipartRequestOptions{override: map[string]string{"continuationOfId": "1"}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "continuation without parent", options: multipartRequestOptions{override: map[string]string{"researchKind": "CONTINUATION"}}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "client ID", options: multipartRequestOptions{unknownField: "id"}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "client status", options: multipartRequestOptions{unknownField: "status"}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "client process", options: multipartRequestOptions{unknownField: "process"}, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "non PDF", options: multipartRequestOptions{contractContentType: "image/png"}, status: http.StatusUnsupportedMediaType, code: "UNSUPPORTED_MEDIA_TYPE"},
		{name: "PDF exceeds 20 MiB", options: multipartRequestOptions{contractData: bytes.Repeat([]byte("p"), 20*1024*1024+1)}, status: http.StatusRequestEntityTooLarge, code: "PAYLOAD_TOO_LARGE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performMultipartCreate(t, multipartCreateRouter(multipartResearchCreatorStub{}), test.options)
			assertMultipartCreateError(t, response, test.status, test.code)
		})
	}
}

func TestCreateResearchMultipartMapsContinuationAndInternalErrors(t *testing.T) {
	tests := []struct {
		name    string
		creator multipartResearchCreatorStub
		status  int
		code    string
	}{
		{name: "missing continuation", creator: multipartResearchCreatorStub{err: service.ErrContinuationNotFound}, status: http.StatusNotFound, code: "CONTINUATION_NOT_FOUND"},
		{name: "internal failure", creator: multipartResearchCreatorStub{err: errors.New("database secret")}, status: http.StatusInternalServerError, code: "INTERNAL_ERROR"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performMultipartCreate(t, multipartCreateRouter(test.creator), multipartRequestOptions{})
			assertMultipartCreateError(t, response, test.status, test.code)
			if strings.Contains(response.Body.String(), "database secret") {
				t.Fatalf("response leaked internal error: %s", response.Body.String())
			}
		})
	}
}

type multipartResearchCreatorStub struct {
	create func(context.Context, domain.CreateResearchInput) (domain.Research, error)
	err    error
}

func (stub multipartResearchCreatorStub) CreateMultipart(ctx context.Context, input domain.CreateResearchInput) (domain.Research, error) {
	if stub.create != nil {
		return stub.create(ctx, input)
	}
	return domain.Research{}, stub.err
}

func multipartCreateRouter(creator multipartResearchCreatorStub) http.Handler {
	return handler.NewRouter(handler.Dependencies{MultipartResearch: handler.NewMultipartResearchHandler(creator)})
}

type multipartRequestOptions struct {
	omit                map[string]bool
	override            map[string]string
	duplicate           string
	unknown             bool
	unknownField        string
	contractContentType string
	contractData        []byte
}

func performMultipartCreate(t *testing.T, router http.Handler, options multipartRequestOptions) *httptest.ResponseRecorder {
	t.Helper()
	request := multipartCreateRequest(t, options)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func multipartCreateRequest(t *testing.T, options multipartRequestOptions) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := validMultipartFields()
	for key, value := range options.override {
		fields[key] = value
	}
	for key, value := range fields {
		if key == "contractFile" || options.omit[key] {
			continue
		}
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write %s: %v", key, err)
		}
		if options.duplicate == key {
			if err := writer.WriteField(key, value); err != nil {
				t.Fatalf("duplicate %s: %v", key, err)
			}
		}
	}
	if options.unknown {
		if err := writer.WriteField("unexpected", "value"); err != nil {
			t.Fatalf("write unknown part: %v", err)
		}
	}
	if options.unknownField != "" {
		if err := writer.WriteField(options.unknownField, "client value"); err != nil {
			t.Fatalf("write prohibited part: %v", err)
		}
	}
	if !options.omit["contractFile"] {
		contentType := options.contractContentType
		if contentType == "" {
			contentType = "application/pdf"
		}
		data := options.contractData
		if data == nil {
			data = []byte("%PDF")
		}
		if err := writeMultipartFile(writer, "contractFile", "contract.pdf", contentType, data); err != nil {
			t.Fatalf("write contract: %v", err)
		}
		if options.duplicate == "contractFile" {
			if err := writeMultipartFile(writer, "contractFile", "copy.pdf", contentType, data); err != nil {
				t.Fatalf("duplicate contract: %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/researches", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func writeMultipartFile(writer *multipart.Writer, name, filename, contentType string, data []byte) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, name, filename))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = part.Write(data)
	return err
}

func validMultipartFields() map[string]string {
	return map[string]string{
		"title": "โครงการใหม่", "isSubsidized": "true",
		"projectMembers": `[{"fullName":"หัวหน้า","email":"lead@example.com","affiliation":"หน่วยงาน","contributionPercent":60,"role":"LEAD"},{"fullName":"ผู้ร่วม","email":"co@example.com","affiliation":"หน่วยงาน","contributionPercent":40,"role":"CO_RESEARCHER"}]`,
		"fundingType":    "INTERNAL", "fundingSourceName": "แหล่งทุน", "contractNumber": "C-001", "contractFile": "",
		"projectType": "RESEARCH", "researchKind": "BUDGET", "responsibleProjectUnit": "หน่วยงานโครงการ", "responsibleBudgetUnit": "หน่วยงานงบประมาณ",
		"startDate": "01/01/2569", "endDate": "31/12/2569", "budgetAmount": "1000.00", "thaiAbstract": "บทคัดย่อไทย", "englishAbstract": "English abstract", "objectives": "วัตถุประสงค์", "keywords": "คำค้น", "continuationOfId": "null",
	}
}

func assertMultipartCreateError(t *testing.T, response *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if response.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, wantStatus, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertErrorResponse(t, response, wantCode)
}
