// Package cli provides the command-line and JSON Lines interfaces.
package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkyanam/agentcalc/internal/calc"
	"github.com/pkyanam/agentcalc/internal/ops"
	"github.com/pkyanam/agentcalc/internal/script"
	"github.com/pkyanam/agentcalc/internal/table"
)

const Help = `agentcalc — precise answers, compact requests

Usage: agentcalc <command> [arguments] [options]

Commands:
  eval 'EXPR' [--var x=2]          Scientific arithmetic; ^ and ** are powers
  exact 'NUMBER OP NUMBER'        Exact rational + - * / and integer powers
  stats [NUMBERS...]              Statistics; JSON array or CSV from stdin
  convert VALUE FROM TO          Unit conversion (case-insensitive unit names)
  root 'EXPR' LOWER UPPER         Find a bracketed root (variable x)
  integrate 'EXPR' LOWER UPPER    Numerical definite integral (variable x)
  derivative 'EXPR' X            Numerical derivative at x
  units                          List supported units by dimension
  matrix OP                      Read {"a":[[...]],"b":[[...]]} from stdin
  python 'EXPR'                   Python expression with data, math, json
  node 'EXPR'                     JavaScript expression with data and Math
  python|node --file PATH         Script defining main(data)
  batch                          JSON Lines requests on stdin, one response each
  table --query JSON             Named CSV/JSON queries: stats, sum, values, ratio
  version                        Show build version
  help                           This guide

Options:
  --data JSON                    Input value instead of stdin
  --input PATH                   Read input from a file (use - for stdin)
  --column NAME                  Select a numeric CSV column for stats
  --var NAME=VALUE                Set an expression variable (repeatable)
  --timeout DURATION             Script timeout, e.g. 2s (default 5s, max 5m)
  --text                         Print only the result (objects remain JSON)
  --pretty                       Indent JSON output (batch requires --collect)
  --collect                      Batch results as one object keyed by string id
  --query JSON                   Named queries for table (or --query-file PATH)
  --help, -h                     Show help

Examples:
  agentcalc eval 'sqrt(144) + 2^10'
  agentcalc eval 'price * (1 + tax)' --var price=49.95 --var tax=0.08
  agentcalc exact '0.1 + 0.2'
  agentcalc stats 1 2 3 4 5
  agentcalc stats --input sales.csv --column revenue
  agentcalc convert 72 f c
  agentcalc matrix inverse --data '{"a":[[4,7],[2,6]]}'
  agentcalc python 'sum(x*x for x in data)' --data '[1,2,3]'
  agentcalc node 'data.map(x => x * 2)' --data '[1,2,3]'
  echo '{"id":"a","command":"eval","expr":"6*7"}' | agentcalc batch

Default response: {"ok":true,"result":...}; errors: {"ok":false,"error":"..."}.
Exit codes: 0 success, 1 calculation/input/runtime error, 2 usage error.
Batch continues after errors and exits 1 if any request failed. IDs are echoed.
Batch commands: eval, exact, stats, convert, matrix, units, root, integrate, derivative.
No script execution in batch.
Batch fields: command, id, expr, vars, values, value, from, to, op, a, b, lower, upper, x.
Optional select picks a result object field (e.g. "fraction" for exact).
Use batch --collect --text to get one keyed JSON object without a wrapper.
Collect requires unique nonempty string IDs; any error returns one error object.
Table example: --query '{"totals":{"op":"sum","column":"revenue","group_by":"region"}}'
Table query fields: op, column, where (equality), group_by, sort [{column,desc}],
limit, numerator, denominator, fields (stats result keys). Filters, sort, limit
apply before aggregation.
Matrix operations: add, subtract, multiply, transpose, determinant, inverse, solve.
For solve, b may be [5,1] or [[5],[1]].
Numerical calculus is approximate and assumes smooth functions; roots need a
continuous function with a sign-changing bracket. Use finite integration bounds.
Floating point uses IEEE-754 float64; use exact for decimal/fraction arithmetic.
Python/Node run trusted code with your local permissions; they are NOT sandboxed.
Input/output limits: 8 MiB input, 1 MiB per batch line. No network or telemetry
in core commands. Quote expressions to prevent shell expansion.
`

