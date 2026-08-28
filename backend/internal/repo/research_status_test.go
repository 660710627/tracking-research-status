package repo

import (
	"context"
	"testing"
)

func TestResearchRepositoryExposesStatusUpdateContract(t *testing.T) {
	var _ interface {
		UpdateStatus(context.Context, UpdateResearchStatusParams) (Research, error)
	} = NewResearchRepository(nil)
}
