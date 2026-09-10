package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/db"
	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

type formPart struct{ name, value, filename, media string }

func postFixture() []formPart {
	return []formPart{
		{name: "title", value: "Research ก"}, {name: "continuationOfId", value: "null"}, {name: "isSubsidized", value: "false"},
		{name: "projectMembers", media: "application/json", value: `[{"fullName":"Lead","email":"lead@example.com","affiliation":"University","contributionPercent":60,"role":"LEAD"},{"fullName":"Co","email":"co@example.com","affiliation":"Institute","contributionPercent":40,"role":"CO_RESEARCHER"}]`},
		{name: "fundingType", value: "INTERNAL"}, {name: "fundingSourceName", value: "Fund"}, {name: "contractNumber", value: "CN-001"},
		{name: "contractFile", filename: "contract.pdf", media: "application/pdf", value: string(postPDF(0))},
		{name: "projectType", value: "RESEARCH"}, {name: "researchKind", value: "BUDGET"},
		{name: "responsibleProjectUnit", value: "Project unit"}, {name: "responsibleBudgetUnit", value: "Budget unit"},
		{name: "startDate", value: "29/02/2567"}, {name: "endDate", value: "01/03/2568"}, {name: "budgetAmount", value: "12345.67"},
		{name: "thaiAbstract", value: "บทคัดย่อ"}, {name: "englishAbstract", value: "Abstract"}, {name: "objectives", value: "Objectives"}, {name: "keywords", value: "water"},
	}
}
func replacePart(parts []formPart, name, value string) []formPart {
	out := append([]formPart(nil), parts...)
	for i := range out {
		if out[i].name == name {
			out[i].value = value
		}
	}
	return out
}
func postPDF(padding int) []byte {
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	if padding > 0 {
		b.WriteByte('%')
		b.WriteString(strings.Repeat("x", padding))
		b.WriteByte('\n')
	}
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R] /Count 1 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>", "<< /Length 0 >>\nstream\n\nendstream"}
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 5\n0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	return b.Bytes()
}
func sizedPostPDF(size int) []byte {
	pad := size - len(postPDF(0)) - 2
	for i := 0; i < 10; i++ {
		b := postPDF(pad)
		if len(b) == size {
			return b
		}
		pad += size - len(b)
	}
	panic("PDF fixture size mismatch")
}
func formBody(t *testing.T, parts []formPart) ([]byte, string) {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if err := w.SetBoundary("t07-boundary"); err != nil {
		t.Fatal(err)
	}
	for _, p := range parts {
		h := textproto.MIMEHeader{}
		disposition := fmt.Sprintf(`form-data; name="%s"`, p.name)
		if p.filename != "" {
			disposition += fmt.Sprintf(`; filename="%s"`, p.filename)
		}
		h.Set("Content-Disposition", disposition)
		if p.media != "" {
			h.Set("Content-Type", p.media)
		}
		part, err := w.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(part, p.value); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes(), w.FormDataContentType()
}
func postRouter(t *testing.T) (http.Handler, *sql.DB, string) {
	t.Helper()
	root := t.TempDir()
	database, err := db.Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	storageRoot := filepath.Join(root, "files")
	store, err := service.NewFileContractStore(storageRoot)
	if err != nil {
		t.Fatal(err)
	}
	return handler.NewRouter(service.NewHealthService(repo.NewHealthRepository(database)), handler.WithResearchCreator(service.NewResearchService(repo.NewResearchRepository(database), store))), database, storageRoot
}
func sendPost(router http.Handler, body []byte, media, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if media != "" {
		request.Header.Set("Content-Type", media)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
func postParts(t *testing.T, router http.Handler, parts []formPart) *httptest.ResponseRecorder {
	t.Helper()
	body, media := formBody(t, parts)
	return sendPost(router, body, media, "/api/v1/researches")
}
func postError(t *testing.T, res *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if res.Code != status {
		t.Fatalf("status=%d want %d body=%s", res.Code, status, res.Body.String())
	}
	if !strings.HasPrefix(res.Header().Get("Content-Type"), "application/json") {
		t.Fatal("non-JSON response")
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 {
		t.Fatal("wrong error envelope")
	}
	var e struct {
		Code        string `json:"code"`
		Message     string `json:"message"`
		FieldErrors []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"fieldErrors"`
	}
	if err := json.Unmarshal(data["error"], &e); err != nil {
		t.Fatal(err)
	}
	if e.Code != code || strings.TrimSpace(e.Message) == "" {
		t.Fatalf("wrong error %+v", e)
	}
	if code == "VALIDATION_ERROR" {
		if len(e.FieldErrors) == 0 {
			t.Fatal("missing fieldErrors")
		}
		for _, f := range e.FieldErrors {
			if f.Field == "" || f.Message == "" {
				t.Fatal("empty field error")
			}
		}
	} else {
		var obj map[string]any
		_ = json.Unmarshal(data["error"], &obj)
		if _, ok := obj["fieldErrors"]; ok {
			t.Fatal("non-validation error contains fieldErrors")
		}
	}
	for _, secret := range []string{"injected-secret", "storage_path", "SQLITE", "CREATE TRIGGER", "test.db"} {
		if strings.Contains(res.Body.String(), secret) {
			t.Fatalf("leaked %s", secret)
		}
	}
}
func noCreate(t *testing.T, database *sql.DB, root string) {
	t.Helper()
	for _, table := range []string{"researches", "research_members", "research_contracts"} {
		var n int
		if err := database.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("partial %s: %d", table, n)
		}
	}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			t.Errorf("unexpected file %s", path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPOSTCreateSuccess(t *testing.T) {
	router, database, root := postRouter(t)
	parts := postFixture()
	res := postParts(t, router, parts)
	if res.Code != 201 {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if !strings.HasPrefix(res.Header().Get("Content-Type"), "application/json") {
		t.Fatal("success response is not JSON")
	}
	var got map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	for _, p := range parts {
		if p.name == "contractFile" {
			continue
		}
		var want any = p.value
		switch p.name {
		case "projectMembers":
			if err := json.Unmarshal([]byte(p.value), &want); err != nil {
				t.Fatal(err)
			}
		case "continuationOfId":
			want = nil
		case "isSubsidized":
			want = false
		case "budgetAmount":
			want = float64(12345.67)
		}
		a, _ := json.Marshal(got[p.name])
		b, _ := json.Marshal(want)
		if !bytes.Equal(a, b) {
			t.Fatalf("field %s=%s want %s", p.name, a, b)
		}
	}
	id, ok := got["id"].(float64)
	if !ok || id <= 0 || id != float64(int64(id)) {
		t.Fatal("invalid generated ID")
	}
	if got["status"] != "กำลังดำเนินการ" || got["process"] != "สัญญาโครงการ" {
		t.Fatal("wrong initial state")
	}
	meta, ok := got["contractFile"].(map[string]any)
	if !ok || len(meta) != 3 || meta["filename"] != "contract.pdf" || meta["contentType"] != "application/pdf" || meta["sizeBytes"] != float64(len(postPDF(0))) {
		t.Fatalf("invalid metadata %v", meta)
	}
	var path string
	if err := database.QueryRow("SELECT storage_path FROM research_contracts WHERE research_id=?", int64(id)).Scan(&path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil || !bytes.Equal(data, postPDF(0)) {
		t.Fatalf("persisted PDF unavailable: %v", err)
	}
	for _, key := range []string{"description", "storagePath", "contractNumberKey"} {
		if _, ok := got[key]; ok {
			t.Fatalf("unexpected field %s", key)
		}
	}
	if strings.Contains(res.Body.String(), path) || strings.Contains(res.Body.String(), root) {
		t.Fatal("internal path exposed")
	}
}

func TestPOSTCreateMissingAndDuplicateParts(t *testing.T) {
	for index, p := range postFixture() {
		for _, duplicate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/duplicate=%v", p.name, duplicate), func(t *testing.T) {
				router, database, root := postRouter(t)
				parts := postFixture()
				if duplicate {
					parts = append(parts, p)
				} else {
					parts = append(parts[:index], parts[index+1:]...)
				}
				postError(t, postParts(t, router, parts), 422, "VALIDATION_ERROR")
				noCreate(t, database, root)
			})
		}
	}
}
func TestPOSTCreateUnknownParts(t *testing.T) {
	for _, name := range []string{"id", "status", "process", "description", "contractNumberKey", "storagePath", "contractFileMetadata", "unexpected"} {
		t.Run(name, func(t *testing.T) {
			router, database, root := postRouter(t)
			postError(t, postParts(t, router, append(postFixture(), formPart{name: name, value: "1"})), 422, "VALIDATION_ERROR")
			noCreate(t, database, root)
		})
	}
}
func TestPOSTCreateScalars(t *testing.T) {
	for field, values := range map[string][]string{"isSubsidized": {"", "null", "1", "TRUE", "\"true\""}, "continuationOfId": {"", "undefined", "1.5", "\"null\"", "{}"}, "budgetAmount": {"", "null", "abc", "NaN", "Infinity", "{}"}} {
		for _, value := range values {
			t.Run(field+value, func(t *testing.T) {
				router, database, root := postRouter(t)
				postError(t, postParts(t, router, replacePart(postFixture(), field, value)), 422, "VALIDATION_ERROR")
				noCreate(t, database, root)
			})
		}
	}
}
func TestPOSTCreateMembers(t *testing.T) {
	var valid string
	for _, p := range postFixture() {
		if p.name == "projectMembers" {
			valid = p.value
		}
	}
	cases := []string{"", "[", "{}", "null", "true", `"members"`, valid + " []", strings.Replace(valid, `"role":"LEAD"`, `"role":"LEAD","role":"CO_RESEARCHER"`, 1), strings.Replace(valid, `"fullName":"Lead"`, `"fullName":"Lead","unknown":1`, 1), strings.Replace(valid, `"contributionPercent":60`, `"contributionPercent":"60"`, 1), strings.Replace(valid, `"email":"lead@example.com"`, `"email":null`, 1), `[1,2]`}
	for i, value := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			router, database, root := postRouter(t)
			postError(t, postParts(t, router, replacePart(postFixture(), "projectMembers", value)), 422, "VALIDATION_ERROR")
			noCreate(t, database, root)
		})
	}
}

func TestPOSTCreateTransport(t *testing.T) {
	for _, tc := range []struct {
		name, media string
		body        []byte
		status      int
		code        string
	}{
		{"missing_media", "", []byte("{}"), 415, "UNSUPPORTED_MEDIA_TYPE"}, {"json", "application/json", []byte("{}"), 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"missing_boundary", "multipart/form-data", []byte("bad"), 422, "VALIDATION_ERROR"}, {"wrong_boundary", "multipart/form-data; boundary=other", []byte("--broken"), 422, "VALIDATION_ERROR"},
		{"empty", "multipart/form-data; boundary=t07-boundary", nil, 422, "VALIDATION_ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router, database, root := postRouter(t)
			postError(t, sendPost(router, tc.body, tc.media, "/api/v1/researches"), tc.status, tc.code)
			noCreate(t, database, root)
		})
	}
	t.Run("truncated", func(t *testing.T) {
		router, database, root := postRouter(t)
		body, media := formBody(t, postFixture())
		postError(t, sendPost(router, body[:len(body)-12], media, "/api/v1/researches"), 422, "VALIDATION_ERROR")
		noCreate(t, database, root)
	})
	t.Run("query", func(t *testing.T) {
		router, database, root := postRouter(t)
		body, media := formBody(t, postFixture())
		postError(t, sendPost(router, body, media, "/api/v1/researches?foo=1"), 422, "VALIDATION_ERROR")
		noCreate(t, database, root)
	})
}

func TestPOSTCreatePDF(t *testing.T) {
	for _, tc := range []struct {
		name   string
		data   []byte
		status int
		code   string
	}{{"empty", nil, 422, "VALIDATION_ERROR"}, {"fake", []byte("not PDF"), 422, "VALIDATION_ERROR"}, {"at_limit", sizedPostPDF(20971520), 201, ""}, {"over_limit", sizedPostPDF(20971521), 413, "PAYLOAD_TOO_LARGE"}} {
		t.Run(tc.name, func(t *testing.T) {
			router, database, root := postRouter(t)
			res := postParts(t, router, replacePart(postFixture(), "contractFile", string(tc.data)))
			if tc.status == 201 {
				if res.Code != 201 {
					t.Fatalf("status=%d %s", res.Code, res.Body.String())
				}
			} else {
				postError(t, res, tc.status, tc.code)
				noCreate(t, database, root)
			}
		})
	}
}
func TestPOSTCreateBodyLimit(t *testing.T) {
	router, database, root := postRouter(t)
	body, media := formBody(t, postFixture())
	body = append(body, bytes.Repeat([]byte{' '}, 22020097-len(body))...)
	request := httptest.NewRequest("POST", "/api/v1/researches", bytes.NewReader(body))
	request.ContentLength = -1
	request.Header.Set("Content-Type", media)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, request)
	postError(t, res, 413, "PAYLOAD_TOO_LARGE")
	noCreate(t, database, root)
}

func TestPOSTCreateBodyBoundary(t *testing.T) {
	for _, size := range []int{22020096, 22020097} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			router, database, root := postRouter(t)
			body, media := formBody(t, replacePart(postFixture(), "contractFile", string(sizedPostPDF(20971520))))
			// Legal part header padding keeps the PDF within its independent 20 MiB cap.
			extra := size - len(body) - len("\r\nX-Padding: ")
			if extra < 0 {
				t.Fatal("invalid size fixture")
			}
			body = bytes.Replace(body, []byte("\r\n\r\n"), []byte("\r\nX-Padding: "+strings.Repeat("x", extra)+"\r\n\r\n"), 1)
			if len(body) != size {
				t.Fatalf("body size=%d want %d", len(body), size)
			}
			res := sendPost(router, body, media, "/api/v1/researches")
			if size == 22020096 {
				if res.Code != 201 {
					t.Fatalf("exact aggregate limit rejected: %d %s", res.Code, res.Body.String())
				}
			} else {
				postError(t, res, 413, "PAYLOAD_TOO_LARGE")
				noCreate(t, database, root)
			}
		})
	}
}

func TestPOSTCreateEncodings(t *testing.T) {
	for _, filename := range []string{"", "project-members.json"} {
		t.Run("json_part_filename="+filename, func(t *testing.T) {
			router, _, _ := postRouter(t)
			parts := postFixture()
			for i := range parts {
				if parts[i].name == "projectMembers" {
					parts[i].filename = filename
				}
			}
			body, _ := formBody(t, parts)
			res := sendPost(router, body, "Multipart/Form-Data; boundary=\"t07-boundary\"", "/api/v1/researches")
			if res.Code != 201 {
				t.Fatalf("valid media/JSON encoding rejected: %d %s", res.Code, res.Body.String())
			}
		})
	}
}
func TestPOSTCreateBusinessErrors(t *testing.T) {
	t.Run("missing_parent", func(t *testing.T) {
		router, database, root := postRouter(t)
		parts := replacePart(replacePart(postFixture(), "researchKind", "CONTINUATION"), "continuationOfId", "99999")
		postError(t, postParts(t, router, parts), 404, "CONTINUATION_NOT_FOUND")
		noCreate(t, database, root)
	})
	for _, field := range []string{"title", "contractNumber"} {
		t.Run(field, func(t *testing.T) {
			router, database, _ := postRouter(t)
			res := postParts(t, router, postFixture())
			if res.Code != 201 {
				t.Fatalf("setup failed %s", res.Body.String())
			}
			parts := postFixture()
			code := "TITLE_ALREADY_EXISTS"
			if field == "title" {
				parts = replacePart(parts, "contractNumber", "CN-002")
			} else {
				parts = replacePart(parts, "title", "Different")
				parts = replacePart(parts, "contractNumber", " cn-001 ")
				code = "CONTRACT_NUMBER_ALREADY_EXISTS"
			}
			postError(t, postParts(t, router, parts), 409, code)
			var n int
			if err := database.QueryRow("SELECT count(*) FROM researches").Scan(&n); err != nil || n != 1 {
				t.Fatalf("duplicate persisted %d %v", n, err)
			}
		})
	}
}
func TestPOSTCreateContinuation(t *testing.T) {
	for _, status := range []string{"โครงการเสร็จสิ้น", "ยุติโครงการ"} {
		t.Run(status, func(t *testing.T) {
			router, database, _ := postRouter(t)
			res := postParts(t, router, postFixture())
			if res.Code != 201 {
				t.Fatal(res.Body.String())
			}
			var parent struct {
				ID int64 `json:"id"`
			}
			if err := json.Unmarshal(res.Body.Bytes(), &parent); err != nil {
				t.Fatal(err)
			}
			if _, err := database.Exec("UPDATE researches SET status=? WHERE id=?", status, parent.ID); err != nil {
				t.Fatal(err)
			}
			parts := replacePart(replacePart(replacePart(postFixture(), "researchKind", "CONTINUATION"), "continuationOfId", fmt.Sprint(parent.ID)), "contractNumber", "CHILD")
			res = postParts(t, router, parts)
			if res.Code != 201 {
				t.Fatal(res.Body.String())
			}
			var child struct {
				ID     int64 `json:"id"`
				Parent int64 `json:"continuationOfId"`
			}
			if err := json.Unmarshal(res.Body.Bytes(), &child); err != nil {
				t.Fatal(err)
			}
			if child.Parent != parent.ID || child.ID == parent.ID {
				t.Fatal("identity lost")
			}
		})
	}
}
func TestPOSTCreateDatabaseFailure(t *testing.T) {
	router, database, root := postRouter(t)
	if _, err := database.Exec("CREATE TRIGGER failure BEFORE INSERT ON research_contracts BEGIN SELECT RAISE(ABORT,'injected-secret test.db'); END"); err != nil {
		t.Fatal(err)
	}
	postError(t, postParts(t, router, postFixture()), 500, "INTERNAL_ERROR")
	noCreate(t, database, root)
}
