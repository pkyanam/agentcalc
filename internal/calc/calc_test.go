package calc

import (
	"math"
	"strings"
	"testing"
)

func TestEvalArithmeticAndPrecedence(t *testing.T) {
	for _, tc := range []struct {
		expr string
		want float64
	}{
		{"1+2*3", 7}, {"2^3^2", 512}, {"-2^2", -4}, {"2^-2", .25},
		{"(1+2)!", 6}, {"5%2", 1}, {"1e-3 + .5", .501},
	} {
		got, e := Eval(tc.expr, nil)
		if e != nil || math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("Eval(%q)=%v,%v want %v", tc.expr, got, e, tc.want)
		}
	}
}

func TestEvalFunctionsVariablesAndConstants(t *testing.T) {
	got, e := Eval("sin(pi/2)+sqrt(x)+mean(1,2,3)+choose(5,2)+clamp(9,0,4)", map[string]float64{"x": 4})
	if e != nil || got != 19 {
		t.Fatalf("got %v, %v", got, e)
	}
}

func TestEvalErrors(t *testing.T) {
	for _, expr := range []string{"1/0", "sqrt(-1)", "log(0)", "factorial(2.5)", "choose(2,3)", "min()", "wat(1)", "1e+", "1+"} {
		if _, e := Eval(expr, nil); e == nil {
			t.Errorf("Eval(%q) unexpectedly succeeded", expr)
		}
	}
}

func TestEvalRejectsNonFiniteInputsAndResults(t *testing.T) {
	if _, e := Eval("x", map[string]float64{"x": math.Inf(1)}); e == nil {
		t.Error("nonfinite variable accepted")
	}
	if _, e := Eval("1e308*1e308", nil); e == nil {
		t.Error("nonfinite result accepted")
	}
}

func TestEvalStressLimits(t *testing.T) {
	if _, err := Eval(strings.Repeat("(", maxDepth+1)+"1"+strings.Repeat(")", maxDepth+1), nil); err == nil {
		t.Fatal("deeply nested expression unexpectedly succeeded")
	}
	if _, err := Eval(strings.Repeat("1+", maxInput/2)+"1", nil); err == nil {
		t.Fatal("oversized expression unexpectedly succeeded")
	}
}

func BenchmarkEval(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := Eval("sqrt(144)+2^10+sin(pi/4)*mean(1,2,3)", nil); err != nil {
			b.Fatal(err)
		}
	}
}
