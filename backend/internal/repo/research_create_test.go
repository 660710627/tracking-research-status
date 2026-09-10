package repo_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/660710627/my-research/internal/db"
	"github.com/660710627/my-research/internal/repo"
)

func createDB(t *testing.T) (*sql.DB, *repo.ResearchRepository) {
	t.Helper()
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	return database, repo.NewResearchRepository(database)
}

func createInput(suffix string) repo.CreateResearchInput {
	return repo.CreateResearchInput{
		Title: "โครงการ " + suffix, IsSubsidized: true, ProjectType: "RESEARCH", ResearchKind: "BUDGET",
		ResponsibleProjectUnit: "คณะวิทยาศาสตร์", ResponsibleBudgetUnit: "กองคลัง",
		StartDate: "2024-02-29", EndDate: "2025-03-01", BudgetAmount: 12345.67,
		ThaiAbstract: "บทคัดย่อไทย", EnglishAbstract: "English abstract", Objectives: "วัตถุประสงค์", Keywords: "น้ำ, สิ่งแวดล้อม",
		Members: []repo.ResearchMemberInput{
			{FullName: "หัวหน้า", Email: "lead@example.com", Affiliation: "มหาวิทยาลัย", ContributionPercent: 60.25, Role: "LEAD"},
			{FullName: "ผู้ร่วม", Email: "co@example.com", Affiliation: "สถาบัน", ContributionPercent: 39.75, Role: "CO_RESEARCHER"},
		},
		Contract: repo.ResearchContractInput{FundingType: "INTERNAL", FundingSourceName: "กองทุน", ContractNumber: "CN-" + suffix, ContractNumberKey: "cn-" + suffix, StoragePath: "contracts/" + suffix + ".pdf", OriginalFilename: "สัญญา.pdf", ContentType: "application/pdf", SizeBytes: 512},
	}
}

func mustCreate(t *testing.T, r *repo.ResearchRepository, input repo.CreateResearchInput) int64 {
	t.Helper()
	id, err := r.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if id <= 0 {
		t.Fatalf("ID = %d, want positive", id)
	}
	return id
}

func execFixture(t *testing.T, database *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := database.Exec(query, args...); err != nil {
		t.Fatal(err)
	}
}

func counts(t *testing.T, database *sql.DB) [3]int {
	t.Helper()
	var result [3]int
	for i, table := range []string{"researches", "research_members", "research_contracts"} {
		if err := database.QueryRow("SELECT count(*) FROM " + table).Scan(&result[i]); err != nil {
			t.Fatal(err)
		}
	}
	return result
}

