import csv
import io
import json
import math


def main(data):
    rows = []
    for row in csv.DictReader(io.StringIO(data)):
        rows.append({
            "id": int(row["id"]),
            "region": row["region"],
            "status": row["status"],
            "units": int(row["units"]),
            "revenue": float(row["revenue"]),
        })

    revenues = sorted(row["revenue"] for row in rows)
    n = len(revenues)
    total = sum(revenues)
    mean = total / n

    def percentile(p):
        pos = (n - 1) * p
        lo = int(math.floor(pos))
        hi = int(math.ceil(pos))
        if lo == hi:
            return revenues[lo]
        return revenues[lo] + (revenues[hi] - revenues[lo]) * (pos - lo)

    median = percentile(0.5)
    sample_stddev = math.sqrt(sum((x - mean) ** 2 for x in revenues) / (n - 1))
    revenue_stats = {
        "count": n,
        "sum": total,
        "mean": mean,
        "median": median,
        "sample_stddev": sample_stddev,
        "p25": percentile(0.25),
        "p75": percentile(0.75),
    }

    region_totals = {}
    paid = [row for row in rows if row["status"] == "paid"]
    for row in paid:
        region_totals[row["region"]] = region_totals.get(row["region"], 0.0) + row["revenue"]
    region_totals = dict(sorted(region_totals.items()))

    largest_paid_ids = [row["id"] for row in sorted(paid, key=lambda row: (-row["revenue"], row["id"]))[:5]]
    weighted_unit_price = total / sum(row["units"] for row in rows)

    return {
        "revenue_stats": revenue_stats,
        "region_totals": region_totals,
        "largest_paid_ids": largest_paid_ids,
        "weighted_unit_price": weighted_unit_price,
    }
