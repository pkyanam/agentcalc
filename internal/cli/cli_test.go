package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
)

func run(t *testing.T, args []string, input string) (int, map[string]any) {
	t.Helper()
	var out, err bytes.Buffer
	code := Main(args, strings.NewReader(input), &out, &err, "test")
	var v map[string]any
	if e := json.Unmarshal(out.Bytes(), &v); e != nil {
		t.Fatalf("invalid output %s: %v", out.String(), e)
	}
	return code, v
}
func TestEndToEnd(t *testing.T) {
	cases := []struct {
		args  []string
		input string
		want  any
	}{
		{[]string{"eval", "sqrt(144)+2^10"}, "", float64(1036)},
		{[]string{"eval", "price*(1+tax)", "--var", "price=100", "--var", "tax=0.08"}, "", float64(108)},
		{[]string{"convert", "1", "km", "m"}, "", float64(1000)},
		{[]string{"matrix", "determinant"}, `{"a":[[4,7],[2,6]]}`, float64(10)},
	}
	for _, c := range cases {
		code, v := run(t, c.args, c.input)
		if code != 0 || v["result"] != c.want {
			t.Errorf("%v: code %d output %v", c.args, code, v)
		}
	}
}
func TestExact(t *testing.T) {
	code, v := run(t, []string{"exact", "0.1 + 0.2"}, "")
	if code != 0 || v["result"].(map[string]any)["fraction"] != "3/10" {
		t.Fatal(v)
	}
}
func TestCSV(t *testing.T) {
	code, v := run(t, []string{"stats", "--column", "revenue"}, "item,revenue\na,10\nb,20\nc,30\n")
	if code != 0 || v["result"].(map[string]any)["mean"] != float64(20) {
		t.Fatal(v)
	}
}
func TestErrors(t *testing.T) {
	for _, args := range [][]string{{"eval", "1/0"}, {"stats", "NaN"}, {"matrix", "inverse", "--data", `{"a":[[1,2],[2,4]]}`}, {"eval"}, {"wat"}, {"exact", "1 / 0"}} {
		code, v := run(t, args, "")
		if code == 0 || v["ok"] != false {
			t.Fatalf("%v: %v", args, v)
		}
	}
}
func TestBatch(t *testing.T) {
	var out, err bytes.Buffer
	code := Main([]string{"batch"}, strings.NewReader("{\"id\":1,\"command\":\"eval\",\"expr\":\"6*7\"}\n{\"id\":2,\"command\":\"eval\",\"expr\":\"1/0\"}\n{\"id\":3,\"command\":\"eval\",\"expr\":\"2+2\"}\n"), &out, &err, "test")
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if code != 1 || len(lines) != 3 {
		t.Fatalf("%d %s", code, out.String())
	}
	var v map[string]any
	json.Unmarshal([]byte(lines[2]), &v)
	if v["result"] != float64(4) || v["id"] != float64(3) {
		t.Fatal(v)
	}
}

func TestRunAtomicAndNativeResults(t *testing.T) {
	var out, errOut bytes.Buffer
	input := "growth = eval 1000*(1+0.05/12)^24\n" +
		"fraction = exact 0.1 + 0.2\n" +
		"bytes = convert 3.75 GiB B\n" +
		"root = root cos(x)-x 0 1\n" +
		"solution = matrix solve {\"a\":[[2,1],[1,-1]],\"b\":[5,1]}\n"
	if code := Main([]string{"run", "--text"}, strings.NewReader(input), &out, &errOut, "test"); code != 0 {
		t.Fatalf("code %d: %s %s", code, out.String(), errOut.String())
	}
	var got map[string]any
	if e := json.Unmarshal(out.Bytes(), &got); e != nil {
		t.Fatal(e)
	}
	if got["fraction"] != "3/10" || got["bytes"] != float64(4026531840) {
		t.Fatal(got)
	}
	if _, ok := got["solution"].([]any); !ok {
		t.Fatalf("matrix result: %#v", got["solution"])
	}

	out.Reset()
	bad := "ok = eval 42\nbad = eval 1/0\nthird = eval 3\n"
	if code := Main([]string{"run", "--text"}, strings.NewReader(bad), &out, &errOut, "test"); code == 0 || out.String() == "{}\n" {
		t.Fatalf("expected atomic failure: code=%d output=%q", code, out.String())
	}
	if strings.Contains(out.String(), "42") {
		t.Fatal("partial result emitted")
	}
}

