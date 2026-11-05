# Pipe language-server support

`TypeScript-d4d.6` integrates pipes with the native compiler's built-in language
server. Select the local build using [VSCodeSetup.md](VSCodeSetup.md), then run
**TypeScript: Restart TS Server** after rebuilding. Building VS Code is not
required.

## Editor behavior

- Diagnostics use the original pipe source, including argument errors and
  recovery for incomplete targets or unpack syntax.
- Hover shows inferred result and callback types, instantiated generic targets,
  and the selected overload. Parentheses and TypeScript assertions around a
  target preserve its invocation context.
- Completion supports target names, members, contextual callback bodies,
  contextual object inputs, and string-literal input choices. RHS completions
  are not given an input parameter's contextual type.
- Go to Definition selects the resolved target overload; Find All References
  reports original source symbols. Generated JavaScript captures are not exposed.
- Semantic tokens classify `|>` as an operator and recognize callable targets
  as invocation sites, including callable values that also have properties.
  Full and range requests retain correct UTF-8 and UTF-16 source positions.
- Document, selection, and semicolon formatting use ordinary binary-operator
  spacing preferences. Unpacked inputs, multiline chains, and TSX expressions
  retain their source syntax.

## Signature help

Use **Trigger Parameter Hints** (normally **Ctrl+Shift+Space**) on a pipe input or
target. The server also advertises `>` as a trigger for completing `|>` when its
target can be resolved. Existing call and type-argument triggers remain enabled.

An ordinary input occupies parameter zero. Explicit unpacked elements occupy
their corresponding positions. A fixed tuple spread advances later inputs by
its expanded length. Positions within a multi-element spread or after a
variable-length spread can be ambiguous; the service does not select an
arbitrary receiving parameter. It reports a null active parameter when the
client supports that capability, and otherwise uses the compiler's existing
out-of-range fallback. How the fallback is displayed depends on the client.

Nested ordinary calls retain their own signature help. A factory-call target
has its own call signature inside the factory arguments and the returned
callable's signature at the enclosing pipe. Inline RHS functions, chains,
generics, members, and contextual callbacks use the existing call resolver.

The checker exposes source-to-effective-argument positions for editor queries;
the language service does not rewrite or reparent the source tree.

## Validation

```powershell
go test ./internal/ls ./internal/lsp ./internal/format ./internal/checker -count=1 -timeout=5m
go test ./internal/fourslash/tests -run '^TestPipeEditor' -count=1 -timeout=3m
npx hereby lint
npm run build
```

Tests cover all the features above, fixed and unknown spreads, optional tuple
positions, cursor whitespace boundaries, both active-parameter capability
variants, malformed pipes, and TSX formatting. Navigation, references, and
diagnostic baselines are stored under `testdata/baselines/reference/fourslash`.
The protocol tests exchange real framed LSP messages for initialization and
semantic-token requests. Representative upstream editor tests provide regression
coverage alongside the compiler's pipe baselines.

Syntax coloring also depends on the editor's grammar and theme. Pipe JavaScript
emission is documented in [PipeEmission.md](PipeEmission.md); fork versioning and
distribution packaging remain delivery work in `.8`.
