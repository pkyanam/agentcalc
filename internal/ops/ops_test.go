package ops

import (
	"math"
	"testing"
)

func TestStats(t *testing.T) {
	s, err := Stats([]float64{4, 1, 3, 2})
	if err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]float64{"count": 4, "sum": 10, "mean": 2.5, "median": 2.5, "p25": 1.75, "p75": 3.25, "population_variance": 1.25, "sample_variance": 5.0 / 3} {
		got, ok := s[k].(float64)
		if k == "count" {
			if int(s[k].(int)) != int(want) {
				t.Errorf("%s", k)
			}
			continue
		}
		if !ok || math.Abs(got-want) > 1e-12 {
			t.Errorf("%s: got %v want %v", k, got, want)
		}
	}
	if _, ok := s["sample_stddev"]; !ok {
		t.Error("missing sample stddev")
	}
	if _, err := Stats([]float64{math.NaN()}); err == nil {
		t.Error("NaN accepted")
	}
	if _, err := Stats([]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}); err == nil {
		t.Error("overflowing stats accepted")
	}
}

func TestConvert(t *testing.T) {
	got, err := Convert(0, "C", "F")
	if err != nil || math.Abs(got-32) > 1e-12 {
		t.Fatalf("C/F: %v %v", got, err)
	}
	got, err = Convert(1, "km", "m")
	if err != nil || got != 1000 {
		t.Fatalf("km/m: %v %v", got, err)
	}
	if _, err = Convert(1, "kg", "m"); err == nil {
		t.Error("incompatible units accepted")
	}
	if _, err = Convert(1, "bogus", "m"); err == nil {
		t.Error("unknown unit accepted")
	}
}

func TestMatrix(t *testing.T) {
	a := [][]float64{{1, 2}, {3, 4}}
	v, err := Matrix("multiply", a, [][]float64{{2, 0}, {1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	got := v.([][]float64)
	if got[0][0] != 4 || got[1][1] != 8 {
		t.Fatalf("multiply: %#v", got)
	}
	d, err := Matrix("determinant", a, nil)
	if err != nil || d.(float64) != -2 {
		t.Fatalf("determinant: %v %v", d, err)
	}
	inv, err := Matrix("inverse", a, nil)
	if err != nil {
		t.Fatal(err)
	}
	im := inv.([][]float64)
	if math.Abs(im[0][0]+2) > 1e-12 || math.Abs(im[1][1]+.5) > 1e-12 {
		t.Fatalf("inverse: %#v", im)
	}
	sol, err := Matrix("solve", a, [][]float64{{5}, {11}})
	if err != nil {
		t.Fatal(err)
	}
	x := sol.([]float64)
	if math.Abs(x[0]-1) > 1e-12 || math.Abs(x[1]-2) > 1e-12 {
		t.Fatalf("solve: %v", x)
	}
	if _, err := Matrix("inverse", [][]float64{{1, 2}, {2, 4}}, nil); err == nil {
		t.Error("singular inverse accepted")
	}
	if _, err := Matrix("inverse", [][]float64{{1e-13}}, nil); err != nil {
		t.Errorf("tiny nonzero pivot rejected: %v", err)
	}
	if _, err := Matrix("multiply", [][]float64{{math.MaxFloat64}}, [][]float64{{2}}); err == nil {
		t.Error("overflowing matrix result accepted")
	}
}

func TestAllMatrixOutputsFinite(t *testing.T) {
	for _, c := range []struct {
		op   string
		a, b [][]float64
	}{
		{"determinant", [][]float64{{1e308, 0}, {0, 1e308}}, nil},
		{"inverse", [][]float64{{1e-320}}, nil},
		{"solve", [][]float64{{1e-300}}, [][]float64{{1e300}}},
	} {
		if v, e := Matrix(c.op, c.a, c.b); e == nil {
			t.Errorf("%s returned nonfinite result %v without error", c.op, v)
		}
	}
}