func TestRunDuplicateAndLimits(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Main([]string{"run", "--text"}, strings.NewReader("x = eval 1\nx = eval 2\n"), &out, &errOut, "test"); code == 0 {
		t.Fatal(out.String())
	}
	var many strings.Builder
	for i := 0; i < 1001; i++ {
		many.WriteString("x")
		many.WriteString(strconv.Itoa(i))
		many.WriteString(" = eval 1\n")
	}
	out.Reset()
	if code := Main([]string{"run", "--text"}, strings.NewReader(many.String()), &out, &errOut, "test"); code == 0 {
		t.Fatal("expected command limit")
	}
}

func TestRunParserTable(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, map[string]any)
	}{
		{"expression spaces and bounds", "r = root cos ( x ) - x 0 1\n", func(t *testing.T, got map[string]any) {
			v, ok := got["r"].(float64)
			if !ok || v < 0.73908 || v > 0.73909 {
				t.Fatalf("root=%v", got["r"])
			}
		}},
		{"derivative", "d = derivative x ^ 3 2\n", func(t *testing.T, got map[string]any) {
			v, ok := got["d"].(float64)
			if !ok || v < 11.999 || v > 12.001 {
				t.Fatalf("derivative=%v", got["d"])
			}
		}},
		{"stats", "s = stats 1 2 3 4\n", func(t *testing.T, got map[string]any) {
			s := got["s"].(map[string]any)
			if s["mean"] != float64(2.5) {
				t.Fatalf("stats=%v", s)
			}
		}},
		{"matrix solution", "m = matrix solve {\"a\":[[2,1],[1,-1]],\"b\":[5,1]}\n", func(t *testing.T, got map[string]any) {
			m := got["m"].([]any)
			if len(m) != 2 || m[0].(float64) < 1.999 || m[0].(float64) > 2.001 || m[1].(float64) < .999 || m[1].(float64) > 1.001 {
				t.Fatalf("matrix=%v", m)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if code := Main([]string{"run", "--text"}, strings.NewReader(tt.input), &out, &errOut, "test"); code != 0 {
				t.Fatalf("code=%d output=%s error=%s", code, out.String(), errOut.String())
			}
			var got map[string]any
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			tt.check(t, got)
		})
	}
}

func TestRunFileAndRejectedInputs(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/commands.txt"
	if err := os.WriteFile(path, []byte("# calculations\n\nanswer = eval 6 * 7\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := Main([]string{"run", "--input", path, "--text"}, strings.NewReader("ignored = eval 0"), &out, &errOut, "test"); code != 0 {
		t.Fatalf("file run failed: %d %s", code, out.String())
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil || got["answer"] != float64(42) {
		t.Fatalf("file result=%s err=%v", out.String(), err)
	}

	bad := []string{
		"\n# comment only\n",
		"1bad = eval 1\n",
		"bad-name = eval 1\n",
		"x = matrix determinant {\"a\":[[1]],\"extra\":1}\n",
		"x = python 1+1\n",
		"x = root x NaN 1\n",
	}
	for _, input := range bad {
		out.Reset()
		errOut.Reset()
		if code := Main([]string{"run", "--text"}, strings.NewReader(input), &out, &errOut, "test"); code == 0 {
			t.Errorf("accepted invalid run input %q: %s", input, out.String())
		}
	}
}

func TestTrailingJSON(t *testing.T) {
	code, v := run(t, []string{"stats", "--data", "[1,2] [3]"}, "")
	if code == 0 {
		t.Fatal(v)
	}
}

func TestNumericalCLI(t *testing.T) {
	for _, c := range []struct {
		args []string
		want float64
	}{
		{[]string{"root", "1e-20*(x-1)", "0", "2"}, 1},
		{[]string{"integrate", "x^2", "0", "3"}, 9},
		{[]string{"derivative", "x^3", "2"}, 12},
	} {
		code, v := run(t, c.args, "")
		if code != 0 {
			t.Fatal(v)
		}
		got := v["result"].(float64)
		if got < c.want-1e-7 || got > c.want+1e-7 {
			t.Fatal(v)
		}
	}
}
func TestNullScriptResult(t *testing.T) {
	var out bytes.Buffer
	if e := emit(&out, nil, nil, nil, false, false); e != nil {
		t.Fatal(e)
	}
	var v map[string]any
	json.Unmarshal(out.Bytes(), &v)
	if _, ok := v["result"]; !ok {
		t.Fatal("success must include null result")
	}
}
