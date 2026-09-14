package calc

import (
	"math"
	"strings"
	"testing"
)

func TestRoot(t *testing.T) {
	x, err := Root("x^2-2", 0, 2)
	if err != nil || math.Abs(x-math.Sqrt2) > 1e-10 {
		t.Fatalf("Root=%v, %v; want sqrt(2)", x, err)
	}
	x, err = Root("1e-20*(x-1)", 0, 2)
	if err != nil || math.Abs(x-1) > 1e-10 {
		t.Fatalf("scaled Root=%v, %v; want 1", x, err)
	}
	if _, err = Root("1e-20", 0, 2); err == nil {
		t.Error("Root accepted a tiny constant as a zero")
	}
}

func TestIntegrate(t *testing.T) {
	v, err := Integrate("sin(x)", 0, math.Pi)
	if err != nil || math.Abs(v-2) > 1e-10 {
		t.Fatalf("Integrate=%v, %v; want 2", v, err)
	}
}

func TestDerivative(t *testing.T) {
	v, err := Derivative("x^3", 2)
	if err != nil || math.Abs(v-12) > 1e-8 {
		t.Fatalf("Derivative=%v, %v; want 12", v, err)
	}
}

func TestNumericalErrors(t *testing.T) {
	if _, err := Root("x^2+1", -1, 1); err == nil {
		t.Error("Root accepted an unbracketed interval")
	}
	if _, err := Integrate("1/x", -1, 1); err == nil {
		t.Error("Integrate accepted a singular integrand")
	}
	if _, err := Integrate("1e308", 0, 2); err == nil {
		t.Error("Integrate accepted an overflowing result")
	}
	if _, err := Integrate("1", -math.MaxFloat64, math.MaxFloat64); err == nil {
		t.Error("Integrate accepted an overflowing interval width")
	}
	if _, err := Derivative("sqrt(x)", 0); err == nil || !strings.Contains(err.Error(), "x-h") {
		t.Errorf("Derivative singular point error=%v", err)
	}
}
