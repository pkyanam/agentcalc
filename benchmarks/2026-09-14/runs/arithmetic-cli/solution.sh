#!/usr/bin/env bash
set -euo pipefail

AGENTCALC="/Users/preetham/.local/bin/agentcalc"

# Each numeric result is computed by agentcalc; Python below only assembles JSON.
growth_json="$($AGENTCALC eval '18750*(1+0.063/12)^37')"
exact_sum_json="$($AGENTCALC exact '123456789.123456789 + 0.000000011')"
combinations_json="$($AGENTCALC eval 'choose(52,7)')"
gib_bytes_json="$($AGENTCALC convert 3.75 gib byte)"
fahrenheit_celsius_json="$($AGENTCALC convert -17.3 f c)"
trig_json="$($AGENTCALC eval 'sin(pi/7)^2 + cos(pi/7)^2 + sqrt(2025)')"

export growth_json exact_sum_json combinations_json gib_bytes_json fahrenheit_celsius_json trig_json
python3 - <<'PY'
import json
import os

def result(name):
    payload = json.loads(os.environ[name])
    if not payload.get("ok"):
        raise SystemExit(payload.get("error", "agentcalc failure"))
    return payload["result"]

exact = result("exact_sum_json")
output = {
    "growth": result("growth_json"),
    "exact_sum": exact["fraction"],
    "combinations": result("combinations_json"),
    "gib_bytes": result("gib_bytes_json"),
    "fahrenheit_celsius": result("fahrenheit_celsius_json"),
    "trig": result("trig_json"),
}
print(json.dumps(output, separators=(",", ":")))
PY
