package repo

import (
	"context"
	"fmt"
)

type Research struct {
	ID                     int64
	Title                  string
	ContinuationOfID       *int64
	IsSubsidized           bool
	ProjectType            string
	ResearchKind           string
	ResponsibleProjectUnit string
	ResponsibleBudgetUnit  string
	StartDate              string
	EndDate                string
	BudgetAmount           float64
	ThaiAbstract           string
	EnglishAbstract        string
	Objectives             string
	Keywords               string
	Status                 string
	Process                string
	Members                []ResearchMember
	Contract               ResearchContract
}

// ResearchMember shares the persisted member value type without exposing storage details.
type ResearchMember = ResearchMemberInput
type ResearchContract struct {
	FundingType, FundingSourceName, ContractNumber string
	OriginalFilename, ContentType                  string
	SizeBytes                                      int64
}

// List reads a single SQL snapshot and groups member rows without losing duplicate titles.
func (r *ResearchRepository) List(ctx context.Context) ([]Research, error) {
	rows, err := r.database.QueryContext(ctx, `SELECT r.id,r.title,r.continuation_of_id,r.is_subsidized,
r.project_type,r.research_kind,r.responsible_project_unit,r.responsible_budget_unit,
r.start_date,r.end_date,r.budget_amount,r.thai_abstract,r.english_abstract,r.objectives,r.keywords,r.status,r.process,
c.funding_type,c.funding_source_name,c.contract_number,c.original_filename,c.content_type,c.size_bytes,
m.full_name,m.email,m.affiliation,m.contribution_percent,m.role
FROM researches r JOIN research_contracts c ON c.research_id=r.id
JOIN research_members m ON m.research_id=r.id
ORDER BY r.title ASC,r.id ASC,m.role DESC`)
	if err != nil {
		return nil, fmt.Errorf("%w: list: %v", ErrInternal, err)
	}
	defer rows.Close()
	result := make([]Research, 0)
	for rows.Next() {
		var item Research
		var member ResearchMember
		c := &item.Contract
		if err := rows.Scan(&item.ID, &item.Title, &item.ContinuationOfID, &item.IsSubsidized, &item.ProjectType, &item.ResearchKind, &item.ResponsibleProjectUnit, &item.ResponsibleBudgetUnit, &item.StartDate, &item.EndDate, &item.BudgetAmount, &item.ThaiAbstract, &item.EnglishAbstract, &item.Objectives, &item.Keywords, &item.Status, &item.Process,
			&c.FundingType, &c.FundingSourceName, &c.ContractNumber, &c.OriginalFilename, &c.ContentType, &c.SizeBytes,
			&member.FullName, &member.Email, &member.Affiliation, &member.ContributionPercent, &member.Role); err != nil {
			return nil, fmt.Errorf("%w: scan list: %v", ErrInternal, err)
		}
		if len(result) == 0 || result[len(result)-1].ID != item.ID {
			result = append(result, item)
		}
		last := &result[len(result)-1]
		last.Members = append(last.Members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: read list: %v", ErrInternal, err)
	}
	return result, nil
}