func TestCreateCompleteAggregate(t *testing.T) {
	database, r := createDB(t)
	input := createInput("complete")
	id := mustCreate(t, r, input)
	var got repo.CreateResearchInput
	var parent sql.NullInt64
	var status, process string
	err := database.QueryRow(`SELECT title,is_subsidized,project_type,research_kind,continuation_of_id,responsible_project_unit,responsible_budget_unit,start_date,end_date,budget_amount,thai_abstract,english_abstract,objectives,keywords,status,process FROM researches WHERE id=?`, id).Scan(
		&got.Title, &got.IsSubsidized, &got.ProjectType, &got.ResearchKind, &parent, &got.ResponsibleProjectUnit, &got.ResponsibleBudgetUnit, &got.StartDate, &got.EndDate, &got.BudgetAmount, &got.ThaiAbstract, &got.EnglishAbstract, &got.Objectives, &got.Keywords, &status, &process)
	if err != nil {
		t.Fatal(err)
	}
	if parent.Valid || status != "กำลังดำเนินการ" || process != "สัญญาโครงการ" {
		t.Fatalf("invalid defaults: %v %q %q", parent, status, process)
	}
	rows, err := database.Query(`SELECT full_name,email,affiliation,contribution_percent,role FROM research_members WHERE research_id=? ORDER BY role DESC`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var m repo.ResearchMemberInput
		if err := rows.Scan(&m.FullName, &m.Email, &m.Affiliation, &m.ContributionPercent, &m.Role); err != nil {
			t.Fatal(err)
		}
		got.Members = append(got.Members, m)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	c := &got.Contract
	err = database.QueryRow(`SELECT funding_type,funding_source_name,contract_number,contract_number_key,storage_path,original_filename,content_type,size_bytes FROM research_contracts WHERE research_id=?`, id).Scan(&c.FundingType, &c.FundingSourceName, &c.ContractNumber, &c.ContractNumberKey, &c.StoragePath, &c.OriginalFilename, &c.ContentType, &c.SizeBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, input) {
		t.Fatalf("persisted aggregate = %#v; want %#v", got, input)
	}
	if got := counts(t, database); got != [3]int{1, 2, 1} {
		t.Fatal(got)
	}
}

func TestCreateIdentity(t *testing.T) {
	database, r := createDB(t)
	typ := reflect.TypeOf(createInput("shape"))
	for _, field := range []string{"ID", "Status", "Process"} {
		if _, ok := typ.FieldByName(field); ok {
			t.Fatalf("create input exposes %s", field)
		}
	}
	first := mustCreate(t, r, createInput("first"))
	highest := mustCreate(t, r, createInput("highest"))
	// Fixture-only deletion, deliberately not a production DELETE method.
	execFixture(t, database, "DELETE FROM research_contracts WHERE research_id=?", highest)
	execFixture(t, database, "DELETE FROM research_members WHERE research_id=?", highest)
	execFixture(t, database, "DELETE FROM researches WHERE id=?", highest)
	next := mustCreate(t, r, createInput("next"))
	if first == highest || next <= highest {
		t.Fatalf("IDs reused/not increasing: %d %d %d", first, highest, next)
	}
}

func TestCreateContinuation(t *testing.T) {
	for _, state := range []string{"กำลังดำเนินการ", "โครงการเสร็จสิ้น", "ยุติโครงการ"} {
		t.Run(state, func(t *testing.T) {
			database, r := createDB(t)
			root := createInput("root")
			parent := mustCreate(t, r, root)
			execFixture(t, database, "UPDATE researches SET status=? WHERE id=?", state, parent)
			for i, title := range []string{root.Title, root.Title, "ชื่อที่ต่างจากต้นทาง"} {
				in := createInput(fmt.Sprint(i))
				in.Title = title
				in.ResearchKind = "CONTINUATION"
				in.ContinuationOfID = &parent
				id := mustCreate(t, r, in)
				var got int64
				if err := database.QueryRow("SELECT continuation_of_id FROM researches WHERE id=?", id).Scan(&got); err != nil || got != parent {
					t.Fatalf("parent=%d err=%v", got, err)
				}
			}
			if got := counts(t, database); got != [3]int{4, 8, 4} {
				t.Fatal(got)
			}
		})
	}
}

func TestCreateTitleRules(t *testing.T) {
	for _, tc := range []struct {
		name, title string
		conflict    bool
	}{
		{"same", "Root", true}, {"unicode_trim", strings.TrimSpace("\u2003Root\u00a0"), true}, {"case_sensitive", "root", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			database, r := createDB(t)
			in := createInput("root")
			in.Title = "Root"
			mustCreate(t, r, in)
			in = createInput("other")
			in.Title = tc.title
			id, err := r.Create(context.Background(), in)
			if tc.conflict {
				if id != 0 || !errors.Is(err, repo.ErrTitleAlreadyExists) {
					t.Fatalf("id=%d err=%v", id, err)
				}
				if got := counts(t, database); got != [3]int{1, 2, 1} {
					t.Fatal(got)
				}
			} else if err != nil || id <= 0 {
				t.Fatalf("id=%d err=%v", id, err)
			}
		})
	}
	t.Run("root_conflicts_with_child", func(t *testing.T) {
		_, r := createDB(t)
		parent := mustCreate(t, r, createInput("parent"))
		child := createInput("child")
		child.ResearchKind = "CONTINUATION"
		child.ContinuationOfID = &parent
		mustCreate(t, r, child)
		root := createInput("new-contract")
		root.Title = child.Title
		if _, err := r.Create(context.Background(), root); !errors.Is(err, repo.ErrTitleAlreadyExists) {
			t.Fatal(err)
		}
	})
}

