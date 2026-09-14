# agentcalc

A fast calculator CLI for agents. One dependency-free binary handles scientific expressions, exact fractions, statistics, units, matrices, and numerical calculus. JSON in, JSON out; Python and Node are optional escape hatches.

## Let your agent set it up

Copy and paste this prompt:

```text
Set up https://github.com/pkyanam/agentcalc for your own use. Identify your agent
host, OS/CPU, active profile, and supported skill discovery locations first.
Inspect the repo README, installer, and skills/agentcalc/SKILL.md.

Install the latest matching binary and verify its SHA-256 checksum. On macOS/Linux
with GitHub CLI, use `bash scripts/install.sh --skill` for the shared
~/.agents/skills/agentcalc location, or `--agent hermes`, `--agent claude`,
`--agent codex`, or `--skill-dir /absolute/host/skills/agentcalc` for your host.
Don't assume every host scans ~/.agents/skills. Respect custom profile homes and
existing installations; inspect before upgrading with --force. If the binary is
already installed, use `bash scripts/install-skill.sh --binary /absolute/agentcalc`
with the appropriate target. On Windows or without gh, use manual release/source
installation and install the skill through your host's supported mechanism.

The skill uses the open Agent Skills format. You can also use
`npx skills add pkyanam/agentcalc --skill agentcalc -g -a <your-agent-id> -y`
for supported agents; this installs the skill, not the binary. Preserve references.
Hermes normally reads ${HERMES_HOME:-$HOME/.hermes}/skills; for a shared copy,
configure skills.external_dirs in the active Hermes config without replacing
existing settings. If your host has no native skill loader, read SKILL.md from
its chosen location as task instructions. Do not claim discovery until checked.

Resolve the binary via PATH or record its absolute path/AGENTCALC_BIN. Check
`agentcalc version`, `agentcalc exact '0.1 + 0.2'` (fraction 3/10), and
`agentcalc eval 'sqrt(144) + 2^10'` (1036). Verify skill discovery using your host's
listing/reload mechanism. Report the binary path, skill path, version, and any
required reload. For subsequent work, answer obvious single-step small-integer
arithmetic directly (3+5, 12*4); use agentcalc for anything more involved or
uncertain. Group independent calculations in one run and use native table queries.
Calculate directly without creating helper files unless the task needs them.
Use successful results without repeating the calculation in another runtime.
```

## Install

