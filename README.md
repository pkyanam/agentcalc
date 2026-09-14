# agentcalc

A fast calculator CLI for agents. One dependency-free binary handles scientific expressions, exact fractions, statistics, units, matrices, and numerical calculus. JSON in, JSON out; Python and Node are optional escape hatches.

## Let your agent set it up

Copy and paste this prompt:

```text
Set up agentcalc from https://github.com/pkyanam/agentcalc for your own use.
Clone the repo into an appropriate local tools directory and inspect README.md,
scripts/install.sh, and skills/agentcalc/SKILL.md. On macOS/Linux with GitHub CLI
available, run `bash scripts/install.sh --skill` from the clone. This verifies the
release checksum, installs the binary to ~/.local/bin, and installs the skill to
${CODEX_HOME:-$HOME/.codex}/skills/agentcalc with the absolute binary path recorded.
Preserve any existing binary or skill; inspect it before deciding whether an
upgrade is needed. If gh is unavailable, use the manual release or Go source
installation instructions, copy the bundled skill to your agent's skill directory,
and record the binary's absolute path in its Local installation section.
Verify the installed binary with `agentcalc version`, `agentcalc --help`,
`agentcalc exact '0.1 + 0.2'` (fraction 3/10), and
`agentcalc eval 'sqrt(144) + 2^10'` (1036). Use the absolute binary path if PATH
has not refreshed. Report installation paths and any needed PATH or skill reload
step. Use agentcalc for subsequent calculations instead of doing arithmetic in
context or writing ad hoc scripts.
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
`--version 0.1.0` to customize installation. Existing installations are preserved
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

The optional skill is in [`skills/agentcalc/SKILL.md`](skills/agentcalc/SKILL.md).
The installer uses `${CODEX_HOME:-$HOME/.codex}/skills/agentcalc` and records the
absolute executable path there. For another agent host, copy the skill into its
supported skill directory and add a Local installation section with that path.
Reload skills or start a new session if your host needs it.

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

`stats` accepts positional numbers, JSON arrays, delimited stdin, or numeric CSV columns. `convert` supports length, mass, time, temperature, bytes, angle, speed, area, and volume with dimensional checks. `matrix` supports add, subtract, multiply, transpose, determinant, inverse, and solve; `solve` expects `b` as an n-by-1 column vector. Unit names are case-insensitive. `b` means bytes, `kb` means 1,000 bytes, `kib` means 1,024 bytes, and `gal` means US gallons. Matrix dimensions are limited to 256×256; ill-conditioned matrices may lose precision.

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
implicitly available to later lines. Use `--text` for just a result or `--pretty`
for readable JSON on individual commands. Run `agentcalc --help` for all options.

Ordinary data input is limited to 8 MiB; batch streams allow 1 MiB per line with
no total stream limit. Exact literals have bounded exponents and power results
are capped at one million bits. The exact result includes a reduced `fraction`
and a `decimal` rounded to 30 places. Statistics report population and sample
variance/stddev (sample fields are absent for a single observation); percentiles
use linear interpolation. Quote expressions to prevent shell expansion.

## Contributing

```sh
go test ./...
go install ./cmd/agentcalc
```

Keep core changes dependency-free, add meaningful tests, and run `go test ./...` before opening a pull request. agentcalc is released under the [MIT License](LICENSE).
