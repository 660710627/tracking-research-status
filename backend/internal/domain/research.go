package domain

import "io"

type FundingType string

const (
	FundingTypeInternal FundingType = "INTERNAL"
	FundingTypeExternal FundingType = "EXTERNAL"
)

type ProjectType string

const (
	ProjectTypeResearch        ProjectType = "RESEARCH"
	ProjectTypeAcademicService ProjectType = "ACADEMIC_SERVICE"
)

type ResearchKind string

const (
	ResearchKindBudget       ResearchKind = "BUDGET"
	ResearchKindContinuation ResearchKind = "CONTINUATION"
)

type MemberRole string

const (
	MemberRoleLead         MemberRole = "LEAD"
	MemberRoleCoResearcher MemberRole = "CO_RESEARCHER"
)

type ProjectMember struct {
	FullName            string     `json:"fullName"`
	Email               string     `json:"email"`
	Affiliation         string     `json:"affiliation"`
	ContributionPercent float64    `json:"contributionPercent"`
	Role                MemberRole `json:"role"`
}

type ResearchData struct {
	Title                  string          `json:"title"`
	ContinuationOfID       *int64          `json:"continuationOfId"`
	IsSubsidized           bool            `json:"isSubsidized"`
	ProjectMembers         []ProjectMember `json:"projectMembers"`
	FundingType            FundingType     `json:"fundingType"`
	FundingSourceName      string          `json:"fundingSourceName"`
	ContractNumber         string          `json:"contractNumber"`
	ProjectType            ProjectType     `json:"projectType"`
	ResearchKind           ResearchKind    `json:"researchKind"`
	ResponsibleProjectUnit string          `json:"responsibleProjectUnit"`
	ResponsibleBudgetUnit  string          `json:"responsibleBudgetUnit"`
	StartDate              string          `json:"startDate"`
	EndDate                string          `json:"endDate"`
	BudgetAmount           float64         `json:"budgetAmount"`
	ThaiAbstract           string          `json:"thaiAbstract"`
	EnglishAbstract        string          `json:"englishAbstract"`
	Objectives             string          `json:"objectives"`
	Keywords               string          `json:"keywords"`
}

type ContractUpload struct {
	Filename    string
	ContentType string
	SizeBytes   int64
	Content     io.Reader
}

type ContractMetadata struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

// StagedContract deliberately keeps the temporary storage identifier opaque to
// callers outside the file-lifecycle port.
type StagedContract struct {
	Token    string
	Metadata ContractMetadata
}

type CreateResearchInput struct {
	ResearchData
	Contract ContractUpload
}

type CreateResearchRecord struct {
	ResearchData
	Contract StagedContract
}

type Research struct {
	ID int64 `json:"id"`
	ResearchData
	Contract ContractMetadata `json:"contractFile"`
	Status   string           `json:"status"`
	Process  string           `json:"process"`
}
