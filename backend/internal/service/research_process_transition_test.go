package service

import (
	"context"
	"errors"
	"github.com/660710627/my-research/internal/repo"
	"testing"
)

func TestResearchProcessServiceValidatesAndMapsTypedErrors(t *testing.T) {
	for _, test := range []struct {
		input          UpdateResearchProcessInput
		returned, want error
	}{
		{UpdateResearchProcessInput{ID: 0, Process: "สัญญาโครงการ"}, nil, ErrValidation},
		{UpdateResearchProcessInput{ID: 1, Process: ""}, nil, ErrValidation},
		{UpdateResearchProcessInput{ID: 1, Process: "other"}, nil, ErrValidation},
		{UpdateResearchProcessInput{ID: 1, Process: "บันทึกข้อตกลง"}, repo.ErrResearchNotFound, ErrResearchNotFound},
		{UpdateResearchProcessInput{ID: 1, Process: "บันทึกข้อตกลง"}, repo.ErrInvalidProcessTransition, ErrInvalidProcessTransition},
		{UpdateResearchProcessInput{ID: 1, Process: "บันทึกข้อตกลง"}, repo.ErrProjectAlreadyEnded, ErrProjectAlreadyEnded},
	} {
		called := false
		svc := NewResearchProcessService(researchProcessStoreStub{update: func(context.Context, repo.UpdateResearchProcessParams) (repo.Research, error) {
			called = true
			return repo.Research{}, test.returned
		}})
		_, err := svc.UpdateProcess(context.Background(), test.input)
		if !errors.Is(err, test.want) {
			t.Fatalf("error=%v want=%v", err, test.want)
		}
		if test.want == ErrValidation && called {
			t.Fatal("store called")
		}
	}
}