type request struct {
	Select  string             `json:"select"`
	Lower   float64            `json:"lower"`
	Upper   float64            `json:"upper"`
	X       float64            `json:"x"`
	ID      any                `json:"id,omitempty"`
	Command string             `json:"command"`
	Expr    string             `json:"expr"`
	Vars    map[string]float64 `json:"vars"`
	Values  []float64          `json:"values"`
	Value   float64            `json:"value"`
	From    string             `json:"from"`
	To      string             `json:"to"`
	Op      string             `json:"op"`
	A       [][]float64        `json:"a"`
	B       rightMatrix        `json:"b"`
}

// rightMatrix accepts a column-vector shorthand, matching conventional Ax=b input.
type rightMatrix [][]float64

func (m *rightMatrix) UnmarshalJSON(data []byte) error {
	var elements []json.RawMessage
	if e := json.Unmarshal(data, &elements); e != nil {
		return errors.New("b must be a numeric vector or matrix")
	}
	if string(data) == "null" {
		*m = nil
		return nil
	}
	rows := make([][]float64, len(elements))
	vector := -1
	for i, raw := range elements {
		isRow := len(raw) > 0 && raw[0] == '['
		shape := 0
		if isRow {
			shape = 1
		}
		if vector >= 0 && vector != shape {
			return errors.New("b cannot mix rows and scalar values")
		}
		vector = shape
		numbers := []json.RawMessage{raw}
		if isRow {
			if e := json.Unmarshal(raw, &numbers); e != nil {
				return e
			}
		}
		rows[i] = make([]float64, len(numbers))
		for j, n := range numbers {
			if string(n) == "null" {
				return errors.New("b entries must be numbers, not null")
			}
			if e := json.Unmarshal(n, &rows[i][j]); e != nil {
				return errors.New("b entries must be numbers")
			}
		}
	}
	*m = rows
	return nil
}

type response struct {
	OK     bool   `json:"ok"`
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"`
	ID     any    `json:"id,omitempty"`
}

func execute(r request) (any, error) {
	result, err := executeCore(r)
	if err != nil || r.Select == "" {
		return result, err
	}
	object, ok := result.(map[string]any)
	if !ok {
		return nil, errors.New("select requires an object result")
	}
	value, ok := object[r.Select]
	if !ok {
		return nil, fmt.Errorf("result has no field %q", r.Select)
	}
	return value, nil
}

func executeCore(r request) (any, error) {
	switch r.Command {
	case "root":
		return calc.Root(r.Expr, r.Lower, r.Upper)
	case "integrate":
		return calc.Integrate(r.Expr, r.Lower, r.Upper)
	case "derivative":
		return calc.Derivative(r.Expr, r.X)
	case "eval":
		return calc.Eval(r.Expr, r.Vars)
	case "exact":
		return exact(r.Expr)
	case "stats":
		return ops.Stats(r.Values)
	case "convert":
		return ops.Convert(r.Value, r.From, r.To)
	case "units":
		return ops.Units(), nil
	case "matrix":
		return ops.Matrix(r.Op, r.A, r.B)
	default:
		return nil, fmt.Errorf("unknown command %q", r.Command)
	}
}

var literalExponent = regexp.MustCompile(`[eEpP]([+-]?[0-9]+)`)

