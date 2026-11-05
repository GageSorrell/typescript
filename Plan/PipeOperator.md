# TypeScript 7.0.2 fork with `|>`

## Summary

Populate `E:\TypeScriptFork` with the native compiler from the official [`typescript/v7.0.2` tag](https://github.com/microsoft/typescript-go/tree/typescript/v7.0.2), then add the pipe operator throughout parsing, type checking, JavaScript emission, and the built-in language server.

Deliver compiler source, tests, a working Windows binary, examples, and build/editor instructions.

Extend this delivery with twelve static helpers on the global `Function` object,
as specified in [Global Function helpers](FunctionHelpers.md). Supply both typed
declarations and working JavaScript runtime support.

## Confirmed language behavior

```ts
value |> fn                      // Pass one argument.
value |> first |> second         // Equivalent result to second(first(value)).
...[a, b] |> fn                   // Pass two arguments.
...[...Foo] |> fn                 // Expand a named array into arguments.
[a, b] |> fn                      // Pass the array as one argument.
value |> object.method           // Preserve object as this.
value |> (x => x + 1)             // Parenthesized arrow functions work.
```

- Evaluate each stage's input once, before evaluating its RHS. Complete each invocation before advancing to the next stage.
- Chains associate left to right. Precedence is below arithmetic, comparison, logical, and nullish-coalescing operators, and above conditional expressions and assignment.
- The RHS is an expression producing a callable value. For example, `value |> factory()` invokes the function returned by `factory()`.
- Multi-argument syntax requires a spread-prefixed array literal. Reject standalone `...Foo |> fn` and an unpack marker without a following pipe.
- Ordinary arrays may supply arguments to fixed-arity functions. Check element compatibility, while leaving unknown array lengths to runtime. Missing arguments become `undefined`; excess arguments follow ordinary JavaScript behavior.
- Pipes introduce no implicit awaiting, currying, or argument placeholders.
- Explicit `Function.dual`, `Function.flow`, and `Function.curry` helpers support
  data-last pipelines and function composition without changing pipe semantics.

## Global Function helpers

These are properties of the global `Function` constructor, not
`Function.prototype` methods. `Function::name` denotes the same proposed static
access in the broader accessor-operator roadmap; this pipe delivery uses the
currently supported `Function.name` spelling and does not depend on `::`.

| Helper | Purpose | Priority | Beads priority |
| --- | --- | --- | --- |
| `Function.dual` | Binary data-first/data-last invocation | Very high | P1 |
| `Function.flow` | Compose a variable number of functions left to right | Very high | P1 |
| `Function.curry` | Convert a function to curried form | Very high | P1 |
| `Function.identity` | Return the given argument | High | P2 |
| `Function.constant` | Create a function returning a fixed value | High | P2 |
| `Function.flip` | Reverse argument order | High | P2 |
| `Function.tupled` | Convert positional arguments to a tuple argument | High | P2 |
| `Function.untupled` | Convert a tuple argument to positional arguments | High | P2 |
| `Function.memoize` | Cache function results | High | P2 |
| `Function.once` | Execute a function at most once | Medium | P3 |
| `Function.compose` | Compose two functions | Medium | P3 |
| `Function.absurd` | Mark unreachable branches using `never` | Medium | P3 |

Priorities express implementation order; all twelve helpers are in scope. P0
remains reserved for critical issues. The linked specification defines typing,
runtime behavior, unresolved arity decisions, and acceptance coverage.

## Implementation

### Repository and build

- Preserve the existing initialization branch, instructions, Beads data, workspace configuration, and `PlannedFeatures.md`.
- Fetch the exact upstream tag, record its resolved commit, and create an upstream-based `pipe-operator` branch. Reconcile overlapping configuration files while retaining local guidance.
- Initialize upstream submodules and dependencies. Restore access to `bd` and use the existing Beads workflow.
- Retain upstream build commands and identify the built compiler as `7.0.2-pipe.0`. Leave changes uncommitted; create no hosted fork or remote publication.

### Syntax and checking

- Add a dedicated `|>` token, represented through the existing binary-expression AST. Update `_scripts/ast.json` and regenerate affected AST, enum, and serialization outputs.
- Extend precedence parsing and expression recognition for unpacked pipe inputs, including nested expressions, call arguments, parentheses, and TSX. Preserve existing spread syntax elsewhere.
- Retain the source AST and source locations through checking and editor operations.
- Reuse call resolution for overloads, generics, contextual typing, optional/rest parameters, receiver checking, and return-type inference.
- Check explicit arguments and fixed tuples using ordinary positional and arity rules. For unknown-length array spreads, check their element types against every receiving parameter position they may occupy, including rest parameters; waive only the unknown-length restriction. Keep this relaxation exclusive to pipes.
- Model invocation in control-flow analysis so call effects, assertion functions, and type predicates receive appropriate treatment.

### Emission and editor support

- Add an unconditional pipe-lowering transform for every supported JavaScript target, including ESNext.
- Use per-evaluation captures to preserve input-first evaluation, argument order, member receivers, exceptions, and recursion safety. Keep `await` and `yield` in their original execution context.
- Preserve declaration inference, comments, and useful source mappings. Emitted JavaScript must contain no pipe syntax.
- Update formatting, semantic tokens, diagnostics, hover, completion, navigation, and signature help to understand pipes and their effective arguments.
- Document the local compiler/LSP executable and VS Code workspace configuration.

### Function helper delivery

- Follow [FunctionHelpers.md](FunctionHelpers.md) for the public contracts.
- Extend the fork's standard-library `FunctionConstructor` declarations and
  account for their generation and packaging in the native compiler build.
- Deliver automatic runtime availability before helper use across supported
  targets, including ESNext; declarations alone do not provide these functions
  in JavaScript hosts. Specify installation, repeated-load and collision behavior
  before implementing the runtime integration.
- Preserve argument tuples, inferred result types, and composition compatibility
  in declarations and editor tooling. Reject invalid compositions at their source.
- Keep helper tasks distinct from completed pipe syntax, checking, emission, and
  editor tasks; integrate them into the remaining conformance, packaging, and
  final regression work.

## Validation

- Add scanner/parser tests for chaining, precedence, multiline input, nesting, TSX, unpack syntax, malformed input, and existing operator/spread regressions.
- Add checker baselines for unary and multi-argument calls, tuples, ordinary arrays, incompatible elements, overloads, generics, optional/rest parameters, receivers, contextual callbacks, and inferred declarations.
- Execute emitted JavaScript to verify single evaluation, left-first ordering, spread iteration, receiver preservation, thrown exceptions, recursive re-entry, default parameters, class fields, `await`, and `yield`.
- Add language-server tests for diagnostics, hover, completion, navigation, signature help, and formatting.
- Verify all twelve Function helpers with positive/negative type cases, inferred
  declarations, editor cases, and execution of emitted JavaScript. Include the
  helper-specific cases in [FunctionHelpers.md](FunctionHelpers.md).
- Run the build, targeted tests, full Go suite with upstream submodule tests enabled, and upstream lint checks. Review generated files and baseline changes for unrelated differences.

## Defaults and boundaries

- Enable pipes by default in this fork, using the compiler's existing TS, TSX, JS, and JSX modes.
- Implement the requested pipe functionality and twelve global Function helpers;
  preserve the remainder of the broader roadmap for future work.
- Preserve existing TypeScript behavior outside pipe expressions.
- Implementation requires network access for source and dependency downloads; shell GitHub access is currently restricted, and `bd` is currently absent from PATH.
