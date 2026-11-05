# Pipe type checking and control flow

`TypeScript-d4d.3` integrates the [pipe AST](PipeSyntax.md) with the compiler's
existing call resolver. Pipes reuse ordinary overload selection, generic
inference, contextual callback typing, optional/rest parameter checks, receiver
checks, and return-type inference. Declaration emission uses those inferred
result types.

## Source AST contract

The checker resolves signatures on the original pipe `BinaryExpression`.
It does not replace the source tree with synthetic calls or change argument
parents. Go AST accessors expose a pipe's RHS through `Expression()` and its
original argument nodes through `Arguments()` / `ArgumentList()`:

- An ordinary LHS supplies one argument.
- A spread-prefixed array literal supplies its element nodes as arguments.
- The existing effective-argument logic expands fixed tuple spreads.
- A member RHS supplies its object as the `this` receiver.

Inline RHS functions participate in immediately invoked function inference and
control flow, with arguments bound before the target. Contextual typing of
unpacked elements uses their effective call positions without reparenting them.

## Flow behavior

Predicate conditions narrow input references, including explicit unpacked
arguments. Standalone assertion calls and never-returning calls participate in
flow analysis under the existing TypeScript assertion-target rules. Array
`push`/`unshift` targets participate in evolving-array inference.

Since a pipe captures inputs before evaluating its RHS, a predicate cannot narrow
an input reference that the RHS reassigns. The checker detects explicit writes,
including property writes and writes inside immediately invoked functions, and
withholds that narrowing.

## Boundaries

Unknown-length array spreads use the pipe-specific checking added in `.4`,
described in [PipeArrayArguments.md](PipeArrayArguments.md). Ordinary calls retain
TypeScript's spread restrictions; explicit pipe arguments and fixed tuples retain
their existing arity checks.
JavaScript lowering is implemented in `.5`; see [PipeEmission.md](PipeEmission.md).
Declaration emission continues to use the original source AST and inferred types.

## Validation

```powershell
go test ./internal/checker -run '^TestPipe' -count=1 -timeout=2m
go test ./internal/testrunner -run 'TestLocal/pipeOperator' -count=1 -timeout=3m
```

Unit tests verify call types and diagnostics, inference, receivers, assertions,
predicates, target side effects, declaration output, and unchanged source parents.
The `pipeOperator*.ts` compiler fixtures have reviewed diagnostic, type,
symbol, and declaration baselines. Regression checks also cover the shared AST,
binder, checker, scanner, parser, and encoder, plus representative upstream call
and flow fixtures.