func exact(expr string) (any, error) {
	if len(expr) > 10000 {
		return nil, errors.New("exact expression too long")
	}
	parts := strings.Fields(expr)
	if len(parts) != 3 {
		return nil, errors.New("exact expects 'NUMBER OP NUMBER' with spaces, e.g. '1/3 + 0.2'")
	}
	for _, literal := range []string{parts[0], parts[2]} {
		for _, exponent := range literalExponent.FindAllStringSubmatch(literal, -1) {
			n, err := strconv.ParseInt(exponent[1], 10, 64)
			if err != nil || n < -10000 || n > 10000 {
				return nil, errors.New("literal exponent limit is 10000")
			}
		}
	}
	a, ok := new(big.Rat).SetString(parts[0])
	if !ok {
		return nil, errors.New("invalid left rational")
	}
	if len(expr) > 10000 {
		return nil, errors.New("exact expression too long")
	}
	b, ok := new(big.Rat).SetString(parts[2])
	if !ok {
		return nil, errors.New("invalid right rational")
	}
	out := new(big.Rat)
	switch parts[1] {
	case "+":
		out.Add(a, b)
	case "-":
		out.Sub(a, b)
	case "*":
		out.Mul(a, b)
	case "/":
		if b.Sign() == 0 {
			return nil, errors.New("division by zero")
		}
		out.Quo(a, b)
	case "^", "**":
		if !b.IsInt() || !b.Num().IsInt64() {
			return nil, errors.New("exponent must be an integer")
		}
		n := b.Num().Int64()
		if n < -10000 || n > 10000 {
			return nil, errors.New("exponent limit is 10000")
		}
		if n < 0 {
			if a.Sign() == 0 {
				return nil, errors.New("zero to negative power")
			}
			a.Inv(a)
			n = -n
		}
		if int64(a.Num().BitLen()+a.Denom().BitLen())*n > 1000000 {
			return nil, errors.New("exact result exceeds one million bits")
		}
		out.SetFrac(new(big.Int).Exp(a.Num(), big.NewInt(n), nil), new(big.Int).Exp(a.Denom(), big.NewInt(n), nil))
	default:
		return nil, errors.New("exact supports + - * / ^ **")
	}
	return map[string]any{"fraction": out.RatString(), "decimal": out.FloatString(30)}, nil
}

func emit(w io.Writer, result any, err error, id any, plain, pretty bool) error {
	if plain && err == nil {
		switch v := result.(type) {
		case string:
			_, e := fmt.Fprintln(w, v)
			return e
		case float64:
			_, e := fmt.Fprintln(w, strconv.FormatFloat(v, 'g', -1, 64))
			return e
		}
	}
	var value any = result
	if !plain || err != nil {
		r := response{OK: err == nil, Result: result, ID: id}
		if err != nil {
			r.Result = nil
			r.Error = err.Error()
		}
		value = r
	}
	var buffer bytes.Buffer
	enc := json.NewEncoder(&buffer)
	enc.SetEscapeHTML(false)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if e := enc.Encode(value); e != nil {
		return e
	}
	_, e := w.Write(buffer.Bytes())
	return e
}
func readLimited(r io.Reader) ([]byte, error) {
	b, e := io.ReadAll(io.LimitReader(r, 8*1024*1024+1))
	if e == nil && len(b) > 8*1024*1024 {
		e = errors.New("input exceeds 8 MiB")
	}
	return b, e
}
func decode(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return errors.New("expected one JSON value")
	}
	return nil
}

