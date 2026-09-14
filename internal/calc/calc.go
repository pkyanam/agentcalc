// Package calc evaluates small, safe mathematical expressions.
package calc

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const (
	maxInput = 1 << 20
	maxDepth = 128
	maxArgs  = 128
)

// Eval evaluates expression using vars as its named variables.
func Eval(expression string, vars map[string]float64) (float64, error) {
	if len(expression) > maxInput {
		return 0, fmt.Errorf("expression is too long")
	}
	p := parser{s: expression, vars: vars}
	p.next()
	v, err := p.add(0)
	if err != nil {
		return 0, err
	}
	if p.t.kind != tokEOF {
		return 0, p.errorf("unexpected %q", p.t.text)
	}
	if !finite(v) {
		return 0, fmt.Errorf("result is not finite")
	}
	return v, nil
}

type tokenKind uint8

const (
	tokEOF tokenKind = iota
	tokInvalid
	tokNumber
	tokIdent
	tokPlus
	tokMinus
	tokMul
	tokDiv
	tokMod
	tokPow
	tokLParen
	tokRParen
	tokComma
	tokFact
)

type token struct {
	kind  tokenKind
	text  string
	value float64
}

type parser struct {
	s     string
	pos   int
	t     token
	vars  map[string]float64
	depth int
}

func (p *parser) next() {
	for p.pos < len(p.s) {
		r := rune(p.s[p.pos])
		if !unicode.IsSpace(r) {
			break
		}
		p.pos++
	}
	if p.pos >= len(p.s) {
		p.t = token{kind: tokEOF}
		return
	}
	start := p.pos
	c := p.s[p.pos]
	switch c {
	case '+':
		p.pos++
		p.t = token{kind: tokPlus, text: "+"}
	case '-':
		p.pos++
		p.t = token{kind: tokMinus, text: "-"}
	case '*':
		p.pos++
		if p.pos < len(p.s) && p.s[p.pos] == '*' {
			p.pos++
			p.t = token{kind: tokPow, text: "**"}
		} else {
			p.t = token{kind: tokMul, text: "*"}
		}
	case '^':
		p.pos++
		p.t = token{kind: tokPow, text: "^"}
	case '/':
		p.pos++
		p.t = token{kind: tokDiv, text: "/"}
	case '%':
		p.pos++
		p.t = token{kind: tokMod, text: "%"}
	case '(':
		p.pos++
		p.t = token{kind: tokLParen, text: "("}
	case ')':
		p.pos++
		p.t = token{kind: tokRParen, text: ")"}
	case ',':
		p.pos++
		p.t = token{kind: tokComma, text: ","}
	case '!':
		p.pos++
		p.t = token{kind: tokFact, text: "!"}
	default:
		if (c >= '0' && c <= '9') || c == '.' {
			p.scanNumber()
			return
		}
		if unicode.IsLetter(rune(c)) || c == '_' {
			p.pos++
			for p.pos < len(p.s) {
				r := rune(p.s[p.pos])
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
					break
				}
				p.pos++
			}
			p.t = token{kind: tokIdent, text: p.s[start:p.pos]}
			return
		}
		p.pos++
		p.t = token{kind: tokInvalid, text: string(c)}
	}
}
func (p *parser) scanNumber() {
	start := p.pos
	if p.s[p.pos] == '.' {
		p.pos++
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			p.pos++
		}
	} else {
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			p.pos++
		}
		if p.pos < len(p.s) && p.s[p.pos] == '.' {
			p.pos++
			for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
				p.pos++
			}
		}
	}
	if p.pos < len(p.s) && (p.s[p.pos] == 'e' || p.s[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.s) && (p.s[p.pos] == '+' || p.s[p.pos] == '-') {
			p.pos++
		}
		e := p.pos
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			p.pos++
		}
		if e == p.pos {
			p.t = token{kind: tokInvalid, text: p.s[start:p.pos]}
			return
		}
	}
	x, err := strconv.ParseFloat(p.s[start:p.pos], 64)
	if err != nil {
		p.t = token{kind: tokInvalid, text: p.s[start:p.pos]}
		return
	}
	p.t = token{kind: tokNumber, text: p.s[start:p.pos], value: x}
}
func (p *parser) errorf(f string, a ...any) error {
	return fmt.Errorf("at position %d: %s", p.pos, fmt.Sprintf(f, a...))
}
func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func checked(x float64) (float64, error) {
	if !finite(x) {
		return 0, fmt.Errorf("operation produced a non-finite result")
	}
	return x, nil
}

