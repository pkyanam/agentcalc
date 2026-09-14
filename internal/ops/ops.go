// Package ops contains the dependency-free operations exposed by agentcalc.
package ops

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Stats returns descriptive statistics for values. Percentiles use linear
// interpolation between adjacent sorted observations.
func Stats(values []float64) (map[string]any, error) {
	if len(values) == 0 {
		return nil, errors.New("stats: empty input")
	}
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("stats: values must be finite")
		}
	}
	x := append([]float64(nil), values...)
	sort.Float64s(x)
	n := float64(len(x))
	sum := 0.0
	compensation := 0.0
	for _, v := range x {
		// Kahan summation keeps small observations from disappearing when the
		// input has a wide range of magnitudes.
		y := v - compensation
		t := sum + y
		compensation = (t - sum) - y
		sum = t
	}
	if !finite(sum) {
		return nil, errors.New("stats: sum is not finite")
	}
	mean := sum / n
	if !finite(mean) {
		return nil, errors.New("stats: mean is not finite")
	}
	ss := 0.0
	for _, v := range x {
		d := v - mean
		ss += d * d
	}
	variance := ss / n
	if !finite(variance) {
		return nil, errors.New("stats: variance is not finite")
	}
	r := map[string]any{
		"count": len(x), "sum": sum, "mean": mean,
		"median": percentile(x, .5), "min": x[0], "max": x[len(x)-1],
		"population_variance": variance, "population_stddev": math.Sqrt(variance),
		// Short names are retained for callers that do not need to distinguish
		// population from sample estimates.
		"variance": variance, "stddev": math.Sqrt(variance),
		"p25": percentile(x, .25), "p75": percentile(x, .75),
	}
	if len(x) > 1 {
		sv := ss / (n - 1)
		if !finite(sv) {
			return nil, errors.New("stats: sample variance is not finite")
		}
		r["sample_variance"], r["sample_stddev"] = sv, math.Sqrt(sv)
	}
	return r, nil
}

func percentile(x []float64, p float64) float64 {
	if len(x) == 1 {
		return x[0]
	}
	pos := p * float64(len(x)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	f := pos - float64(lo)
	// Avoid overflow in x[hi]-x[lo] for opposite-signed extreme values.
	d := x[hi] - x[lo]
	if math.IsInf(d, 0) {
		return x[lo]*(1-f) + x[hi]*f
	}
	return x[lo] + d*f
}

// Units lists accepted unit names, grouped by dimension. Names are
// case-insensitive; Convert also accepts common spelled-out aliases.
func Units() map[string][]string {
	return map[string][]string{
		"length":      {"m", "km", "cm", "mm", "mi", "yd", "ft", "in"},
		"mass":        {"kg", "g", "mg", "lb", "oz"},
		"time":        {"s", "ms", "us", "ns", "min", "h", "day"},
		"temperature": {"c", "f", "k"}, "bytes": {"b", "kb", "mb", "gb", "tb", "kib", "mib", "gib", "tib"},
		"angle": {"rad", "deg", "grad", "turn"}, "speed": {"m/s", "km/h", "mph", "ft/s", "knot"},
		"area":   {"m2", "km2", "cm2", "ft2", "in2", "acre", "hectare"},
		"volume": {"m3", "l", "ml", "cm3", "ft3", "in3", "gal"},
	}
}

type unitDef struct {
	dim            string
	factor, offset float64
}

var units = map[string]unitDef{
	"m": {"length", 1, 0}, "km": {"length", 1000, 0}, "cm": {"length", .01, 0}, "mm": {"length", .001, 0}, "mi": {"length", 1609.344, 0}, "yd": {"length", .9144, 0}, "ft": {"length", .3048, 0}, "in": {"length", .0254, 0},
	"kg": {"mass", 1, 0}, "g": {"mass", .001, 0}, "mg": {"mass", 1e-6, 0}, "lb": {"mass", .45359237, 0}, "oz": {"mass", .028349523125, 0},
	"s": {"time", 1, 0}, "ms": {"time", 1e-3, 0}, "us": {"time", 1e-6, 0}, "ns": {"time", 1e-9, 0}, "min": {"time", 60, 0}, "h": {"time", 3600, 0}, "day": {"time", 86400, 0},
	"c": {"temperature", 1, 0}, "f": {"temperature", 5.0 / 9, -32}, "k": {"temperature", 1, -273.15},
	"b": {"bytes", 1, 0}, "kb": {"bytes", 1e3, 0}, "mb": {"bytes", 1e6, 0}, "gb": {"bytes", 1e9, 0}, "tb": {"bytes", 1e12, 0}, "kib": {"bytes", 1024, 0}, "mib": {"bytes", 1024 * 1024, 0}, "gib": {"bytes", 1024 * 1024 * 1024, 0}, "tib": {"bytes", 1024 * 1024 * 1024 * 1024, 0},
	"rad": {"angle", 1, 0}, "deg": {"angle", math.Pi / 180, 0}, "grad": {"angle", math.Pi / 200, 0}, "turn": {"angle", 2 * math.Pi, 0},
	"m/s": {"speed", 1, 0}, "km/h": {"speed", 1000.0 / 3600, 0}, "mph": {"speed", 1609.344 / 3600, 0}, "ft/s": {"speed", .3048, 0}, "knot": {"speed", 1852.0 / 3600, 0},
	"m2": {"area", 1, 0}, "km2": {"area", 1e6, 0}, "cm2": {"area", 1e-4, 0}, "ft2": {"area", .09290304, 0}, "in2": {"area", .00064516, 0}, "acre": {"area", 4046.8564224, 0}, "hectare": {"area", 10000, 0},
	"m3": {"volume", 1, 0}, "l": {"volume", .001, 0}, "ml": {"volume", 1e-6, 0}, "cm3": {"volume", 1e-6, 0}, "ft3": {"volume", .028316846592, 0}, "in3": {"volume", 1.6387064e-5, 0}, "gal": {"volume", .003785411784, 0},
}

func init() {
	aliases := map[string]string{
		"meter": "m", "meters": "m", "kilometer": "km", "kilometers": "km", "centimeter": "cm", "centimeters": "cm", "millimeter": "mm", "millimeters": "mm", "mile": "mi", "miles": "mi", "yard": "yd", "yards": "yd", "foot": "ft", "feet": "ft", "inch": "in", "inches": "in",
		"kilogram": "kg", "kilograms": "kg", "gram": "g", "grams": "g", "milligram": "mg", "milligrams": "mg", "pound": "lb", "pounds": "lb", "ounce": "oz", "ounces": "oz",
		"second": "s", "seconds": "s", "millisecond": "ms", "milliseconds": "ms", "microsecond": "us", "microseconds": "us", "nanosecond": "ns", "nanoseconds": "ns", "minute": "min", "minutes": "min", "hour": "h", "hours": "h", "d": "day", "days": "day",
		"celsius": "c", "fahrenheit": "f", "kelvin": "k", "byte": "b", "bytes": "b",
		"radian": "rad", "radians": "rad", "degree": "deg", "degrees": "deg", "revolution": "turn", "revolutions": "turn", "turns": "turn", "knot": "knot", "knots": "knot",
		"liter": "l", "liters": "l", "litre": "l", "litres": "l", "gallon": "gal", "gallons": "gal", "hectares": "hectare", "acre": "acre", "acres": "acre",
	}
	for alias, canonical := range aliases {
		units[alias] = units[canonical]
	}
}

func normalizeUnit(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "²", "2")
	s = strings.ReplaceAll(s, "³", "3")
	return s
}