func Main(args []string, in io.Reader, out, stderr io.Writer, version string) int {
	if len(args) == 0 {
		fmt.Fprint(out, Help)
		return 0
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(out, Help)
		return 0
	}
	if args[0] == "version" || args[0] == "--version" {
		fmt.Fprintln(out, "agentcalc "+version)
		return 0
	}
	cmd := args[0]
	var pos []string
	vars := map[string]float64{}
	data, input, column, file := "", "", "", ""
	query, queryFile := "", ""
	collect := false
	plain, pretty, hasData := false, false, false
	timeout := 5 * time.Second
	fail := func(e error, code int) int {
		if err := emit(out, nil, e, nil, false, pretty); err != nil {
			fmt.Fprintln(stderr, err)
		}
		return code
	}
	for i := 1; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if a == "--help" || a == "-h" {
			fmt.Fprint(out, Help)
			return 0
		}
		if a == "--text" {
			plain = true
			continue
		}
		if a == "--collect" {
			collect = true
			continue
		}
		if a == "--pretty" {
			pretty = true
			continue
		}
		switch a {
		case "--data", "--input", "--column", "--file", "--var", "--timeout", "--query", "--query-file":
			if i+1 == len(args) {
				return fail(fmt.Errorf("%s needs a value", a), 2)
			}
			i++
			v := args[i]
			switch a {
			case "--query":
				query = v
			case "--query-file":
				queryFile = v
			case "--data":
				data = v
				hasData = true
			case "--input":
				input = v
			case "--column":
				column = v
			case "--file":
				file = v
			case "--timeout":
				d, e := time.ParseDuration(v)
				if e != nil || d <= 0 || d > 5*time.Minute {
					return fail(errors.New("timeout must be positive and at most 5m"), 2)
				}
				timeout = d
			case "--var":
				k, val, ok := strings.Cut(v, "=")
				f, e := strconv.ParseFloat(val, 64)
				if !ok || k == "" || e != nil || math.IsNaN(f) || math.IsInf(f, 0) {
					return fail(errors.New("--var expects NAME=finite-number"), 2)
				}
				vars[k] = f
			}
		default:
			if strings.HasPrefix(a, "--") {
				return fail(fmt.Errorf("unknown option %s", a), 2)
			}
			pos = append(pos, a)
		}
	}
	if collect && cmd != "batch" {
		return fail(errors.New("--collect requires batch"), 2)
	}
	if (query != "" || queryFile != "") && cmd != "table" {
		return fail(errors.New("--query and --query-file require table"), 2)
	}
	if query != "" && queryFile != "" {
		return fail(errors.New("choose --query or --query-file"), 2)
	}
	if column != "" && cmd != "stats" {
		return fail(errors.New("--column requires stats"), 2)
	}
	if file != "" && cmd != "python" && cmd != "node" {
		return fail(errors.New("--file requires python or node"), 2)
	}
	if len(vars) > 0 && cmd != "eval" {
		return fail(errors.New("--var requires eval"), 2)
	}
	if (hasData || input != "") && cmd != "stats" && cmd != "matrix" && cmd != "python" && cmd != "node" && cmd != "batch" && cmd != "table" {
		return fail(errors.New("this command does not accept input data"), 2)
	}
	if hasData && input != "" {
		return fail(errors.New("--data and --input are mutually exclusive"), 2)
	}
	getInput := func() ([]byte, error) {
		if hasData {
			return readLimited(strings.NewReader(data))
		}
		if input != "" && input != "-" {
			f, e := os.Open(input)
			if e != nil {
				return nil, e
			}
			defer f.Close()
			return readLimited(f)
		}
		return readLimited(in)
	}
	if cmd == "batch" {
		if ((plain || pretty) && !collect) || len(pos) > 0 || hasData {
			return fail(errors.New("batch uses JSON Lines stdin/--input; --text/--pretty require --collect; no --data"), 2)
		}
		reader := in
		if input != "" && input != "-" {
			f, e := os.Open(input)
			if e != nil {
				return fail(e, 1)
			}
			defer f.Close()
			reader = f
		}
		if collect {
			return batchCollect(reader, out, plain, pretty)
		}
		return batch(reader, out)
	}
	r := request{Command: cmd, Vars: vars}
	var result any
	var err error
	switch cmd {
	case "table":
		if len(pos) != 0 || (query == "" && queryFile == "") {
			return fail(errors.New("table requires --query JSON or --query-file PATH"), 2)
		}
		queryData := []byte(query)
		if queryFile != "" {
			f, e := os.Open(queryFile)
			if e != nil {
				return fail(e, 1)
			}
			queryData, e = readLimited(f)
			f.Close()
			if e != nil {
				return fail(e, 1)
			}
		}
		var queries map[string]table.Query
		if e := decode(queryData, &queries); e != nil {
			return fail(e, 1)
		}
		b, e := getInput()
		if e != nil {
			return fail(e, 1)
		}
		result, err = table.Evaluate(b, queries)
	case "root", "integrate", "derivative":
		want := 3
		if cmd == "derivative" {
			want = 2
		}
		if len(pos) != want {
			return fail(fmt.Errorf("%s expects a quoted expression and %d numeric argument(s)", cmd, want-1), 2)
		}
		r.Expr = pos[0]
		lo, e := strconv.ParseFloat(pos[1], 64)
		if e != nil {
			return fail(e, 2)
		}
		if cmd == "derivative" {
			r.X = lo
		} else {
			r.Lower = lo
			r.Upper, e = strconv.ParseFloat(pos[2], 64)
			if e != nil {
				return fail(e, 2)
			}
		}
		result, err = execute(r)
	case "eval", "exact":
		if len(pos) == 0 {
			return fail(errors.New("missing expression"), 2)
		}
		r.Expr = strings.Join(pos, " ")
		result, err = execute(r)
	case "units":
		if len(pos) != 0 {
			return fail(errors.New("units takes no arguments"), 2)
		}
		result, err = execute(r)
	case "stats":
		if len(pos) > 0 {
			if column != "" {
				return fail(errors.New("--column requires CSV input"), 2)
			}
			if hasData || input != "" {
				return fail(errors.New("choose positional numbers or input data"), 2)
			}
			for _, p := range pos {
				v, e := strconv.ParseFloat(p, 64)
				if e != nil {
					return fail(e, 1)
				}
				r.Values = append(r.Values, v)
			}
		} else {
			b, e := getInput()
			if e != nil {
				return fail(e, 1)
			}
			r.Values, err = parseValues(b, column)
			if err != nil {
				return fail(err, 1)
			}
		}
		result, err = execute(r)
	case "convert":
		if len(pos) != 3 {
			return fail(errors.New("convert expects VALUE FROM TO"), 2)
		}
		r.Value, err = strconv.ParseFloat(pos[0], 64)
		if err != nil {
			return fail(err, 1)
		}
		r.From = pos[1]
		r.To = pos[2]
		result, err = execute(r)
	case "matrix":
		if len(pos) != 1 {
			return fail(errors.New("matrix expects one operation"), 2)
		}
		b, e := getInput()
		if e != nil {
			return fail(e, 1)
		}
		var matrices struct {
			A [][]float64 `json:"a"`
			B rightMatrix `json:"b"`
		}
		if e = decode(b, &matrices); e != nil {
			return fail(e, 1)
		}
		r.A = matrices.A
		r.B = matrices.B
		r.Op = pos[0]
		result, err = execute(r)
	case "python", "node":
		if (file == "" && len(pos) == 0) || (file != "" && len(pos) > 0) {
			return fail(errors.New("provide an expression or --file PATH"), 2)
		}
		var d any
		if hasData || input != "" {
			b, e := getInput()
			if e != nil {
				return fail(e, 1)
			}
			if e = decode(b, &d); e != nil {
				return fail(e, 1)
			}
		} else if f, ok := in.(*os.File); !ok {
			b, e := getInput()
			if e != nil {
				return fail(e, 1)
			}
			if len(bytes.TrimSpace(b)) > 0 {
				if e = decode(b, &d); e != nil {
					return fail(e, 1)
				}
			}
		} else if st, e := f.Stat(); e == nil && st.Mode()&os.ModeCharDevice == 0 {
			b, e := getInput()
			if e != nil {
				return fail(e, 1)
			}
			if len(bytes.TrimSpace(b)) > 0 {
				if e = decode(b, &d); e != nil {
					return fail(e, 1)
				}
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if file != "" {
			result, err = script.RunFile(ctx, cmd, file, d)
		} else {
			result, err = script.Run(ctx, cmd, strings.Join(pos, " "), d)
		}
	default:
		return fail(fmt.Errorf("unknown command %q; run agentcalc --help", cmd), 2)
	}
	if err != nil {
		return fail(err, 1)
	}
	if e := emit(out, result, nil, nil, plain, pretty); e != nil {
		fmt.Fprintln(stderr, e)
		return 1
	}
	return 0
}

func parseValues(b []byte, column string) ([]float64, error) {
	b = bytes.TrimSpace(b)
	if len(b) == 0 {
		return nil, errors.New("empty data")
	}
	var vals []float64
	if b[0] == '[' {
		if column != "" {
			return nil, errors.New("--column applies only to CSV")
		}
		e := decode(b, &vals)
		return vals, e
	}
	if column != "" {
		rows, e := csv.NewReader(bytes.NewReader(b)).ReadAll()
		if e != nil {
			return nil, e
		}
		idx := -1
		for i, h := range rows[0] {
			if h == column {
				idx = i
			}
		}
		if idx < 0 {
			return nil, fmt.Errorf("CSV column %q not found", column)
		}
		for n, row := range rows[1:] {
			v, e := strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
			if e != nil {
				return nil, fmt.Errorf("CSV row %d: %w", n+2, e)
			}
			vals = append(vals, v)
		}
		return vals, nil
	}
	for _, s := range strings.FieldsFunc(string(b), func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r' }) {
		v, e := strconv.ParseFloat(s, 64)
		if e != nil {
			return nil, fmt.Errorf("invalid number %q (CSV headers need --column)", s)
		}
		vals = append(vals, v)
	}
	return vals, nil
}

var requiredFields = map[string][]string{"eval": {"expr"}, "exact": {"expr"}, "stats": {"values"}, "convert": {"value", "from", "to"}, "matrix": {"op", "a"}, "root": {"expr", "lower", "upper"}, "integrate": {"expr", "lower", "upper"}, "derivative": {"expr", "x"}}

func parseRequest(line []byte) (request, error) {
	var r request
	if e := decode(line, &r); e != nil {
		return r, e
	}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(line, &fields); e != nil {
		return r, e
	}
	if raw, ok := fields["id"]; ok {
		r.ID = raw
	}
	for _, name := range requiredFields[r.Command] {
		if raw, ok := fields[name]; !ok || string(raw) == "null" {
			return r, fmt.Errorf("%s requires field %q", r.Command, name)
		}
	}
	return r, nil
}

func batch(in io.Reader, out io.Writer) int {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	code := 0
	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		r, e := parseRequest(scanner.Bytes())
		var value any
		if e == nil {
			value, e = execute(r)
		}
		if e != nil {
			code = 1
		}
		if e := emit(out, value, e, r.ID, false, false); e != nil {
			return 1
		}
	}
	if e := scanner.Err(); e != nil {
		emit(out, nil, e, nil, false, false)
		return 1
	}
	return code
}

