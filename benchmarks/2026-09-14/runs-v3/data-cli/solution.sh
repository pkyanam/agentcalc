#!/bin/bash
set -euo pipefail

run_dir="$(cd "$(dirname "$0")" && pwd)"
csv_path="$run_dir/../../tasks/sales.csv"
bin_path="/tmp/agentcalc-benchmark-v3/agentcalc"

query='{"revenue_stats":{"op":"stats","column":"revenue","fields":["count","sum","mean","median","sample_stddev","p25","p75"]},"region_totals":{"op":"sum","column":"revenue","group_by":"region","where":{"status":"paid"}},"largest_paid_ids":{"op":"values","column":"id","where":{"status":"paid"},"sort":[{"column":"revenue","desc":true},{"column":"id"}],"limit":5},"weighted_unit_price":{"op":"ratio","numerator":"revenue","denominator":"units"}}'
"$bin_path" table --input "$csv_path" --query "$query" --text
