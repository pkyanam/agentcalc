#!/usr/bin/env bash
set -euo pipefail
python3 - <<'PY'
import json
import math

A = [[10.0, 2.0, -1.0, 0.0],
     [2.0, 11.0, 3.0, -1.0],
     [-1.0, 3.0, 12.0, 2.0],
     [0.0, -1.0, 2.0, 9.0]]
b = [5.0, 25.0, -2.0, 19.0]

# Gaussian elimination with partial pivoting solves A*x=b and tracks det(A).
n = len(A)
m = [row[:] + [rhs] for row, rhs in zip(A, b)]
det = 1.0
for col in range(n):
    pivot = max(range(col, n), key=lambda r: abs(m[r][col]))
    if pivot != col:
        m[col], m[pivot] = m[pivot], m[col]
        det = -det
    p = m[col][col]
    det *= p
    for r in range(col + 1, n):
        q = m[r][col] / p
        for c in range(col, n + 1):
            m[r][c] -= q * m[col][c]
x = [0.0] * n
for r in range(n - 1, -1, -1):
    x[r] = (m[r][n] - sum(m[r][c] * x[c] for c in range(r + 1, n))) / m[r][r]

def bisect(f, lo, hi, iterations=220):
    flo = f(lo)
    for _ in range(iterations):
        mid = (lo + hi) / 2.0
        fm = f(mid)
        if fm == 0.0:
            return mid
        if flo * fm <= 0.0:
            hi = mid
        else:
            lo, flo = mid, fm
    return (lo + hi) / 2.0

root = bisect(lambda z: math.cos(z) - z, 0.0, 1.0)

def simpson(f, a, b):
    c = (a + b) / 2.0
    return (b - a) * (f(a) + 4.0*f(c) + f(b)) / 6.0

def adaptive(f, a, b, eps):
    whole = simpson(f, a, b)
    def rec(left, right, estimate, tol, depth):
        mid = (left + right) / 2.0
        lval = simpson(f, left, mid)
        rval = simpson(f, mid, right)
        if depth <= 0 or abs(lval + rval - estimate) <= 15.0 * tol:
            return lval + rval + (lval + rval - estimate) / 15.0
        return rec(left, mid, lval, tol/2.0, depth-1) + rec(mid, right, rval, tol/2.0, depth-1)
    return rec(a, b, whole, eps, 30)

integral = adaptive(lambda z: math.exp(-z*z), 0.0, 2.0, 1e-13)
derivative = math.exp(0.7) * (math.sin(0.7) + math.cos(0.7))
tiny_root = bisect(lambda z: 1e-20 * (z - 1.25), 0.0, 3.0)

print(json.dumps({"solution": x, "determinant": det, "root": root,
                  "integral": integral, "derivative": derivative,
                  "tiny_root": tiny_root}, separators=(',', ':')))
PY
