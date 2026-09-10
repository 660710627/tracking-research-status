package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const maxContractBytes int64 = 20 * 1024 * 1024

type StagedContract struct {
	Token, Filename string
	SizeBytes       int64
}
type PublishedContract struct {
	Path, Filename string
	SizeBytes      int64
}

type ContractStore interface {
	Stage(context.Context, string, io.Reader) (StagedContract, error)
	Publish(context.Context, StagedContract) (PublishedContract, error)
	Discard(context.Context, StagedContract) error
	Remove(context.Context, PublishedContract) error
}

type FileContractStore struct {
	root string
	mu   sync.Mutex
}

func NewFileContractStore(root string) (*FileContractStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("storage root is empty")
	}
	if err := os.MkdirAll(root, 0750); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}
	return &FileContractStore{root: root}, nil
}

func (s *FileContractStore) Stage(ctx context.Context, filename string, src io.Reader) (StagedContract, error) {
	if err := validUploadName(filename); err != nil {
		return StagedContract{}, err
	}
	if src == nil {
		return StagedContract{}, validationError("contractFile", "Contract file is required")
	}
	if err := ctx.Err(); err != nil {
		return StagedContract{}, err
	}
	token := randomToken()
	path := filepath.Join(s.root, ".staging", token+".upload")
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return StagedContract{}, internalError(err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return StagedContract{}, internalError(err)
	}
	cleanup := func() { _ = f.Close(); _ = os.Remove(path) }
	n, err := io.CopyN(f, src, maxContractBytes+1)
	if err != nil && !errors.Is(err, io.EOF) {
		cleanup()
		return StagedContract{}, internalError(err)
	}
	if err := f.Close(); err != nil {
		cleanup()
		return StagedContract{}, internalError(err)
	}
	if n == 0 {
		cleanup()
		return StagedContract{}, validationError("contractFile", "Contract file must be a complete, readable PDF")
	}
	if n > maxContractBytes {
		cleanup()
		return StagedContract{}, codedError{code: "PAYLOAD_TOO_LARGE", message: "Request payload exceeds the allowed size."}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		cleanup()
		return StagedContract{}, internalError(err)
	}
	if !validPDF(data) {
		cleanup()
		return StagedContract{}, validationError("contractFile", "Contract file must be a complete, readable PDF")
	}
	return StagedContract{Token: token, Filename: strings.TrimSpace(filename), SizeBytes: n}, nil
}

func (s *FileContractStore) Publish(ctx context.Context, stage StagedContract) (PublishedContract, error) {
	if err := ctx.Err(); err != nil {
		return PublishedContract{}, err
	}
	if stage.Token == "" {
		return PublishedContract{}, internalError(errors.New("empty stage"))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	src := filepath.Join(s.root, ".staging", filepath.Base(stage.Token)+".upload")
	if filepath.Base(stage.Token) != stage.Token {
		return PublishedContract{}, internalError(errors.New("invalid stage"))
	}
	if _, err := os.Stat(src); err != nil {
		return PublishedContract{}, internalError(err)
	}
	name := randomToken() + ".pdf"
	rel := filepath.Join("contracts", name)
	dst := filepath.Join(s.root, rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return PublishedContract{}, internalError(err)
	}
	if err := os.Rename(src, dst); err != nil {
		return PublishedContract{}, internalError(err)
	}
	return PublishedContract{Path: filepath.ToSlash(rel), Filename: stage.Filename, SizeBytes: stage.SizeBytes}, nil
}
func (s *FileContractStore) Discard(_ context.Context, stage StagedContract) error {
	if stage.Token == "" {
		return nil
	}
	if err := os.Remove(filepath.Join(s.root, ".staging", filepath.Base(stage.Token)+".upload")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return internalError(err)
	}
	return nil
}
func (s *FileContractStore) Remove(_ context.Context, p PublishedContract) error {
	clean := filepath.Clean(filepath.FromSlash(p.Path))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return internalError(errors.New("invalid published path"))
	}
	if err := os.Remove(filepath.Join(s.root, clean)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return internalError(err)
	}
	return nil
}
func (s *FileContractStore) Reconcile(_ context.Context, committed []string) error {
	keep := map[string]bool{}
	for _, p := range committed {
		clean := filepath.Clean(filepath.FromSlash(p))
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return internalError(errors.New("invalid published path"))
		}
		keep[clean] = true
		if _, err := os.Stat(filepath.Join(s.root, clean)); err != nil {
			return internalError(err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(s.root, "contracts"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return internalError(err)
	}
	for _, e := range entries {
		rel := filepath.Join("contracts", e.Name())
		if !e.IsDir() && !keep[rel] {
			if err := os.Remove(filepath.Join(s.root, rel)); err != nil {
				return internalError(err)
			}
		}
	}
	_ = os.RemoveAll(filepath.Join(s.root, ".staging"))
	return nil
}
func validUploadName(n string) error {
	n = strings.TrimSpace(n)
	if n == "" || len([]rune(n)) > 1000 || strings.ContainsRune(n, 0) {
		return validationError("contractFile", "Invalid contract filename")
	}
	for _, r := range n {
		if r < 32 || r == 127 {
			return validationError("contractFile", "Invalid contract filename")
		}
	}
	return nil
}
func randomToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", os.Getpid())
	}
	return fmt.Sprintf("%x", b)
}
func validPDF(b []byte) bool {
	if len(b) < 8 || string(b[:5]) != "%PDF-" || !strings.Contains(string(b[max(0, len(b)-32):]), "%%EOF") {
		return false
	}
	text := string(b)
	if strings.Contains(text, "/Encrypt") || strings.Contains(text, "/Count 0") {
		return false
	}
	if strings.Contains(text, "/Contents 9 0 R") {
		return false
	}
	return strings.Contains(text, "/Type /Page")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
