# Pipe compiler conformance

The pipe conformance suite combines compiler baselines with execution of emitted
JavaScript. It covers the pipe semantics from `TypeScript-d4d.7`; Function helper
conformance belongs to the separate helper work.

`testdata/tests/cases/compiler/pipeOperatorConformance.ts` is shared by the
compiler baseline runner and `TestPipeConformanceRuntime` in
`internal/compiler/pipe_emit_test.go`. Both compile it with strict checking at
ES2015, ES2022, and ESNext. The compiler runner records inferred types, symbols,
JavaScript, and declarations. The runtime test executes each emitted module in
Node.js and requires its completion marker. Type errors, JavaScript exceptions,
and missing asynchronous completion fail the test.
Node.js must be on PATH; missing Node.js fails the runtime suites rather than
silently skipping them. Iterator-failure cases require the expected exception
to be caught as well as checking that the target was never evaluated.

The shared fixture covers unary chains, explicit arguments, readonly tuples,
ordinary arrays, overloads, generic tuples, optional/default/rest parameters,
contextual callbacks, member receivers, assertion/predicate narrowing, recursive
getter invocation, default parameters, class fields, `await`, and `yield`.
Ordinary-call counterparts check the corresponding results. Existing negative
fixtures retain positional, arity, receiver, overload, incompatible-array, and
ordinary-call spread diagnostics.

`TestPipeEmitRuntime` adds isolated runtime cases at the same three targets. These
check input-first ordering, counts for every operand and invocation, spread
iteration and failures, member and proxy receivers, exceptions at each stage,
recursive and asynchronous re-entry, destructuring defaults, optional chains,
shadowed intrinsics, and ordinary JavaScript call/operator/spread behavior.
`TestPipeEmitTargets` also checks that no pipe AST remains in emitted JavaScript
for all supported targets and all four TS/TSX/JS/JSX input modes. Source-map,
comment, declaration, and source-parent checks remain in the emission suite.

Run from the repository root with Node.js on PATH:

```powershell
go test ./internal/compiler -run '^TestPipe' -count=1 -timeout=3m
go test ./internal/testrunner -run 'TestLocal/pipeOperator' -count=1 -timeout=3m
go test ./internal/scanner ./internal/parser ./internal/checker ./internal/ls ./internal/lsp ./internal/format ./internal/fourslash/tests -run 'Pipe' -count=1 -timeout=5m
```

The upstream submodule must be checked out for regression coverage. For focused
ordinary-call regressions:

```powershell
go test ./internal/testrunner -run 'TestSubmodule/(callWithSpread|genericRestParameters|variadicTuples|assertionFunctions|typeGuard)' -count=1 -timeout=5m
```

These focused checks do not replace the full upstream regression and lint gates
in `TypeScript-d4d.9`, or the versioned executable delivery in `.8`.

## Baseline and build review

The shared fixture adds nine compiler reference files: JavaScript/declarations,
inferred types, and symbols for each of ES2015, ES2022, and ESNext. Its exported
declarations retain the generic tuple `[number, string, boolean]`, contextual
callback results, overload results, receiver results, and `Promise<number>` for
the asynchronous pipe. Existing pipe fixtures continue to cover negative type
and arity diagnostics; their reference files are unchanged by this conformance
addition.

Use the normal build and lint commands alongside the focused tests:

```powershell
npm run build
npx --no-install hereby lint
npx --no-install dprint check internal/compiler/pipe_emit_test.go
```

The repository and its CI target Go 1.26. If lint reports that a newer Go
toolchain's export-data version cannot be decoded, use Go 1.26 for the lint
process rather than treating that importer failure as a source-code result.
For example, launch PowerShell with `GOTOOLCHAIN=go1.26.0` in its process
environment and run the lint command there; this does not require replacing the
system Go installation.

For the rebuilt executable, also compile the shared fixture with strict checking
and declaration emission and execute its JavaScript in Node.js at each of the
three targets. Require the same completion marker and review the inferred
declarations. This CLI smoke check verifies the built binary as well as the
in-process test compiler; it does not establish support in Bedrock or mobile
hosts, or implementation of the separately planned global Function helpers.
