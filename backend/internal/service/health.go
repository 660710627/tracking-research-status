package service

import (
	"context"
	"fmt"
)

const CodeServiceUnavailable = "SERVICE_UNAVAILABLE"

type HealthRepository interface {
	Ping(context.Context) error
}

type HealthService struct {
	repository HealthRepository
}

func NewHealthService(repository HealthRepository) *HealthService {
	return &HealthService{repository: repository}
}

func (s *HealthService) CheckHealth(ctx context.Context) error {
	if err := s.repository.Ping(ctx); err != nil {
		return &HealthError{code: CodeServiceUnavailable, cause: err}
	}
	return nil
}

type HealthError struct {
	code  string
	cause error
}

func (e *HealthError) Error() string {
	return fmt.Sprintf("health check failed: %v", e.cause)
}

func (e *HealthError) Unwrap() error {
	return e.cause
}

func (e *HealthError) ErrorCode() string {
	return e.code
}
