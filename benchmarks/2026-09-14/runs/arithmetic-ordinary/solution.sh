#!/usr/bin/env bash
set -euo pipefail

TASK_FILE="$(cd "$(dirname "$0")/../../tasks" && pwd)/arithmetic.txt"
python3 - "$TASK_FILE" <<'PY'
import json
import math
import re
import sys
from decimal import Decimal
from fractions import Fraction

text = open(sys.argv[1], encoding="utf-8").read()

def number(pattern):
    match = re.search(pattern, text)
    if not match:
        raise SystemExit(f"missing task input: {pattern}")
    return match.group(1)

principal = Decimal(number(r"growth:\s*([0-9]+(?:\.[0-9]+)?)"))
annual_rate = Decimal(number(r"nominal annual rate\s*([0-9]+(?:\.[0-9]+)?)%")) / 100
months = int(number(r"for\s+([0-9]+)\s+months"))
growth = float(principal) * (1.0 + float(annual_rate) / 12.0) ** months

sum_match = re.search(
    r"exact_sum:.*?for\s+([0-9]+\.[0-9]+)\s*\+\s*([0-9]+\.[0-9]+)",
    text,
    re.S,
)
if not sum_match:
    raise SystemExit("missing exact sum inputs")
exact_sum = Fraction(sum_match.group(1)) + Fraction(sum_match.group(2))

choose_k = int(number(r"choose\s+([0-9]+)\s+items\s+from"))
choose_n = int(number(r"items\s+from\s+([0-9]+)"))
gib = Decimal(number(r"gib_bytes: convert\s+([0-9]+(?:\.[0-9]+)?)\s+GiB"))
fahrenheit = float(number(r"fahrenheit_celsius: convert\s+(-?[0-9]+(?:\.[0-9]+)?)\s+degrees Fahrenheit"))

result = {
    "growth": growth,
    "exact_sum": str(exact_sum.numerator) if exact_sum.denominator == 1 else f"{exact_sum.numerator}/{exact_sum.denominator}",
    "combinations": math.comb(choose_n, choose_k),
    "gib_bytes": int(gib * (2 ** 30)),
    "fahrenheit_celsius": (fahrenheit - 32.0) * 5.0 / 9.0,
    "trig": math.sin(math.pi / 7.0) ** 2 + math.cos(math.pi / 7.0) ** 2 + math.sqrt(2025.0),
}
print(json.dumps(result, separators=(",", ":"), allow_nan=False))
PY
