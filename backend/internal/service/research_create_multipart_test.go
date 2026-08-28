package service_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/660710627/my-research/internal/domain"
	"github.com/660710627/my-research/internal/service"
)

func TestMultipartResearchServiceStagesPersistsPublishesAndReturnsCompleteResearch(t *testing.T) {
	storage := newContractStagerStub(t)
	store := &multipartResearchStoreStub{create: func(_ context.Context, record domain.CreateResearchRecord) (domain.Research, error) {
		if record.Contract.Token == "" {
			t.Fatal("repository received an unstaged contract")
		}
		return domain.Research{ID: 41, ResearchData: record.ResearchData, Contract: record.Contract.Metadata, Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ"}, nil
	}}
	serviceUnderTest := service.NewMultipartResearchService(store, storage)

	created, err := serviceUnderTest.CreateMultipart(context.Background(), validMultipartCreateInput())
	if err != nil {
		t.Fatalf("create multipart research: %v", err)
	}
	if !storage.staged || !storage.published || storage.discarded {
		t.Fatalf("file lifecycle stage/publish/discard = %t/%t/%t, want true/true/false", storage.staged, storage.published, storage.discarded)
	}
	if created.ID != 41 || created.Contract.Filename != "contract.pdf" || created.Status != "กำลังดำเนินการ" || created.Process != "สัญญาโครงการ" {
		t.Fatalf("created research = %#v, want complete latest record", created)
	}
}

func TestMultipartResearchServiceRejectsInvalidProjectDataBeforeStaging(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*domain.CreateResearchInput)
	}{
		{name: "blank title", mutate: func(input *domain.CreateResearchInput) { input.Title = " " }},
		{name: "title over 1000", mutate: func(input *domain.CreateResearchInput) { input.Title = string(bytes.Repeat([]byte("ก"), 1001)) }},
		{name: "invalid funding type", mutate: func(input *domain.CreateResearchInput) { input.FundingType = "OTHER" }},
		{name: "invalid project type", mutate: func(input *domain.CreateResearchInput) { input.ProjectType = "OTHER" }},
		{name: "invalid research kind", mutate: func(input *domain.CreateResearchInput) { input.ResearchKind = "OTHER" }},
		{name: "invalid Buddhist date", mutate: func(input *domain.CreateResearchInput) { input.StartDate = "2569-01-01" }},
		{name: "non-positive budget", mutate: func(input *domain.CreateResearchInput) { input.BudgetAmount = 0 }},
		{name: "more than two decimal budget", mutate: func(input *domain.CreateResearchInput) { input.BudgetAmount = 1.001 }},
		{name: "missing lead", mutate: func(input *domain.CreateResearchInput) { input.ProjectMembers[0].Role = domain.MemberRoleCoResearcher }},
		{name: "missing co-researcher", mutate: func(input *domain.CreateResearchInput) { input.ProjectMembers[1].Role = domain.MemberRoleLead }},
		{name: "invalid member email", mutate: func(input *domain.CreateResearchInput) { input.ProjectMembers[0].Email = "not-an-email" }},
		{name: "invalid contribution", mutate: func(input *domain.CreateResearchInput) { input.ProjectMembers[0].ContributionPercent = 100.001 }},
		{name: "not a PDF", mutate: func(input *domain.CreateResearchInput) { input.Contract.ContentType = "image/png" }},
		{name: "missing continuation for continuation kind", mutate: func(input *domain.CreateResearchInput) { input.ResearchKind = domain.ResearchKindContinuation }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := newContractStagerStub(t)
			store := &multipartResearchStoreStub{}
			serviceUnderTest := service.NewMultipartResearchService(store, storage)
			input := validMultipartCreateInput()
			test.mutate(&input)

			_, err := serviceUnderTest.CreateMultipart(context.Background(), input)
			if !errors.Is(err, service.ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
			if storage.staged || store.called {
				t.Fatal("invalid input must not stage a file or call the repository")
			}
		})
	}
}

func TestMultipartResearchServiceDiscardsStagedFileWhenPersistenceFails(t *testing.T) {
	storage := newContractStagerStub(t)
	store := &multipartResearchStoreStub{err: errors.New("database unavailable")}
	serviceUnderTest := service.NewMultipartResearchService(store, storage)

	_, err := serviceUnderTest.CreateMultipart(context.Background(), validMultipartCreateInput())
	if !errors.Is(err, service.ErrInternal) {
		t.Fatalf("error = %v, want ErrInternal", err)
	}
	if !storage.staged || !storage.discarded || storage.published {
		t.Fatalf("file lifecycle stage/discard/publish = %t/%t/%t, want true/true/false", storage.staged, storage.discarded, storage.published)
	}
}

