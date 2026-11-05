# Windows pipe compiler delivery

The local fork identifies itself as **7.0.2-pipe.0** through the upstream
`internal/core/version.go` mechanism. `VersionMajorMinor()` remains `7.0`, and
CLI help, the LSP initialization response, and incremental-build metadata use
the fork's full version. The normal upstream `npm run build` command is retained.

## Build and verify

Use Go 1.26 or higher, Node.js 20.19 or higher, npm, and the pinned TypeScript
submodule. Use Go 1.26 for the pinned linter, as described in
[PipeConformance.md](PipeConformance.md). From `C:\Typescript` (or the checkout's
actual root):

```powershell
git submodule update --init --recursive --depth 1
npm ci --no-audit --no-fund
go mod download
npm run build
./built/local/tsgo.exe --version
node ./scripts/Verify-PipeDelivery.mjs
```

The version command prints `Version 7.0.2-pipe.0`. Keep the generated library
declarations in `built/local` alongside `tsgo.exe` when copying the SDK.
The verification script strictly compiles [examples/pipe.ts](../examples/pipe.ts)
with declarations at ES2015, ES2022, and ESNext, runs each output in Node.js, and
checks its inferred declarations. It also launches the actual compiler with
`--lsp --stdio`, checks its reported version and a pipe result's hover, and
requires a clean shutdown. Requests have time limits and the child process is
terminated if verification fails.

## Run the example

```powershell
./built/local/tsgo.exe --ignoreConfig --strict --target esnext --module commonjs --outDir built/examples examples/pipe.ts
node ./built/examples/pipe.js
```

The program prints `pipe examples ok`. It demonstrates left-to-right unary
chains, explicit multi-argument input, readonly tuples, ordinary arrays, member
receivers, single evaluation, and promises. A plain array on the left is one
argument; `...[...values]` supplies its elements as arguments. Unknown-length
arrays are checked for compatible element types, with runtime lengths left to
JavaScript: missing values become `undefined`, defaults apply, and excess values
are passed normally. Ordinary call spread restrictions retain upstream behavior.

The pipe has lower precedence than arithmetic, comparison, logical, and nullish
operators, and higher precedence than conditional expressions and assignment.
Each stage evaluates the input once before evaluating the target, preserves a
member target's receiver, and finishes before the next stage. The RHS is any
expression yielding a callable, so `value |> factory()` invokes the returned
function. There is no implicit awaiting, currying, or argument placeholder.
See [PipeOperator.md](PipeOperator.md), [PipeArrayArguments.md](PipeArrayArguments.md),
and [PipeEmission.md](PipeEmission.md) for the full language contracts.

## Local Windows SDK archive

On Windows, build, run the delivery checks, and create a local archive:

```powershell
./scripts/Build-PipeWindows.ps1
```

The script uses the ordinary upstream build, verifies the binary, and archives
the contents of `built/local`, including the required library declarations, into
`built/pipe-compiler-7.0.2-pipe.0-win32-x64.zip` on Windows x64 (or the corresponding
`arm64` name for a native ARM64 build). It prints the archive's SHA-256 hash.
The target must be executable on the build machine because verification runs it.
The archive layout places `tsgo.exe` and library declarations at its root.

After extracting it to a directory of your choice, pass that executable to the
same delivery check from this checkout:

```powershell
node ./scripts/Verify-PipeDelivery.mjs C:/path/to/extracted-sdk/tsgo.exe
```

This creates a local SDK artifact. It does not invoke npm publishing, Microsoft
signing, VSIX creation, or hosted fork publication.

## Editor and completion boundaries

Follow [VSCodeSetup.md](VSCodeSetup.md) to select the directory containing the
local `tsgo.exe` and library declarations. The same executable serves CLI and LSP
requests. The process-level verification proves startup, version reporting, pipe
hover, and clean shutdown; visual operation in an installed VS Code client is
separate evidence and must not be inferred from that check.

`TypeScript-d4d.8` also requires the twelve global Function helpers after the
scope extension. Their implementations and integration checks remain tracked
in `TypeScript-d4d.11` through `.24`; the current pipe example and SDK checks
verify the implemented pipe functionality. Helper delivery, examples, and runtime
installation must pass those dependencies before the expanded `.8` can close.
The full regression review remains `TypeScript-d4d.9`. Bedrock, mobile, and other
host execution require their own integration evidence.
