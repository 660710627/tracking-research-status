package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/660710627/my-research/internal/domain"
)

type MultipartResearchWriter interface {
	CreateMultipart(context.Context, domain.CreateResearchRecord) (domain.Research, error)
}

func (repository *ResearchRepository) CreateMultipart(ctx context.Context, record domain.CreateResearchRecord) (domain.Research, error) {
	repository.mutationMu.Lock()
	defer repository.mutationMu.Unlock()
	tx, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return domain.Research{}, fmt.Errorf("begin create research: %w", err)
	}
	defer tx.Rollback()
	data := record.ResearchData
	if data.ContinuationOfID != nil {
		var exists int
		err := tx.QueryRowContext(ctx, "SELECT 1 FROM researches WHERE id=?", *data.ContinuationOfID).Scan(&exists)
		if err == sql.ErrNoRows {
			return domain.Research{}, ErrContinuationNotFound
		}
		if err != nil {
			return domain.Research{}, err
		}
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO researches (title,continuation_of_id,is_subsidized,funding_type,funding_source_name,contract_number,contract_filename,contract_content_type,contract_size_bytes,contract_path,project_type,research_kind,responsible_project_unit,responsible_budget_unit,start_date,end_date,budget_amount,thai_abstract,english_abstract,objectives,keywords) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, data.Title, data.ContinuationOfID, data.IsSubsidized, data.FundingType, data.FundingSourceName, data.ContractNumber, record.Contract.Metadata.Filename, record.Contract.Metadata.ContentType, record.Contract.Metadata.SizeBytes, record.Contract.Token, data.ProjectType, data.ResearchKind, data.ResponsibleProjectUnit, data.ResponsibleBudgetUnit, data.StartDate, data.EndDate, data.BudgetAmount, data.ThaiAbstract, data.EnglishAbstract, data.Objectives, data.Keywords)
	if err != nil {
		return domain.Research{}, mapCreateError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Research{}, err
	}
	for _, member := range data.ProjectMembers {
		if _, err := tx.ExecContext(ctx, `INSERT INTO research_members (research_id,full_name,email,affiliation,contribution_percent,role) VALUES (?,?,?,?,?,?)`, id, member.FullName, member.Email, member.Affiliation, member.ContributionPercent, member.Role); err != nil {
			return domain.Research{}, fmt.Errorf("insert research member: %w", err)
		}
	}
	created, err := loadResearch(ctx, tx, id)
	if err != nil {
		return domain.Research{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Research{}, err
	}
	return created, nil
}
