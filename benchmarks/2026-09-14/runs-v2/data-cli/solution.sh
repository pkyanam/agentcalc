#!/usr/bin/env bash
set -euo pipefail

bin=/tmp/agentcalc-benchmark-v2/agentcalc
input=/Users/preetham/Code/agentcalc/benchmarks/2026-09-14/tasks/sales.csv

raw="$($bin table --input "$input" --query '{"revenue_stats":{"op":"stats","column":"revenue"},"region_totals":{"op":"sum","column":"revenue","group_by":"region","where":{"status":"paid"}},"largest_paid_ids":{"op":"values","column":"id","where":{"status":"paid"},"sort":[{"column":"revenue","desc":true},{"column":"id"}],"limit":5},"weighted_unit_price":{"op":"ratio","numerator":"revenue","denominator":"units"}}' --text)"

# The table query performs all calculations. jq only selects the requested
# fields and emits the required compact JSON object.
printf '%s\n' "$raw" | jq -c '{revenue_stats: {count: .revenue_stats.count, sum: .revenue_stats.sum, mean: .revenue_stats.mean, median: .revenue_stats.median, sample_stddev: .revenue_stats.sample_stddev, p25: .revenue_stats.p25, p75: .revenue_stats.p75}, region_totals: .region_totals, largest_paid_ids: .largest_paid_ids, weighted_unit_price: .weighted_unit_price}'
