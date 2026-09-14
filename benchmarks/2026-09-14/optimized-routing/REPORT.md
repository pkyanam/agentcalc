# v0.3.0: routing and total-token optimization

The final matched sample used **13.7% fewer total model tokens** and **34.0%
fewer output tokens** with agentcalc. All 12 final answers passed independent
checks. Savings were concentrated in numerical work; arithmetic and data cost
more in aggregate. This is a small, tuned sample, not a universal savings claim.

## Final measurements

Two fresh `gpt-5.6-luna` agents per arm per suite, recorded reasoning effort
`medium`. These are rounds 5 and 6, using the final skill. All numbers below are
sums across the two repetitions. Positive savings mean fewer CLI tokens.

| Workload | Ordinary total | CLI total | Total savings | Ordinary output | CLI output |
| --- | ---: | ---: | ---: | ---: | ---: |
| Arithmetic | 169,525 | 172,466 | -1.7% | 1,008 | 900 |
| CSV aggregation | 176,891 | 202,927 | -14.7% | 1,514 | 1,489 |
| Numerical | 289,123 | 173,059 | 40.1% | 2,799 | 1,125 |
| **All three** | **635,539** | **548,452** | **13.7%** | **5,321** | **3,514** |

| Aggregate metric | Ordinary | CLI |
| --- | ---: | ---: |
| Input, including cache hits | 630,218 | 544,938 |
| Cached input (subset of input) | 598,528 | 517,376 |
| Uncached input | 31,690 | 27,562 |
| Output, including reasoning | 5,321 | 3,514 |
| Outer tool calls | 16 | 13 |
| Correct final answers | 6/6 | 6/6 |

Total means input plus output, counting repeated context on every response.
Cached input is included in total, not added again. Reasoning output is a subset
of output. These are actual provider counters, not character estimates or
estimates from an agent. This is not a dollar-cost comparison: cache/output
pricing and provider billing differ. Host system/tool context is substantial,
so avoiding one extra response can matter more than shortening a command.

The ordinary numerical agents tried unavailable libraries (mpmath, NumPy,
SciPy) before falling back to standard-library code. Those attempts are counted;
avoiding dependency discovery is part of this result. An environment with those
libraries preinstalled, or an agent told their availability up front, may show
smaller savings. One CLI data agent mistyped its skill path, reread it, and
continued; that extra call is also counted. We did not remove unlucky attempts.

## What changed

- `run` accepts short `name = command arguments` lines and returns one named
  JSON object. It reuses the existing numerical engine, requires no runtime,
  and removes repeated JSON request fields.
- The skill description excludes obvious single-step small-integer arithmetic.
  Examples such as `3+5` or `12*4` can be answered directly; larger or uncertain
  calculations, chained operations, exact decimals, units, functions and data
  go through the CLI.
- Guidance favors one invocation for independent calculations, direct answers
  without unnecessary helper files, and no redundant second-runtime check
  after success. Errors still require correction.
- Each table query must explicitly carry its requested filters. Sibling queries
  do not inherit them. This fixes a semantic failure found during iteration.

A same-binary microbenchmark of equivalent six-answer arithmetic requests used
242 input bytes for `run` versus 464 for JSON batch (47.8% less payload).
Across 100 sequential, randomly interleaved repetitions after 3 warmups, median
process times were 2.67 ms and 2.73 ms respectively. This shows comparable
runtime, not a meaningful runtime speedup claim. Payload bytes are not tokens.
See [native-results.json](native-results.json) and [native.py](native.py).

## Protocol and limits

The [protocol](protocol.txt) requests final answers only. Each agent reads the
same task and protocol together; the CLI arm additionally loads the skill in
that read, so cold skill-loading overhead is included. The three tasks are the
original [arithmetic](../tasks/arithmetic.txt), [data](../tasks/data.txt), and
[numerical](../tasks/numerical.txt) workloads. Data agents read the same 240-row
CSV from disk. Both arms receive the same efficiency guidance and available
shell tools, without a tool-call limit. Ordinary agents may use Python, Node,
and installed libraries; CLI agents use native operations or adapters. No
network, installs, other agents, or source/gold inspection is permitted.

This differs from the v0.2.0 protocol, which required reusable `solution.sh`
artifacts. Do not attribute a cross-version token difference entirely to the
binary: the new task format, instructions, and skill all changed. The comparison
above is between matched arms within the new protocol. It tests a combined
workflow, not an isolated causal effect of `run` versus `batch`.

The three suites were used during tuning. There is no held-out evaluation or
confidence interval; two repetitions per arm are too few to establish general
savings. Model runs overlapped, so wall time is reported in JSON for context,
not used as a stable performance claim. Exact values, units, statistics and
numerical tolerances are checked with the independent reference in
[../evaluate.py](../evaluate.py).

Four separate `3+5` routing smoke runs (two per arm) all returned 8 with zero
tool calls. They are excluded from every aggregate savings figure above. The
skill description itself still has a small context cost; skipping tools does
not make installed skill metadata free.

## Iterations retained

Forty fresh Luna runs are retained: 36 main-suite runs and 4 tiny smoke runs.

1. Rounds 1–2: exploratory, using [protocol-exploratory.txt](protocol-exploratory.txt)
   and [skill.md](skill.md). Some ordinary agents read an installed skill against
   instructions; some agents made unrelated communication-discovery calls, an
   invalid empty message attempt, or redundant calls. All costs remain in
   [exploratory-results.json](exploratory-results.json). Its 30.8% aggregate
   reduction is **not** the primary claim.
2. Rounds 3–4: clearer initial instructions prohibited skill reads in the ordinary
   arm and explained automatic delivery of finals. One CLI data answer omitted
   the paid filter for top IDs; one ordinary arithmetic answer used strings for
   numeric fields. Those failures remain in [iteration-results.json](iteration-results.json).
3. Rounds 5–6: the skill explicitly explained independent filters and included
   the filter in the top-ID example. Both repetitions were run afresh for all
   three suites. All 12 passed; this is [final-results.json](final-results.json)
   and the [final skill snapshot](skill-final.md).

A local trace audit found no forbidden ordinary-arm skill reads or unrelated
communication calls in the final rounds. Model-made errors/retries remain in
the measurements. No private transcripts or reasoning text are published.

## Reproduce

Use [launch.py](launch.py) to print the matched agent launch messages for a fresh
round, then launch each with Luna, medium effort, and a fresh context in your
agent host. The host must supply its own tool environment and actual usage logs.
Source checkout paths and the binary path are parameters. Repeating a shell
recipe cannot reproduce model-token usage.

```sh
python3 benchmarks/2026-09-14/optimized-routing/launch.py --round 7 --binary /absolute/agentcalc
python3 benchmarks/2026-09-14/optimized-routing/collect.py \
  --sessions /path/to/local/session/logs --rounds 7 --output fresh-results.json
python3 benchmarks/2026-09-14/optimized-routing/native.py \
  --binary /absolute/agentcalc --output native-results.json
```

The collector supports this host's JSONL schema, deduplicates per-response usage,
requires completed final responses, and emits only counters and correctness
facts. Other hosts need a usage-schema adapter. [all-results.json](all-results.json)
retains all 40 runs; rounds, individual counters, and errors are available for
independent reaggregation. [versions.json](versions.json) records source/input
hashes. Preserve the protocol and skill snapshots when comparing new rounds.
