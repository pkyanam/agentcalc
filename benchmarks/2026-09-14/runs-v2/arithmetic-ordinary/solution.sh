#!/usr/bin/env bash
set -euo pipefail
python3 - <<'PY'
import json
import math
from fractions import Fraction

principal = 18750.0
nominal_rate = 0.063
months = 37
growth = principal * (1.0 + nominal_rate / 12.0) ** months

exact_sum = Fraction("123456789.123456789") + Fraction("0.000000011")
combinations = math.comb(52, 7)
gib_bytes = int(3.75 * (2 ** 30))
fahrenheit = -17.3
fahrenheit_celsius = (fahrenheit - 32.0) * 5.0 / 9.0
trig = math.sin(math.pi / 7.0) ** 2 + math.cos(math.pi / 7.0) ** 2 + math.sqrt(2025.0)

result = {
    "growth": growth,
    "exact_sum": str(exact_sum.numerator) if exact_sum.denominator == 1 else f"{exact_sum.numerator}/{exact_sum.denominator}",
    "combinations": combinations,
    "gib_bytes": gib_bytes,
    "fahrenheit_celsius": fahrenheit_celsius,
    "trig": trig,
}
print(json.dumps(result, separators=(",", ":"), allow_nan=False))
PY
