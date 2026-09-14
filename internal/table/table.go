// Package table provides small, declarative operations over tabular data.
package table

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/pkyanam/agentcalc/internal/ops"
)

type Sort struct {
	Column string `json:"column"`
	Desc   bool   `json:"desc"`
}

type Query struct {
	Op          string         `json:"op"`
	Column      string         `json:"column"`
	GroupBy     string         `json:"group_by"`
	Where       map[string]any `json:"where"`
	Sort        []Sort         `json:"sort"`
	Limit       *int           `json:"limit"`
	Numerator   string         `json:"numerator"`
	Denominator string         `json:"denominator"`
	Fields      []string       `json:"fields"`
}

type row map[string]any

// Evaluate parses CSV (with a header) or a JSON array of objects and evaluates
// each named query. Filtering, sorting, and limiting happen before an op.
func Evaluate(input []byte, queries map[string]Query) (map[string]any, error) {
	if len(queries) == 0 {
		return nil, errors.New("table: at least one query is required")
	}
	if len(queries) > 128 {
		return nil, errors.New("table: too many queries (maximum 128)")
	}
	if len(strings.TrimSpace(string(input))) == 0 {
		return nil, errors.New("table: empty input")
	}
	rows, err := parse(input)
	if err != nil {
		return nil, err
	}
	result := make(map[string]any, len(queries))
	for name, q := range queries {
		v, err := evaluate(rows, q)
		if err != nil {
			return nil, fmt.Errorf("query %q: %w", name, err)
		}
		result[name] = v
	}
	return result, nil
}

func parse(input []byte) ([]row, error) {
	s := strings.TrimSpace(string(input))
	if strings.HasPrefix(s, "[") {
		return parseJSON(input)
	}
	return parseCSV(strings.NewReader(string(input)))
}

func parseJSON(input []byte) ([]row, error) {
	var values []map[string]any
	dec := json.NewDecoder(strings.NewReader(string(input)))
	dec.UseNumber()
	if err := dec.Decode(&values); err != nil {
		return nil, fmt.Errorf("table: invalid JSON: %w", err)
	}
	if values == nil {
		return nil, errors.New("table: JSON root must be an array")
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, errors.New("table: JSON must contain one array")
	}
	rows := make([]row, len(values))
	for i, v := range values {
		rows[i] = row(v)
	}
	return rows, nil
}

func parseCSV(r io.Reader) ([]row, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	head, err := cr.Read()
	if err == io.EOF {
		return nil, errors.New("table: CSV has no header")
	}
	if err != nil {
		return nil, fmt.Errorf("table: invalid CSV header: %w", err)
	}
	if len(head) == 0 {
		return nil, errors.New("table: CSV has empty header")
	}
	seen := make(map[string]bool, len(head))
	for i, h := range head {
		if h == "" {
			return nil, errors.New("table: CSV header contains empty column")
		}
		if seen[h] {
			return nil, fmt.Errorf("table: duplicate CSV column %q", h)
		}
		seen[h] = true
		_ = i
	}
	var records [][]string
	for {
		rec, e := cr.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, fmt.Errorf("table: invalid CSV: %w", e)
		}
		if len(rec) != len(head) {
			return nil, fmt.Errorf("table: CSV row has %d fields, want %d", len(rec), len(head))
		}
		records = append(records, rec)
	}
	numeric := make([]bool, len(head))
	for c := range head {
		numeric[c] = len(records) > 0
		for _, rec := range records {
			if rec[c] == "" {
				continue
			}
			f, e := strconv.ParseFloat(strings.TrimSpace(rec[c]), 64)
			if e != nil || math.IsNaN(f) || math.IsInf(f, 0) {
				numeric[c] = false
				break
			}
		}
	}
	rows := make([]row, len(records))
	for i, rec := range records {
		rows[i] = make(row, len(head))
		for c, h := range head {
			if numeric[c] && rec[c] != "" {
				f, _ := strconv.ParseFloat(strings.TrimSpace(rec[c]), 64)
				rows[i][h] = f
			} else {
				rows[i][h] = rec[c]
			}
		}
	}
	return rows, nil
}