func (p *parser) add(d int) (float64, error) {
	a, e := p.mul(d)
	if e != nil {
		return 0, e
	}
	for p.t.kind == tokPlus || p.t.kind == tokMinus {
		k := p.t.kind
		p.next()
		b, e := p.mul(d)
		if e != nil {
			return 0, e
		}
		if k == tokPlus {
			a += b
		} else {
			a -= b
		}
		if !finite(a) {
			return 0, fmt.Errorf("result is not finite")
		}
	}
	return a, nil
}
func (p *parser) mul(d int) (float64, error) {
	a, e := p.unary(d)
	if e != nil {
		return 0, e
	}
	for p.t.kind == tokMul || p.t.kind == tokDiv || p.t.kind == tokMod {
		k := p.t.kind
		p.next()
		b, e := p.unary(d)
		if e != nil {
			return 0, e
		}
		if (k == tokDiv || k == tokMod) && b == 0 {
			return 0, p.errorf("division by zero")
		}
		switch k {
		case tokMul:
			a *= b
		case tokDiv:
			a /= b
		case tokMod:
			a = math.Mod(a, b)
		}
		if !finite(a) {
			return 0, fmt.Errorf("result is not finite")
		}
	}
	return a, nil
}
func (p *parser) unary(d int) (float64, error) {
	if d > maxDepth {
		return 0, fmt.Errorf("expression nesting exceeds %d", maxDepth)
	}
	if p.t.kind == tokPlus || p.t.kind == tokMinus {
		k := p.t.kind
		p.next()
		a, e := p.unary(d + 1)
		if e != nil {
			return 0, e
		}
		if k == tokMinus {
			a = -a
		}
		return a, nil
	}
	return p.power(d)
}
func (p *parser) power(d int) (float64, error) {
	a, e := p.primary(d)
	if e != nil {
		return 0, e
	}
	if p.t.kind == tokPow {
		p.next()
		b, e := p.unary(d + 1)
		if e != nil {
			return 0, e
		}
		a, e = checked(math.Pow(a, b))
		if e != nil {
			return 0, e
		}
	}
	return a, nil
}
func (p *parser) primary(d int) (float64, error) {
	if d > maxDepth {
		return 0, fmt.Errorf("expression nesting exceeds %d", maxDepth)
	}
	var a float64
	switch p.t.kind {
	case tokNumber:
		a = p.t.value
		p.next()
	case tokIdent:
		name := p.t.text
		p.next()
		var e error
		if p.t.kind == tokLParen {
			a, e = p.call(name, d+1)
		} else {
			a, e = p.variable(name)
		}
		if e != nil {
			return 0, e
		}
	case tokLParen:
		p.next()
		var e error
		a, e = p.add(d + 1)
		if e != nil {
			return 0, e
		}
		if p.t.kind != tokRParen {
			return 0, p.errorf("expected closing parenthesis")
		}
		p.next()
	default:
		return 0, p.errorf("expected number, variable, or function")
	}
	for p.t.kind == tokFact {
		p.next()
		var e error
		a, e = factorial(a)
		if e != nil {
			return 0, e
		}
	}
	return checked(a)
}
func (p *parser) variable(name string) (float64, error) {
	switch strings.ToLower(name) {
	case "pi":
		return math.Pi, nil
	case "e":
		return math.E, nil
	case "tau":
		return 2 * math.Pi, nil
	}
	v, ok := p.vars[name]
	if !ok {
		return 0, p.errorf("unknown variable %q", name)
	}
	return checked(v)
}

