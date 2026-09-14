package calc

import (
	"fmt"
	"math"
)

const (
	numericTolerance = 1e-12
	rootIterations   = 200
	integralMaxEvals = 200000
	integralMaxDepth = 24
)

// Root finds a zero of expression on [lower, upper] using bisection. The
// expression is evaluated with x set to each trial point.
func Root(expression string, lower, upper float64) (float64, error) {
	if !finite(lower) || !finite(upper) {
		return 0, fmt.Errorf("root bounds must be finite")
	}
	if lower > upper {
		lower, upper = upper, lower
	}
	f := func(x float64) (float64, error) { return Eval(expression, map[string]float64{"x": x}) }
	fl, err := f(lower)
	if err != nil {
		return 0, fmt.Errorf("root at lower bound: %w", err)
	}
	if fl == 0 {
		return lower, nil
	}
	if lower == upper {
		return 0, fmt.Errorf("root is not bracketed")
	}
	fu, err := f(upper)
	if err != nil {
		return 0, fmt.Errorf("root at upper bound: %w", err)
	}
	if fu == 0 {
		return upper, nil
	}
	if math.Signbit(fl) == math.Signbit(fu) {
		return 0, fmt.Errorf("root is not bracketed: endpoint values have the same sign")
	}
	for i := 0; i < rootIterations; i++ {
		mid := midpoint(lower, upper)
		fm, err := f(mid)
		if err != nil {
			return 0, fmt.Errorf("root evaluation: %w", err)
		}
		if fm == 0 || halfWidth(lower, upper) <= numericTolerance*(1+math.Abs(mid)) {
			return mid, nil
		}
		if math.Signbit(fl) != math.Signbit(fm) {
			upper, fu = mid, fm
		} else {
			lower, fl = mid, fm
		}
	}
	return 0, fmt.Errorf("root did not converge after %d iterations", rootIterations)
}

// Integrate estimates the definite integral of expression on [lower, upper]
// using adaptive Simpson quadrature. The expression is evaluated with x as
// its variable. It returns an error when the finite evaluation or depth limits
// are reached before the requested tolerance is met.
func Integrate(expression string, lower, upper float64) (float64, error) {
	if !finite(lower) || !finite(upper) {
		return 0, fmt.Errorf("integration bounds must be finite")
	}
	if lower == upper {
		return 0, nil
	}
	sign := 1.0
	if lower > upper {
		lower, upper = upper, lower
		sign = -1
	}
	if !finite(upper - lower) {
		return 0, fmt.Errorf("integration bounds span is too large")
	}
	n := &integrator{expression: expression}
	mid := midpoint(lower, upper)
	fa, err := n.eval(lower)
	if err != nil {
		return 0, fmt.Errorf("integrand at lower bound: %w", err)
	}
	fm, err := n.eval(mid)
	if err != nil {
		return 0, fmt.Errorf("integrand at midpoint: %w", err)
	}
	fb, err := n.eval(upper)
	if err != nil {
		return 0, fmt.Errorf("integrand at upper bound: %w", err)
	}
	whole := simpson(lower, upper, fa, fm, fb)
	if !finite(whole) {
		return 0, fmt.Errorf("integration produced a non-finite intermediate result")
	}
	tol := numericTolerance * (1 + math.Abs(whole))
	value, err := n.adapt(lower, upper, fa, fm, fb, whole, tol, 0)
	if err != nil {
		return 0, err
	}
	return sign * value, nil
}

type integrator struct {
	expression string
	evals      int
}

func (n *integrator) eval(x float64) (float64, error) {
	if n.evals >= integralMaxEvals {
		return 0, fmt.Errorf("integration did not converge: evaluation limit reached")
	}
	n.evals++
	return Eval(n.expression, map[string]float64{"x": x})
}

func simpson(a, b, fa, fm, fb float64) float64 {
	return (b - a) * (fa + 4*fm + fb) / 6
}

func midpoint(a, b float64) float64 { return a/2 + b/2 }

func halfWidth(a, b float64) float64 { return math.Abs(b/2 - a/2) }

func (n *integrator) adapt(a, b, fa, fm, fb, whole, tol float64, depth int) (float64, error) {
	mid := midpoint(a, b)
	leftMid := midpoint(a, mid)
	rightMid := midpoint(mid, b)
	fl, err := n.eval(leftMid)
	if err != nil {
		return 0, err
	}
	fr, err := n.eval(rightMid)
	if err != nil {
		return 0, err
	}
	left := simpson(a, mid, fa, fl, fm)
	right := simpson(mid, b, fm, fr, fb)
	if !finite(left) || !finite(right) {
		return 0, fmt.Errorf("integration produced a non-finite intermediate result")
	}
	delta := left + right - whole
	if !finite(delta) {
		return 0, fmt.Errorf("integration produced a non-finite intermediate result")
	}
	if math.Abs(delta) <= 15*tol {
		result := left + right + delta/15
		if !finite(result) {
			return 0, fmt.Errorf("integration produced a non-finite result")
		}
		return result, nil
	}
	if depth >= integralMaxDepth {
		return 0, fmt.Errorf("integration did not converge: maximum recursion depth reached")
	}
	vl, err := n.adapt(a, mid, fa, fl, fm, left, tol/2, depth+1)
	if err != nil {
		return 0, err
	}
	vr, err := n.adapt(mid, b, fm, fr, fb, right, tol/2, depth+1)
	if err != nil {
		return 0, err
	}
	result := vl + vr
	if !finite(result) {
		return 0, fmt.Errorf("integration produced a non-finite result")
	}
	return result, nil
}

// Derivative estimates the first derivative of expression at x using a
// central finite difference. It is an approximation and is subject to the
// expression's numerical precision and domain.
func Derivative(expression string, x float64) (float64, error) {
	if !finite(x) {
		return 0, fmt.Errorf("derivative point must be finite")
	}
	h := math.Cbrt(2.220446049250313e-16) * math.Max(1, math.Abs(x))
	if x-h == x || x+h == x {
		return 0, fmt.Errorf("derivative step is below floating-point resolution")
	}
	fp, err := Eval(expression, map[string]float64{"x": x + h})
	if err != nil {
		return 0, fmt.Errorf("derivative at x+h: %w", err)
	}
	fm, err := Eval(expression, map[string]float64{"x": x - h})
	if err != nil {
		return 0, fmt.Errorf("derivative at x-h: %w", err)
	}
	d := (fp - fm) / (2 * h)
	if !finite(d) {
		return 0, fmt.Errorf("derivative result is not finite")
	}
	return d, nil
}