// Convert converts value between units in the same dimension.
func Convert(value float64, from, to string) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, errors.New("convert: value must be finite")
	}
	f, fok := units[normalizeUnit(from)]
	t, tok := units[normalizeUnit(to)]
	if !fok || !tok {
		return 0, fmt.Errorf("convert: unknown unit %q or %q", from, to)
	}
	if f.dim != t.dim {
		return 0, fmt.Errorf("convert: incompatible dimensions %s and %s", f.dim, t.dim)
	}
	base := (value + f.offset) * f.factor
	out := base/t.factor - t.offset
	if math.IsInf(out, 0) || math.IsNaN(out) {
		return 0, errors.New("convert: result is not finite")
	}
	return out, nil
}

func validMatrix(m [][]float64) (int, int, error) {
	const maxMatrixDimension = 256
	if len(m) == 0 {
		return 0, 0, errors.New("matrix: empty matrix")
	}
	c := len(m[0])
	if len(m) > maxMatrixDimension || c > maxMatrixDimension {
		return 0, 0, fmt.Errorf("matrix: dimensions exceed %d", maxMatrixDimension)
	}
	if c == 0 {
		return 0, 0, errors.New("matrix: empty row")
	}
	for _, row := range m {
		if len(row) != c {
			return 0, 0, errors.New("matrix: rows have different lengths")
		}
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return 0, 0, errors.New("matrix: entries must be finite")
			}
		}
	}
	return len(m), c, nil
}
func clone(m [][]float64) [][]float64 {
	r := make([][]float64, len(m))
	for i := range m {
		r[i] = append([]float64(nil), m[i]...)
	}
	return r
}

// Matrix performs add, subtract, multiply, transpose, determinant, inverse,
// and solve. solve treats b as an n-by-1 column and returns []float64.
func Matrix(op string, a, b [][]float64) (any, error) {
	result, err := matrix(op, a, b)
	if err != nil {
		return nil, err
	}
	valid := true
	switch v := result.(type) {
	case float64:
		valid = finite(v)
	case []float64:
		for _, x := range v {
			valid = valid && finite(x)
		}
	case [][]float64:
		for _, row := range v {
			for _, x := range row {
				valid = valid && finite(x)
			}
		}
	}
	if !valid {
		return nil, errors.New("matrix: result is not finite")
	}
	return result, nil
}

