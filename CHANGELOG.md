# Changelog

## 0.2.0

- Added collected batch JSON objects and per-result field selection.
- Added native CSV/JSON table queries with filters, grouping, sorting, ratios, and stats projections.
- Accepted flat right-hand-side vectors for matrix solving.
- Made skill installation portable: shared ~/.agents by default, explicit Hermes/Claude/Codex targets, arbitrary paths, and a skill-only installer.
- Validated the bundled skill against the Agent Skills reference validator, skill-creator checks, and skills.sh CLI discovery/install.
- Published an 18-run Luna benchmark with actual token counts, correctness checks, original recipes, all iterations, and a portable replay harness.

## 0.1.0

Initial MIT-licensed release of agentcalc.

- Scientific expression evaluator with variables, constants, combinatorics, and domain checks.
- Exact rational arithmetic for decimal, fraction, and large integer calculations.
- Descriptive statistics from numbers, JSON, delimited text, or named CSV columns.
- Unit conversion across ten dimensions.
- Matrix arithmetic, determinants, inversion, and linear system solving.
- Bracketed roots, adaptive numerical integration, and finite-difference derivatives.
- Optional Python and Node expressions and script files with JSON data and timeouts.
- JSON output, structured errors, JSON Lines batching, and plain/pretty output modes.
- Bundled agent skill, guided setup prompt, checksum-verifying installer, and six platform builds.
