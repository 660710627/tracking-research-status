package repo

import (
	"context"
	"testing"
)

func TestResearchRepositoryExposesProcessUpdateContract(t *testing.T) {
	var _ interface {
		UpdateProcess(context.Context, UpdateResearchProcessParams) (Research, error)
	} = NewResearchRepository(nil)
}
