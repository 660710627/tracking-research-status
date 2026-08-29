package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"modernc.org/sqlite"
)

var registerOnce sync.Once
var registerErr error

func Open(path string) (*sql.DB, error) { return sql.Open("sqlite", path) }

func Initialize(ctx context.Context, database *sql.DB) error {
	registerOnce.Do(func() { registerErr = registerValidationFunctions() })
	if registerErr != nil {
		return fmt.Errorf("register SQLite validation functions: %w", registerErr)
	}
	if err := replaceLegacySchema(ctx, database); err != nil {
		return err
	}
	_, err := database.ExecContext(ctx, researchSchema)
	if err != nil {
		return fmt.Errorf("initialize research schema: %w", err)
	}
	return nil
}

func replaceLegacySchema(ctx context.Context, database *sql.DB) error {
	rows, err := database.QueryContext(ctx, "PRAGMA table_info(researches)")
	if err != nil {
		return fmt.Errorf("inspect research schema: %w", err)
	}
	defer rows.Close()
	legacy := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == "description" {
			legacy = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !legacy {
		return nil
	}
	if _, err := database.ExecContext(ctx, `DROP TABLE IF EXISTS research_members; DROP TABLE IF EXISTS researches;`); err != nil {
		return fmt.Errorf("remove legacy research data: %w", err)
	}
	return nil
}

func registerValidationFunctions() error {
	if err := sqlite.RegisterDeterministicScalarFunction("research_title_valid", 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		value, ok := sqliteText(args)
		return boolInteger(ok && validText(value, true)), nil
	}); err != nil {
		return err
	}
	if err := sqlite.RegisterDeterministicScalarFunction("research_text_valid", 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		value, ok := sqliteText(args)
		return boolInteger(ok && validText(value, false)), nil
	}); err != nil {
		return err
	}
	return sqlite.RegisterDeterministicScalarFunction("research_date_valid", 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		value, ok := sqliteText(args)
		return boolInteger(ok && validBuddhistDate(value)), nil
	})
}

func sqliteText(args []driver.Value) (string, bool) {
	if len(args) != 1 {
		return "", false
	}
	value, ok := args[0].(string)
	return value, ok
}
func boolInteger(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
func validText(value string, title bool) bool {
	if !utf8.ValidString(value) || value != strings.TrimFunc(value, unicode.IsSpace) || utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > 1000 || title && strings.ContainsRune(value, '/') {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
func validBuddhistDate(value string) bool {
	if len(value) != 10 || value[2] != '/' || value[5] != '/' {
		return false
	}
	var day, month, year int
	if _, err := fmt.Sscanf(value, "%02d/%02d/%04d", &day, &month, &year); err != nil || day < 1 || day > 31 || month < 1 || month > 12 || year < 1 {
		return false
	}
	return true
}

const researchSchema = `
PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS researches (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 title TEXT NOT NULL CHECK (research_title_valid(title) = 1),
 continuation_of_id INTEGER REFERENCES researches(id),
 is_subsidized INTEGER NOT NULL CHECK (is_subsidized IN (0,1)),
 funding_type TEXT NOT NULL CHECK (funding_type IN ('INTERNAL','EXTERNAL')),
 funding_source_name TEXT NOT NULL CHECK (research_text_valid(funding_source_name) = 1),
 contract_number TEXT NOT NULL CHECK (research_text_valid(contract_number) = 1),
 contract_filename TEXT NOT NULL CHECK (research_text_valid(contract_filename) = 1),
 contract_content_type TEXT NOT NULL CHECK (contract_content_type = 'application/pdf'),
 contract_size_bytes INTEGER NOT NULL CHECK (contract_size_bytes BETWEEN 1 AND 20971520),
 contract_path TEXT NOT NULL CHECK (research_text_valid(contract_path) = 1),
 project_type TEXT NOT NULL CHECK (project_type IN ('RESEARCH','ACADEMIC_SERVICE')),
 research_kind TEXT NOT NULL CHECK (research_kind IN ('BUDGET','CONTINUATION')),
 responsible_project_unit TEXT NOT NULL CHECK (research_text_valid(responsible_project_unit) = 1),
 responsible_budget_unit TEXT NOT NULL CHECK (research_text_valid(responsible_budget_unit) = 1),
 start_date TEXT NOT NULL CHECK (research_date_valid(start_date) = 1),
 end_date TEXT NOT NULL CHECK (research_date_valid(end_date) = 1),
 budget_amount REAL NOT NULL CHECK (budget_amount > 0 AND abs(budget_amount * 100 - round(budget_amount * 100)) < 0.000001),
 thai_abstract TEXT NOT NULL CHECK (research_text_valid(thai_abstract) = 1),
 english_abstract TEXT NOT NULL CHECK (research_text_valid(english_abstract) = 1),
 objectives TEXT NOT NULL CHECK (research_text_valid(objectives) = 1),
 keywords TEXT NOT NULL CHECK (research_text_valid(keywords) = 1),
 status TEXT NOT NULL DEFAULT 'กำลังดำเนินการ', process TEXT NOT NULL DEFAULT 'สัญญาโครงการ',
 CHECK ((research_kind = 'BUDGET' AND continuation_of_id IS NULL) OR (research_kind = 'CONTINUATION' AND continuation_of_id IS NOT NULL))
);
CREATE TABLE IF NOT EXISTS research_members (
 research_id INTEGER NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
 full_name TEXT NOT NULL CHECK (research_text_valid(full_name) = 1), email TEXT NOT NULL CHECK (research_text_valid(email) = 1), affiliation TEXT NOT NULL CHECK (research_text_valid(affiliation) = 1),
 contribution_percent REAL NOT NULL CHECK (contribution_percent > 0 AND contribution_percent <= 100 AND abs(contribution_percent * 100 - round(contribution_percent * 100)) < 0.000001), role TEXT NOT NULL CHECK (role IN ('LEAD','CO_RESEARCHER'))
);
CREATE TRIGGER IF NOT EXISTS researches_root_title_unique BEFORE INSERT ON researches WHEN NEW.continuation_of_id IS NULL AND EXISTS (SELECT 1 FROM researches WHERE title = NEW.title) BEGIN SELECT RAISE(ABORT, 'TITLE_ALREADY_EXISTS'); END;
CREATE TRIGGER IF NOT EXISTS researches_continuation_exists BEFORE INSERT ON researches WHEN NEW.continuation_of_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM researches WHERE id = NEW.continuation_of_id) BEGIN SELECT RAISE(ABORT, 'CONTINUATION_NOT_FOUND'); END;
CREATE TRIGGER IF NOT EXISTS researches_id_immutable BEFORE UPDATE OF id ON researches WHEN NEW.id <> OLD.id BEGIN SELECT RAISE(ABORT, 'IMMUTABLE_ID'); END;
CREATE TRIGGER IF NOT EXISTS researches_continuation_immutable BEFORE UPDATE OF continuation_of_id ON researches WHEN NEW.continuation_of_id IS NOT OLD.continuation_of_id BEGIN SELECT RAISE(ABORT, 'IMMUTABLE_CONTINUATION'); END;
CREATE TRIGGER IF NOT EXISTS researches_parent_delete_restricted BEFORE DELETE ON researches WHEN EXISTS (SELECT 1 FROM researches WHERE continuation_of_id = OLD.id) BEGIN SELECT RAISE(ABORT, 'RESEARCH_HAS_CONTINUATIONS'); END;
`