func matrix(op string, a, b [][]float64) (any, error) {
	op = strings.ToLower(strings.TrimSpace(op))
	if op == "sub" {
		op = "subtract"
	}
	if op == "mul" {
		op = "multiply"
	}
	if op == "det" {
		op = "determinant"
	}
	if op == "inv" {
		op = "inverse"
	}
	ar, ac, e := validMatrix(a)
	if e != nil {
		return nil, e
	}
	switch op {
	case "transpose":
		r := make([][]float64, ac)
		for j := range r {
			r[j] = make([]float64, ar)
			for i := range r[j] {
				r[j][i] = a[i][j]
			}
		}
		return r, nil
	case "add", "subtract":
		br, bc, e := validMatrix(b)
		if e != nil {
			return nil, e
		}
		if ar != br || ac != bc {
			return nil, errors.New("matrix: dimensions do not match")
		}
		r := make([][]float64, ar)
		for i := range r {
			r[i] = make([]float64, ac)
			for j := range r[i] {
				r[i][j] = a[i][j]
				if op == "add" {
					r[i][j] += b[i][j]
				} else {
					r[i][j] -= b[i][j]
				}
				if math.IsInf(r[i][j], 0) || math.IsNaN(r[i][j]) {
					return nil, errors.New("matrix: result is not finite")
				}
			}
		}
		return r, nil
	case "multiply":
		br, bc, e := validMatrix(b)
		if e != nil {
			return nil, e
		}
		if ac != br {
			return nil, errors.New("matrix: incompatible multiplication dimensions")
		}
		r := make([][]float64, ar)
		for i := range r {
			r[i] = make([]float64, bc)
			for j := range r[i] {
				for k := 0; k < ac; k++ {
					r[i][j] += a[i][k] * b[k][j]
				}
				if math.IsInf(r[i][j], 0) || math.IsNaN(r[i][j]) {
					return nil, errors.New("matrix: result is not finite")
				}
			}
		}
		return r, nil
	case "determinant":
		if ar != ac {
			return nil, errors.New("matrix: determinant requires square matrix")
		}
		return determinant(a)
	case "inverse":
		if ar != ac {
			return nil, errors.New("matrix: inverse requires square matrix")
		}
		return inverse(a)
	case "solve":
		br, bc, e := validMatrix(b)
		if e != nil {
			return nil, e
		}
		if ar != ac || br != ar || bc != 1 {
			return nil, errors.New("matrix: solve requires square a and n-by-1 b")
		}
		return solve(a, b)
	default:
		return nil, fmt.Errorf("matrix: unknown operation %q", op)
	}
}

func determinant(a [][]float64) (float64, error) {
	m := clone(a)
	n := len(m)
	d := 1.0
	for k := 0; k < n; k++ {
		p := k
		for i := k + 1; i < n; i++ {
			if math.Abs(m[i][k]) > math.Abs(m[p][k]) {
				p = i
			}
		}
		if m[p][k] == 0 {
			return 0, nil
		}
		if p != k {
			m[p], m[k] = m[k], m[p]
			d = -d
		}
		pivot := m[k][k]
		d *= pivot
		for i := k + 1; i < n; i++ {
			q := m[i][k] / pivot
			for j := k + 1; j < n; j++ {
				m[i][j] -= q * m[k][j]
			}
		}
	}
	return d, nil
}
func inverse(a [][]float64) ([][]float64, error) {
	n := len(a)
	m := make([][]float64, n)
	for i := range m {
		m[i] = make([]float64, 2*n)
		copy(m[i], a[i])
		m[i][n+i] = 1
	}
	for k := 0; k < n; k++ {
		p := k
		for i := k + 1; i < n; i++ {
			if math.Abs(m[i][k]) > math.Abs(m[p][k]) {
				p = i
			}
		}
		if m[p][k] == 0 {
			return nil, errors.New("matrix: singular matrix")
		}
		m[p], m[k] = m[k], m[p]
		q := m[k][k]
		for j := 0; j < 2*n; j++ {
			m[k][j] /= q
		}
		for i := 0; i < n; i++ {
			if i == k {
				continue
			}
			q = m[i][k]
			for j := 0; j < 2*n; j++ {
				m[i][j] -= q * m[k][j]
			}
		}
	}
	r := make([][]float64, n)
	for i := range r {
		r[i] = append([]float64(nil), m[i][n:]...)
	}
	return r, nil
}
func solve(a, b [][]float64) ([]float64, error) {
	inv, e := inverse(a)
	if e != nil {
		return nil, e
	}
	r := make([]float64, len(a))
	for i := range r {
		for j := range r {
			r[i] += inv[i][j] * b[j][0]
		}
	}
	return r, nil
}
