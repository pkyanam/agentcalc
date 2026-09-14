#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
CSV_PATH="${CSV_PATH:-$SCRIPT_DIR/../../tasks/sales.csv}"

python3 - "$CSV_PATH" <<'PY'
import csv
import json
import sys
from decimal import Decimal, getcontext

getcontext().prec = 50
path = sys.argv[1]
with open(path, newline="") as f:
    rows = list(csv.DictReader(f))

revenues = [Decimal(row["revenue"]) for row in rows]
ordered = sorted(revenues)
n = len(revenues)
total = sum(revenues, Decimal(0))
mean = total / Decimal(n)

def percentile(p):
    position = Decimal(n - 1) * Decimal(str(p))
    lower = int(position)
    upper = min(lower + 1, n - 1)
    fraction = position - Decimal(lower)
    return ordered[lower] + fraction * (ordered[upper] - ordered[lower])

variance = sum((value - mean) ** 2 for value in revenues) / Decimal(n - 1)
sample_stddev = variance.sqrt()
if n % 2:
    median = ordered[n // 2]
else:
    median = (ordered[n // 2 - 1] + ordered[n // 2]) / Decimal(2)

region_totals = {}
for row in rows:
    if row["status"] == "paid":
        region = row["region"]
        region_totals[region] = region_totals.get(region, Decimal(0)) + Decimal(row["revenue"])

paid = [row for row in rows if row["status"] == "paid"]
paid.sort(key=lambda row: (-Decimal(row["revenue"]), int(row["id"])))
units = sum(int(row["units"]) for row in rows)

def number(value):
    return float(value) if isinstance(value, Decimal) else value

result = {
    "revenue_stats": {
        "count": n,
        "sum": number(total),
        "mean": number(mean),
        "median": number(median),
        "sample_stddev": number(sample_stddev),
        "p25": number(percentile(0.25)),
        "p75": number(percentile(0.75)),
    },
    "region_totals": {region: number(value) for region, value in region_totals.items()},
    "largest_paid_ids": [int(row["id"]) for row in paid[:5]],
    "weighted_unit_price": number(total / Decimal(units)),
}
print(json.dumps(result, separators=(",", ":"), allow_nan=False))
PY