func TestCreateContractUniqueness(t *testing.T) {
	database, r := createDB(t)
	in := createInput("one")
	in.Contract.ContractNumber = "Straße"
	in.Contract.ContractNumberKey = "strasse"
	mustCreate(t, r, in)
	in = createInput("two")
	in.Contract.ContractNumber = "STRASSE"
	in.Contract.ContractNumberKey = "strasse"
	if id, err := r.Create(context.Background(), in); id != 0 || !errors.Is(err, repo.ErrContractNumberAlreadyExists) {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if got := counts(t, database); got != [3]int{1, 2, 1} {
		t.Fatal(got)
	}
}

func TestCreateRejectedAggregate(t *testing.T) {
	cases := []struct {
		name   string
		change func(*repo.CreateResearchInput)
		want   error
	}{
		{"missing_parent", func(in *repo.CreateResearchInput) {
			id := int64(999)
			in.ContinuationOfID = &id
			in.ResearchKind = "CONTINUATION"
		}, repo.ErrContinuationNotFound},
		{"continuation_without_parent", func(in *repo.CreateResearchInput) { in.ResearchKind = "CONTINUATION" }, repo.ErrConstraintViolation},
		{"no_members", func(in *repo.CreateResearchInput) { in.Members = nil }, repo.ErrConstraintViolation},
		{"one_member", func(in *repo.CreateResearchInput) { in.Members = in.Members[:1] }, repo.ErrConstraintViolation},
		{"three_members", func(in *repo.CreateResearchInput) { in.Members = append(in.Members, in.Members[0]) }, repo.ErrConstraintViolation},
		{"two_leads", func(in *repo.CreateResearchInput) { in.Members[1].Role = "LEAD" }, repo.ErrConstraintViolation},
		{"two_collaborators", func(in *repo.CreateResearchInput) { in.Members[0].Role = "CO_RESEARCHER" }, repo.ErrConstraintViolation},
		{"duplicate_email", func(in *repo.CreateResearchInput) { in.Members[1].Email = in.Members[0].Email }, repo.ErrConstraintViolation},
		{"duplicate_email_case", func(in *repo.CreateResearchInput) { in.Members[1].Email = "LEAD@EXAMPLE.COM" }, repo.ErrConstraintViolation},
		{"sum_under", func(in *repo.CreateResearchInput) { in.Members[1].ContributionPercent = 39.74 }, repo.ErrConstraintViolation},
		{"sum_over", func(in *repo.CreateResearchInput) { in.Members[1].ContributionPercent = 39.76 }, repo.ErrConstraintViolation},
		{"invalid_second_member", func(in *repo.CreateResearchInput) { in.Members[1].FullName = "" }, repo.ErrConstraintViolation},
		{"missing_contract", func(in *repo.CreateResearchInput) { in.Contract = repo.ResearchContractInput{} }, repo.ErrConstraintViolation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			database, r := createDB(t)
			in := createInput(tc.name)
			tc.change(&in)
			id, err := r.Create(context.Background(), in)
			if id != 0 || !errors.Is(err, tc.want) {
				t.Fatalf("id=%d err=%v want %v", id, err, tc.want)
			}
			if got := counts(t, database); got != [3]int{} {
				t.Fatalf("partial write %v", got)
			}
		})
	}
}

func TestCreateImmutableAndForeignKeys(t *testing.T) {
	for _, q := range []string{
		"UPDATE researches SET id=id+100 WHERE id=?",
		"UPDATE researches SET continuation_of_id=NULL WHERE id=?",
		"UPDATE researches SET continuation_of_id=999999 WHERE id=?",
		"DELETE FROM researches WHERE id=(SELECT continuation_of_id FROM researches WHERE id=?)",
		"UPDATE research_members SET research_id=999999 WHERE research_id=?",
		"UPDATE research_contracts SET research_id=999999 WHERE research_id=?",
	} {
		t.Run(q, func(t *testing.T) {
			database, r := createDB(t)
			parent := mustCreate(t, r, createInput("parent"))
			in := createInput("child")
			in.ResearchKind = "CONTINUATION"
			in.ContinuationOfID = &parent
			child := mustCreate(t, r, in)
			before := counts(t, database)
			if _, err := database.Exec(q, child); err == nil {
				t.Fatal("invariant accepted invalid SQL")
			}
			if got := counts(t, database); got != before {
				t.Fatal(got)
			}
			var got int64
			if err := database.QueryRow("SELECT continuation_of_id FROM researches WHERE id=?", child).Scan(&got); err != nil || got != parent {
				t.Fatalf("changed identity: %d %v", got, err)
			}
		})
	}
}

func TestCreateRollback(t *testing.T) {
	for _, table := range []string{"researches", "research_members", "research_contracts"} {
		t.Run(table, func(t *testing.T) {
			database, r := createDB(t)
			mustCreate(t, r, createInput("existing"))
			before := counts(t, database)
			execFixture(t, database, "CREATE TRIGGER fail_create BEFORE INSERT ON "+table+" BEGIN SELECT RAISE(ABORT, 'injected failure'); END")
			id, err := r.Create(context.Background(), createInput("failure"))
			if id != 0 || !errors.Is(err, repo.ErrInternal) {
				t.Fatalf("id=%d err=%v", id, err)
			}
			if got := counts(t, database); got != before {
				t.Fatalf("rollback %v want %v", got, before)
			}
			execFixture(t, database, "DROP TRIGGER fail_create")
			mustCreate(t, r, createInput("failure"))
			if got := counts(t, database); got != [3]int{2, 4, 2} {
				t.Fatal(got)
			}
		})
	}
}

