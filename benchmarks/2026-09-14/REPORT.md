# agentcalc × GPT-5.6-luna benchmark

Newer results: [v0.3.0 routing and total-token optimization](optimized-routing/REPORT.md). The measurements below describe v0.2.0 and its reusable-script protocol.

18 fresh Luna subagent runs across three paired workload suites and three development rounds. Every completed answer passed independent checks. After two iterations, native CLI recipes ran 3.4–4.2× faster than the ordinary-tool recipes in the final round. Output-token savings were about 26–27% for CSV and numerical work; simple arithmetic still used 35% more output tokens.

These are exploratory, cold-start agent measurements, not a statistically established model benchmark. We retained the losing runs.

## Final paired results

Lower is better for latency and tokens. Runtime medians include Bash, process startup, parsing, calculation, and JSON output. Model wall time is separate and includes tool calls and verification.

| Workload | Ordinary runtime | CLI runtime | Runtime speedup | Ordinary output tokens | CLI output tokens | Output-token change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| arithmetic | 21.52 ms | 5.13 ms | 4.19× | 1,263 | 1,708 | +35.2% |
| data | 25.51 ms | 7.47 ms | 3.41× | 2,256 | 1,664 | -26.2% |
| numerical | 21.03 ms | 5.26 ms | 4.00× | 2,472 | 1,809 | -26.8% |

| Workload | Ordinary model wall time | CLI model wall time | Ordinary tool calls | CLI tool calls |
| --- | ---: | ---: | ---: | ---: |
| arithmetic | 34.1 s | 43.8 s | 5 | 7 |
| data | 49.1 s | 37.1 s | 5 | 5 |
| numerical | 54.3 s | 41.8 s | 6 | 5 |

Across this particular final mix: **5,181 CLI output tokens versus 5,991 ordinary output tokens (13.5% fewer)**. This mix is not representative of every agent workload.

## Actual input, cache, and total usage

Counts below come from local Codex `token_usage_record` records for each isolated subagent. We sum per-response `usage`, deduplicated by response ID; we do **not** sum cumulative turn/thread totals. Output tokens include reasoning tokens, so the reasoning subset is not added again. All repeated system/tool context is included in input totals. Cache hits are a subset of input. No dollar-cost claim is made: cached and uncached tokens have different economics, and the recorded cache behavior depends on scheduling.

| Workload / arm | Input tokens | Cached input | Uncached input | Output | Total |
| --- | ---: | ---: | ---: | ---: | ---: |
| arithmetic / ordinary | 171,805 | 163,328 | 8,477 | 1,263 | 173,068 |
| arithmetic / cli | 242,412 | 228,352 | 14,060 | 1,708 | 244,120 |
| data / ordinary | 195,822 | 178,688 | 17,134 | 2,256 | 198,078 |
| data / cli | 178,957 | 169,472 | 9,485 | 1,664 | 180,621 |
| numerical / ordinary | 216,208 | 195,840 | 20,368 | 2,472 | 218,680 |
| numerical / cli | 177,381 | 170,496 | 6,885 | 1,809 | 179,190 |

Across all three final suites, aggregate total tokens were **603,931 CLI versus 589,826 ordinary (+2.4%)**. Thus this pilot supports output-token savings for the heavier workloads, **not an overall reduction in all tokens consumed**. Large repeated context and the extra arithmetic turns offset the gains.

## What the initial benchmark found and what changed

1. **JSON reshaping erased startup gains.** The original agents launched multiple calculator processes and Python/jq to assemble one object. Added `batch --collect --text`, unique string IDs, and `select` for result fields. Raw JSON Lines behavior remains available. Collected failures are atomic.
2. **CSV processing was just a Python wrapper.** Added native named table queries for stats, sums, grouped/filtered totals, sorted values, and ratios. This removes the need to write a CSV calculation script for those operations.
3. **An unnatural vector shape caused retries.** In iteration 2 the numerical agent repeatedly supplied flat `b` to `matrix solve`, receiving type errors before falling back to separate calls. Flat vectors now work alongside column matrices.
4. **Extra stats keys triggered another formatter.** Added `fields` projections to native table stats.
5. **Discovery overhead matters.** The skill now includes compact batch and table examples and moves less common detail into a reference. It is portable across skill loaders.

## All development rounds

Each row is a separate fresh subagent; later rounds are not continuations of earlier agents.

