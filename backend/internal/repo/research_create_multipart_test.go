package repo_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/660710627/my-research/internal/domain"
	"github.com/660710627/my-research/internal/repo"
)

func TestResearchRepositoryCreateMultipartPersistsCompleteResearch(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)

	created, err := repository.CreateMultipart(context.Background(), multipartCreateRecord("โครงการหลัก", nil))
	if err != nil {
		t.Fatalf("create multipart research: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("id = %d, want a positive server-generated ID", created.ID)
	}
	if created.Title != "โครงการหลัก" || len(created.ProjectMembers) != 2 {
		t.Fatalf("created research = %#v, want complete persisted project data", created)
	}
	if created.Contract != (domain.ContractMetadata{Filename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 4}) {
		t.Fatalf("contract metadata = %#v, want persisted PDF metadata", created.Contract)
	}
	if created.Status != initialStatus || created.Process != initialProcess {
		t.Fatalf("status/process = %q/%q, want initial %q/%q", created.Status, created.Process, initialStatus, initialProcess)
	}
}

func TestResearchRepositoryCreateMultipartEnforcesContinuationAndRootTitleRules(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)

	parent, err := repository.CreateMultipart(context.Background(), multipartCreateRecord("ชื่อซ้ำ", nil))
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	_, err = repository.CreateMultipart(context.Background(), multipartCreateRecord("ชื่อซ้ำ", nil))
	if !errors.Is(err, repo.ErrTitleAlreadyExists) {
		t.Fatalf("duplicate root error = %v, want ErrTitleAlreadyExists", err)
	}

	missingID := int64(999)
	_, err = repository.CreateMultipart(context.Background(), multipartCreateRecord("งานต่อเนื่อง", &missingID))
	if !errors.Is(err, repo.ErrContinuationNotFound) {
		t.Fatalf("missing continuation error = %v, want ErrContinuationNotFound", err)
	}
	assertResearchCount(t, db, 1)

	child, err := repository.CreateMultipart(context.Background(), multipartCreateRecord("ชื่อซ้ำ", &parent.ID))
	if err != nil {
		t.Fatalf("create continuation: %v", err)
	}
	if child.ContinuationOfID == nil || *child.ContinuationOfID != parent.ID {
		t.Fatalf("continuationOfId = %v, want %d", child.ContinuationOfID, parent.ID)
	}
}

func TestResearchRepositoryCreateMultipartKeepsCreateAtomicAndIDsUnique(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)

	const workers = 4
	var wait sync.WaitGroup
	results := make(chan domain.Research, workers)
	errorsFound := make(chan error, workers)
	for index := 0; index < workers; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			created, err := repository.CreateMultipart(context.Background(), multipartCreateRecord("โครงการพร้อมกัน-"+string(rune('A'+index)), nil))
			if err != nil {
				errorsFound <- err
				return
			}
			results <- created
		}(index)
	}
	wait.Wait()
	close(results)
	close(errorsFound)

	seen := map[int64]bool{}
	for err := range errorsFound {
		t.Errorf("concurrent create: %v", err)
	}
	for created := range results {
		if created.ID <= 0 || seen[created.ID] {
			t.Errorf("id %d is non-positive or reused", created.ID)
		}
		seen[created.ID] = true
	}
	if len(seen) != workers {
		t.Fatalf("created IDs = %d, want %d", len(seen), workers)
	}
}

func multipartCreateRecord(title string, continuationOfID *int64) domain.CreateResearchRecord {
	return domain.CreateResearchRecord{
		ResearchData: domain.ResearchData{
			Title: title, ContinuationOfID: continuationOfID, IsSubsidized: true,
			ProjectMembers: []domain.ProjectMember{
				{FullName: "หัวหน้า", Email: "lead@example.com", Affiliation: "หน่วยงาน", ContributionPercent: 60, Role: domain.MemberRoleLead},
				{FullName: "ผู้ร่วม", Email: "co@example.com", Affiliation: "หน่วยงาน", ContributionPercent: 40, Role: domain.MemberRoleCoResearcher},
			},
			FundingType: domain.FundingTypeInternal, FundingSourceName: "แหล่งทุน", ContractNumber: "C-001",
			ProjectType: domain.ProjectTypeResearch,
			ResearchKind: func() domain.ResearchKind {
				if continuationOfID == nil {
					return domain.ResearchKindBudget
				}
				return domain.ResearchKindContinuation
			}(),
			ResponsibleProjectUnit: "หน่วยงานโครงการ", ResponsibleBudgetUnit: "หน่วยงานงบประมาณ",
			StartDate: "01/01/2569", EndDate: "31/12/2569", BudgetAmount: 1000,
			ThaiAbstract: "บทคัดย่อไทย", EnglishAbstract: "English abstract", Objectives: "วัตถุประสงค์", Keywords: "คำค้น",
		},
		Contract: domain.StagedContract{Token: "staged-contract", Metadata: domain.ContractMetadata{Filename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 4}},
	}
}
