# Pipe JavaScript emission

`TypeScript-d4d.5` adds an unconditional pipe transform to JavaScript emission,
after type erasure and JSX transformation and before target-specific transforms.
It runs for TS, TSX, JS, and JSX, including the ESNext target. The source AST
remains available unchanged for checking, declarations, and editor requests.

## Evaluation and captures

Each pipe first evaluates and captures its arguments, then evaluates its target
as a pair containing the receiver and callable, then invokes that callable.
An ordinary LHS supplies one captured value. An unpacked array supplies its
elements, expanding iterables before target evaluation. Elisions are captured
as `undefined`. Captured arguments are applied without iterating the capture
array again, so changing array iteration in the RHS cannot change those values.

The captures are parameters of generated arrow invocations. They belong to
each evaluation, including recursive calls, getters, default parameters, field
initializers, and suspended async/generator evaluations. There is no shared
pipe temporary or runtime length check. Chained stages finish one invocation
before evaluating the next target.

Target access records the receiver before reading the method, including
computed properties, private members, and `super`. Invocation uses `apply`
through a fresh function's prototype methods; it does not consult a target's own
`call` or `apply`, or identifiers named `Function` or `Reflect` in user code.

Optional-chain targets retain their own short-circuit rules for property keys
and intermediate call arguments. The pipe invocation itself is unconditional:
if its target evaluates to `undefined`, invocation throws. No implicit await,
currying, placeholders, or optional pipe invocation is introduced.

Source `await` and `yield` expressions remain in call arguments in their
original execution context. The normal target transforms subsequently lower
async functions, optional chains, and class fields as needed.

## Output and validation

The transform preserves original-node links and source locations. Tests check
operand source mappings, comments, unchanged source parents, and inferred
declarations. Declaration emission does not expose generated captures.

```powershell
go test ./internal/compiler -run '^TestPipe' -count=1 -timeout=3m
go test ./internal/testrunner -run 'TestLocal/pipeOperator' -count=1 -timeout=3m
npm run build
```

`internal/compiler/pipe_emit_test.go` reparses output for every supported target
from ES2015 through ESNext, in all four input modes. It executes runtime cases
with Node.js using ES2015, ES2022, and ESNext output, including ordering, spreads,
receivers, exceptions, recursive re-entry, missing/excess arguments, default
parameters, class fields, `await`, and `yield`. The `pipeOperatorEmit.ts` fixture
records JavaScript, declarations, types, and symbols at three representative
targets. Existing call/checker and upstream emitter fixtures provide regression
coverage.

After building, pipe programs can now be compiled and run normally:

```powershell
./built/local/tsgo.exe --ignoreConfig --target esnext --outDir built/verification/pipe-emit testdata/tests/cases/compiler/pipeOperatorEmit.ts
```

The sample fixture exports values and functions; importing its generated module
allows inspecting the results. Editor integration is documented in
[PipeEditor.md](PipeEditor.md).
The local compiler identifies as `7.0.2-pipe.0`. See
[PipeDelivery.md](PipeDelivery.md) for the runnable example, SDK archive, and
CLI/LSP verification. The expanded `.8` still requires Function helper delivery.