func TestMultipartResearchServiceMapsStageFailureWithoutPersistence(t *testing.T) {
	storage := newContractStagerStub(t)
	storage.stageErr = errors.New("temporary storage unavailable")
	store := &multipartResearchStoreStub{}
	serviceUnderTest := service.NewMultipartResearchService(store, storage)

	_, err := serviceUnderTest.CreateMultipart(context.Background(), validMultipartCreateInput())
	if !errors.Is(err, service.ErrContractStage) {
		t.Fatalf("error = %v, want ErrContractStage", err)
	}
	if store.called || storage.discarded || storage.published {
		t.Fatal("stage failure must not persist, publish, or discard an unstaged file")
	}
}

func TestMultipartResearchServiceDiscardsStagedFileWhenPublishFails(t *testing.T) {
	storage := newContractStagerStub(t)
	storage.publishErr = errors.New("final storage unavailable")
	store := &multipartResearchStoreStub{create: func(_ context.Context, record domain.CreateResearchRecord) (domain.Research, error) {
		return domain.Research{ID: 42, ResearchData: record.ResearchData, Contract: record.Contract.Metadata, Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ"}, nil
	}}
	serviceUnderTest := service.NewMultipartResearchService(store, storage)

	_, err := serviceUnderTest.CreateMultipart(context.Background(), validMultipartCreateInput())
	if !errors.Is(err, service.ErrContractPublish) {
		t.Fatalf("error = %v, want ErrContractPublish", err)
	}
	if !storage.staged || !storage.discarded || !storage.published {
		t.Fatalf("file lifecycle stage/discard/publish = %t/%t/%t, want true/true/true", storage.staged, storage.discarded, storage.published)
	}
}

type multipartResearchStoreStub struct {
	create func(context.Context, domain.CreateResearchRecord) (domain.Research, error)
	err    error
	called bool
}

func (stub *multipartResearchStoreStub) CreateMultipart(ctx context.Context, record domain.CreateResearchRecord) (domain.Research, error) {
	stub.called = true
	if stub.create != nil {
		return stub.create(ctx, record)
	}
	return domain.Research{}, stub.err
}

type contractStagerStub struct {
	staged, published, discarded bool
	stageErr, publishErr         error
}

func newContractStagerStub(t *testing.T) *contractStagerStub {
	t.Helper()
	_ = t.TempDir()
	return &contractStagerStub{}
}

func (stub *contractStagerStub) Stage(_ context.Context, upload domain.ContractUpload) (domain.StagedContract, error) {
	stub.staged = true
	if stub.stageErr != nil {
		return domain.StagedContract{}, stub.stageErr
	}
	return domain.StagedContract{Token: "temporary-contract", Metadata: domain.ContractMetadata{Filename: upload.Filename, ContentType: upload.ContentType, SizeBytes: upload.SizeBytes}}, nil
}

func (stub *contractStagerStub) Publish(context.Context, domain.StagedContract) (domain.ContractMetadata, error) {
	stub.published = true
	if stub.publishErr != nil {
		return domain.ContractMetadata{}, stub.publishErr
	}
	return domain.ContractMetadata{Filename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 4}, nil
}

func (stub *contractStagerStub) Discard(context.Context, domain.StagedContract) error {
	stub.discarded = true
	return nil
}

func validMultipartCreateInput() domain.CreateResearchInput {
	return domain.CreateResearchInput{ResearchData: domain.ResearchData{
		Title: "โครงการใหม่", IsSubsidized: true,
		ProjectMembers: []domain.ProjectMember{
			{FullName: "หัวหน้า", Email: "lead@example.com", Affiliation: "หน่วยงาน", ContributionPercent: 60, Role: domain.MemberRoleLead},
			{FullName: "ผู้ร่วม", Email: "co@example.com", Affiliation: "หน่วยงาน", ContributionPercent: 40, Role: domain.MemberRoleCoResearcher},
		},
		FundingType: domain.FundingTypeInternal, FundingSourceName: "แหล่งทุน", ContractNumber: "C-001", ProjectType: domain.ProjectTypeResearch, ResearchKind: domain.ResearchKindBudget,
		ResponsibleProjectUnit: "หน่วยงานโครงการ", ResponsibleBudgetUnit: "หน่วยงานงบประมาณ", StartDate: "01/01/2569", EndDate: "31/12/2569", BudgetAmount: 1000,
		ThaiAbstract: "บทคัดย่อไทย", EnglishAbstract: "English abstract", Objectives: "วัตถุประสงค์", Keywords: "คำค้น",
	}, Contract: domain.ContractUpload{Filename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 4, Content: bytes.NewReader([]byte("%PDF"))}}
}