func evaluate(source []row, q Query) (any, error) {
	op := strings.ToLower(strings.TrimSpace(q.Op))
	if op != "stats" && op != "sum" && op != "values" && op != "ratio" {
		return nil, fmt.Errorf("unknown op %q", q.Op)
	}
	if len(q.Fields) > 0 && op != "stats" {
		return nil, errors.New("fields is only supported for stats")
	}
	if q.GroupBy != "" && op != "sum" {
		return nil, fmt.Errorf("group_by is only supported for sum")
	}
	if err := validateQueryColumns(source, q); err != nil {
		return nil, err
	}
	rows := make([]row, 0, len(source))
	for _, r := range source {
		ok, err := matches(r, q.Where)
		if err != nil {
			return nil, err
		}
		if ok {
			rows = append(rows, r)
		}
	}
	if err := validateSort(rows, q.Sort); err != nil {
		return nil, err
	}
	if len(q.Sort) > 0 {
		sort.SliceStable(rows, func(i, j int) bool { return lessRows(rows[i], rows[j], q.Sort) })
	}
	if q.Limit != nil {
		if *q.Limit < 0 {
			return nil, errors.New("limit must be nonnegative")
		}
		if *q.Limit < len(rows) {
			rows = rows[:*q.Limit]
		}
	}
	switch op {
	case "stats":
		if q.Column == "" {
			return nil, errors.New("stats requires column")
		}
		vals, err := numericColumn(rows, q.Column)
		if err != nil {
			return nil, err
		}
		stats, err := ops.Stats(vals)
		if err != nil || len(q.Fields) == 0 {
			return stats, err
		}
		selected := make(map[string]any, len(q.Fields))
		seen := make(map[string]bool, len(q.Fields))
		for _, field := range q.Fields {
			if field == "" {
				return nil, errors.New("stats fields cannot be empty")
			}
			if seen[field] {
				return nil, fmt.Errorf("duplicate stats field %q", field)
			}
			value, ok := stats[field]
			if !ok {
				return nil, fmt.Errorf("unknown stats field %q", field)
			}
			seen[field] = true
			selected[field] = value
		}
		return selected, nil
	case "sum":
		if q.Column == "" {
			return nil, errors.New("sum requires column")
		}
		if q.GroupBy != "" {
			return groupedSum(rows, q.Column, q.GroupBy)
		}
		return sumColumn(rows, q.Column)
	case "values":
		if q.Column == "" {
			return nil, errors.New("values requires column")
		}
		out := make([]any, len(rows))
		for i, r := range rows {
			v, ok := r[q.Column]
			if !ok {
				return nil, fmt.Errorf("missing column %q", q.Column)
			}
			out[i] = v
		}
		return out, nil
	case "ratio":
		if q.Numerator == "" || q.Denominator == "" {
			return nil, errors.New("ratio requires numerator and denominator")
		}
		n, err := sumColumn(rows, q.Numerator)
		if err != nil {
			return nil, err
		}
		d, err := sumColumn(rows, q.Denominator)
		if err != nil {
			return nil, err
		}
		if d == 0 || math.IsNaN(d) || math.IsInf(d, 0) {
			return nil, errors.New("ratio denominator is zero or nonfinite")
		}
		v := n / d
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("ratio result is nonfinite")
		}
		return v, nil
	}
	return nil, nil
}

func matches(r row, where map[string]any) (bool, error) {
	for k, want := range where {
		got, ok := r[k]
		if !ok {
			return false, fmt.Errorf("missing column %q", k)
		}
		if !equal(got, want) {
			return false, nil
		}
	}
	return true, nil
}
func equal(a, b any) bool {
	af, aok := number(a)
	bf, bok := number(b)
	if aok && bok {
		return af == bf
	}
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	}
	return false
}

