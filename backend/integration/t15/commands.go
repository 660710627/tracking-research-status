package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/660710627/my-research/internal/db"
)

// Fixture controls only. Refuse paths outside the runner's temporary directory.
func fixtureCommand(args []string) error {
	if args[0] == "pdfs" {
		root, err := os.MkdirTemp("", "research-t15-pdfs-")
		if err != nil {
			return err
		}
		emptyPassword, password := "", "secret"
		files := map[string][]byte{
			"valid.pdf": pdfFixture(1, 0, nil), "empty.pdf": {}, "fake.pdf": []byte("not a PDF"),
			"malformed.pdf": []byte("%PDF-1.4\n1 0 obj\n"), "zero-pages.pdf": pdfFixture(0, 0, nil),
			"encrypted.pdf": pdfFixture(1, 0, &emptyPassword), "password.pdf": pdfFixture(1, 0, &password),
			"20MiB.pdf": sizedPDF(20 * 1024 * 1024), "20MiB-plus-one.pdf": sizedPDF(20*1024*1024 + 1),
			"21MiB-plus-one.pdf": sizedPDF(21*1024*1024 + 1),
		}
		for name, data := range files {
			if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
				return err
			}
		}
		fmt.Println(root)
		return nil
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: audit|completed|terminated|remove-parent|fail-db|restore-db TEMP_ROOT")
	}
	root, err := filepath.Abs(args[1])
	if err != nil {
		return err
	}
	temp, err := filepath.Abs(os.TempDir())
	if err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Dir(root), temp) || !strings.HasPrefix(filepath.Base(root), "research-t15-") {
		return fmt.Errorf("not an isolated T15 fixture directory")
	}
	if _, err := os.Stat(filepath.Join(root, "test.db")); err != nil {
		return err
	}
	database, err := db.Open(filepath.Join(root, "test.db"))
	if err != nil {
		return err
	}
	defer database.Close()
	statements := map[string]string{
		"completed":     "UPDATE researches SET status='โครงการเสร็จสิ้น' WHERE id=1",
		"terminated":    "UPDATE researches SET status='ยุติโครงการ' WHERE id=2",
		"remove-parent": "DELETE FROM research_contracts WHERE research_id=(SELECT id FROM researches WHERE title='Disposable parent'); DELETE FROM research_members WHERE research_id=(SELECT id FROM researches WHERE title='Disposable parent'); DELETE FROM researches WHERE title='Disposable parent'",
		"fail-db":       "CREATE TRIGGER t15_fail BEFORE INSERT ON researches BEGIN SELECT RAISE(ABORT, 'T15 fixture persistence failure'); END",
		"restore-db":    "DROP TRIGGER IF EXISTS t15_fail",
	}
	if statement, ok := statements[args[0]]; ok {
		_, err := database.ExecContext(context.Background(), statement)
		return err
	}
	if args[0] != "audit" {
		return fmt.Errorf("unknown fixture action")
	}
	counts := map[string]int{}
	for _, table := range []string{"researches", "research_members", "research_contracts"} {
		var count int
		if err := database.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			return err
		}
		counts[table] = count
	}
	storedPaths := []string{}
	rows, err := database.Query("SELECT storage_path FROM research_contracts ORDER BY research_id")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return err
		}
		storedPaths = append(storedPaths, path)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	files := map[string]string{}
	err = filepath.WalkDir(filepath.Join(root, "files"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		files[relative] = fmt.Sprintf("%d bytes sha256=%x", len(data), sha256.Sum256(data))
		return nil
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"counts": counts, "storedPaths": storedPaths, "files": files})
}
