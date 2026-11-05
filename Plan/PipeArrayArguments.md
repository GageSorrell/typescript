# Unknown-length pipe arguments

`TypeScript-d4d.4` allows ordinary arrays inside an unpacked pipe input:

```ts
declare function add(a: number, b: number): number;
const Foo = [1, 2];
const result = ...[...Foo] |> add; // number
```

Readonly arrays, mixed explicit arguments and spreads, multiple spreads, and
generic arrays work too. An ordinary array LHS still supplies one argument.
Ordinary calls such as `add(...Foo)` retain TypeScript's spread restrictions.

## Compatibility and length

Each spread's element type must fit every receiving parameter position it may
occupy, including optional parameters, array rests, and tuple rests. Explicit
values after an unknown-length spread are checked at every position they may
reach. Fixed suffixes of variadic tuple rests retain their end alignment.

```ts
declare function prefixed(head: string, a: number, b: number): void;
...["head", ...Foo] |> prefixed; // accepted

declare function heterogeneous(a: number, b: string): void;
...[...Foo] |> heterogeneous; // error: number can reach the string parameter
const mixed = [1, "x"];       // (string | number)[]
...[...mixed] |> heterogeneous; // error: the union does not fit either slot
const tuple = [1, "x"] as const;
...[...tuple] |> heterogeneous; // accepted: fixed positional types
```

Unknown lengths are left to JavaScript: missing values become `undefined` and
excess values are ignored by a fixed-arity receiver. There are no runtime length
checks. Explicit arguments and fixed tuples retain ordinary arity checks,
including provable excess explicit arguments when mixed with unknown spreads.
Ignored excess expressions still receive normal expression checking.

## Inference and flow

Pipes retain the existing call resolver's overload selection, generic inference,
receiver checks, contextual callbacks, and declaration inference. Inline RHS
function parameters infer from the possible inputs at each position. Generic
rest arguments use the existing spread/tuple inference machinery.

The checker measures possible argument positions from both ends of the input.
These ranges also limit control-flow narrowing: an assertion or predicate can
narrow a source reference only when its parameter position is certain. A fixed
prefix before an array spread can narrow; a suffix shifted by that spread cannot.
The original pipe AST and source parents remain intact.

## Validation and remaining stages

`internal/checker/pipe_spread_test.go` covers compatibility, arity, inference,
contextual functions, receivers, and flow. The `pipeOperatorArrayArguments.ts`
and `pipeOperatorArrayErrors.ts` compiler fixtures record type, symbol, and
diagnostic baselines. Declaration and flow fixtures cover the added inference
and narrowing behavior. Ordinary spread, generic-rest, assertion, and variadic
tuple fixtures provide upstream regression coverage.

```powershell
go test ./internal/checker -run '^TestPipe' -count=1 -timeout=2m
go test ./internal/testrunner -run 'TestLocal/pipeOperator' -count=1 -timeout=3m
```

JavaScript lowering is implemented in `.5`; see [PipeEmission.md](PipeEmission.md).
Language-server integration is documented in [PipeEditor.md](PipeEditor.md);
see [VSCodeSetup.md](VSCodeSetup.md) for selecting the local executable.
