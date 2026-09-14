---
name: agentcalc
description: Use agentcalc for multi-step arithmetic, exact decimals/fractions, statistics, unit conversion, matrices, calculus, or tabular aggregation. Skip this skill and tools for obvious single-step small-integer arithmetic (3+5, 12*4); answer directly. Batch nontrivial calculations in one invocation.
license: MIT
---

# agentcalc

Use `AGENTCALC_BIN` if set, otherwise `agentcalc` on PATH or the Local installation path. Requires shell access.

## Route efficiently

Answer obvious single-step small-integer arithmetic directly, without loading references or calling tools: `3+5`, `20-7`, `12*4`, `100/4`. If uncertain, use the CLI. Use it for everything more involved: chained operations, large numbers, decimals requiring exactness, units, data, functions, statistics, and calculus. Do not calculate tool results again in context.

Use one invocation for independent answers. Read only the needed reference section if syntax below is insufficient. Run the calculation directly; create scripts/files only when requested or needed for reuse. After a successful result, answer from it; do not repeat it with a second runtime just to confirm. Fix actual errors.

## Calculate

Single: `agentcalc eval 'sqrt(144)+2^10' --text`. Float64; `pi,e,tau`, `^,**`, `choose(n,k)`; radians. For exact rational/decimal arithmetic use `exact '0.1 + 0.2'` (one binary operator, surrounding spaces).

Multiple: `agentcalc run --text` reads named lines from stdin:

```sh
agentcalc run --text <<'CALC'
growth = eval 1000*(1+0.05/12)^24
fraction = exact 0.1 + 0.2
root = root cos(x)-x 0 1
solution = matrix solve {"a":[[2,1],[1,-1]],"b":[5,1]}
CALC
```

Returns one object keyed by name; exact yields its fraction. Also: `convert 3.75 GiB B`, `integrate exp(-x^2) 0 2`, `derivative sin(x)*exp(x) 0.7`, `matrix determinant {"a":[[2,1],[1,-1]]}`, `stats 1 2 3`. Numerical calculus is approximate; roots require a continuous sign-changing bracket.

## Aggregate files

```sh
agentcalc table --input sales.csv --query '{"stats":{"op":"stats","column":"revenue","fields":["count","sum","mean","median","sample_stddev","p25","p75"]},"totals":{"op":"sum","column":"revenue","group_by":"region","where":{"status":"paid"}},"ids":{"op":"values","column":"id","where":{"status":"paid"},"sort":[{"column":"revenue","desc":true},{"column":"id"}],"limit":5},"price":{"op":"ratio","numerator":"revenue","denominator":"units"}}' --text
```

Each query is independent: apply its requested filters explicitly; sibling queries do not share filters. Match output names to the task. CSV headers or JSON object arrays. Equality `where`, then `sort`, then `limit`, then aggregation. `fields` projects stats. Only sum supports group_by. Python/Node adapters handle unsupported transformations using trusted local code.

`--text` unwraps successful JSON. Errors remain JSON with `ok:false,error` and nonzero exit status. Never present an error as an answer. Additional syntax, JSON batches, script adapters and limits: [reference](references/commands.md).
