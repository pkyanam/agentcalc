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

# Gaussian elimination with partial pivoting, used both for the solve and det.
M = [row[:] + [rhs] for row, rhs in zip(A, b)]
det = 1.0
n = len(A)
for col in range(n):
    pivot = max(range(col, n), key=lambda r: abs(M[r][col]))
    if pivot != col:
        M[col], M[pivot] = M[pivot], M[col]
        det = -det
    p = M[col][col]
    det *= p
    for row in range(col + 1, n):
        q = M[row][col] / p
        for k in range(col, n + 1):
            M[row][k] -= q * M[col][k]
x = [0.0] * n
for row in range(n - 1, -1, -1):
    x[row] = (M[row][n] - sum(M[row][k] * x[k] for k in range(row + 1, n))) / M[row][row]

# Bisection gives a deterministic, certified bracket for cos(x)-x on [0, 1].
lo, hi = 0.0, 1.0
for _ in range(100):
    mid = (lo + hi) / 2.0
    if math.cos(mid) - mid > 0.0:
        lo = mid
    else:
        hi = mid
root = (lo + hi) / 2.0

integral = math.sqrt(math.pi) * math.erf(2.0) / 2.0
derivative = math.exp(0.7) * (math.sin(0.7) + math.cos(0.7))

# Newton's formula avoids loss of significance for the tiny supplied scale.
tiny_scale = 1e-20
tiny_offset = 1.25
tiny_guess = 0.0
tiny_root = tiny_guess - (tiny_scale * (tiny_guess - tiny_offset)) / tiny_scale

print(json.dumps({"solution": x, "determinant": det, "root": root,
                  "integral": integral, "derivative": derivative,
                  "tiny_root": tiny_root}, separators=(',', ':')))
PY
