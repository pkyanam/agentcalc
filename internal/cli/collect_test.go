package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestCollectOutput(t *testing.T) {
	input := `{"id":"sum","command":"exact","expr":"0.1 + 0.2","select":"fraction"}` + "\n" + `{"id":"answer","command":"eval","expr":"6*7"}`
	var out, err bytes.Buffer
	if code := Main([]string{"batch", "--collect", "--text"}, strings.NewReader(input), &out, &err, "test"); code != 0 {
		t.Fatal(out.String(), err.String())
	}
	var got map[string]any
	if e := json.Unmarshal(out.Bytes(), &got); e != nil {
		t.Fatal(e)
	}
	if got["sum"] != "3/10" || got["answer"] != float64(42) {
		t.Fatal(got)
	}
}
func TestCollectFailureAtomicity(t *testing.T) {
	for _, bad := range []string{
		`{"id":"ok","command":"eval","expr":"3"}`,
		`{"id":"bad","command":"eval","expr":"1/0"}`,
		`{"id":5,"command":"eval","expr":"3"}`,
		`{"command":"eval","expr":"3"}`,
		`{"id":"bad","command":"exact","expr":"1 + 2","select":"wrong"}`,
	} {
		input := `{"id":"ok","command":"eval","expr":"42"}` + "\n" + bad
		code, got := run(t, []string{"batch", "--collect", "--text"}, input)
		if code != 1 || got["ok"] != false || got["result"] != nil {
			t.Fatalf("partial/invalid success: %v", got)
		}
	}
}
func TestNativeTableCLI(t *testing.T) {
	data := "id,region,status,units,revenue\n1,east,paid,2,10\n2,west,pending,1,8\n3,east,paid,4,24\n"
	query := `{"s":{"op":"stats","column":"revenue"},"g":{"op":"sum","column":"revenue","group_by":"region","where":{"status":"paid"}},"ids":{"op":"values","column":"id","sort":[{"column":"revenue","desc":true}],"limit":2},"ratio":{"op":"ratio","numerator":"revenue","denominator":"units"}}`
	code, got := run(t, []string{"table", "--query", query, "--text"}, data)
	if code != 0 {
		t.Fatal(got)
	}
	if got["ratio"] != float64(6) || got["g"].(map[string]any)["east"] != float64(34) {
		t.Fatal(got)
	}
	ids := got["ids"].([]any)
	if ids[0] != float64(3) || ids[1] != float64(1) {
		t.Fatal(ids)
	}
}
func TestNativeTableNumericFilter(t *testing.T) {
	code, got := run(t, []string{"table", "--query", `{"v":{"op":"sum","column":"x","where":{"id":1.0}}}`, "--text"}, `[{"id":1,"x":3},{"id":2,"x":8}]`)
	if code != 0 || got["v"] != float64(3) {
		t.Fatal(got)
	}
}

func TestSolveFlatVector(t *testing.T) {
	for _, args := range [][]string{
		{"matrix", "solve", "--data", `{"a":[[2,1],[1,-1]],"b":[5,1]}`},
		{"batch", "--collect"},
	} {
		input := `{"id":"solution","command":"matrix","op":"solve","a":[[2,1],[1,-1]],"b":[5,1]}`
		code, got := run(t, args, input)
		if code != 0 {
			t.Fatal(got)
		}
	}
}
func TestInvalidVectorEntries(t *testing.T) {
	for _, b := range []string{`[5,null]`, `[5,[1]]`, `["5",1]`, `[[5],[null]]`} {
		code, got := run(t, []string{"matrix", "solve", "--data", `{"a":[[2,1],[1,-1]],"b":` + b + `}`}, "")
		if code == 0 {
			t.Fatal(got)
		}
	}
}
