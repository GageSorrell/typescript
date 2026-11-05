# TypeScript 7.0.2 upstream baseline

This fork starts from `microsoft/typescript-go`, the native compiler source for
TypeScript 7.0.2. The Git remote `upstream` points to
`https://github.com/microsoft/typescript-go.git`.

| Item | Pinned value |
| --- | --- |
| Release tag | `typescript/v7.0.2` |
| Compiler commit | `2bd066d87f5bafd315be9f40889d0a60b9e58e0b` |
| TypeScript submodule commit | `4d4f005c8541e0255a9d8791205fdce326e462bc` |
| Working branch | `pipe-operator` |
| Preserved initialization branch | `Master` |
| Initialization commit | `61a29773c321721023441b7a999265cb02c2750a` |

The upstream tag and submodule were fetched shallowly. The original initialization
branches remain intact. Existing project instructions, Beads configuration and
hooks, workspace file, roadmap, and pipe plan are retained as local additions.
The root `.gitignore` combines upstream rules with the original Beads/Dolt rules.
No source commits, pushes, or Dolt remote synchronization are part of this setup.

## Development tools

The baseline uses Go 1.26.4, Node.js 24.19.0, and npm 12.1.0. Upstream requires
Go 1.26 or higher and Node.js 20.19 or higher. Dependencies come from the unchanged
`go.mod`, `go.sum`, and `package-lock.json` files.

In a PowerShell session, expose the locally installed Beads CLI and npm tools:

```powershell
. ./scripts/Enter-DevShell.ps1
bd prime
```

The helper changes PATH only in the current process. It uses an existing `bd` on
PATH, or the installation at `$env:LOCALAPPDATA/Programs/bd/bd.exe`.

## Build and verify

From the repository root:

```powershell
git submodule update --init --recursive --depth 1
npm ci --no-audit --no-fund
go mod download
npm run build
./built/local/tsgo.exe --version
```

`npm run build` retains the upstream build task and produces `built/local/tsgo.exe`
with library declarations alongside it. This initial build is the unmodified
upstream compiler; pipe support and the `7.0.2-pipe.0` version are later epic tasks.
See [PipeOperator.md](PipeOperator.md) for the full feature specification.
