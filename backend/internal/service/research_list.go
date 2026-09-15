package service

import (
	"context"
	"github.com/660710627/my-research/internal/repo"
)

type Research struct {
	ID                     int64                `json:"id"`
	Title                  string               `json:"title"`
	ContinuationOfID       *int64               `json:"continuationOfId"`
	IsSubsidized           bool                 `json:"isSubsidized"`
	ProjectType            string               `json:"projectType"`
	ResearchKind           string               `json:"researchKind"`
	ResponsibleProjectUnit string               `json:"responsibleProjectUnit"`
	ResponsibleBudgetUnit  string               `json:"responsibleBudgetUnit"`
	StartDate              string               `json:"startDate"`
	EndDate                string               `json:"endDate"`
	BudgetAmount           float64              `json:"budgetAmount"`
	ThaiAbstract           string               `json:"thaiAbstract"`
	EnglishAbstract        string               `json:"englishAbstract"`
	Objectives             string               `json:"objectives"`
	Keywords               string               `json:"keywords"`
	Status                 string               `json:"status"`
	Process                string               `json:"process"`
	ProjectMembers         []ResearchMember     `json:"projectMembers"`
	FundingType            string               `json:"fundingType"`
	FundingSourceName      string               `json:"fundingSourceName"`
	ContractNumber         string               `json:"contractNumber"`
	ContractFile           ContractFileMetadata `json:"contractFile"`
}
type ResearchMember struct {
	FullName            string  `json:"fullName"`
	Email               string  `json:"email"`
	Affiliation         string  `json:"affiliation"`
	ContributionPercent float64 `json:"contributionPercent"`
	Role                string  `json:"role"`
}
type ContractFileMetadata struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}
type researchListRepository interface {
	List(context.Context) ([]repo.Research, error)
}
type ResearchListService struct{ repository researchListRepository }

func NewResearchListService(r researchListRepository) *ResearchListService {
	return &ResearchListService{repository: r}
}
func (s *ResearchListService) List(ctx context.Context) ([]Research, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	result := make([]Research, 0, len(items))
	for _, item := range items {
		members := make([]ResearchMember, 0, len(item.Members))
		for _, m := range item.Members {
			members = append(members, ResearchMember{FullName: m.FullName, Email: m.Email, Affiliation: m.Affiliation, ContributionPercent: m.ContributionPercent, Role: m.Role})
		}
		result = append(result, Research{
			ID:                     item.ID,
			Title:                  item.Title,
			ContinuationOfID:       item.ContinuationOfID,
			IsSubsidized:           item.IsSubsidized,
			ProjectType:            item.ProjectType,
			ResearchKind:           item.ResearchKind,
			ResponsibleProjectUnit: item.ResponsibleProjectUnit,
			ResponsibleBudgetUnit:  item.ResponsibleBudgetUnit,
			StartDate:              formatBuddhist(item.StartDate),
			EndDate:                formatBuddhist(item.EndDate),
			BudgetAmount:           item.BudgetAmount,
			ThaiAbstract:           item.ThaiAbstract,
			EnglishAbstract:        item.EnglishAbstract,
			Objectives:             item.Objectives,
			Keywords:               item.Keywords,
			Status:                 item.Status,
			Process:                item.Process,
			ProjectMembers:         members, FundingType: item.Contract.FundingType, FundingSourceName: item.Contract.FundingSourceName, ContractNumber: item.Contract.ContractNumber,
			ContractFile: ContractFileMetadata{Filename: item.Contract.OriginalFilename, ContentType: item.Contract.ContentType, SizeBytes: item.Contract.SizeBytes},
		})
	}
	return result, nil
}
