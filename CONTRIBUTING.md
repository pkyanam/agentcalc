# Contributing

Use Go 1.24 or newer. Python 3 and Node are optional, but install both to exercise
all script-adapter tests.

```sh
go test -race ./...
go vet ./...
go build ./cmd/agentcalc
```

Keep core commands dependency-free and deterministic. Add tests for mathematical
invariants and error behavior when changing calculations. Preserve the JSON
response contract and document CLI changes in the README and bundled skill.
Numerical approximations should state their assumptions and reject nonfinite
results. Never describe the Python/Node adapters as a sandbox.

Open an issue or pull request on GitHub. Contributions are licensed under MIT.

Maintainers can build the six release archives with `scripts/release.sh 0.1.0`.
Pushing a `vX.Y.Z` tag runs the release workflow and publishes archives plus
`checksums.txt`. GitHub also provides source archives for each release.
