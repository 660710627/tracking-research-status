package repo_test

import (
	"fmt"
	"strings"
	"testing"
)

func TestCreateSQLConstraints(t *testing.T) {
	type constraintCase struct {
		name, table, column string
		value               any
	}
	var cases []constraintCase
	textFields := map[string][]string{
		"researches":         {"title", "responsible_project_unit", "responsible_budget_unit", "thai_abstract", "english_abstract", "objectives", "keywords"},
		"research_members":   {"full_name", "affiliation"},
		"research_contracts": {"funding_source_name", "contract_number", "original_filename"},
	}
	for table, columns := range textFields {
		for _, column := range columns {
			for name, value := range map[string]any{"null": nil, "empty": "", "whitespace": "\u2003\u00a0", "too_long": strings.Repeat("ก", 1001), "nul": "a\x00b", "control": "a\x01b", "del": "a\x7fb"} {
				cases = append(cases, constraintCase{column + "/" + name, table, column, value})
			}
		}
	}
	for name, value := range map[string]any{"slash": "a/b", "newline": "a\nb", "tab": "a\tb"} {
		cases = append(cases, constraintCase{"title/" + name, "researches", "title", value})
	}
	fields := []struct {
		table, column string
		values        []any
	}{
		{"researches", "is_subsidized", []any{nil, 2, -1, "yes"}},
		{"researches", "project_type", []any{nil, "", "OTHER"}},
		{"researches", "research_kind", []any{nil, "OTHER", "CONTINUATION"}},
		{"researches", "status", []any{nil, "", "UNKNOWN"}},
		{"researches", "process", []any{nil, "", "UNKNOWN"}},
		{"researches", "budget_amount", []any{nil, 0, -1, 1.001, "money"}},
		{"researches", "start_date", []any{nil, "", "2023-02-29", "2024-04-31", "2024-13-01", "2024-00-01", "29/02/2567", "2025-03-02"}},
		{"researches", "end_date", []any{nil, "", "2025-02-29", "2025-02-28", "2024-03-01", "2024-02-28", "01/03/2568"}},
		{"research_members", "email", []any{nil, "", strings.Repeat("a", 243) + "@example.com", "no-at-sign", "a@", "@example.com", "a b@example.com", "co@example.com", "CO@EXAMPLE.COM"}},
		{"research_members", "role", []any{nil, "", "UNKNOWN"}},
		{"research_members", "contribution_percent", []any{nil, 0, -1, 100.01, 60.251, "percent"}},
		{"research_contracts", "funding_type", []any{nil, "", "OTHER"}},
		{"research_contracts", "contract_number_key", []any{nil}},
		{"research_contracts", "storage_path", []any{nil, ""}},
		{"research_contracts", "content_type", []any{nil, "", "text/plain"}},
		{"research_contracts", "size_bytes", []any{nil, 0, -1, 20971521, 1.5}},
	}
	for _, field := range fields {
		for i, value := range field.values {
			cases = append(cases, constraintCase{fmt.Sprintf("%s/%d", field.column, i), field.table, field.column, value})
		}
	}
	for _, tc := range cases {
		t.Run(tc.table+"/"+tc.name, func(t *testing.T) {
			database, r := createDB(t)
			id := mustCreate(t, r, createInput("constraint"))
			key := "research_id"
			if tc.table == "researches" {
				key = "id"
			}
			where := " WHERE " + key + "=?"
			if tc.table == "research_members" {
				where += " AND role='LEAD'"
			}
			var before any
			if err := database.QueryRow("SELECT "+tc.column+" FROM "+tc.table+where, id).Scan(&before); err != nil {
				t.Fatal(err)
			}
			if _, err := database.Exec("UPDATE "+tc.table+" SET "+tc.column+"=?"+where, tc.value, id); err == nil {
				t.Fatalf("SQL accepted invalid %s=%v", tc.column, tc.value)
			}
			var after any
			if err := database.QueryRow("SELECT "+tc.column+" FROM "+tc.table+where, id).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(before) != fmt.Sprint(after) {
				t.Fatalf("failed statement changed value: %v -> %v", before, after)
			}
		})
	}
	t.Run("one_contract_per_project", func(t *testing.T) {
		database, r := createDB(t)
		id := mustCreate(t, r, createInput("one"))
		_, err := database.Exec(`INSERT INTO research_contracts (research_id,funding_type,funding_source_name,contract_number,contract_number_key,storage_path,original_filename,content_type,size_bytes)
   VALUES (?,'EXTERNAL','fund','different','different','contracts/new.pdf','new.pdf','application/pdf',100)`, id)
		if err == nil {
			t.Fatal("accepted second contract")
		}
		if got := counts(t, database); got != [3]int{1, 2, 1} {
			t.Fatal(got)
		}
	})
	t.Run("unique_contract_key", func(t *testing.T) {
		database, r := createDB(t)
		first := mustCreate(t, r, createInput("first"))
		second := mustCreate(t, r, createInput("second"))
		if _, err := database.Exec("UPDATE research_contracts SET contract_number_key=(SELECT contract_number_key FROM research_contracts WHERE research_id=?) WHERE research_id=?", first, second); err == nil {
			t.Fatal("accepted duplicate normalized contract key")
		}
	})
	t.Run("root_cannot_reference_parent", func(t *testing.T) {
		database, r := createDB(t)
		parent := mustCreate(t, r, createInput("parent"))
		in := createInput("invalid-root")
		in.ContinuationOfID = &parent
		if _, err := r.Create(t.Context(), in); err == nil {
			t.Fatal("accepted root with parent")
		}
		if got := counts(t, database); got != [3]int{1, 2, 1} {
			t.Fatal(got)
		}
	})
}

