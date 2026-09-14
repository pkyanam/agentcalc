# Command reference

All commands emit JSON by default: success is `{"ok":true,"result":...}` and failure is `{"ok":false,"error":"..."}`. Exit status 0 means success, 1 means calculation, input, or runtime failure, and 2 means usage error. Core input is limited to 8 MiB; batch lines are limited to 1 MiB.

## Core commands

```sh
agentcalc eval 'sqrt(144) + 2^10'
agentcalc exact '1/3 + 0.2'
agentcalc stats 1 2 3 4 5
agentcalc stats --input sales.csv --column revenue
agentcalc convert 72 f c
agentcalc units
agentcalc matrix inverse --data '{"a":[[4,7],[2,6]]}'
agentcalc root 'x^2-2' 0 2
agentcalc integrate 'sin(x)' 0 3.141592653589793
agentcalc derivative 'x^3' 2
```

`eval` uses IEEE-754 float64. Supported functions are `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `sqrt`, `cbrt`, `abs`, `ln`, `log`, `log10`, `log2`, `exp`, `floor`, `ceil`, `round`, `pow`, `hypot`, `atan2`, `clamp`, `min`, `max`, `sum`, `mean`, `factorial`, and `choose`. `exact` supports one binary `+`, `-`, `*`, `/`, `^`, or `**` operation. `stats` accepts positional numbers, a JSON array, delimited input, or a numeric CSV column. `convert` checks dimensions across length, mass, time, temperature, bytes, angle, speed, area, and volume. Matrix operations are add, subtract, multiply, transpose, determinant, inverse, and solve; solve accepts `b` as a flat numeric vector or an n-by-1 column.

## Short named runs

`run --text` accepts `name = command arguments` lines through stdin or `--input PATH`.
It returns one named object atomically; duplicate names or errors fail the run.
Supported commands: eval, exact (fraction result), convert, stats, root, integrate,
derivative, matrix. Expressions are unquoted inside the run text. Root/integrate
consume two trailing numeric bounds; derivative consumes one trailing point.
Matrix syntax is `matrix OP {"a":[...],"b":[...]}`. No shell code is executed.
Use JSON batch requests for variables or explicit result field selection.

## Collected batches

Raw `batch` accepts one JSON request per line and echoes optional ids. Collection mode is useful when several calculations are independent:

```sh
agentcalc batch --collect --text <<'JSONL'
{"id":"total","command":"exact","expr":"1/3 + 0.2","select":"fraction"}
{"id":"root","command":"root","expr":"x^2-2","lower":0,"upper":2}
JSONL
```

Collection requires unique string ids and returns one object keyed by those ids. `select` is optional and selects an immediate result field, for example `fraction` from `exact`.

## Table queries

`table` reads CSV or JSON and applies a declarative query. Query operations are `stats`, `sum`, `values`, and `ratio` (with `numerator` and `denominator`). `stats` accepts `fields:["mean","sum"]` to project named result keys. Options include `sort` (an array of `{column,desc}`), `limit` (nonnegative integer), and `where` (AND equality predicates). Only `sum` accepts `group_by`. Filtering, sorting, and limiting happen before the operation. CSV columns containing numbers are inferred as float64, including numeric IDs; quote/preserve identifiers through JSON strings when leading zeros matter:

```sh
agentcalc table --input sales.csv \
  --query '{"totals":{"op":"sum","column":"revenue","group_by":"region","where":{"status":"paid"}}}' --text
```

## Scripts

```sh
agentcalc python 'sum(x*x for x in data)' --data '[1,2,3]'
agentcalc node 'data.map(x => x * 2)' --data '[1,2,3]'
agentcalc python --file transform.py --data '{"x":2}'
```

Expression input is `data`; Python also has `math` and `json`, and Node has `Math` and `JSON`. Script files define `main(data)`. Python 3 and Node.js are optional local runtimes and execute trusted code with local permissions.

Collected input/output are capped at 8 MiB; raw batches stream with no total limit. Each line is limited to 1 MiB. Matrix dimensions are capped at 256. Core scripts require Python 3 or Node only when those commands are used. `--timeout` defaults to 5s, maximum 5m.
