#!/usr/bin/env bash
set -euo pipefail

AC=/Users/preetham/.local/bin/agentcalc

# Run the complete numerical suite in one inspectable agentcalc batch.
responses=$($AC batch <<'REQUESTS'
{"id":"solution","command":"matrix","op":"solve","a":[[10,2,-1,0],[2,11,3,-1],[-1,3,12,2],[0,-1,2,9]],"b":[[5],[25],[-2],[19]]}
{"id":"determinant","command":"matrix","op":"determinant","a":[[10,2,-1,0],[2,11,3,-1],[-1,3,12,2],[0,-1,2,9]]}
{"id":"root","command":"root","expr":"cos(x)-x","lower":0,"upper":1}
{"id":"integral","command":"integrate","expr":"exp(-x^2)","lower":0,"upper":2}
{"id":"derivative","command":"derivative","expr":"sin(x)*exp(x)","x":0.7}
{"id":"tiny_root","command":"root","expr":"1e-20*(x-1.25)","lower":0,"upper":3}
REQUESTS
)

# Convert the JSON Lines batch response into the requested single JSON object.
json_array=$(printf '%s\n' "$responses" | paste -sd, - | sed 's/^/[/' | sed 's/$/]/')
printf '%s' "$json_array" | $AC python '(lambda r: {"solution": r["solution"], "determinant": r["determinant"], "root": r["root"], "integral": r["integral"], "derivative": r["derivative"], "tiny_root": r["tiny_root"]})(dict((item["id"], item["result"]) for item in data if item.get("ok")))' --input - --text