func TestCreateSQLValidBoundaries(t *testing.T) {
	for _, size := range []int{1, 1000} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			database, r := createDB(t)
			in := createInput(fmt.Sprint(size))
			text := strings.Repeat("ก", size)
			in.Title = text
			in.ResponsibleProjectUnit = text
			in.ResponsibleBudgetUnit = text
			in.ThaiAbstract = text
			in.EnglishAbstract = text
			in.Objectives = text
			in.Keywords = text
			for i := range in.Members {
				in.Members[i].FullName = text
				in.Members[i].Affiliation = text
			}
			in.Contract.FundingSourceName = text
			in.Contract.ContractNumber = text
			in.Contract.ContractNumberKey = text
			in.Contract.OriginalFilename = text
			in.IsSubsidized = false
			in.ProjectType = "ACADEMIC_SERVICE"
			in.Contract.FundingType = "EXTERNAL"
			in.BudgetAmount = 0.01
			in.Contract.SizeBytes = 20971520
			in.Members[0].ContributionPercent = 0.01
			in.Members[1].ContributionPercent = 99.99
			mustCreate(t, r, in)
			if got := counts(t, database); got != [3]int{1, 2, 1} {
				t.Fatal(got)
			}
		})
	}
	t.Run("normal_calendar_anniversary", func(t *testing.T) {
		_, r := createDB(t)
		in := createInput("anniversary")
		in.StartDate = "2025-06-30"
		in.EndDate = "2026-06-30"
		mustCreate(t, r, in)
	})
	t.Run("general_text_allows_newline_tab", func(t *testing.T) {
		_, r := createDB(t)
		in := createInput("text")
		in.ThaiAbstract = "line1\nline2\tend"
		mustCreate(t, r, in)
	})
	t.Run("email_254", func(t *testing.T) {
		_, r := createDB(t)
		in := createInput("email")
		in.Members[0].Email = strings.Repeat("a", 64) + "@" + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 57) + ".com"
		if len(in.Members[0].Email) != 254 {
			t.Fatal("invalid boundary fixture")
		}
		mustCreate(t, r, in)
	})
}
