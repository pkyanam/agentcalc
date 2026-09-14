#!/usr/bin/env bash
set -euo pipefail

exec /tmp/agentcalc-benchmark-v3/agentcalc batch --collect --text <<'JSONL'
{"id":"growth","command":"eval","expr":"18750*(1+0.063/12)^37"}
{"id":"exact_sum","command":"exact","expr":"123456789.123456789 + 0.000000011","select":"fraction"}
{"id":"combinations","command":"eval","expr":"choose(52,7)"}
{"id":"gib_bytes","command":"convert","value":3.75,"from":"gib","to":"bytes"}
{"id":"fahrenheit_celsius","command":"convert","value":-17.3,"from":"f","to":"c"}
{"id":"trig","command":"eval","expr":"sin(pi/7)^2 + cos(pi/7)^2 + sqrt(2025)"}
JSONL
