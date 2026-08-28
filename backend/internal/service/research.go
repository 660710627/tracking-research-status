package service

import (
	"errors"

	"github.com/660710627/my-research/internal/domain"
)

var (
	ErrValidation               = errors.New("validation failed")
	ErrContinuationNotFound     = errors.New("continuation research not found")
	ErrResearchHasContinuations = errors.New("research has continuations")
	ErrResearchNotFound         = errors.New("research not found")
	ErrTitleAlreadyExists       = errors.New("research title already exists")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrProjectAlreadyEnded      = errors.New("project already ended")
	ErrInvalidProcessTransition = errors.New("invalid process transition")
	ErrInternal                 = errors.New("internal service error")
)

type Research = domain.Research
