# Global Function helpers for the pipe operator

This extends [PipeOperator.md](PipeOperator.md). It specifies planned work;
runtime helpers and their declarations are not implemented by this document.
All twelve helpers are static properties of the global `Function` constructor.
The plan's `Function.name` spelling corresponds to the requested `Function::name`
static access when the separate accessor-operator feature becomes available.

## Required contracts

### Function.dual (very high / P1)

Accept a function with exactly two required positional arguments, shaped like
`(a: A, b: B) => C`, and return a callable with both signatures:

```ts
(a: A, b: B) => C;
(b: B) => (a: A) => C;
```

Calling the wrapper with two arguments invokes the original as `fn(a, b)`;
calling it with one captures `b` and returns a unary function accepting `a`.
The original function runs once when both values are available. Dispatch uses
the number of supplied arguments, so an explicit `undefined` is still a supplied
argument. Reject source signatures with zero, one, three or more parameters,
optional parameters, or open rest parameters. A fixed two-element required rest
tuple is an equivalent two-argument signature. Do not infer this type contract
solely from JavaScript `Function.length`.

### Function.flow (very high / P1)

Accept a variable number of functions and compose them left to right:
`flow(f0, f1, f2)(...args)` evaluates `f2(f1(f0(...args)))`. Preserve the complete
argument tuple of `f0` (including zero arguments) and the final stage's return
type. Every later stage must have a single-argument signature, and each stage's
return type must be assignable to the next stage's input type. A tuple or array
result is one value, not an implicitly unpacked argument vector. Reject
incompatible adjacent stages and later zero-argument or multi-argument stages.

One function composes to equivalent invocation of that function. For zero
functions, use the generic unary identity function. There must be no arbitrary
public overload-count limit. Preserve inference across generic stages where
their signatures support it; test this instead of widening failures to `any`.
Evaluate stages once in order, stop on a throw, and do not implicitly await.

### Function.curry (very high / P1)

For a function `(a: A, b: B, c: C) => R`, return a function shaped like
`(a: A) => (b: B) => (c: C) => R`, preserving parameter order and result type.
Each stage accepts one argument and invokes the original only when the complete
argument vector is available. For source functions taking zero or one argument,
return the exact original function: `Function.curry(fn) === fn`.

Required fixed-arity signatures are the initial contract. Before implementation,
settle and document how optional/default/rest parameters, overloaded signatures,
and generic functions interact with arity and inference. An unbounded rest vector
has no inherent final unary stage; do not silently choose a stopping point or
claim support based on `Function.length`. Add explicit diagnostics for unsupported
forms and acceptance cases for every supported form.

### Function.identity (high / P2)

Use exactly these generic parameters and callable shape:

```ts
<In, Out extends In = In>(value: In) => Out;
```

Return the supplied value unchanged, preserving object reference identity.
Default `Out` to `In`; permit an explicit subtype `Out` and reject types that do
not extend `In`. An explicitly narrower `Out` is a caller-supplied type assertion;
it does not validate or transform the runtime value. Do not substitute the more
restrictive one-type-parameter identity signature for this requested contract.

### Function.constant (high / P2)

`constant(value)` returns a callable that always returns that same value,
including the same reference for objects. It may ignore any arguments supplied
when called, making it usable as a callback. Preserve the captured value's type.

### Function.flip (high / P2)

Reverse the complete positional argument vector: for `fn(a, b, c)`, the returned
function accepts `(c, b, a)` and invokes `fn(a, b, c)`. Preserve the reversed
parameter tuple and original return type. Zero- and one-argument functions retain
equivalent behavior. Specify optional/rest tuple handling before implementation;
do not silently interpret "reverse" as swapping only the first two arguments.

### Function.tupled and Function.untupled (high / P2)

`tupled(fn)` converts `fn(...args)` to a function taking that argument tuple as
one argument. `untupled(fn)` converts a function accepting one tuple argument
into a function accepting the tuple's elements positionally. Preserve element
order, tuple types, optional/rest information where supported, and return types.
Support zero-argument/empty-tuple cases. Accept compatible readonly input tuples
for `tupled` without mutating them. Round trips must preserve call results and
argument order, though wrapper reference equality is not required.

### Function.memoize (high / P2)

Return a wrapper with the source function's argument and result types. Cache by
the complete argument vector and the dynamic receiver, with arity included.
Use SameValueZero equality for primitive keys and reference identity for object
and function keys. Do not serialize arguments or conflate distinct objects.
Cache `undefined` results as well as other values. Repeated calls with the same
key return the same cached value; changed arguments or receiver invoke the source
again. Propagate synchronous exceptions without caching them. Cache a returned
promise itself, including a subsequently rejected promise; do not implicitly
await or retry it. Document cache lifetime and retention in the implementation.