**macOS or Linux, with [GitHub CLI](https://cli.github.com/) installed:**

```sh
git clone https://github.com/pkyanam/agentcalc.git
cd agentcalc
bash scripts/install.sh --skill
export PATH="$HOME/.local/bin:$PATH"
agentcalc eval '6*7'
# {"ok":true,"result":42}
```

The installer verifies SHA-256 checksums and needs no sudo. Omit `--skill` to
install only the binary. Use `--bin-dir DIR`, `--skill-dir DIR`, or
`--version 0.3.0` to customize installation. Existing installations are preserved
unless you explicitly pass `--force` to upgrade. GitHub CLI may require `gh auth
login` or a `GH_TOKEN` depending on your environment. No GitHub CLI is needed after
installation.

**Manual download, including Windows:** get the archive matching your OS and CPU
from [Releases](https://github.com/pkyanam/agentcalc/releases/latest), verify it
against `checksums.txt`, extract it, and place `agentcalc` (Windows:
`agentcalc.exe`) in a directory on your `PATH`. Builds cover macOS, Linux, and
Windows, each on amd64 (Intel/AMD) and arm64 (Apple Silicon/ARM). Each archive
includes the README, MIT license, and agent skill. GitHub also supplies source
ZIP and tar archives.

**From source, with Go 1.24 or newer:**

```sh
go install github.com/pkyanam/agentcalc/cmd/agentcalc@latest
# Go normally installs into $(go env GOPATH)/bin; add that directory to PATH.
agentcalc version
```

Or build a clone with `go build -o agentcalc ./cmd/agentcalc`. Source builds report
`dev` as the version; official release binaries embed their release version.
There are no third-party Go dependencies and no runtime requirements for core
commands.

## Portable agent skill

The [bundled skill](skills/agentcalc/SKILL.md) follows the
[Agent Skills specification](https://agentskills.io/specification). It uses
standard `name`, `description`, and `license` frontmatter, with additional command
instructions in a relative `references/` directory. Both the specification's
`skills-ref` validator and the skill-creator validator pass.

`--skill` now defaults to **`~/.agents/skills/agentcalc`**. This is a shared
convention, not a promise that every host scans that directory. Choose a target:

| Installer target | Skill directory |
| --- | --- |
| `--agent shared` (default) | `~/.agents/skills/agentcalc` |
| `--agent hermes` | `${HERMES_HOME:-$HOME/.hermes}/skills/agentcalc` |
| `--agent claude` | `${CLAUDE_CONFIG_DIR:-$HOME/.claude}/skills/agentcalc` |
| `--agent codex` | `${CODEX_HOME:-$HOME/.codex}/skills/agentcalc` |
| `--skill-dir DIR` | Exact directory supplied, including `agentcalc` |

For an already installed binary, install only the skill without redownloading:

```sh
bash scripts/install-skill.sh --agent shared --binary "$(command -v agentcalc)"
bash scripts/install-skill.sh --agent hermes --binary "$(command -v agentcalc)"
# Or a project-local directory your host recognizes:
bash scripts/install-skill.sh --skill-dir "$PWD/.agents/skills/agentcalc"
```

The installer appends an absolute binary path when one is provided. `--force`
updates a reviewed existing installation; otherwise it preserves existing files.
`--print-path` on the skill-only installer previews its resolved destination.
`AGENTS_HOME` is an agentcalc installer override for the shared root, not a
universal host setting. Host-specific overrides are honored as shown above.

**skills.sh:** its open CLI discovers this repository's `skills/agentcalc` folder
and can copy or link it into supported hosts. These commands install only the
skill; install the binary separately. See the [skills CLI documentation](https://github.com/vercel-labs/skills).

```sh
npx skills add pkyanam/agentcalc --list
npx skills add pkyanam/agentcalc --skill agentcalc -g -a hermes-agent -y
npx skills add pkyanam/agentcalc --skill agentcalc -g -a claude-code -y
```

**Hermes:** native skills belong to the active profile's Hermes home. To reuse
one shared copy instead, add its parent directory to `skills.external_dirs` in
the active profile's `config.yaml`, merging existing entries:

```yaml
skills:
  external_dirs:
    - ~/.agents/skills
```

Hermes also supports project `.hermes/skills` and `.agents/skills` directories,
subject to its project trust mechanism. Its hub can install directly with
`hermes skills install pkyanam/agentcalc/skills/agentcalc`. Discovery and profile
behavior are documented in [Hermes Skills System](https://hermes-agent.nousresearch.com/docs/user-guide/features/skills).
Claude's paths and override are documented in [Claude's directory guide](https://code.claude.com/docs/en/claude-directory).
For other hosts, use their supported paths or the skills CLI's agent registry;
read the skill explicitly when the host has no loader. Format portability does
not imply identical discovery rules.

## Commands

```sh
agentcalc eval 'sqrt(144) + 2^10'
agentcalc root 'x^2-2' 0 2
agentcalc integrate 'sin(x)' 0 3.141592653589793
agentcalc derivative 'x^3' 2
agentcalc eval 'price * (1 + tax)' --var price=49.95 --var tax=0.08
agentcalc exact '0.1 + 0.2'
agentcalc stats 1 2 3 4 5
agentcalc stats --input sales.csv --column revenue
agentcalc convert 72 f c
agentcalc units
agentcalc matrix inverse --data '{"a":[[4,7],[2,6]]}'
agentcalc python 'sum(x*x for x in data)' --data '[1,2,3]'
agentcalc node 'data.map(x => x * 2)' --data '[1,2,3]'
agentcalc python --file transform.py --data '{"x":2}'
agentcalc node --file transform.js --data '{"x":2}'
```

`eval` supports scientific notation, variables, constants `pi`, `e`, `tau`, postfix factorial `!`, and `^`/`**`. Powers associate right; `-2^2` is `-4`. Functions include `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `sqrt`, `cbrt`, `abs`, `ln`/`log`, `log10`, `log2`, `exp`, `floor`, `ceil`, `round`, `pow`, `hypot`, `atan2`, `clamp`, `min`, `max`, `sum`, `mean`, `factorial`, and `choose`. Angles use radians and `round` is half away from zero. `root`, `integrate`, and `derivative` are approximate numerical operations: roots require a continuous function and sign changing bracket, integration expects a smooth finite interval, and derivative expects a smooth function at `x`. `exact` requires spaces around its operator and performs one binary operation (`+`, `-`, `*`, `/`, `^`, or `**`) on rational numbers. Core arithmetic uses IEEE-754 `float64`; use `exact` for decimal or fraction arithmetic.

`stats` accepts positional numbers, JSON arrays, delimited stdin, or numeric CSV columns. `convert` supports length, mass, time, temperature, bytes, angle, speed, area, and volume with dimensional checks. `matrix` supports add, subtract, multiply, transpose, determinant, inverse, and solve; `solve` accepts `b` as a flat numeric vector or an n-by-1 column vector. Unit names are case-insensitive. `b` means bytes, `kb` means 1,000 bytes, `kib` means 1,024 bytes, and `gal` means US gallons. Matrix dimensions are limited to 256×256; ill-conditioned matrices may lose precision.

Python and Node expressions receive JSON input as `data`; Python also has `math` and `json`, and JavaScript has `Math` and `JSON`. Script files must define `main(data)`. These optional runtimes execute trusted local code with local permissions and are not sandboxed. Use `--timeout` (default 5s, maximum 5m). Core commands make no network requests or telemetry. Script timeouts stop the
runtime process; they are not an isolation boundary for subprocesses it launches.
Script output must be one JSON-serializable value; logging to stdout will break
the protocol. Use stderr for logs. Stdout is limited to 8 MiB and captured stderr
to 64 KiB.

Example `transform.py`:

```python
def main(data):
    return {"total": sum(data), "squares": [x*x for x in data]}
```

Run it with `agentcalc python --file transform.py --data '[1,2,3]'`.
A JavaScript file can define `function main(data) { return data.map(x => x*x); }`.
Synchronous JSON-serializable return values are expected.

## Short named calculations

Use `run` to return multiple answers in one process without JSON request boilerplate:

```sh
agentcalc run --text <<'CALC'
growth = eval 1000*(1+0.05/12)^24
fraction = exact 0.1 + 0.2
bytes = convert 3.75 GiB B
root = root cos(x)-x 0 1
solution = matrix solve {"a":[[2,1],[1,-1]],"b":[5,1]}
CALC
```

Each line is `name = command arguments`. Results form one JSON object; `exact`
returns the reduced fraction in this shorthand. Use ordinary `exact` for both
fraction and decimal. Names must be unique. Requests are independent, and a
failed request makes the whole run fail. Use `--input PATH` to read a saved run.
Expressions occupy the rest of the line; for root/integrate the final two values
are bounds, and for derivative the final value is the evaluation point.

The skill's routing rule is deliberately conservative: answer obvious one-step
small-integer arithmetic directly, such as `3+5` or `12*4`. Use the CLI when
uncertain or for larger numbers, chained operations, exact decimals, conversions,
functions, or data. Batch independent work, request only needed output, and avoid
writing helper scripts or repeating successful calculations unless the task
requires it. These choices reduce both generated code and repeated context.

## JSON and batch

Successful commands return `{"ok":true,"result":...}`. Errors return `{"ok":false,"error":"..."}`. Exit code 0 means success, 1 means calculation, input, or runtime error, and 2 means usage error. Batch continues after an individual failure and exits 1 if any request failed.

Batch consumes one JSON request per line. Supported commands are `eval`, `exact`, `stats`, `convert`, `matrix`, `units`, `root`, `integrate`, and `derivative`; script execution is unavailable. Root and integrate require `expr`, `lower`, and `upper`; derivative requires `expr` and `x`.

```sh
printf '%s\n' \
  '{"id":"a","command":"eval","expr":"6*7"}' \
  '{"id":"b","command":"convert","value":72,"from":"f","to":"c"}' \
  '{"id":"c","command":"matrix","op":"solve","a":[[2,1],[1,3]],"b":[[5],[6]]}' \
  | agentcalc batch
```

| Command | Required batch fields | Optional fields |
| --- | --- | --- |
| `eval` | `expr` | `vars` (object of numeric variables) |
| `exact` | `expr` | — |
| `stats` | `values` (numeric array) | — |
| `convert` | `value`, `from`, `to` | — |
| `matrix` | `op`, `a`; `b` for binary operations and solve | — |
| `root`, `integrate` | `expr`, `lower`, `upper` | — |
| `derivative` | `expr`, `x` | — |
| `units` | — | — |

Every request also needs `command`; optional `id` is echoed in its response.
Unknown fields are rejected. Batch requests are independent; results are not
implicitly available to later lines. Use `--collect --text` on a batch to return one object keyed by nonempty string
IDs, without launching Python or jq to reshape JSON. Collected batches require
unique IDs and fail as one error object if any request fails. `select` picks a
single result field, such as `fraction` from `exact`:

```sh
agentcalc batch --collect --text <<'JSONL'
{"id":"fraction","command":"exact","expr":"0.1 + 0.2","select":"fraction"}
{"id":"root","command":"root","expr":"cos(x)-x","lower":0,"upper":1}
JSONL
```

Use `--text` for just a result or `--pretty` for readable JSON on individual
commands or collected batches. Run `agentcalc --help` for all options.

Ordinary data and collected batches are limited to 8 MiB. Raw batch streams allow
1 MiB per line with no total stream limit. Exact literals have bounded exponents and power results
are capped at one million bits. The exact result includes a reduced `fraction`
and a `decimal` rounded to 30 places. Statistics report population and sample
variance/stddev (sample fields are absent for a single observation); percentiles
use linear interpolation. Quote expressions to prevent shell expansion.

## Native table queries

Read CSV with headers or a JSON array of objects. A named query produces one
output key; combine summaries in one process:

```sh
agentcalc table --input sales.csv --query '{"summary":{"op":"stats","column":"revenue","fields":["count","sum","mean"]},"paid_by_region":{"op":"sum","column":"revenue","group_by":"region","where":{"status":"paid"}},"top_ids":{"op":"values","column":"id","sort":[{"column":"revenue","desc":true},{"column":"id"}],"limit":5},"unit_price":{"op":"ratio","numerator":"revenue","denominator":"units"}}' --text
```

Use `--query-file PATH` for longer query objects. `where` applies AND equality
filters; `sort` and nonnegative `limit` apply before the operation. `stats` can
project named `fields`; only `sum` accepts `group_by`. At most 128 queries are
accepted. Numeric CSV columns are inferred as float64, so use JSON string fields
when identifiers must retain leading zeros. Numeric aggregation rejects empty
inputs; `values` and grouped sums return empty containers when no rows match.

## Benchmarks

The [v0.3.0 routing benchmark](benchmarks/2026-09-14/optimized-routing/REPORT.md)
measured **13.7% fewer total model tokens** and **34.0% fewer output tokens**
across arithmetic, data, and numerical workloads (two fresh Luna runs per arm
per workload; all 12 final answers correct). Total includes repeated and cached
input, skill loading, and retries. Tiny `3+5` checks used no tools and are excluded
from those savings totals.

Savings were concentrated in numerical work, including avoided dependency
lookup failures. Arithmetic and data consumed more total tokens in aggregate;
this is a small, tuned workload sample, not a universal or dollar-cost claim.
The report retains earlier failures and explains the changed answer-only protocol.
See also the [v0.2.0 benchmark](benchmarks/2026-09-14/REPORT.md), which required
reusable scripts and did not reduce aggregate total tokens.

## Contributing

```sh
go test ./...
go install ./cmd/agentcalc
```

Keep core changes dependency-free, add meaningful tests, and run `go test ./...` before opening a pull request. agentcalc is released under the [MIT License](LICENSE).
