package db

import (
	"database/sql/driver"
	"math"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"modernc.org/sqlite"
)

// Register before Open so every pooled connection can enforce the same CHECKs.
// These SQL functions cover rules SQLite's ASCII-only built-ins cannot express.
func init() {
	registerPredicate("research_text_valid", 2, func(a []driver.Value) bool {
		s, ok := a[0].(string)
		if !ok || !utf8.ValidString(s) {
			return false
		}
		n := utf8.RuneCountInString(strings.TrimSpace(s))
		if n < 1 || n > 1000 {
			return false
		}
		title := a[1] == int64(1)
		for _, r := range s {
			if title && r == '/' || unicode.IsControl(r) && (title || r != '\n' && r != '\t') {
				return false
			}
		}
		return true
	})
	registerPredicate("research_email_valid", 1, func(a []driver.Value) bool {
		s, ok := a[0].(string)
		if !ok || !utf8.ValidString(s) {
			return false
		}
		s = strings.TrimSpace(s)
		if n := utf8.RuneCountInString(s); n < 1 || n > 254 {
			return false
		}
		address, err := mail.ParseAddress(s)
		return err == nil && address.Address == s
	})
	registerPredicate("research_decimal_valid", 1, func(a []driver.Value) bool {
		var s string
		switch v := a[0].(type) {
		case int64:
			return v > 0
		case float64:
			if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
				return false
			}
			// Use the round-trip decimal representation, not an epsilon that
			// would silently accept values with more than two decimal places.
			s = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			return false
		}
		dot := strings.IndexByte(s, '.')
		return dot < 0 || len(s)-dot-1 <= 2
	})
	registerPredicate("research_dates_valid", 2, func(a []driver.Value) bool {
		start, ok1 := a[0].(string)
		end, ok2 := a[1].(string)
		if !ok1 || !ok2 {
			return false
		}
		s, err1 := time.Parse(time.DateOnly, start)
		e, err2 := time.Parse(time.DateOnly, end)
		// AddDate normalizes February 29 to March 1 in a non-leap year.
		return err1 == nil && err2 == nil && s.Format(time.DateOnly) == start &&
			e.Format(time.DateOnly) == end && !e.Before(s.AddDate(1, 0, 0))
	})
	sqlite.MustRegisterDeterministicScalarFunction("research_trim", 1, func(_ *sqlite.FunctionContext, a []driver.Value) (driver.Value, error) {
		s, _ := a[0].(string)
		return strings.TrimSpace(s), nil
	})
	sqlite.MustRegisterDeterministicScalarFunction("research_email_key", 1, func(_ *sqlite.FunctionContext, a []driver.Value) (driver.Value, error) {
		s, _ := a[0].(string)
		// Canonicalize each Unicode simple-fold equivalence class, including
		// non-ASCII addresses, without changing the stored display value.
		return strings.Map(func(r rune) rune {
			key := r
			for next := unicode.SimpleFold(r); next != r; next = unicode.SimpleFold(next) {
				if next < key {
					key = next
				}
			}
			return key
		}, strings.TrimSpace(s)), nil
	})
}

func registerPredicate(name string, n int32, valid func([]driver.Value) bool) {
	sqlite.MustRegisterDeterministicScalarFunction(name, n, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		if valid(args) {
			return int64(1), nil
		}
		return int64(0), nil
	})
}
