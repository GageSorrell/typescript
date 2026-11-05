# Pipe syntax implementation

This describes the syntax and AST implemented by `TypeScript-d4d.2`. Call checking,
control-flow effects, JavaScript lowering, and language-server behavior are
separate parts of the [pipe operator plan](PipeOperator.md).

## Grammar and AST

The scanner emits `BarGreaterThanToken` for adjacent `|>` characters. A pipe is a
`BinaryExpression` with its original left operand, operator token, and right
operand. Chains associate left to right. Pipe precedence is below logical and
nullish-coalescing operators, and above conditional expressions and assignment.

```ts
input + other |> first |> second;
...[a, b] |> combine;
...[...values] |> combine;
consume(...[a, b] |> combine);
value |> (x => x + 1);
```

For unpacked input, the binary expression's `Left` is a `SpreadElement` whose
`Expression` is the array literal. The literal's elements retain their original
nodes and source positions, including any inner spread elements. A plain array
on the left remains an ordinary `ArrayLiteralExpression`.

In argument lists, array literals, and JSX child expressions, a spread-prefixed
array immediately followed by a pipe is recognized as unpacked pipe input.
Existing ordinary spreads retain their original AST. In particular,
`fn(...[a, b])` remains a spread call, and `fn(...(value |> makeArray))` spreads the
pipe's result. Standalone `...values |> fn` and unpacked inputs without an
immediately following pipe produce parse diagnostics.

Parenthesized unpacked pipes are distinguished from arrow rest parameters.
Lookahead classifications are cached by source position and parser context to
avoid exponential work on deeply nested spread literals. The cache is cleared
when the pooled parser is released.

## API compatibility

The token is defined in `_scripts/ast.json`; generated Go and TypeScript AST
types, token names, enum values, and decoder classifications use that definition.
The TypeScript-facing scanner also recognizes the token.

Adding a token shifts existing `SyntaxKind` numbers. Both sides of this fork's
binary AST protocol therefore use version **6**, and reject upstream version 5
payloads. Use the compiler and API package built from this same fork together.

## Validation

```powershell
go test ./internal/scanner ./internal/parser ./internal/ast ./internal/api/encoder -count=1 -timeout=5m
node --experimental-strip-types --no-warnings --conditions=@typescript/source --test _packages/native-preview/test/pipe.test.ts _packages/native-preview/test/encoder.test.ts
node node_modules/typescript/bin/tsc -b _packages/native-preview/tsconfig.json _packages/native-preview/test/tsconfig.json
npm run build
```

Tests cover precedence and grouping, nested and multiline pipes, all four script
modes, JSX, malformed syntax, ordinary spread/operator regressions, source ranges
and parent pointers, nested-spread lookahead, serialization, and protocol mismatch
rejection. Parser regression tests also parse the pinned upstream source and test
corpus.
