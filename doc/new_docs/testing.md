# Compiler testdir testing

The `cmd/internal/testdir` package runs the `.go` files under `$GOROOT/test` (compile, run, errorcheck, etc.). These are the main compiler integration tests.

Run all commands from `$GOROOT/src` with `GOROOT` set to this tree:

```bash
cd /root/go/src
export GOROOT=/root/go
```

Rebuild the compiler first if you changed it:

```bash
./make.bash
```

## Timeouts

Vanilla Go does not apply a default subprocess timeout in testdir; only recipes with an explicit `-t N` flag are timed.

This fork adds **`-cmdtimeout=N`** (seconds). When a test recipe does not specify `-t`, subprocesses use `-cmdtimeout` if it is greater than zero. The default is **300** seconds (5 minutes); set **`-cmdtimeout=0`** for vanilla behavior (no timeout).

Example: cap per-command runtime at 5 minutes during a shard run:

```bash
go test -count=1 -timeout=45m -parallel=8 -v cmd/internal/testdir \
  -shard=0 -shards=8 -cmdtimeout=300 2>&1 | tee /root/go/testdir_output.txt
```

Use a generous **package timeout** (`-timeout=…` on `go test`) so the shard can finish. Individual hangs are capped by `-cmdtimeout` when set.

## Quick check (one test)

Run a single file by name:

```bash
go test -count=1 -timeout=2m -v cmd/internal/testdir \
  -run='Test/fixedbugs/issue8947.go'
```

Other useful filters:

```bash
# All tests in a directory
go test -count=1 -timeout=10m -v cmd/internal/testdir -run='Test/fixedbugs/'

# One top-level test file
go test -count=1 -timeout=2m -v cmd/internal/testdir -run='Test/escape.go'
```

Save output:

```bash
go test -count=1 -timeout=2m -v cmd/internal/testdir \
  -run='Test/fixedbugs/issue8947.go' 2>&1 | tee /root/go/testdir_output.txt
```

## Small batch (one shard)

Split the suite into **N shards** and run one shard. This is much faster than the full suite and still exercises a broad slice of tests.

Example: **N=8**, shard **0**, **8** parallel subtests, **45m** package timeout:

```bash
go test -count=1 -timeout=45m -parallel=8 -v cmd/internal/testdir \
  -shard=0 -shards=8 -cmdtimeout=45 2>&1 | tee /root/go/testdir_output.txt
```

Run other shards by changing `-shard` (0 through 7 when `-shards=8`):

```bash
go test -count=1 -timeout=45m -parallel=8 -v cmd/internal/testdir \
  -shard=3 -shards=8 2>&1 | tee /root/go/testdir_shard3.txt
```

A recent shard-0 run with these flags completed in **~49s** with **exit code 1** (failures present). Inspect failures with:

```bash
grep -E '^--- FAIL:|testdir_test.go:147:|unexpected success' testdir_output.txt
```

Or keep a separate summary file:

```bash
grep -E '^--- FAIL:|testdir_test.go:147:|testdir_test.go:145:|unexpected success' \
  testdir_output.txt > testdir_failures.txt
```

## Full suite (all tests)

With `-shards=0` (the default), **all** testdir cases run in one invocation. Expect a long run (often 30–90+ minutes depending on hardware).

```bash
go test -count=1 -timeout=3h -parallel=8 -v cmd/internal/testdir \
  2>&1 | tee /root/go/testdir_output.txt
```

To run every shard explicitly (same total coverage, easier to parallelize across machines or terminals):

```bash
for i in $(seq 0 7); do
  go test -count=1 -timeout=45m -parallel=8 cmd/internal/testdir \
    -shard=$i -shards=8 2>&1 | tee "/root/go/testdir_shard${i}.txt"
done
```

## Flag reference

| Flag | Purpose |
|------|---------|
| `-shard=N` | Which shard to run (0-based) |
| `-shards=N` | Split suite into N shards; `0` means no sharding (all tests) |
| `-cmdtimeout=N` | Default subprocess timeout in seconds when a recipe has no `-t` (default **300**; `0` = none) |
| `-parallel=N` | Max concurrent subtests (`go test` flag) |
| `-timeout=…` | Max wall time for the whole `go test` invocation |
| `-run='Test/…'` | Run only matching subtests |
| `-v` | Verbose; print each subtest and failure details |
| `-count=1` | Disable test cache |

Testdir-only flags (passed after the package name):

| Flag | Purpose |
|------|---------|
| `-f` | Ignore expected-failure lists (treat known failures as real failures) |
| `-l=N` | Limit parallel `runoutput` tests |

Environment (upstream Go also uses these):

| Variable | Purpose |
|----------|---------|
| `GO_TEST_SHARDS` | Used by `cmd/dist test` to set shard count (default 1; 10 on builders) |
| `GO_TEST_TIMEOUT_SCALE` | Scales `-t` timeouts in individual test recipes |

## Related tests

Unit tests inside the compiler tree (faster, narrower):

```bash
go test -count=1 -timeout=10m cmd/compile/...
```

Syntax parser regression for switch/composite-literal cases:

```bash
go test -count=1 -run='TestIssue8947SwitchCaseParse' cmd/compile/internal/syntax
```
