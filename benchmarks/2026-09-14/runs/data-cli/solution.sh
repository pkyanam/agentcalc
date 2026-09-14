#!/usr/bin/env bash
set -euo pipefail

AGENTCALC="/Users/preetham/.local/bin/agentcalc"
TASK_DIR="/Users/preetham/Code/agentcalc/benchmarks/2026-09-14/tasks"
RUN_DIR="/Users/preetham/Code/agentcalc/benchmarks/2026-09-14/runs/data-cli"

# Python here only serializes the source file as a JSON string for the adapter.
csv_json="$(python3 -c 'import json,sys; print(json.dumps(open(sys.argv[1]).read()))' "$TASK_DIR/sales.csv")"
"$AGENTCALC" python --file "$RUN_DIR/transform.py" --data "$csv_json" --text
