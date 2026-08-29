package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/660710627/my-research/internal/domain"
	"github.com/660710627/my-research/internal/repo"
)

var (
	ErrContractStage   = errors.New("contract staging failed")
	ErrContractPublish = errors.New("contract publish failed")
)

type MultipartResearchCreator interface {
	CreateMultipart(context.Context, domain.CreateResearchInput) (domain.Research, error)
}
type MultipartResearchStore interface {
	CreateMultipart(context.Context, domain.CreateResearchRecord) (domain.Research, error)
}
type ContractStager interface {
	Stage(context.Context, domain.ContractUpload) (domain.StagedContract, error)
	Publish(context.Context, domain.StagedContract) (domain.ContractMetadata, error)
	Discard(context.Context, domain.StagedContract) error
}
type MultipartResearchService struct {
	store  MultipartResearchStore
	stager ContractStager
}

func NewMultipartResearchService(store MultipartResearchStore, stager ContractStager) *MultipartResearchService {
	return &MultipartResearchService{store: store, stager: stager}
}

func (service *MultipartResearchService) CreateMultipart(ctx context.Context, input domain.CreateResearchInput) (domain.Research, error) {
	normalizeResearchData(&input.ResearchData)
	if !validCreateInput(input) {
		return domain.Research{}, ErrValidation
	}
	if service.store == nil || service.stager == nil {
		return domain.Research{}, fmt.Errorf("%w: create dependencies", ErrInternal)
	}
	staged, err := service.stager.Stage(ctx, input.Contract)
	if err != nil {
		return domain.Research{}, fmt.Errorf("%w: %v", ErrContractStage, err)
	}
	created, err := service.store.CreateMultipart(ctx, domain.CreateResearchRecord{ResearchData: input.ResearchData, Contract: staged})
	if err != nil {
		_ = service.stager.Discard(ctx, staged)
		switch {
		case errors.Is(err, repo.ErrContinuationNotFound):
			return domain.Research{}, ErrContinuationNotFound
		case errors.Is(err, repo.ErrTitleAlreadyExists):
			return domain.Research{}, ErrTitleAlreadyExists
		default:
			return domain.Research{}, fmt.Errorf("%w: create research", ErrInternal)
		}
	}
	metadata, err := service.stager.Publish(ctx, staged)
	if err != nil {
		_ = service.stager.Discard(ctx, staged)
		if rollback, ok := service.store.(interface {
			Delete(context.Context, int64) error
		}); ok {
			_ = rollback.Delete(ctx, created.ID)
		}
		return domain.Research{}, fmt.Errorf("%w: %v", ErrContractPublish, err)
	}
	created.Contract = metadata
	return created, nil
}

func normalizeResearchData(data *domain.ResearchData) {
	data.Title = normalizeText(data.Title)
	data.FundingSourceName = normalizeText(data.FundingSourceName)
	data.ContractNumber = normalizeText(data.ContractNumber)
	data.ResponsibleProjectUnit = normalizeText(data.ResponsibleProjectUnit)
	data.ResponsibleBudgetUnit = normalizeText(data.ResponsibleBudgetUnit)
	data.ThaiAbstract = normalizeText(data.ThaiAbstract)
	data.EnglishAbstract = normalizeText(data.EnglishAbstract)
	data.Objectives = normalizeText(data.Objectives)
	data.Keywords = normalizeText(data.Keywords)
	for index := range data.ProjectMembers {
		data.ProjectMembers[index].FullName = normalizeText(data.ProjectMembers[index].FullName)
		data.ProjectMembers[index].Email = normalizeText(data.ProjectMembers[index].Email)
		data.ProjectMembers[index].Affiliation = normalizeText(data.ProjectMembers[index].Affiliation)
	}
}
func normalizeText(value string) string { return strings.TrimFunc(value, unicode.IsSpace) }
func validText(value string, title bool) bool {
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > 1000 || title && strings.ContainsRune(value, '/') {
		return false
	}
	for _, value := range value {
		if unicode.IsControl(value) {
			return false
		}
	}
	return true
}
func validCreateInput(input domain.CreateResearchInput) bool {
	d := input.ResearchData
	if !validText(d.Title, true) || !validText(d.FundingSourceName, false) || !validText(d.ContractNumber, false) || !validText(d.ResponsibleProjectUnit, false) || !validText(d.ResponsibleBudgetUnit, false) || !validText(d.ThaiAbstract, false) || !validText(d.EnglishAbstract, false) || !validText(d.Objectives, false) || !validText(d.Keywords, false) || (d.FundingType != domain.FundingTypeInternal && d.FundingType != domain.FundingTypeExternal) || (d.ProjectType != domain.ProjectTypeResearch && d.ProjectType != domain.ProjectTypeAcademicService) || (d.ResearchKind != domain.ResearchKindBudget && d.ResearchKind != domain.ResearchKindContinuation) || !validDate(d.StartDate) || !validDate(d.EndDate) || d.BudgetAmount <= 0 || !twoDecimal(d.BudgetAmount) || input.Contract.Filename == "" || input.Contract.ContentType != "application/pdf" || input.Contract.SizeBytes < 1 || input.Contract.SizeBytes > 20*1024*1024 || input.Contract.Content == nil {
		return false
	}
	if d.ResearchKind == domain.ResearchKindBudget && d.ContinuationOfID != nil || d.ResearchKind == domain.ResearchKindContinuation && (d.ContinuationOfID == nil || *d.ContinuationOfID <= 0) {
		return false
	}
	lead, co := false, false
	for _, member := range d.ProjectMembers {
		if !validText(member.FullName, false) || !validText(member.Affiliation, false) || !validText(member.Email, false) || !twoDecimal(member.ContributionPercent) || member.ContributionPercent <= 0 || member.ContributionPercent > 100 {
			return false
		}
		if _, err := mail.ParseAddress(member.Email); err != nil {
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
func validDate(value string) bool {
	parsed, err := time.Parse("02/01/2006", value)
	return err == nil && parsed.Year() > 0
}
func twoDecimal(value float64) bool { return math.Abs(value*100-math.Round(value*100)) < 0.000001 }