func (p *parser) call(name string, d int) (float64, error) {
	p.next()
	args := make([]float64, 0, 4)
	if p.t.kind != tokRParen {
		for {
			if len(args) >= maxArgs {
				return 0, p.errorf("too many arguments")
			}
			v, e := p.add(d)
			if e != nil {
				return 0, e
			}
			args = append(args, v)
			if p.t.kind != tokComma {
				break
			}
			p.next()
		}
	}
	if p.t.kind != tokRParen {
		return 0, p.errorf("expected comma or closing parenthesis")
	}
	p.next()
	n := strings.ToLower(name)
	var v float64
	arity := func(min, max int) error {
		if len(args) < min || (max >= 0 && len(args) > max) {
			if max == min {
				return p.errorf("%s expects %d arguments", name, min)
			}
			return p.errorf("%s expects %d or more arguments", name, min)
		}
		return nil
	}
	switch n {
	case "sin":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Sin(args[0])
	case "cos":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Cos(args[0])
	case "tan":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Tan(args[0])
	case "asin":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Asin(args[0])
	case "acos":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Acos(args[0])
	case "atan":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Atan(args[0])
	case "sqrt":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Sqrt(args[0])
	case "cbrt":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Cbrt(args[0])
	case "abs":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Abs(args[0])
	case "ln", "log":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Log(args[0])
	case "log10":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Log10(args[0])
	case "log2":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Log2(args[0])
	case "exp":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Exp(args[0])
	case "floor":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Floor(args[0])
	case "ceil":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Ceil(args[0])
	case "round":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		v = math.Round(args[0])
	case "pow":
		if e := arity(2, 2); e != nil {
			return 0, e
		}
		v = math.Pow(args[0], args[1])
	case "hypot":
		if e := arity(2, 2); e != nil {
			return 0, e
		}
		v = math.Hypot(args[0], args[1])
	case "atan2":
		if e := arity(2, 2); e != nil {
			return 0, e
		}
		v = math.Atan2(args[0], args[1])
	case "clamp":
		if e := arity(3, 3); e != nil {
			return 0, e
		}
		if args[1] > args[2] {
			return 0, p.errorf("clamp lower bound exceeds upper bound")
		}
		v = math.Min(math.Max(args[0], args[1]), args[2])
	case "min", "max", "sum", "mean":
		if e := arity(1, -1); e != nil {
			return 0, e
		}
		v = args[0]
		if n == "sum" {
			v = 0
		}
		for i := range args {
			if n == "min" {
				v = math.Min(v, args[i])
			}
			if n == "max" {
				v = math.Max(v, args[i])
			}
			if n == "sum" {
				v += args[i]
			}
		}
		if n == "mean" {
			v = 0
			for _, x := range args {
				v += x
			}
			v /= float64(len(args))
		}
	case "factorial":
		if e := arity(1, 1); e != nil {
			return 0, e
		}
		vv, err := factorial(args[0])
		if err != nil {
			return 0, err
		}
		v = vv
	case "choose":
		if e := arity(2, 2); e != nil {
			return 0, e
		}
		vv, err := choose(args[0], args[1])
		if err != nil {
			return 0, err
		}
		v = vv
	default:
		return 0, p.errorf("unknown function %q", name)
	}
	return checked(v)
}

func integer(x float64) bool { return finite(x) && x >= 0 && x == math.Trunc(x) }
func factorial(x float64) (float64, error) {
	if !integer(x) || x > 170 {
		return 0, fmt.Errorf("factorial requires an integer from 0 to 170")
	}
	v := 1.0
	for i := 2.0; i <= x; i++ {
		v *= i
	}
	return v, nil
}
func choose(n, k float64) (float64, error) {
	if !integer(n) || !integer(k) || k > n || n > 170 {
		return 0, fmt.Errorf("choose requires integers with 0 <= k <= n <= 170")
	}
	if k > n-k {
		k = n - k
	}
	v := 1.0
	for i := 1.0; i <= k; i++ {
		v = v * (n - k + i) / i
	}
	return v, nil
}
