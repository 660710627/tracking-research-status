package service

import (
	"context"
	"errors"
	"fmt"
)

var ErrServiceUnavailable = errors.New("service unavailable")

type HealthRepository interface {
	Check(context.Context) error
}

type HealthService struct {
	repository HealthRepository
}

func NewHealthService(repository HealthRepository) *HealthService {
	return &HealthService{repository: repository}
}

func (service *HealthService) Check(ctx context.Context) error {
	if err := service.repository.Check(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrServiceUnavailable, err)
	}
	return nil
}
