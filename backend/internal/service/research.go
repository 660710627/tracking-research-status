package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/660710627/my-research/internal/repo"
)

var (
	ErrValidation               = errors.New("validation failed")
	ErrContinuationNotFound     = errors.New("continuation research not found")
	ErrResearchHasContinuations = errors.New("research has continuations")
	ErrResearchNotFound         = errors.New("research not found")
	ErrTitleAlreadyExists       = errors.New("research title already exists")
	ErrInternal                 = errors.New("internal service error")
)

type Research struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	ContinuationOfID *int64 `json:"continuationOfId"`
	Status           string `json:"status"`
	Process          string `json:"process"`
}

type CreateResearchInput struct {
	Title            string
	Description      string
	ContinuationOfID *int64
}

type ResearchStore interface {
	Create(context.Context, repo.CreateResearchParams) (repo.Research, error)
}

type ResearchService struct {
	store ResearchStore
}

func NewResearchService(store ResearchStore) *ResearchService {
	return &ResearchService{store: store}
}

func (service *ResearchService) Create(ctx context.Context, input CreateResearchInput) (Research, error) {
	title := strings.TrimFunc(input.Title, unicode.IsSpace)
	description := strings.ReplaceAll(input.Description, "\r\n", "\n")
	description = strings.TrimFunc(description, unicode.IsSpace)

	if !validResearchTitle(title) || !validResearchDescription(description) || input.ContinuationOfID != nil && *input.ContinuationOfID <= 0 {
		return Research{}, ErrValidation
	}

	created, err := service.store.Create(ctx, repo.CreateResearchParams{
		Title: title, Description: description, ContinuationOfID: input.ContinuationOfID,
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrContinuationNotFound):
			return Research{}, ErrContinuationNotFound
		case errors.Is(err, repo.ErrTitleAlreadyExists):
			return Research{}, ErrTitleAlreadyExists
		default:
			return Research{}, fmt.Errorf("%w: create research", ErrInternal)
		}
	}
	return Research{
		ID: created.ID, Title: created.Title, Description: created.Description,
		ContinuationOfID: created.ContinuationOfID, Status: created.Status, Process: created.Process,
	}, nil
}

func validResearchTitle(value string) bool {
	length := utf8.RuneCountInString(value)
	if !utf8.ValidString(value) || length < 1 || length > 200 || strings.ContainsRune(value, '/') {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validResearchDescription(value string) bool {
	length := utf8.RuneCountInString(value)
	if !utf8.ValidString(value) || length < 1 || length > 5000 {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) && character != '\n' && character != '\t' {
			return false
		}
	}
	return true
}
