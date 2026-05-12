package filter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alive-worker/logsift/internal/parser"
)

// Expr is a single --where clause: field<op>value.
type Expr struct {
	Field string
	Op    string
	Value string
}

var opOrder = []string{"==", "!=", ">=", "<=", ">", "<"}

func ParseExpr(s string) (*Expr, error) {
	for _, op := range opOrder {
		if i := strings.Index(s, op); i > 0 {
			field := strings.TrimSpace(s[:i])
			val := strings.TrimSpace(s[i+len(op):])
			if field == "" || val == "" {
				return nil, fmt.Errorf("malformed expression %q", s)
			}
			return &Expr{Field: field, Op: op, Value: val}, nil
		}
	}
	return nil, fmt.Errorf("no operator in expression %q", s)
}

// Keep evaluates the expression against an entry. String comparisons are
// case-sensitive; numeric comparisons activate when both sides parse as
// float64.
func (x *Expr) Keep(e *parser.Entry) bool {
	got := lookup(e, x.Field)
	if got == nil {
		return false
	}
	gotStr := fmt.Sprint(got)
	if gotF, err1 := toFloat(got); err1 == nil {
		if wantF, err2 := strconv.ParseFloat(x.Value, 64); err2 == nil {
			return numericCompare(gotF, x.Op, wantF)
		}
	}
	return stringCompare(gotStr, x.Op, x.Value)
}

func lookup(e *parser.Entry, field string) any {
	switch field {
	case "level":
		return e.Level
	case "service":
		return e.Service
	case "msg", "message":
		return e.Message
	}
	if v, ok := e.Extra[field]; ok {
		return v
	}
	return nil
}

func toFloat(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case int:
		return float64(x), nil
	case string:
		return strconv.ParseFloat(x, 64)
	}
	return 0, fmt.Errorf("not numeric")
}

func numericCompare(a float64, op string, b float64) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	}
	return false
}

func stringCompare(a, op, b string) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	}
	return false
}
