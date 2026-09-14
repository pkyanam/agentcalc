#!/usr/bin/env bash
set -euo pipefail

agentcalc_bin="/tmp/agentcalc-benchmark-v2/agentcalc"

printf '%s\n' \
  '{"id":"growth","command":"eval","expr":"18750*(1+0.063/12)^37"}' \
  '{"id":"exact_sum","command":"exact","expr":"123456789.123456789 + 0.000000011","select":"fraction"}' \
  '{"id":"combinations","command":"eval","expr":"choose(52,7)"}' \
  '{"id":"gib_bytes","command":"convert","value":3.75,"from":"GiB","to":"bytes"}' \
  '{"id":"fahrenheit_celsius","command":"convert","value":-17.3,"from":"F","to":"C"}' \
  '{"id":"trig","command":"eval","expr":"sin(pi/7)^2 + cos(pi/7)^2 + sqrt(2025)"}' \
  | "$agentcalc_bin" batch --collect --text
