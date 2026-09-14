#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
csv_path="${script_dir}/../../tasks/sales.csv"

python3 - "$csv_path" <<'PY'
import csv
import json
import sys
from decimal import Decimal

path = sys.argv[1]
rows = []
with open(path, newline="") as f:
    for row in csv.DictReader(f):
        rows.append({
            "id": int(row["id"]),
            "region": row["region"],
            "status": row["status"],
            "units": Decimal(row["units"]),
            "revenue": Decimal(row["revenue"]),
        })

revenues = sorted(r["revenue"] for r in rows)
n = len(revenues)
total = sum(revenues, Decimal(0))
mean = total / n
if n % 2:
    median = revenues[n // 2]
else:
    median = (revenues[n // 2 - 1] + revenues[n // 2]) / 2

def percentile(p):
    index = Decimal(n - 1) * Decimal(str(p))
    lower = int(index)
    upper = min(lower + 1, n - 1)
    fraction = index - lower
    return revenues[lower] + (revenues[upper] - revenues[lower]) * fraction

variance = sum((x - mean) ** 2 for x in revenues) / (n - 1)
sample_stddev = variance.sqrt()

paid = [r for r in rows if r["status"] == "paid"]
region_totals = {}
for r in paid:
    region_totals[r["region"]] = region_totals.get(r["region"], Decimal(0)) + r["revenue"]
largest_paid_ids = [r["id"] for r in sorted(paid, key=lambda r: (-r["revenue"], r["id"]))[:5]]
weighted_unit_price = total / sum((r["units"] for r in rows), Decimal(0))

def number(value):
    return float(value)

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
    "region_totals": {k: number(v) for k, v in sorted(region_totals.items())},
    "largest_paid_ids": largest_paid_ids,
    "weighted_unit_price": number(weighted_unit_price),
}
print(json.dumps(result, separators=(",", ":"), allow_nan=False))
PY
