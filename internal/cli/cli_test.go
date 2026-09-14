package cli

import (
	"bytes"
	"encoding/json"
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
