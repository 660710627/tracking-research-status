package service

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/660710627/my-research/internal/domain"
)

type LocalContractStager struct{ directory string }

func NewLocalContractStager(directory string) *LocalContractStager {
	return &LocalContractStager{directory: directory}
}

func (stager *LocalContractStager) Stage(_ context.Context, upload domain.ContractUpload) (domain.StagedContract, error) {
	if err := os.MkdirAll(stager.directory, 0o750); err != nil {
		return domain.StagedContract{}, err
	}
	file, err := os.CreateTemp(stager.directory, "contract-*.pdf")
	if err != nil {
		return domain.StagedContract{}, err
	}
	name := file.Name()
	if _, err := io.Copy(file, io.LimitReader(upload.Content, 20*1024*1024+1)); err != nil {
		_ = file.Close()
		_ = os.Remove(name)
		return domain.StagedContract{}, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(name)
		return domain.StagedContract{}, err
	}
	return domain.StagedContract{Token: name, Metadata: domain.ContractMetadata{Filename: upload.Filename, ContentType: upload.ContentType, SizeBytes: upload.SizeBytes}}, nil
}
func (stager *LocalContractStager) Publish(_ context.Context, staged domain.StagedContract) (domain.ContractMetadata, error) {
	if _, err := os.Stat(staged.Token); err != nil {
		return domain.ContractMetadata{}, err
	}
	return staged.Metadata, nil
}
func (stager *LocalContractStager) Discard(_ context.Context, staged domain.StagedContract) error {
	if staged.Token == "" {
		return nil
	}
	if err := os.Remove(staged.Token); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("discard staged contract: %w", err)
	}
	return nil
}
