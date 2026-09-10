package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

type creatorSpy struct {
	calls int
	input repo.CreateResearchInput
	run   func(context.Context, repo.CreateResearchInput) (int64, error)
}

func (r *creatorSpy) Create(ctx context.Context, in repo.CreateResearchInput) (int64, error) {
	r.calls++
	r.input = in
	if r.run != nil {
		return r.run(ctx, in)
	}
	return 42, nil
}

type storeSpy struct {
	events    []string
	fail      string
	published bool
}

func (s *storeSpy) Stage(_ context.Context, _ string, _ io.Reader) (service.StagedContract, error) {
	s.events = append(s.events, "stage")
	if s.fail == "stage" {
		return service.StagedContract{}, errors.New("secret /private/staging SQL")
	}
	return service.StagedContract{Token: "private-token", Filename: "contract.pdf", SizeBytes: 512}, nil
}
func (s *storeSpy) Publish(_ context.Context, _ service.StagedContract) (service.PublishedContract, error) {
	s.events = append(s.events, "publish")
	if s.fail == "publish" || s.fail == "publish_discard" {
		return service.PublishedContract{}, errors.New("secret /private/publish")
	}
	s.published = true
	return service.PublishedContract{Path: "contracts/server-generated.pdf", Filename: "contract.pdf", SizeBytes: 512}, nil
}
func (s *storeSpy) Discard(_ context.Context, _ service.StagedContract) error {
	s.events = append(s.events, "discard")
	if s.fail == "discard" || s.fail == "publish_discard" {
		return errors.New("secret discard")
	}
	return nil
}
func (s *storeSpy) Remove(_ context.Context, _ service.PublishedContract) error {
	s.events = append(s.events, "remove")
	if s.fail == "remove" {
		return errors.New("secret remove")
	}
	s.published = false
	return nil
}

func validCreate(t *testing.T) service.CreateResearchInput {
	t.Helper()
	var in service.CreateResearchInput
	err := json.Unmarshal([]byte(`{"title":"  Research ก  ","continuationOfId":null,"isSubsidized":false,"projectMembers":[{"fullName":"หัวหน้า","email":" Lead@example.com ","affiliation":"มหาวิทยาลัย","contributionPercent":60.25,"role":"LEAD"},{"fullName":"ผู้ร่วม","email":"co@example.com","affiliation":"สถาบัน","contributionPercent":39.75,"role":"CO_RESEARCHER"}],"fundingType":"INTERNAL","fundingSourceName":"กองทุน","contractNumber":" Straße ","projectType":"RESEARCH","researchKind":"BUDGET","responsibleProjectUnit":"โครงการ","responsibleBudgetUnit":"งบประมาณ","startDate":"29/02/2567","endDate":"01/03/2568","budgetAmount":12345.67,"thaiAbstract":"ไทย","englishAbstract":"English","objectives":"เป้าหมาย","keywords":"น้ำ"}`), &in)
	if err != nil {
		t.Fatal(err)
	}
	in.ContinuationOfIDPresent = true
	in.ContractFile = &service.ContractUpload{Filename: "contract.pdf", Reader: strings.NewReader("PDF validated by store double")}
	return in
}

func requireCode(t *testing.T, err error, code, field string) {
	t.Helper()
	var coded interface{ ErrorCode() string }
	if !errors.As(err, &coded) || coded.ErrorCode() != code {
		t.Fatalf("error=%v want code=%s", err, code)
	}
	if field != "" {
		var validation interface{ FieldErrors() []service.FieldError }
		if !errors.As(err, &validation) {
			t.Fatalf("no field errors: %v", err)
		}
		found := false
		for _, f := range validation.FieldErrors() {
			if f.Field == field && strings.TrimSpace(f.Message) != "" {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing field %s: %+v", field, validation.FieldErrors())
		}
	}
	for _, secret := range []string{"secret", "/private", "server-generated", "private-token", "SQL"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error exposes %q", secret)
		}
	}
}

func invalidCreate(t *testing.T, in service.CreateResearchInput, field string) {
	t.Helper()
	r := &creatorSpy{}
	s := &storeSpy{}
	before, _ := json.Marshal(in)
	_, err := service.NewResearchService(r, s).Create(context.Background(), in)
	requireCode(t, err, "VALIDATION_ERROR", field)
	after, _ := json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("validation mutated caller input")
	}
	if r.calls != 0 || len(s.events) != 0 {
		t.Fatalf("invalid input caused side effects: %d %v", r.calls, s.events)
	}
}

