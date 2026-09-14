#!/usr/bin/env bash
set -euo pipefail

bin=/tmp/agentcalc-benchmark-v2/agentcalc
a='[[10,2,-1,0],[2,11,3,-1],[-1,3,12,2],[0,-1,2,9]]'

solution=$($bin matrix solve --data "{\"a\":$a,\"b\":[[5],[25],[-2],[19]]}" --text)
determinant=$($bin matrix determinant --data "{\"a\":$a}" --text)
calculus=$($bin batch --collect --text <<'JSONL'
{"id":"root","command":"root","expr":"cos(x)-x","lower":0,"upper":1}
{"id":"integral","command":"integrate","expr":"exp(-x^2)","lower":0,"upper":2}
{"id":"derivative","command":"derivative","expr":"sin(x)*exp(x)","x":0.7}
{"id":"tiny_root","command":"root","expr":"1e-20*(x-1.25)","lower":0,"upper":3}
JSONL
)

jq -n --argjson solution "$solution" --argjson determinant "$determinant" --argjson calculus "$calculus" \
  '{solution:$solution, determinant:$determinant, root:$calculus.root, integral:$calculus.integral, derivative:$calculus.derivative, tiny_root:$calculus.tiny_root}'