// Collected batches are bounded and atomic: no partial successes hide errors.
func batchCollect(in io.Reader, out io.Writer, plain, pretty bool) int {
	fail := func(e error) int { emit(out, nil, e, nil, false, pretty); return 1 }
	data, e := readLimited(in)
	if e != nil {
		return fail(e)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	results := map[string]any{}
	collectedBytes := 2
	line := 0
	for scanner.Scan() {
		line++
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		r, e := parseRequest(scanner.Bytes())
		if e != nil {
			return fail(fmt.Errorf("line %d: %w", line, e))
		}
		raw, ok := r.ID.(json.RawMessage)
		var id string
		if !ok || json.Unmarshal(raw, &id) != nil || id == "" {
			return fail(fmt.Errorf("line %d: --collect requires a nonempty string id", line))
		}
		if _, exists := results[id]; exists {
			return fail(fmt.Errorf("duplicate id %q", id))
		}
		value, e := execute(r)
		if e != nil {
			return fail(fmt.Errorf("%s: %w", id, e))
		}
		encoded, e := json.Marshal(value)
		if e != nil {
			return fail(e)
		}
		key, _ := json.Marshal(id)
		collectedBytes += len(encoded) + len(key) + 2
		if collectedBytes > 8*1024*1024 {
			return fail(errors.New("collected output exceeds 8 MiB"))
		}
		results[id] = value
	}
	if e := scanner.Err(); e != nil {
		return fail(e)
	}
	// Bound accumulated result size as well as input size.
	b, e := json.Marshal(results)
	if e != nil {
		return fail(e)
	}
	if len(b) > 8*1024*1024 {
		return fail(errors.New("collected output exceeds 8 MiB"))
	}
	if e := emit(out, results, nil, nil, plain, pretty); e != nil {
		return 1
	}
	return 0
}