func TestCreateServiceSuccess(t *testing.T) {
	r := &creatorSpy{}
	s := &storeSpy{}
	r.run = func(_ context.Context, in repo.CreateResearchInput) (int64, error) {
		if !s.published {
			t.Error("metadata committed before PDF publication")
		}
		if in.Contract.StoragePath != "contracts/server-generated.pdf" {
			t.Error("wrong published metadata")
		}
		return 42, nil
	}
	in := validCreate(t)
	before, _ := json.Marshal(in)
	got, err := service.NewResearchService(r, s).Create(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if r.calls != 1 || !s.published {
		t.Fatalf("calls=%d published=%v", r.calls, s.published)
	}
	if r.input.Title != "Research ก" || r.input.Contract.ContractNumber != "Straße" || r.input.Contract.ContractNumberKey != "strasse" || r.input.StartDate != "2024-02-29" || r.input.EndDate != "2025-03-01" || r.input.IsSubsidized || r.input.BudgetAmount != 12345.67 {
		t.Fatalf("incorrect normalization: %+v", r.input)
	}
	if r.input.Members[0].Email != "Lead@example.com" && r.input.Members[0].Email != "lead@example.com" {
		t.Fatal("email not trimmed")
	}
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var response map[string]any
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{"id": float64(42), "title": "Research ก", "startDate": "29/02/2567", "endDate": "01/03/2568", "status": "กำลังดำเนินการ", "process": "สัญญาโครงการ"} {
		if response[key] != want {
			t.Fatalf("response %s=%v want %v", key, response[key], want)
		}
	}
	for _, secret := range []string{"storagePath", "server-generated", "private-token", "contractNumberKey"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("response leaked %s", secret)
		}
	}
	after, _ := json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("create mutated caller input")
	}
}

