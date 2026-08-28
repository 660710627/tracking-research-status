package handler

import (
	"context"
	"errors"

	"github.com/660710627/my-research/internal/domain"
)

var ErrMultipartCreateNotImplemented = errors.New("multipart research create is not implemented")

// MultipartResearchCreator is the HTTP-handler seam for the multipart create
// contract. Request parsing and route wiring are intentionally deferred to T-06.
type MultipartResearchCreator interface {
	CreateMultipart(context.Context, domain.CreateResearchInput) (domain.Research, error)
}

// MultipartResearchHandler is an HTTP-layer seam. Multipart parsing and route
// registration belong to T-06.
type MultipartResearchHandler struct {
	creator MultipartResearchCreator
}

func NewMultipartResearchHandler(creator MultipartResearchCreator) *MultipartResearchHandler {
	return &MultipartResearchHandler{creator: creator}
}

func (handler *MultipartResearchHandler) Create(ctx context.Context, input domain.CreateResearchInput) (domain.Research, error) {
	if handler.creator == nil {
		return domain.Research{}, ErrMultipartCreateNotImplemented
	}
	return handler.creator.CreateMultipart(ctx, input)
}
