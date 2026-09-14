#!/usr/bin/env bash
set -euo pipefail

python3 - <<'PY'
import json
import math
from fractions import Fraction

growth = 18750 * (1 + 0.063 / 12) ** 37
exact_sum = Fraction("123456789.123456789") + Fraction("0.000000011")
result = {
    "growth": growth,
    "exact_sum": str(exact_sum.numerator) if exact_sum.denominator == 1 else f"{exact_sum.numerator}/{exact_sum.denominator}",
    "combinations": math.comb(52, 7),
    "gib_bytes": int(3.75 * (2 ** 30)),
    "fahrenheit_celsius": (-17.3 - 32) * 5 / 9,
    "trig": math.sin(math.pi / 7) ** 2 + math.cos(math.pi / 7) ** 2 + math.sqrt(2025),
}
print(json.dumps(result, separators=(",", ":")))
PY