func TestCreateServiceText(t *testing.T) {
	fields := []string{"Title", "FundingSourceName", "ContractNumber", "ResponsibleProjectUnit", "ResponsibleBudgetUnit", "ThaiAbstract", "EnglishAbstract", "Objectives", "Keywords"}
	paths := []string{"title", "fundingSourceName", "contractNumber", "responsibleProjectUnit", "responsibleBudgetUnit", "thaiAbstract", "englishAbstract", "objectives", "keywords"}
	setters := map[string]func(*service.CreateResearchInput, string){}
	for i, name := range fields {
		setters[paths[i]] = func(in *service.CreateResearchInput, v string) {
			reflect.ValueOf(in).Elem().FieldByName(name).SetString(v)
		}
	}
	for i := 0; i < 2; i++ {
		for _, field := range []string{"FullName", "Affiliation", "Email"} {
			path := fmt.Sprintf("projectMembers[%d].%s", i, map[string]string{"FullName": "fullName", "Affiliation": "affiliation", "Email": "email"}[field])
			setters[path] = func(in *service.CreateResearchInput, v string) {
				reflect.ValueOf(&in.ProjectMembers[i]).Elem().FieldByName(field).SetString(v)
			}
		}
	}
	for path, set := range setters {
		for name, value := range map[string]string{"empty": "", "whitespace": "\u2003\u00a0\t ", "too_long": strings.Repeat("ก", 1001), "nul": "a\x00b", "control": "a\x01b", "del": "a\x7fb"} {
			t.Run(path+"/"+name, func(t *testing.T) { in := validCreate(t); set(&in, value); invalidCreate(t, in, path) })
		}
		if strings.HasSuffix(path, "email") {
			continue
		}
		for _, size := range []int{1, 1000} {
			t.Run(path+fmt.Sprint(size), func(t *testing.T) {
				in := validCreate(t)
				set(&in, "\u2003"+strings.Repeat("ก", size)+"\u00a0")
				r := &creatorSpy{}
				if _, err := service.NewResearchService(r, &storeSpy{}).Create(context.Background(), in); err != nil {
					t.Fatal(err)
				}
				if r.calls != 1 {
					t.Fatal("boundary not persisted")
				}
			})
		}
	}
	for _, v := range []string{"a/b", "a\nb", "a\tb"} {
		t.Run("title_control"+fmt.Sprintf("%q", v), func(t *testing.T) { in := validCreate(t); in.Title = v; invalidCreate(t, in, "title") })
	}
	t.Run("abstract_newline_tab", func(t *testing.T) {
		in := validCreate(t)
		in.ThaiAbstract = "line1\nline2\tend"
		if _, err := service.NewResearchService(&creatorSpy{}, &storeSpy{}).Create(context.Background(), in); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCreateServiceValidation(t *testing.T) {
	cases := []struct {
		name, field string
		mutate      func(*service.CreateResearchInput)
	}{
		{"missing_boolean", "isSubsidized", func(in *service.CreateResearchInput) { in.IsSubsidized = nil }},
		{"missing_parent_part", "continuationOfId", func(in *service.CreateResearchInput) { in.ContinuationOfIDPresent = false }},
		{"continuation_null", "continuationOfId", func(in *service.CreateResearchInput) { in.ResearchKind = "CONTINUATION" }},
		{"root_parent", "continuationOfId", func(in *service.CreateResearchInput) { id := int64(1); in.ContinuationOfID = &id }},
		{"missing_pdf", "contractFile", func(in *service.CreateResearchInput) { in.ContractFile = nil }},
		{"nil_pdf_reader", "contractFile", func(in *service.CreateResearchInput) { in.ContractFile.Reader = nil }},
	}
	for _, field := range []string{"FundingType", "ProjectType", "ResearchKind"} {
		for _, v := range []string{"", "OTHER"} {
			cases = append(cases, struct {
				name, field string
				mutate      func(*service.CreateResearchInput)
			}{field + v, map[string]string{"FundingType": "fundingType", "ProjectType": "projectType", "ResearchKind": "researchKind"}[field], func(in *service.CreateResearchInput) { reflect.ValueOf(in).Elem().FieldByName(field).SetString(v) }})
		}
	}
	for _, id := range []int64{0, -1} {
		cases = append(cases, struct {
			name, field string
			mutate      func(*service.CreateResearchInput)
		}{fmt.Sprint("parent", id), "continuationOfId", func(in *service.CreateResearchInput) { in.ResearchKind = "CONTINUATION"; in.ContinuationOfID = &id }})
	}
	for _, v := range []string{"", "0", "-1", "1.001", "NaN", "Infinity", "abc"} {
		cases = append(cases, struct {
			name, field string
			mutate      func(*service.CreateResearchInput)
		}{"budget" + v, "budgetAmount", func(in *service.CreateResearchInput) { in.BudgetAmount = json.Number(v) }})
	}
	for _, field := range []string{"StartDate", "EndDate"} {
		for _, v := range []string{"", "29/02/2568", "31/04/2567", "01/13/2567", "00/01/2567", "2024-02-29", "1/3/2568"} {
			cases = append(cases, struct {
				name, field string
				mutate      func(*service.CreateResearchInput)
			}{field + v, map[string]string{"StartDate": "startDate", "EndDate": "endDate"}[field], func(in *service.CreateResearchInput) { reflect.ValueOf(in).Elem().FieldByName(field).SetString(v) }})
		}
	}
	for _, v := range []string{"28/02/2568", "29/02/2567", "28/02/2567"} {
		cases = append(cases, struct {
			name, field string
			mutate      func(*service.CreateResearchInput)
		}{"short_year" + v, "endDate", func(in *service.CreateResearchInput) { in.EndDate = v }})
	}
	for _, n := range []int{0, 1, 3} {
		cases = append(cases, struct {
			name, field string
			mutate      func(*service.CreateResearchInput)
		}{fmt.Sprint("members", n), "projectMembers", func(in *service.CreateResearchInput) {
			if n == 3 {
				in.ProjectMembers = append(in.ProjectMembers, in.ProjectMembers[0])
			} else {
				in.ProjectMembers = in.ProjectMembers[:n]
			}
		}})
	}
	for i := 0; i < 2; i++ {
		for _, email := range []string{"not-email", "a@", "@example.com", "a b@example.com", strings.Repeat("a", 243) + "@example.com"} {
			cases = append(cases, struct {
				name, field string
				mutate      func(*service.CreateResearchInput)
			}{fmt.Sprintf("email%d%s", i, email), fmt.Sprintf("projectMembers[%d].email", i), func(in *service.CreateResearchInput) { in.ProjectMembers[i].Email = email }})
		}
		for _, v := range []string{"", "0", "-1", "100.01", "39.751", "NaN"} {
			cases = append(cases, struct {
				name, field string
				mutate      func(*service.CreateResearchInput)
			}{fmt.Sprintf("percent%d%s", i, v), fmt.Sprintf("projectMembers[%d].contributionPercent", i), func(in *service.CreateResearchInput) { in.ProjectMembers[i].ContributionPercent = json.Number(v) }})
		}
		for _, role := range []string{"", "OTHER"} {
			cases = append(cases, struct {
				name, field string
				mutate      func(*service.CreateResearchInput)
			}{fmt.Sprintf("role%d%s", i, role), fmt.Sprintf("projectMembers[%d].role", i), func(in *service.CreateResearchInput) { in.ProjectMembers[i].Role = role }})
		}
	}
	for _, role := range []string{"LEAD", "CO_RESEARCHER"} {
		cases = append(cases, struct {
			name, field string
			mutate      func(*service.CreateResearchInput)
		}{"duplicate_role" + role, "projectMembers", func(in *service.CreateResearchInput) {
			in.ProjectMembers[0].Role = role
			in.ProjectMembers[1].Role = role
		}})
	}
	for _, v := range []string{"39.74", "39.76"} {
		cases = append(cases, struct {
			name, field string
			mutate      func(*service.CreateResearchInput)
		}{"sum" + v, "projectMembers", func(in *service.CreateResearchInput) { in.ProjectMembers[1].ContributionPercent = json.Number(v) }})
	}
	cases = append(cases, struct {
		name, field string
		mutate      func(*service.CreateResearchInput)
	}{"duplicate_email", "projectMembers[1].email", func(in *service.CreateResearchInput) { in.ProjectMembers[1].Email = "\u2003LEAD@example.com\u00a0" }})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { in := validCreate(t); tc.mutate(&in); invalidCreate(t, in, tc.field) })
	}
}

func TestCreateServiceValidDates(t *testing.T) {
	for _, dates := range [][2]string{{"29/02/2567", "01/03/2568"}, {"30/06/2568", "30/06/2569"}, {"01/01/2568", "02/01/2570"}} {
		t.Run(dates[0], func(t *testing.T) {
			in := validCreate(t)
			in.StartDate = dates[0]
			in.EndDate = dates[1]
			in.BudgetAmount = "0.01"
			in.ProjectMembers[0].ContributionPercent = "0.01"
			in.ProjectMembers[1].ContributionPercent = "99.99"
			if _, err := service.NewResearchService(&creatorSpy{}, &storeSpy{}).Create(context.Background(), in); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCreateServiceFailures(t *testing.T) {
	for _, tc := range []struct {
		name, storeFail string
		repoErr         error
		code            string
		events          []string
	}{
		{"stage", "stage", nil, "INTERNAL_ERROR", []string{"stage"}},
		{"publish", "publish", nil, "INTERNAL_ERROR", []string{"stage", "publish", "discard"}},
		{"discard_failure", "publish_discard", nil, "INTERNAL_ERROR", []string{"stage", "publish", "discard"}},
		{"repo", "", repo.ErrInternal, "INTERNAL_ERROR", []string{"stage", "publish", "remove"}},
		{"parent", "", repo.ErrContinuationNotFound, "CONTINUATION_NOT_FOUND", []string{"stage", "publish", "remove"}},
		{"title", "", repo.ErrTitleAlreadyExists, "TITLE_ALREADY_EXISTS", []string{"stage", "publish", "remove"}},
		{"contract", "", repo.ErrContractNumberAlreadyExists, "CONTRACT_NUMBER_ALREADY_EXISTS", []string{"stage", "publish", "remove"}},
		{"cleanup", "remove", repo.ErrInternal, "INTERNAL_ERROR", []string{"stage", "publish", "remove"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &creatorSpy{run: func(context.Context, repo.CreateResearchInput) (int64, error) { return 0, tc.repoErr }}
			s := &storeSpy{fail: tc.storeFail}
			_, err := service.NewResearchService(r, s).Create(context.Background(), validCreate(t))
			requireCode(t, err, tc.code, "")
			if !reflect.DeepEqual(s.events, tc.events) {
				t.Fatalf("events=%v want %v", s.events, tc.events)
			}
			wantCalls := 1
			if tc.storeFail == "stage" || tc.storeFail == "publish" || tc.storeFail == "publish_discard" {
				wantCalls = 0
			}
			if r.calls != wantCalls {
				t.Fatalf("mutation retried/called prematurely: %d", r.calls)
			}
			if tc.storeFail != "remove" && s.published {
				t.Fatal("orphan after compensated failure")
			}
		})
	}
	t.Run("canceled_before_create", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		r := &creatorSpy{}
		s := &storeSpy{}
		_, err := service.NewResearchService(r, s).Create(ctx, validCreate(t))
		if err == nil || r.calls != 0 || len(s.events) != 0 {
			t.Fatalf("cancellation caused side effects: %v", err)
		}
	})
}

func TestCreateServiceValidValues(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*service.CreateResearchInput)
	}{
		{"true_external_academic", func(in *service.CreateResearchInput) {
			v := true
			in.IsSubsidized = &v
			in.FundingType = "EXTERNAL"
			in.ProjectType = "ACADEMIC_SERVICE"
		}},
		{"continuation", func(in *service.CreateResearchInput) {
			id := int64(17)
			in.ContinuationOfID = &id
			in.ResearchKind = "CONTINUATION"
		}},
		{"email_boundary", func(in *service.CreateResearchInput) {
			in.ProjectMembers[0].Email = strings.Repeat("a", 64) + "@" + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 57) + ".com"
		}},
		{"decimal_exact", func(in *service.CreateResearchInput) {
			in.BudgetAmount = "1.01"
			in.ProjectMembers[0].ContributionPercent = "33.33"
			in.ProjectMembers[1].ContributionPercent = "66.67"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := validCreate(t)
			tc.change(&in)
			r := &creatorSpy{}
			if _, err := service.NewResearchService(r, &storeSpy{}).Create(context.Background(), in); err != nil {
				t.Fatal(err)
			}
			if r.calls != 1 {
				t.Fatal("valid input not persisted")
			}
			if !reflect.DeepEqual(r.input.ContinuationOfID, in.ContinuationOfID) || r.input.ResearchKind != in.ResearchKind || r.input.ProjectType != in.ProjectType || r.input.Contract.FundingType != in.FundingType {
				t.Fatal("valid enum/parent data not preserved")
			}
		})
	}
}

func TestCreateServicePDFErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
		code string
	}{
		{"invalid_pdf", []byte("not PDF"), "VALIDATION_ERROR"},
		{"oversized", sizedPDF(20971521), "PAYLOAD_TOO_LARGE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			store, err := service.NewFileContractStore(root)
			if err != nil {
				t.Fatal(err)
			}
			r := &creatorSpy{}
			in := validCreate(t)
			in.ContractFile.Reader = bytes.NewReader(tc.data)
			_, err = service.NewResearchService(r, store).Create(context.Background(), in)
			field := ""
			if tc.code == "VALIDATION_ERROR" {
				field = "contractFile"
			}
			requireCode(t, err, tc.code, field)
			if r.calls != 0 || len(storageFiles(t, root)) != 0 {
				t.Fatal("PDF rejection left data/files")
			}
		})
	}
}

func TestCreateServiceCanceledAfterPublish(t *testing.T) {
	root := t.TempDir()
	store, err := service.NewFileContractStore(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := &creatorSpy{run: func(context.Context, repo.CreateResearchInput) (int64, error) { cancel(); return 0, ctx.Err() }}
	in := validCreate(t)
	in.ContractFile.Reader = bytes.NewReader(pdfFixture(1, 0, nil))
	if _, err := service.NewResearchService(r, store).Create(ctx, in); err == nil {
		t.Fatal("canceled commit reported success")
	}
	if r.calls != 1 || len(storageFiles(t, root)) != 0 {
		t.Fatal("canceled context prevented compensation")
	}
}

func TestCreateServiceWaitsForCommit(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r := &creatorSpy{run: func(ctx context.Context, _ repo.CreateResearchInput) (int64, error) {
		close(entered)
		select {
		case <-release:
			return 42, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}}
	s := &storeSpy{}
	in := validCreate(t)
	go func() { _, err := service.NewResearchService(r, s).Create(ctx, in); done <- err }()
	select {
	case <-entered:
	case err := <-done:
		t.Fatalf("returned before repo commit: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	select {
	case err := <-done:
		t.Fatalf("premature success: %v", err)
	default:
	}
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