### Function.once (medium / P3)

Return a wrapper with the original call shape that invokes the source at most
once, using the first invocation's arguments and receiver. Cache and return the
first result, including `undefined` or a promise. If the first invocation throws,
cache and rethrow that exception on later calls without invoking the source
again. Mark invocation as started before entering the source; synchronous
re-entry before a result exists must throw a documented `TypeError` instead of
executing the source a second time. Calls after a promise is returned reuse it.

### Function.compose (medium / P3)

Compose exactly two functions right to left:
`compose(outer, inner)(...args)` evaluates `outer(inner(...args))`. Preserve
`inner`'s argument tuple and `outer`'s return type. Require a unary `outer` whose
input accepts `inner`'s return type. Use the same compatibility, inference,
exception, and no-implicit-await rules as `flow`.

### Function.absurd (medium / P3)

Use a callable shape `<Out = never>(value: never) => Out` for exhaustive branches.
Reject reachable non-`never` arguments statically. Throw an error if invoked at
runtime through unchecked JavaScript or a type assertion; there is no valid
`never` value to return. Allow callers to select the enclosing branch's output
type without widening the input from `never`.

## Runtime and declaration delivery

Extend `FunctionConstructor` in the fork's standard libraries and deliver actual
runtime definitions for TS, TSX, JS, and JSX consumers across supported emission
targets. No helper should require an application import just to make the global
property exist. Decide and document how the compiler installs the runtime before
use, handles repeated loading and existing conflicting properties, and preserves
support in modules, scripts, bundled builds, and declaration-only consumers.
Use non-enumerable static properties and leave `Function.prototype` unchanged.
Check the runtime strategy against the Bedrock and mobile build roadmap without
claiming host validation that has not occurred.

For wrappers around receiver-sensitive functions, preserve the dynamic receiver
on invocation rather than accidentally binding the global `Function` object.
For `dual`'s deferred form and multi-stage `curry`, specify which invocation
provides the source receiver before implementation. Define consistent receiver
behavior for composition stages and test explicit `this` constraints. None of
these helpers changes the existing pipe member-receiver semantics.

## Acceptance and integration

Beads tracks runtime/declaration infrastructure in `TypeScript-d4d.11` and
integrated helper conformance in `TypeScript-d4d.24`. The individual helper tasks
depend on the infrastructure; integrated conformance depends on all twelve.

| Helper | Implementation issue |
| --- | --- |
| `Function.dual` | `TypeScript-d4d.12` |
| `Function.flow` | `TypeScript-d4d.13` |
| `Function.curry` | `TypeScript-d4d.14` |
| `Function.identity` | `TypeScript-d4d.15` |
| `Function.constant` | `TypeScript-d4d.16` |
| `Function.flip` | `TypeScript-d4d.17` |
| `Function.tupled` | `TypeScript-d4d.18` |
| `Function.untupled` | `TypeScript-d4d.19` |
| `Function.memoize` | `TypeScript-d4d.20` |
| `Function.once` | `TypeScript-d4d.21` |
| `Function.compose` | `TypeScript-d4d.22` |
| `Function.absurd` | `TypeScript-d4d.23` |

The original pipe-only conformance work remains `TypeScript-d4d.7`; its closure
does not establish helper conformance. Packaging/documentation
(`TypeScript-d4d.8`) and final regression review (`TypeScript-d4d.9`) also require
the added helper integration coverage. Planning changes are tracked separately
in `TypeScript-d4d.10` and do not mark any helper as implemented.

- Verify precise parameter tuples, inference, final return types, declaration
  output, and hover/signature help for the public helpers; include invalid input
  and adjacent-composition diagnostics and long `flow` chains.
- Test both `dual` call forms, including explicit `undefined`, and data-last pipe
  use such as `value |> Function.dual(fn)(option)`.
- Test `curry` for three or more arguments, evaluation timing, repeated use of
  partial stages, and exact reference equality for zero/unary functions.
- Test `identity` inference, explicit constrained `Out`, invalid unrelated `Out`,
  object reference equality, and unchanged primitive/`undefined` values.
- Test `constant`, complete argument reversal, tuple adapters and round trips,
  cache hits/misses and receiver keys, once-only execution including exceptions
  and re-entry, composition order, and exhaustive `absurd` branches.
- Execute emitted JavaScript for runtime assertions, including ESNext, multiple
  files/repeated installation, scripts/modules, and ordinary pipes using helpers.
  Declarations or output-text inspection alone do not prove runtime availability.
- Extend the remaining conformance, packaging/documentation, and final regression
  Beads tasks with the helper scope. Keep implementation work open until its
  runtime, typing, and integration acceptance criteria pass.