| Round | Workload / arm | Correct | Output tokens | Total tokens | Model wall s | Calls | Runtime median ms | Runtime p95 ms |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Initial: v0.1.0 | arithmetic / ordinary | yes | 1,662 | 145,787 | 37.8 | 4 | 26.78 | 28.44 |
| Initial: v0.1.0 | arithmetic / cli | yes | 2,349 | 212,935 | 53.0 | 6 | 36.78 | 47.02 |
| Initial: v0.1.0 | data / ordinary | yes | 2,843 | 212,067 | 61.3 | 6 | 26.67 | 28.23 |
| Initial: v0.1.0 | data / cli | yes | 2,865 | 274,504 | 64.8 | 7 | 45.46 | 47.12 |
| Initial: v0.1.0 | numerical / ordinary | yes | 2,456 | 178,256 | 53.7 | 5 | 23.57 | 26.47 |
| Initial: v0.1.0 | numerical / cli | yes | 3,287 | 369,523 | 76.1 | 10 | 31.06 | 32.44 |
| Iteration 2: collect + tables | arithmetic / ordinary | yes | 1,243 | 173,068 | 32.3 | 5 | 21.48 | 22.50 |
| Iteration 2: collect + tables | arithmetic / cli | yes | 1,811 | 223,052 | 41.7 | 6 | 5.42 | 6.18 |
| Iteration 2: collect + tables | data / ordinary | yes | 3,599 | 213,866 | 79.4 | 6 | 25.12 | 28.31 |
| Iteration 2: collect + tables | data / cli | yes | 2,388 | 273,177 | 53.7 | 7 | 7.84 | 8.39 |
| Iteration 2: collect + tables | numerical / ordinary | yes | 3,275 | 213,387 | 67.4 | 6 | 21.73 | 23.48 |
| Iteration 2: collect + tables | numerical / cli | yes | 3,664 | 490,245 | 86.7 | 14 | 12.97 | 13.58 |
| Final: vector shorthand + field projection | arithmetic / ordinary | yes | 1,263 | 173,068 | 34.1 | 5 | 21.52 | 22.39 |
| Final: vector shorthand + field projection | arithmetic / cli | yes | 1,708 | 244,120 | 43.8 | 7 | 5.13 | 5.59 |
| Final: vector shorthand + field projection | data / ordinary | yes | 2,256 | 198,078 | 49.1 | 5 | 25.51 | 26.35 |
| Final: vector shorthand + field projection | data / cli | yes | 1,664 | 180,621 | 37.1 | 5 | 7.47 | 8.17 |
| Final: vector shorthand + field projection | numerical / ordinary | yes | 2,472 | 218,680 | 54.3 | 6 | 21.03 | 22.05 |
| Final: vector shorthand + field projection | numerical / cli | yes | 1,809 | 179,190 | 41.8 | 5 | 5.26 | 5.83 |

## Protocol and limitations

- Model: **gpt-5.6-luna**, recorded reasoning effort **medium**, fresh `fork_turns=none` agents. Three suites × two arms × three rounds = 18 runs. Each round has only one run per arm per suite; the three product versions are not independent replicates of the same treatment.
- Both arms received the same calculation tasks and were allowed normal batching, shell tools, and already installed local runtimes/libraries. No network, package installation, or additional subagents during a trial. Ordinary agents could not use agentcalc or read its documentation. CLI agents had to route calculations through agentcalc; Python/Node adapters were allowed.
- The CLI arm includes a skill read, so these are **cold discovery/setup** measurements, not the cost of an already learned command. Both arms inherited the harness tool/skill metadata. The benchmark required a reusable script, result file, and execution, which adds artifact-writing overhead to both arms.
- Arithmetic checks cover compound growth, an exact decimal sum, combinations, binary byte conversion, temperature conversion, and trigonometry. Data checks cover 240 deterministic CSV rows, seven descriptive statistics, paid totals by region, top-five IDs, and weighted unit price. Numerical checks cover a 4×4 solve/determinant, two roots (including a scaled function), an integral, and a derivative.
- Gold checks are independent: exact Fraction Gauss–Jordan, permutation determinant, analytic erf integral and derivative, bisection root, and Decimal/statistics CSV aggregates. Tolerances are recorded in `evaluate.py`; saved answers and every replay are checked. Correctness means the **final answer after any retries**, not first-attempt success.
- Replay uses three warmups and 30 measurements per recipe, sequential execution in seeded randomized order. Host: Apple M3 / macOS arm64, Python 3.14.6 and Node 24.18.0. The ordinary solutions used Python standard-library code. Raw sample times and platform details are in the JSON files.
- Model tasks ran concurrently, so model wall time is noisy and cannot isolate service latency from tool execution. Counts include all retries and orchestration mistakes. One iteration-2 CLI run also hit an unrelated shell-policy rejection and attempted an invalid task-notification call; those costs were retained.
- The protocol gained one clarification after the first round: supplied literal inputs may be embedded; only computed answers must not be hardcoded. Ordinary arithmetic initially spent effort parsing the task text. Rounds 2 and 3 use the same clarified protocol.
- Changes were motivated by these workloads, so the final result is in-sample. This is evidence of useful engineering improvements, not a claim of broad statistical superiority.

## Practical recommendation

Use agentcalc for repeated calculations, batches, CSV aggregation, and numerical operations where it replaces implementation work. If an agent already has a tiny correct Python expression or a running numerical environment, forcing a new CLI/skill can increase model usage. The Python/Node adapters provide compatibility; merely wrapping an existing script is not expected to speed it up.

## Reproduce and inspect

Install agentcalc v0.2.0 and run from the repository:

```sh
python3 benchmarks/2026-09-14/replay.py --binary "$(command -v agentcalc)" --round final --repeat 30
```

The replay tool copies saved recipes into a temporary directory and relocates machine-specific paths without editing the originals. The ordinary recipes require Python 3; earlier CLI rounds additionally used jq/Node/Python as recorded in the scripts. To replay the initial treatment faithfully, supply a v0.1.0 binary and `--round initial`. This replays process performance and correctness; it does not rerun a model or recreate historical token counts.

- [Initial measurements](initial-results.json), [iteration 2](optimized-results.json), [final measurements](final-results.json).
- [Task protocol](tasks/protocol.txt), [clarified protocol](tasks/protocol-v2.txt), task text and CSV in `tasks/`.
- Original agent-authored recipes and outputs in `runs/`, `runs-v2/`, and `runs-v3/`.
- [Evaluator and telemetry collector](evaluate.py), [portable replay](replay.py), [version hashes](versions.json), and `skill-snapshots/`.
- Token records contain timestamps and usage counters only. Private transcripts, hidden reasoning text, and credentials are not bundled.

For a new model comparison, launch fresh Luna agents using the archived protocol/task files, vary order and seeds across several replicates, keep skill/binary versions fixed, and collect the same per-response usage records. Do not present replay-only timing as a model speedup.
