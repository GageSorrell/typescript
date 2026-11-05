# Using the fork in VS Code

VS Code can load the fork's native language server through Microsoft's
[TypeScript 7 extension](https://marketplace.visualstudio.com/items?itemName=typescriptteam.native-preview).
Building VS Code is not required.

## Setup

1. Install **TypeScript 7** (`TypeScriptTeam.native-preview`) in a compatible VS Code release.
2. Build the fork from `C:\Typescript` (or the checkout's actual root):

   ```powershell
   npm run build
   ```

3. In the workspace where you want to use the fork, add these settings:

   ```json
   {
     "js/ts.experimental.useTsgo": true,
     "js/ts.tsdk.path": "C:/Typescript/built/local"
   }
   ```

   For a folder workspace, use `.vscode/settings.json`. For a saved multi-folder
   workspace, put them inside the `.code-workspace` file's `settings` object.

4. Trust the workspace, open a TypeScript or JavaScript file, and choose **Allow**
   if the extension asks to use the custom TypeScript SDK. Run
   **TypeScript: Restart TS Server** after changing the setting or rebuilding.

The SDK path names the directory containing `tsgo.exe`, not the executable itself.
The extension's `resolveTsdkPathToExe` implementation supports this local build
layout, and starts the executable with `--lsp --stdio`. Keep the library declarations
alongside the executable; the upstream build task copies them there.

Use **TypeScript: Select TypeScript Version...** to inspect the selected SDK.
The extension's output channel is named **TypeScript 7**. The locally built SDK
may display as `(local)`; use the selected path to distinguish it from the
extension's bundled compiler. Its command-line and LSP-reported version is
`7.0.2-pipe.0`. See [PipeDelivery.md](PipeDelivery.md) for build, example, local SDK
archive, and executable-level verification instructions.

## Current boundaries

The language server supports pipe diagnostics, inferred hovers, completion,
navigation, formatting, semantic tokens, and signature help. See
[PipeEditor.md](PipeEditor.md) for behavior, including ambiguous spread positions.
Use **Ctrl+Shift+Space** to request a pipe target's parameter hints.
Syntax coloring can also depend on VS Code's grammar and other extensions.

This selects the editor's language service. Project build tools must be configured
to invoke the fork separately. JavaScript lowering is implemented in `.5`, so
the compiler can emit pipe programs for all supported targets, including ESNext.
See [PipeEmission.md](PipeEmission.md) for behavior and validation.

These instructions do not change the user's installed extensions or settings.
