---
name: agentcalc
description: Use agentcalc for local numeric calculations, exact rational arithmetic, unit conversions, statistics, matrices, CSV summaries, and trusted Python or Node expressions.
---

# agentcalc

Use the installed `agentcalc` binary for small, inspectable calculations instead of reproducing arithmetic in prose. Resolve `agentcalc` through `PATH`; if it is unavailable, use the configured absolute binary path. The skill is normally installed at `~/.codex/skills/agentcalc` and should not overwrite an existing installation.

Core commands return JSON objects with `ok` and `result`; failures return `ok:false` and `error`. Exit code 0 is success, 1 is an input/calculation/runtime failure, and 2 is a usage error.

```sh
agentcalc eval 'sqrt(144) + 2^10'
agentcalc root 'x^2-2' 0 2
agentcalc integrate 'sin(x)' 0 3.141592653589793
agentcalc derivative 'x^3' 2
agentcalc exact '1/3 + 0.2'
agentcalc stats 1 2 3 4 5
agentcalc convert 72 f c
agentcalc matrix determinant --data '{"a":[[1,2],[3,4]]}'
```

Use `exact` for one binary rational operation when decimal representation must be exact. Other numeric operations use IEEE-754 `float64`. `root`, `integrate`, and `derivative` return approximate numerical results: roots need a continuous function and sign-changing `lower`/`upper` bracket, integration needs a smooth finite interval, and derivatives need a smooth function at `x`. The evaluator supports `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `sqrt`, `cbrt`, `abs`, `ln`/`log`, `log10`, `log2`, `exp`, `floor`, `ceil`, `round`, `pow`, `hypot`, `atan2`, `clamp`, `min`, `max`, `sum`, `mean`, `factorial`, and `choose`; angles are radians and `round` is half away from zero. `stats` also accepts JSON arrays, delimited stdin, and numeric CSV columns (`--input file.csv --column name`). `convert` supports dimensions listed by `agentcalc units`; unit names are case-insensitive. Matrix `solve` expects `b` as an n-by-1 column vector.

For trusted local data transformations, use `agentcalc python 'EXPR'` or `agentcalc node 'EXPR'`; input is available as `data`. Python additionally provides `math` and `json`, while Node provides `Math` and `JSON`. Script files passed with `--file` must define `main(data)`. Python and Node are optional local runtimes, execute with local permissions, and are not sandboxed.

For multiple core operations, send one JSON request per line to `agentcalc batch`:

```sh
printf '%s\n' '{"id":1,"command":"stats","values":[1,2,3]}' \
  '{"id":2,"command":"convert","value":1,"from":"km","to":"m"}' \
  | agentcalc batch
```

After installation, verify with `agentcalc version`, `agentcalc exact '0.1 + 0.2'`, and `agentcalc eval 'sqrt(144) + 2^10'`. For a manual installation, append a `Local installation` section recording the absolute binary path.

The installer appends a `Local installation` section containing the absolute binary path. Use that path when `agentcalc` is not on `PATH`, and reload skills when the host requires it.
