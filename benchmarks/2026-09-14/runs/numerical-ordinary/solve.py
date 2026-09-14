#!/usr/bin/env python3
import json
import math


def eliminate(a, b):
    """Gaussian elimination with partial pivoting, returning the solution."""
    n = len(a)
    m = [list(map(float, row)) + [float(rhs)] for row, rhs in zip(a, b)]
    for k in range(n):
        p = max(range(k, n), key=lambda i: abs(m[i][k]))
        m[k], m[p] = m[p], m[k]
        pivot = m[k][k]
        for i in range(k + 1, n):
            q = m[i][k] / pivot
            for j in range(k, n + 1):
                m[i][j] -= q * m[k][j]
    x = [0.0] * n
    for i in range(n - 1, -1, -1):
        x[i] = (m[i][n] - sum(m[i][j] * x[j] for j in range(i + 1, n))) / m[i][i]
    return x


def determinant(a):
    n = len(a)
    m = [list(map(float, row)) for row in a]
    sign = 1.0
    d = 1.0
    for k in range(n):
        p = max(range(k, n), key=lambda i: abs(m[i][k]))
        if p != k:
            m[k], m[p] = m[p], m[k]
            sign = -sign
        pivot = m[k][k]
        d *= pivot
        for i in range(k + 1, n):
            q = m[i][k] / pivot
            for j in range(k + 1, n):
                m[i][j] -= q * m[k][j]
    return sign * d


def bisect(f, lo, hi, tol=1e-14):
    flo = f(lo)
    fhi = f(hi)
    if flo == 0.0:
        return lo
    if fhi == 0.0:
        return hi
    if flo * fhi > 0.0:
        raise ValueError("interval does not bracket a root")
    for _ in range(300):
        mid = (lo + hi) / 2.0
        fm = f(mid)
        if fm == 0.0 or (hi - lo) <= tol:
            return mid
        if flo * fm <= 0.0:
            hi = mid
            fhi = fm
        else:
            lo = mid
            flo = fm
    return (lo + hi) / 2.0


def adaptive_simpson(f, a, b, eps=1e-13):
    def simp(x0, x1, fx0, fxm, fx1):
        return (x1 - x0) * (fx0 + 4.0 * fxm + fx1) / 6.0

    fa, fb = f(a), f(b)
    mid = (a + b) / 2.0
    fm = f(mid)
    whole = simp(a, b, fa, fm, fb)

    def rec(x0, x1, fx0, fxm, fx1, estimate, budget, depth):
        xm = (x0 + x1) / 2.0
        leftm, rightm = (x0 + xm) / 2.0, (xm + x1) / 2.0
        flm, frm = f(leftm), f(rightm)
        left = simp(x0, xm, fx0, flm, fxm)
        right = simp(xm, x1, fxm, frm, fx1)
        if depth <= 0 or abs(left + right - estimate) <= 15.0 * budget:
            return left + right + (left + right - estimate) / 15.0
        return rec(x0, xm, fx0, flm, fxm, left, budget / 2.0, depth - 1) + rec(xm, x1, fxm, frm, fx1, right, budget / 2.0, depth - 1)

    return rec(a, b, fa, fm, fb, whole, eps, 30)


def main():
    A = [[10, 2, -1, 0], [2, 11, 3, -1], [-1, 3, 12, 2], [0, -1, 2, 9]]
    b = [5, 25, -2, 19]
    result = {
        "solution": eliminate(A, b),
        "determinant": determinant(A),
        "root": bisect(lambda x: math.cos(x) - x, 0.0, 1.0),
        "integral": adaptive_simpson(lambda x: math.exp(-x * x), 0.0, 2.0),
        "derivative": math.exp(0.7) * (math.sin(0.7) + math.cos(0.7)),
        "tiny_root": bisect(lambda x: 1e-20 * (x - 1.25), 0.0, 3.0),
    }
    print(json.dumps(result, separators=(",", ":"), allow_nan=False))


if __name__ == "__main__":
    main()
