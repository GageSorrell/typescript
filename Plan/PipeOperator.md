# TypeScript 7.0.2 fork with `|>`

## Summary

Populate `E:\TypeScriptFork` with the native compiler from the official [`typescript/v7.0.2` tag](https://github.com/microsoft/typescript-go/tree/typescript/v7.0.2), then add the pipe operator throughout parsing, type checking, JavaScript emission, and the built-in language server.

Deliver compiler source, tests, a working Windows binary, examples, and build/editor instructions.

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

## Validation

- Add scanner/parser tests for chaining, precedence, multiline input, nesting, TSX, unpack syntax, malformed input, and existing operator/spread regressions.
- Add checker baselines for unary and multi-argument calls, tuples, ordinary arrays, incompatible elements, overloads, generics, optional/rest parameters, receivers, contextual callbacks, and inferred declarations.
- Execute emitted JavaScript to verify single evaluation, left-first ordering, spread iteration, receiver preservation, thrown exceptions, recursive re-entry, default parameters, class fields, `await`, and `yield`.
- Add language-server tests for diagnostics, hover, completion, navigation, signature help, and formatting.
- Run the build, targeted tests, full Go suite with upstream submodule tests enabled, and upstream lint checks. Review generated files and baseline changes for unrelated differences.

## Defaults and boundaries

- Enable pipes by default in this fork, using the compiler's existing TS, TSX, JS, and JSX modes.
- Implement only the requested pipe functionality; preserve the broader roadmap for future work.
- Preserve existing TypeScript behavior outside pipe expressions.
- Implementation requires network access for source and dependency downloads; shell GitHub access is currently restricted, and `bd` is currently absent from PATH.
