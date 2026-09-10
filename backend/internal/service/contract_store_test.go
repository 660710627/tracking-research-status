package service_test

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/service"
)

// Test fixtures only: classic PDF xref with actual byte offsets, optionally
// standard revision-2 encryption. Empty content streams need no encrypted bytes.
func pdfFixture(pages int, padding int, password *string) []byte {
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [] /Count 0 >>"}
	if pages > 0 {
		objects[1] = "<< /Type /Pages /Kids [3 0 R] /Count 1 >>"
		objects = append(objects, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>", "<< /Length 0 >>\nstream\n\nendstream")
	}
	encryptRef := ""
	id := bytes.Repeat([]byte{0x17}, 16)
	if password != nil {
		pad := []byte{0x28, 0xbf, 0x4e, 0x5e, 0x4e, 0x75, 0x8a, 0x41, 0x64, 0x00, 0x4e, 0x56, 0xff, 0xfa, 0x01, 0x08, 0x2e, 0x2e, 0x00, 0xb6, 0xd0, 0x68, 0x3e, 0x80, 0x2f, 0x0c, 0xa9, 0xfe, 0x64, 0x53, 0x69, 0x7a}
		padded := func(p string) []byte { b := append([]byte(p), pad...); return b[:32] }
		ownerHash := md5.Sum(padded("owner-password"))
		cipher, _ := rc4.NewCipher(ownerHash[:5])
		owner := make([]byte, 32)
		cipher.XORKeyStream(owner, padded(*password))
		permissions := make([]byte, 4)
		binary.LittleEndian.PutUint32(permissions, 0xfffffffc)
		material := append(padded(*password), owner...)
		material = append(material, permissions...)
		material = append(material, id...)
		key := md5.Sum(material)
		cipher, _ = rc4.NewCipher(key[:5])
		user := make([]byte, 32)
		cipher.XORKeyStream(user, pad)
		objects = append(objects, fmt.Sprintf("<< /Filter /Standard /V 1 /R 2 /O <%x> /U <%x> /P -4 >>", owner, user))
		encryptRef = fmt.Sprintf(" /Encrypt %d 0 R", len(objects))
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	if padding > 0 {
		out.WriteByte('%')
		out.Write(bytes.Repeat([]byte{'x'}, padding))
		out.WriteByte('\n')
	}
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R /ID [<%s><%s>]%s >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), hex.EncodeToString(id), hex.EncodeToString(id), encryptRef, xref)
	return out.Bytes()
}

func sizedPDF(size int) []byte {
	// Fixed-point padding accounts for the varying number of startxref digits.
	padding := size - len(pdfFixture(1, 0, nil)) - 2
	for i := 0; i < 10; i++ {
		b := pdfFixture(1, padding, nil)
		delta := size - len(b)
		if delta == 0 {
			return b
		}
		padding += delta
	}
	panic("PDF size fixture did not converge")
}

func storageFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestFileContractValidation(t *testing.T) {
	emptyPassword := ""
	password := "secret"
	valid := pdfFixture(1, 0, nil)
	cases := []struct {
		name string
		data []byte
		code string
	}{
		{"valid", valid, ""}, {"empty", nil, "VALIDATION_ERROR"}, {"renamed_text", []byte("not a PDF"), "VALIDATION_ERROR"},
		{"header_only", []byte("%PDF-1.4\n"), "VALIDATION_ERROR"}, {"truncated", valid[:len(valid)/2], "VALIDATION_ERROR"},
		{"broken_page_reference", bytes.Replace(valid, []byte("/Contents 4 0 R"), []byte("/Contents 9 0 R"), 1), "VALIDATION_ERROR"},
		{"zero_pages", pdfFixture(0, 0, nil), "VALIDATION_ERROR"},
		{"encrypted_empty_password", pdfFixture(1, 0, &emptyPassword), "VALIDATION_ERROR"}, {"password_protected", pdfFixture(1, 0, &password), "VALIDATION_ERROR"},
		{"exact_limit", sizedPDF(20971520), ""}, {"over_limit", sizedPDF(20971521), "PAYLOAD_TOO_LARGE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			store, err := service.NewFileContractStore(root)
			if err != nil {
				t.Fatal(err)
			}
			stage, err := store.Stage(context.Background(), "contract.pdf", bytes.NewReader(tc.data))
			if tc.code != "" {
				field := ""
				if tc.code == "VALIDATION_ERROR" {
					field = "contractFile"
				}
				requireCode(t, err, tc.code, field)
				if got := storageFiles(t, root); len(got) != 0 {
					t.Fatalf("invalid upload left files: %v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if stage.SizeBytes != int64(len(tc.data)) {
				t.Fatalf("size=%d want %d", stage.SizeBytes, len(tc.data))
			}
			if err := store.Discard(context.Background(), stage); err != nil {
				t.Fatal(err)
			}
			if len(storageFiles(t, root)) != 0 {
				t.Fatal("discard left files")
			}
		})
	}
	t.Run("content_not_extension", func(t *testing.T) {
		root := t.TempDir()
		store, err := service.NewFileContractStore(root)
		if err != nil {
			t.Fatal(err)
		}
		stage, err := store.Stage(context.Background(), "contract.txt", bytes.NewReader(valid))
		if err != nil {
			t.Fatal("valid PDF rejected based on extension:", err)
		}
		if err := store.Discard(context.Background(), stage); err != nil {
			t.Fatal(err)
		}
	})
}

func TestFileContractLifecycle(t *testing.T) {
	root := t.TempDir()
	store, err := service.NewFileContractStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(storageFiles(t, root)) != 0 {
		t.Fatal("idle store created files before submission")
	}
	data := pdfFixture(1, 0, nil)
	paths := map[string]bool{}
	for i := 0; i < 2; i++ {
		stage, err := store.Stage(context.Background(), "contract.pdf", bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		published, err := store.Publish(context.Background(), stage)
		if err != nil {
			t.Fatal(err)
		}
		clean := filepath.Clean(published.Path)
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			t.Fatalf("unsafe published path %q", published.Path)
		}
		if paths[published.Path] || filepath.Base(clean) == "contract.pdf" {
			t.Fatal("filename reused or supplied by client")
		}
		paths[published.Path] = true
		actual, err := os.ReadFile(filepath.Join(root, clean))
		if err != nil || !bytes.Equal(actual, data) {
			t.Fatalf("published bytes unusable: %v", err)
		}
	}
	if got := storageFiles(t, root); len(got) != 2 {
		t.Fatalf("staging not cleaned or overwrite: %v", got)
	}
}

func TestFileContractFilename(t *testing.T) {
	for _, name := range []string{"", "\u2003\u00a0", strings.Repeat("ก", 1001), "bad\x00.pdf", "bad\x01.pdf"} {
		t.Run(fmt.Sprintf("invalid_%q", name), func(t *testing.T) {
			root := t.TempDir()
			store, err := service.NewFileContractStore(root)
			if err != nil {
				t.Fatal(err)
			}
			_, err = store.Stage(context.Background(), name, bytes.NewReader(pdfFixture(1, 0, nil)))
			requireCode(t, err, "VALIDATION_ERROR", "contractFile")
			if len(storageFiles(t, root)) != 0 {
				t.Fatal("invalid filename left files")
			}
		})
	}
	t.Run("untrusted_path_not_used", func(t *testing.T) {
		root := t.TempDir()
		store, err := service.NewFileContractStore(root)
		if err != nil {
			t.Fatal(err)
		}
		stage, err := store.Stage(context.Background(), "../outside.pdf", bytes.NewReader(pdfFixture(1, 0, nil)))
		if err != nil {
			requireCode(t, err, "VALIDATION_ERROR", "contractFile")
			if len(storageFiles(t, root)) != 0 {
				t.Fatal("rejected filename left files")
			}
			return
		}
		published, err := store.Publish(context.Background(), stage)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Clean(published.Path)
		if filepath.IsAbs(path) || strings.HasPrefix(path, "..") {
			t.Fatalf("client path escaped root: %q", path)
		}
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatal(err)
		}
	})
}

type brokenReader struct{ sent bool }

func (r *brokenReader) Read(p []byte) (int, error) {
	if !r.sent {
		r.sent = true
		return copy(p, "%PDF-1.4\n"), nil
	}
	return 0, errors.New("secret /private/read failure")
}

func TestFileContractReadFailure(t *testing.T) {
	root := t.TempDir()
	store, err := service.NewFileContractStore(root)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Stage(context.Background(), "contract.pdf", &brokenReader{})
	requireCode(t, err, "INTERNAL_ERROR", "")
	if got := storageFiles(t, root); len(got) != 0 {
		t.Fatalf("partial staging left: %v", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.Stage(ctx, "contract.pdf", bytes.NewReader(pdfFixture(1, 0, nil))); err == nil {
		t.Fatal("canceled upload succeeded")
	}
	if len(storageFiles(t, root)) != 0 {
		t.Fatal("canceled upload left files")
	}
}

func TestFileContractRecovery(t *testing.T) {
	root := t.TempDir()
	store, err := service.NewFileContractStore(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	data := pdfFixture(1, 0, nil)
	stage, err := store.Stage(ctx, "committed.pdf", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	committed, err := store.Publish(ctx, stage)
	if err != nil {
		t.Fatal(err)
	}
	stage, err = store.Stage(ctx, "uncommitted.pdf", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	orphan, err := store.Publish(ctx, stage)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Stage(ctx, "interrupted.pdf", bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	// Simulate a restart after publish-before-commit, plus interrupted staging.
	restarted, err := service.NewFileContractStore(root)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := restarted.Reconcile(ctx, []string{committed.Path}); err != nil {
			t.Fatal(err)
		}
	}
	if got := storageFiles(t, root); len(got) != 1 {
		t.Fatalf("recovery left orphan/staging or lost committed PDF: %v", got)
	}
	if _, err := os.Stat(filepath.Join(root, orphan.Path)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("orphan still exists: %v", err)
	}
	actual, err := os.ReadFile(filepath.Join(root, committed.Path))
	if err != nil || !bytes.Equal(actual, data) {
		t.Fatalf("committed PDF lost: %v", err)
	}
	if err := restarted.Reconcile(ctx, []string{committed.Path, "contracts/missing.pdf"}); err == nil {
		t.Fatal("recovery reported ready with missing referenced PDF")
	}
}

// Keep io.Reader contract explicit without importing HTTP/multipart in tests.
var _ io.Reader = (*brokenReader)(nil)
