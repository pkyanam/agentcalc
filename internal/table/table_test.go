package table

import (
	"encoding/json"
	"testing"
)

func TestEvaluateCSV(t *testing.T) {
	input := []byte("id,team,paid,weight\n1,A,10,2\n2,B,5,1\n3,A,7,3\n4,B,8,4\n")
	limit := 2
	got, err := Evaluate(input, map[string]Query{
		"stats":  {Op: "stats", Column: "paid", Where: map[string]any{"team": "A"}},
		"groups": {Op: "sum", Column: "paid", GroupBy: "team"},
		"top":    {Op: "values", Column: "id", Sort: []Sort{{Column: "paid", Desc: true}}, Limit: &limit},
		"ratio":  {Op: "ratio", Numerator: "paid", Denominator: "weight"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["groups"].(map[string]float64)["A"] != 17 {
		t.Fatal(got)
	}
	if got["top"].([]any)[0] != float64(1) {
		t.Fatal(got)
	}
	if got["ratio"].(float64) != 30.0/10.0 {
		t.Fatal(got)
	}
	if got["stats"].(map[string]any)["count"] != 2 {
		t.Fatal(got)
	}
}

func TestStatsFields(t *testing.T) {
	input := []byte("x\n1\n2\n3\n")
	got, err := Evaluate(input, map[string]Query{"q": {Op: "stats", Column: "x", Fields: []string{"count", "mean"}}})
	if err != nil {
		t.Fatal(err)
	}
	stats := got["q"].(map[string]any)
	if len(stats) != 2 || stats["count"] != 3 || stats["mean"] != 2.0 {
		t.Fatalf("unexpected projection: %#v", stats)
	}
	for _, q := range []Query{
		{Op: "stats", Column: "x", Fields: []string{"count", "count"}},
		{Op: "stats", Column: "x", Fields: []string{"bogus"}},
		{Op: "sum", Column: "x", Fields: []string{"sum"}},
	} {
		if _, err := Evaluate(input, map[string]Query{"q": q}); err == nil {
			t.Errorf("accepted invalid fields query: %#v", q)
		}
	}
}

func TestEvaluateJSONAndErrors(t *testing.T) {
	got, err := Evaluate([]byte(`[{"name":"a","x":2},{"name":"b","x":3}]`), map[string]Query{"v": {Op: "values", Column: "name"}})
	if err != nil || len(got["v"].([]any)) != 2 {
		t.Fatalf("%v %#v", err, got)
	}
	if _, err := Evaluate([]byte(`[{"name":"a","x":2},{"name":"b","x":3}]`), map[string]Query{"s": {Op: "sum", Column: "x", Where: map[string]any{"x": json.Number("2")}}}); err != nil {
		t.Fatalf("json number filter: %v", err)
	}
	for _, in := range [][]byte{nil, []byte("a,a\n1,2\n"), []byte("a\n1,2\n")} {
		if _, err := Evaluate(in, map[string]Query{"q": {Op: "values", Column: "a"}}); err == nil {
			t.Errorf("accepted invalid input %q", in)
		}
	}
	if _, err := Evaluate([]byte("x\nfoo\n"), map[string]Query{"q": {Op: "stats", Column: "x"}}); err == nil {
		t.Error("accepted nonnumeric stats")
	}
	if _, err := Evaluate([]byte("n,d\n1,0\n"), map[string]Query{"q": {Op: "ratio", Numerator: "n", Denominator: "d"}}); err == nil {
		t.Error("accepted zero denominator")
	}
}