func number(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func numericColumn(rows []row, col string) ([]float64, error) {
	vals := make([]float64, len(rows))
	for i, r := range rows {
		v, ok := r[col]
		if !ok {
			return nil, fmt.Errorf("missing column %q", col)
		}
		f, ok := number(v)
		if !ok || math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("column %q must contain finite numbers", col)
		}
		vals[i] = f
	}
	if len(vals) == 0 {
		return nil, fmt.Errorf("column %q has no rows", col)
	}
	return vals, nil
}
func sumColumn(rows []row, col string) (float64, error) {
	vals, err := numericColumn(rows, col)
	if err != nil {
		return 0, err
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
		if math.IsNaN(sum) || math.IsInf(sum, 0) {
			return 0, fmt.Errorf("sum of %q is nonfinite", col)
		}
	}
	return sum, nil
}
func groupedSum(rows []row, col, group string) (map[string]float64, error) {
	out := map[string]float64{}
	for _, r := range rows {
		g, ok := r[group]
		if !ok {
			return nil, fmt.Errorf("missing column %q", group)
		}
		v, ok := number(r[col])
		if !ok || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("column %q must contain finite numbers", col)
		}
		key := fmt.Sprint(g)
		out[key] += v
		if math.IsNaN(out[key]) || math.IsInf(out[key], 0) {
			return nil, fmt.Errorf("sum of %q is nonfinite", col)
		}
	}
	return out, nil
}

func validateSort(rows []row, ss []Sort) error {
	for _, s := range ss {
		if s.Column == "" {
			return errors.New("sort requires column")
		}
		for _, r := range rows {
			if _, ok := r[s.Column]; !ok {
				return fmt.Errorf("missing column %q", s.Column)
			}
		}
		if len(rows) > 1 {
			kind := scalarKind(rows[0][s.Column])
			if kind == "complex" {
				return fmt.Errorf("sort column %q contains nested values", s.Column)
			}
			for _, r := range rows[1:] {
				if scalarKind(r[s.Column]) != kind {
					return fmt.Errorf("sort column %q contains mixed types", s.Column)
				}
			}
		}
	}
	return nil
}
func lessRows(a, b row, ss []Sort) bool {
	for _, s := range ss {
		c := compare(a[s.Column], b[s.Column])
		if c != 0 {
			if s.Desc {
				return c > 0
			}
			return c < 0
		}
	}
	return false
}
func compare(a, b any) int {
	af, aok := number(a)
	bf, bok := number(b)
	if aok && bok {
		if af < bf {
			return -1
		}
		if af > bf {
			return 1
		}
		return 0
	}
	as, bs := fmt.Sprint(a), fmt.Sprint(b)
	if as < bs {
		return -1
	}
	if as > bs {
		return 1
	}
	return 0
}

func scalarKind(v any) string {
	if _, ok := number(v); ok {
		return "number"
	}
	switch v.(type) {
	case string:
		return "string"
	case bool:
		return "bool"
	case nil:
		return "nil"
	default:
		return "complex"
	}
}

func validateQueryColumns(rows []row, q Query) error {
	refs := []string{q.Column, q.GroupBy, q.Numerator, q.Denominator}
	for k := range q.Where {
		refs = append(refs, k)
	}
	for _, s := range q.Sort {
		refs = append(refs, s.Column)
	}
	for _, col := range refs {
		if col == "" {
			continue
		}
		found := false
		for _, r := range rows {
			if _, ok := r[col]; ok {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing column %q", col)
		}
	}
	for _, col := range []string{q.GroupBy} {
		if col != "" {
			for _, r := range rows {
				if scalarKind(r[col]) == "complex" {
					return fmt.Errorf("group column %q contains nested values", col)
				}
			}
		}
	}
	for _, s := range q.Sort {
		for _, r := range rows {
			if scalarKind(r[s.Column]) == "complex" {
				return fmt.Errorf("sort column %q contains nested values", s.Column)
			}
		}
	}
	return nil
}
