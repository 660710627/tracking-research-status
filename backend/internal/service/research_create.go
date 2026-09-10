package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/660710627/my-research/internal/repo"
)

type ContractUpload struct {
	Filename string
	Reader   io.Reader
}
type ResearchMemberInput struct {
	FullName            string      `json:"fullName"`
	Email               string      `json:"email"`
	Affiliation         string      `json:"affiliation"`
	ContributionPercent json.Number `json:"contributionPercent"`
	Role                string      `json:"role"`
}
type CreateResearchInput struct {
	Title                   string                `json:"title"`
	ContinuationOfID        *int64                `json:"continuationOfId"`
	ContinuationOfIDPresent bool                  `json:"-"`
	IsSubsidized            *bool                 `json:"isSubsidized"`
	ProjectMembers          []ResearchMemberInput `json:"projectMembers"`
	FundingType             string                `json:"fundingType"`
	FundingSourceName       string                `json:"fundingSourceName"`
	ContractNumber          string                `json:"contractNumber"`
	ContractFile            *ContractUpload       `json:"-"`
	ProjectType             string                `json:"projectType"`
	ResearchKind            string                `json:"researchKind"`
	ResponsibleProjectUnit  string                `json:"responsibleProjectUnit"`
	ResponsibleBudgetUnit   string                `json:"responsibleBudgetUnit"`
	StartDate               string                `json:"startDate"`
	EndDate                 string                `json:"endDate"`
	BudgetAmount            json.Number           `json:"budgetAmount"`
	ThaiAbstract            string                `json:"thaiAbstract"`
	EnglishAbstract         string                `json:"englishAbstract"`
	Objectives              string                `json:"objectives"`
	Keywords                string                `json:"keywords"`
}
type CreatedResearch struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Status    string `json:"status"`
	Process   string `json:"process"`
}
type FieldError struct{ Field, Message string }
type codedError struct {
	code, message string
	fields        []FieldError
}

func (e codedError) Error() string             { return e.message }
func (e codedError) ErrorCode() string         { return e.code }
func (e codedError) FieldErrors() []FieldError { return e.fields }
func validationError(f, m string) error {
	return codedError{code: "VALIDATION_ERROR", message: "Request validation failed.", fields: []FieldError{{f, m}}}
}
func internalError(err error) error {
	return codedError{code: "INTERNAL_ERROR", message: "An unexpected internal error occurred."}
}

type creator interface {
	Create(context.Context, repo.CreateResearchInput) (int64, error)
}
type ResearchService struct {
	creator creator
	store   ContractStore
}