func TestCreateRollbackAtCommit(t *testing.T) {
	database, r := createDB(t)
	mustCreate(t, r, createInput("existing"))
	// A deferred foreign key fails only at COMMIT, after all aggregate inserts.
	execFixture(t, database, "CREATE TABLE commit_failure (parent_id INTEGER REFERENCES researches(id) DEFERRABLE INITIALLY DEFERRED)")
	execFixture(t, database, "CREATE TRIGGER fail_commit AFTER INSERT ON research_contracts BEGIN INSERT INTO commit_failure(parent_id) VALUES (-1); END")
	id, err := r.Create(context.Background(), createInput("commit"))
	if id != 0 || err == nil {
		t.Fatalf("reported success before commit: id=%d err=%v", id, err)
	}
	if got := counts(t, database); got != [3]int{1, 2, 1} {
		t.Fatalf("commit left partial aggregate %v", got)
	}
	var orphanCount int
	if err := database.QueryRow("SELECT count(*) FROM commit_failure").Scan(&orphanCount); err != nil {
		t.Fatal(err)
	}
	if orphanCount != 0 {
		t.Fatal("failed commit left orphan rows")
	}
	execFixture(t, database, "DROP TRIGGER fail_commit")
	mustCreate(t, r, createInput("commit"))
}

func TestCreateUnavailable(t *testing.T) {
	t.Run("canceled", func(t *testing.T) {
		database, r := createDB(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		id, err := r.Create(ctx, createInput("cancel"))
		if id != 0 || err == nil {
			t.Fatalf("id=%d err=%v", id, err)
		}
		if got := counts(t, database); got != [3]int{} {
			t.Fatal(got)
		}
	})
	t.Run("closed", func(t *testing.T) {
		database, r := createDB(t)
		if err := database.Close(); err != nil {
			t.Fatal(err)
		}
		id, err := r.Create(context.Background(), createInput("closed"))
		if id != 0 || !errors.Is(err, repo.ErrInternal) {
			t.Fatalf("id=%d err=%v", id, err)
		}
	})
}

func TestCreateConcurrent(t *testing.T) {
	for _, mode := range []string{"distinct", "same_title", "same_contract", "children"} {
		t.Run(mode, func(t *testing.T) {
			database, r := createDB(t)
			var parent int64
			if mode == "children" {
				parent = mustCreate(t, r, createInput("parent"))
			}
			const n = 8
			type result struct {
				id  int64
				err error
			}
			results := make(chan result, n)
			start := make(chan struct{})
			var wg sync.WaitGroup
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					in := createInput(fmt.Sprint(i))
					switch mode {
					case "same_title":
						in.Title = "same"
					case "same_contract":
						in.Contract.ContractNumber = "SAME"
						in.Contract.ContractNumberKey = "same"
					case "children":
						in.Title = "same"
						in.ResearchKind = "CONTINUATION"
						in.ContinuationOfID = &parent
					}
					<-start
					id, err := r.Create(ctx, in)
					results <- result{id, err}
				}(i)
			}
			close(start)
			wg.Wait()
			close(results)
			ids := map[int64]bool{}
			conflicts := 0
			for got := range results {
				if got.err == nil {
					if got.id <= 0 || ids[got.id] {
						t.Fatalf("invalid/duplicate ID %d", got.id)
					}
					ids[got.id] = true
				} else {
					want := repo.ErrTitleAlreadyExists
					if mode == "same_contract" {
						want = repo.ErrContractNumberAlreadyExists
					}
					if (mode != "same_title" && mode != "same_contract") || !errors.Is(got.err, want) || got.id != 0 {
						t.Fatalf("unexpected concurrent result %+v", got)
					}
					conflicts++
				}
			}
			want := n
			if mode == "same_title" || mode == "same_contract" {
				want = 1
			}
			if len(ids) != want || conflicts != n-want {
				t.Fatalf("success=%d conflicts=%d", len(ids), conflicts)
			}
			total := want
			if mode == "children" {
				total++
			}
			if got := counts(t, database); got != [3]int{total, total * 2, total} {
				t.Fatal(got)
			}
		})
	}
}
