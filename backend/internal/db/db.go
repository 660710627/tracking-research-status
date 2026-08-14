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

var (
	registerFunctionsOnce sync.Once
	registerFunctionsErr  error
)

func Open(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path)
}

func Initialize(ctx context.Context, database *sql.DB) error {
	registerFunctionsOnce.Do(func() {
		registerFunctionsErr = registerValidationFunctions()
	})
	if registerFunctionsErr != nil {
		return fmt.Errorf("register SQLite validation functions: %w", registerFunctionsErr)
	}

	if _, err := database.ExecContext(ctx, researchSchema); err != nil {
		return fmt.Errorf("initialize research schema: %w", err)
	}
	return nil
}

func registerValidationFunctions() error {
	if err := sqlite.RegisterDeterministicScalarFunction("research_title_valid", 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		value, ok := sqliteText(args)
		if !ok {
			return int64(0), nil
		}
		return boolInteger(validTitle(value)), nil
	}); err != nil {
		return err
	}
	return sqlite.RegisterDeterministicScalarFunction("research_description_valid", 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		value, ok := sqliteText(args)
		if !ok {
			return int64(0), nil
		}
		return boolInteger(validDescription(value)), nil
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

func validTitle(value string) bool {
	if !utf8.ValidString(value) || value != strings.TrimFunc(value, unicode.IsSpace) {
		return false
	}
	length := utf8.RuneCountInString(value)
	if length < 1 || length > 200 || strings.ContainsRune(value, '/') {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validDescription(value string) bool {
	if !utf8.ValidString(value) || value != strings.TrimFunc(value, unicode.IsSpace) {
		return false
	}
	length := utf8.RuneCountInString(value)
	if length < 1 || length > 5000 {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) && character != '\n' && character != '\t' {
			return false
		}
	}
	return true
}

const researchSchema = `
CREATE TABLE IF NOT EXISTS researches (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  title              TEXT NOT NULL CHECK (research_title_valid(title) = 1),
  description        TEXT NOT NULL CHECK (research_description_valid(description) = 1),
  continuation_of_id INTEGER REFERENCES researches(id),
  status             TEXT NOT NULL DEFAULT 'กำลังดำเนินการ'
                     CHECK (status IN (
                       'กำลังดำเนินการ',
                       'กำลังดำเนินการ (ขยายเวลาครั้งที่ 1)',
                       'กำลังดำเนินการ (ขยายเวลาครั้งที่ 2)',
                       'กำลังดำเนินการ (ขยายเวลามากกว่า 2 ครั้ง)',
                       'โครงการเสร็จสิ้น',
                       'ยุติโครงการ'
                     )),
  process            TEXT NOT NULL DEFAULT 'สัญญาโครงการ'
                     CHECK (process IN (
                       'สัญญาโครงการ',
                       'บันทึกข้อตกลง',
                       'เปิดบัญชีธนาคาร',
                       'การเบิกจ่ายเงิน',
                       'การจัดสรรค่าธรรมเนียม',
                       'การติดตามส่งรายงาน',
                       'รายงานสรุปการใช้เงิน',
                       'การปิดบัญชีธนาคาร'
                     ))
);

CREATE TRIGGER IF NOT EXISTS researches_continuation_exists
BEFORE INSERT ON researches
WHEN NEW.continuation_of_id IS NOT NULL
 AND NOT EXISTS (SELECT 1 FROM researches WHERE id = NEW.continuation_of_id)
BEGIN
  SELECT RAISE(ABORT, 'CONTINUATION_NOT_FOUND');
END;

CREATE TRIGGER IF NOT EXISTS researches_root_title_unique
BEFORE INSERT ON researches
WHEN NEW.continuation_of_id IS NULL
 AND EXISTS (SELECT 1 FROM researches WHERE title = NEW.title)
BEGIN
  SELECT RAISE(ABORT, 'TITLE_ALREADY_EXISTS');
END;

CREATE TRIGGER IF NOT EXISTS researches_root_title_unique_on_update
BEFORE UPDATE OF title ON researches
WHEN OLD.continuation_of_id IS NULL
 AND NEW.title <> OLD.title
 AND EXISTS (
   SELECT 1 FROM researches
   WHERE id <> OLD.id AND title = NEW.title
 )
BEGIN
  SELECT RAISE(ABORT, 'TITLE_ALREADY_EXISTS');
END;

CREATE TRIGGER IF NOT EXISTS researches_id_immutable
BEFORE UPDATE OF id ON researches
WHEN NEW.id <> OLD.id
BEGIN
  SELECT RAISE(ABORT, 'IMMUTABLE_ID');
END;

CREATE TRIGGER IF NOT EXISTS researches_continuation_immutable
BEFORE UPDATE OF continuation_of_id ON researches
WHEN NEW.continuation_of_id IS NOT OLD.continuation_of_id
BEGIN
  SELECT RAISE(ABORT, 'IMMUTABLE_CONTINUATION');
END;

CREATE TRIGGER IF NOT EXISTS researches_parent_delete_restricted
BEFORE DELETE ON researches
WHEN EXISTS (
  SELECT 1 FROM researches WHERE continuation_of_id = OLD.id
)
BEGIN
  SELECT RAISE(ABORT, 'RESEARCH_HAS_CONTINUATIONS');
END;
`