func NewResearchService(c creator, s ContractStore) *ResearchService {
	return &ResearchService{creator: c, store: s}
}
func (s *ResearchService) Create(ctx context.Context, in CreateResearchInput) (CreatedResearch, error) {
	if err := ctx.Err(); err != nil {
		return CreatedResearch{}, err
	}
	normalized, err := validate(in)
	if err != nil {
		return CreatedResearch{}, err
	}
	stage, err := s.store.Stage(ctx, in.ContractFile.Filename, in.ContractFile.Reader)
	if err != nil {
		return CreatedResearch{}, sanitize(err)
	}
	published, err := s.store.Publish(ctx, stage)
	if err != nil {
		_ = s.store.Discard(context.Background(), stage)
		return CreatedResearch{}, sanitize(err)
	}
	normalized.Contract.StoragePath = published.Path
	normalized.Contract.OriginalFilename = published.Filename
	normalized.Contract.SizeBytes = published.SizeBytes
	id, err := s.creator.Create(ctx, normalized)
	if err != nil {
		_ = s.store.Remove(context.Background(), published)
		return CreatedResearch{}, mapRepoError(err)
	}
	return CreatedResearch{ID: id, Title: normalized.Title, StartDate: formatBuddhist(normalized.StartDate), EndDate: formatBuddhist(normalized.EndDate), Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ"}, nil
}
func validate(in CreateResearchInput) (repo.CreateResearchInput, error) {
	if in.IsSubsidized == nil {
		return repo.CreateResearchInput{}, validationError("isSubsidized", "This field is required")
	}
	if !in.ContinuationOfIDPresent {
		return repo.CreateResearchInput{}, validationError("continuationOfId", "This field is required")
	}
	if err := text(in.Title, true, "title"); err != nil {
		return repo.CreateResearchInput{}, err
	}
	for _, x := range []struct{ v, f string }{{in.FundingSourceName, "fundingSourceName"}, {in.ResponsibleProjectUnit, "responsibleProjectUnit"}, {in.ResponsibleBudgetUnit, "responsibleBudgetUnit"}, {in.ThaiAbstract, "thaiAbstract"}, {in.EnglishAbstract, "englishAbstract"}, {in.Objectives, "objectives"}, {in.Keywords, "keywords"}} {
		if err := text(x.v, false, x.f); err != nil {
			return repo.CreateResearchInput{}, err
		}
	}
	if in.FundingType != "INTERNAL" && in.FundingType != "EXTERNAL" {
		return repo.CreateResearchInput{}, validationError("fundingType", "Invalid funding type")
	}
	if in.ProjectType != "RESEARCH" && in.ProjectType != "ACADEMIC_SERVICE" {
		return repo.CreateResearchInput{}, validationError("projectType", "Invalid project type")
	}
	if in.ResearchKind != "BUDGET" && in.ResearchKind != "CONTINUATION" {
		return repo.CreateResearchInput{}, validationError("researchKind", "Invalid research kind")
	}
	if in.ResearchKind == "BUDGET" && in.ContinuationOfID != nil {
		return repo.CreateResearchInput{}, validationError("continuationOfId", "Budget projects cannot have a parent")
	}
	if in.ResearchKind == "CONTINUATION" && (in.ContinuationOfID == nil || *in.ContinuationOfID <= 0) {
		return repo.CreateResearchInput{}, validationError("continuationOfId", "A continuation project requires a parent")
	}
	start, end, err := parseDates(in.StartDate, in.EndDate)
	if err != nil {
		return repo.CreateResearchInput{}, err
	}
	budget, err := decimal(in.BudgetAmount, "budgetAmount")
	if err != nil {
		return repo.CreateResearchInput{}, err
	}
	if len(in.ProjectMembers) != 2 {
		return repo.CreateResearchInput{}, validationError("projectMembers", "Exactly two members are required")
	}
	members := make([]repo.ResearchMemberInput, 2)
	roles := map[string]bool{}
	sum := 0.0
	emails := map[string]bool{}
	for i, m := range in.ProjectMembers {
		prefix := fmt.Sprintf("projectMembers[%d]", i)
		if err := text(m.FullName, false, prefix+".fullName"); err != nil {
			return repo.CreateResearchInput{}, err
		}
		if err := text(m.Affiliation, false, prefix+".affiliation"); err != nil {
			return repo.CreateResearchInput{}, err
		}
		email := strings.TrimSpace(m.Email)
		if _, e := mail.ParseAddress(email); e != nil || mailAddress(email) != email || len([]rune(email)) > 254 {
			return repo.CreateResearchInput{}, validationError(prefix+".email", "Invalid email")
		}
		key := strings.ToLower(email)
		if emails[key] {
			return repo.CreateResearchInput{}, validationError(prefix+".email", "Duplicate email")
		}
		emails[key] = true
		if m.Role != "LEAD" && m.Role != "CO_RESEARCHER" {
			return repo.CreateResearchInput{}, validationError(prefix+".role", "Invalid role")
		}
		if roles[m.Role] {
			return repo.CreateResearchInput{}, validationError("projectMembers", "One LEAD and one CO_RESEARCHER are required")
		}
		roles[m.Role] = true
		p, err := decimal(m.ContributionPercent, prefix+".contributionPercent")
		if err != nil {
			return repo.CreateResearchInput{}, err
		}
		if p > 100 {
			return repo.CreateResearchInput{}, validationError(prefix+".contributionPercent", "Invalid contribution")
		}
		sum += p
		members[i] = repo.ResearchMemberInput{FullName: strings.TrimSpace(m.FullName), Email: email, Affiliation: strings.TrimSpace(m.Affiliation), ContributionPercent: p, Role: m.Role}
	}
	if math.Abs(sum-100) > 0.0001 {
		return repo.CreateResearchInput{}, validationError("projectMembers", "Contributions must total 100")
	}
	if in.ContractFile == nil || in.ContractFile.Reader == nil {
		return repo.CreateResearchInput{}, validationError("contractFile", "Contract file is required")
	}
	cn := strings.TrimSpace(in.ContractNumber)
	if err := text(cn, false, "contractNumber"); err != nil {
		return repo.CreateResearchInput{}, err
	}
	return repo.CreateResearchInput{Title: strings.TrimSpace(in.Title), IsSubsidized: *in.IsSubsidized, ProjectType: in.ProjectType, ResearchKind: in.ResearchKind, ContinuationOfID: in.ContinuationOfID, ResponsibleProjectUnit: strings.TrimSpace(in.ResponsibleProjectUnit), ResponsibleBudgetUnit: strings.TrimSpace(in.ResponsibleBudgetUnit), StartDate: start, EndDate: end, BudgetAmount: budget, ThaiAbstract: strings.TrimSpace(in.ThaiAbstract), EnglishAbstract: strings.TrimSpace(in.EnglishAbstract), Objectives: strings.TrimSpace(in.Objectives), Keywords: strings.TrimSpace(in.Keywords), Members: members, Contract: repo.ResearchContractInput{FundingType: in.FundingType, FundingSourceName: strings.TrimSpace(in.FundingSourceName), ContractNumber: cn, ContractNumberKey: contractKey(cn)}}, nil
}

func contractKey(s string) string {
	return strings.NewReplacer("ß", "ss", "ẞ", "ss").Replace(strings.ToLower(s))
}
func text(v string, title bool, f string) error {
	t := strings.TrimSpace(v)
	if !utf8.ValidString(v) || utf8.RuneCountInString(t) < 1 || utf8.RuneCountInString(t) > 1000 {
		return validationError(f, "Must contain 1-1000 characters")
	}
	for _, r := range v {
		if r == 0 || unicode.IsControl(r) && (title || r != '\n' && r != '\t') || title && r == '/' {
			return validationError(f, "Contains invalid characters")
		}
	}
	return nil
}
func decimal(n json.Number, f string) (float64, error) {
	v, err := strconv.ParseFloat(string(n), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || math.Abs(v*100-math.Round(v*100)) > 1e-8 {
		return 0, validationError(f, "Must be a positive number with at most two decimals")
	}
	return v, nil
}
func parseDates(a, b string) (string, string, error) {
	parse := func(v string) (time.Time, error) {
		if len(v) != 10 {
			return time.Time{}, errors.New("bad")
		}
		day, e1 := strconv.Atoi(v[0:2])
		month, e2 := strconv.Atoi(v[3:5])
		year, e3 := strconv.Atoi(v[6:10])
		if e1 != nil || e2 != nil || e3 != nil || v[2] != '/' || v[5] != '/' {
			return time.Time{}, errors.New("bad")
		}
		t := time.Date(year-543, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if t.Day() != day || int(t.Month()) != month || t.Year() != year-543 {
			return time.Time{}, errors.New("bad")
		}
		return t, nil
	}
	s, e1 := parse(a)
	e, e2 := parse(b)
	if e1 != nil {
		return "", "", validationError("startDate", "Invalid date")
	}
	if e2 != nil {
		return "", "", validationError("endDate", "Invalid date")
	}
	if e.Before(s.AddDate(1, 0, 0)) {
		return "", "", validationError("endDate", "Must be at least one year after start date")
	}
	return s.Format("2006-01-02"), e.Format("2006-01-02"), nil
}
func formatBuddhist(v string) string {
	t, _ := time.Parse("2006-01-02", v)
	return t.Format("02/01/") + strconv.Itoa(t.Year()+543)
}
func mailAddress(s string) string {
	a, e := mail.ParseAddress(s)
	if e != nil {
		return ""
	}
	return a.Address
}
func sanitize(err error) error {
	var c codedError
	if errors.As(err, &c) {
		return c
	}
	return internalError(err)
}
func mapRepoError(err error) error {
	switch {
	case errors.Is(err, repo.ErrContinuationNotFound):
		return codedError{code: "CONTINUATION_NOT_FOUND", message: "Continuation research was not found."}
	case errors.Is(err, repo.ErrTitleAlreadyExists):
		return codedError{code: "TITLE_ALREADY_EXISTS", message: "A research with this title already exists."}
	case errors.Is(err, repo.ErrContractNumberAlreadyExists):
		return codedError{code: "CONTRACT_NUMBER_ALREADY_EXISTS", message: "A research with this contract number already exists."}
	default:
		return internalError(err)
	}
}
