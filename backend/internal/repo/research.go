package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/660710627/my-research/internal/domain"
)

var (
	ErrContinuationNotFound     = errors.New("continuation research not found")
	ErrResearchHasContinuations = errors.New("research has continuations")
	ErrResearchNotFound         = errors.New("research not found")
	ErrTitleAlreadyExists       = errors.New("research title already exists")
)

type Research = domain.Research

type ResearchRepository struct {
	database   *sql.DB
	mutationMu sync.Mutex
}

func NewResearchRepository(database *sql.DB) *ResearchRepository {
	return &ResearchRepository{database: database}
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

const researchColumns = `id,title,continuation_of_id,is_subsidized,funding_type,funding_source_name,contract_number,contract_filename,contract_content_type,contract_size_bytes,project_type,research_kind,responsible_project_unit,responsible_budget_unit,start_date,end_date,budget_amount,thai_abstract,english_abstract,objectives,keywords,status,process`

func loadResearch(ctx context.Context, source queryer, id int64) (domain.Research, error) {
	row := source.QueryRowContext(ctx, `SELECT `+researchColumns+` FROM researches WHERE id=?`, id)
	var research domain.Research
	var continuation sql.NullInt64
	var subsidized int
	err := row.Scan(&research.ID, &research.Title, &continuation, &subsidized, &research.FundingType, &research.FundingSourceName, &research.ContractNumber, &research.Contract.Filename, &research.Contract.ContentType, &research.Contract.SizeBytes, &research.ProjectType, &research.ResearchKind, &research.ResponsibleProjectUnit, &research.ResponsibleBudgetUnit, &research.StartDate, &research.EndDate, &research.BudgetAmount, &research.ThaiAbstract, &research.EnglishAbstract, &research.Objectives, &research.Keywords, &research.Status, &research.Process)
	if err != nil {
		return domain.Research{}, err
	}
	research.IsSubsidized = subsidized == 1
	if continuation.Valid {
		value := continuation.Int64
		research.ContinuationOfID = &value
	}
	rows, err := source.QueryContext(ctx, `SELECT full_name,email,affiliation,contribution_percent,role FROM research_members WHERE research_id=? ORDER BY rowid`, id)
	if err != nil {
		return domain.Research{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var member domain.ProjectMember
		if err := rows.Scan(&member.FullName, &member.Email, &member.Affiliation, &member.ContributionPercent, &member.Role); err != nil {
			return domain.Research{}, err
		}
		research.ProjectMembers = append(research.ProjectMembers, member)
	}
	return research, rows.Err()
}

func mapCreateError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "CONTINUATION_NOT_FOUND"):
		return ErrContinuationNotFound
	case strings.Contains(message, "TITLE_ALREADY_EXISTS"):
		return ErrTitleAlreadyExists
	default:
		return fmt.Errorf("insert research: %w", err)
	}
}
