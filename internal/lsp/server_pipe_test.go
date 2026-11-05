package lsp_test

import (
	"context"
	"io"
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/internal/bundled"
	"github.com/microsoft/typescript-go/internal/lsp"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/testutil/lsptestutil"
	"github.com/microsoft/typescript-go/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

func TestPipeServerCapabilitiesAndTokens(t *testing.T) {
	t.Parallel()
	for _, encoding := range []lsproto.PositionEncodingKind{lsproto.PositionEncodingKindUTF8, lsproto.PositionEncodingKindUTF16} {
		t.Run(string(encoding), func(t *testing.T) {
			t.Parallel()
			const source = "\"😀\" |> fn;\n...[1,2] |> fn;"
			fs := bundled.WrapFS(vfstest.FromMap(map[string]string{
				"/home/projects/tsconfig.json": `{}`,
				"/home/projects/pipe.ts":       source,
			}, false))
			client, closeClient := lsptestutil.NewLSPClient(t, lsp.ServerOptions{
				Err: io.Discard, Cwd: "/home/projects", FS: fs, DefaultLibraryPath: bundled.LibPath(),
			}, func(_ context.Context, req *lsproto.RequestMessage) *lsproto.ResponseMessage {
				if req.Method == lsproto.MethodClientRegisterCapability || req.Method == lsproto.MethodClientUnregisterCapability {
					return &lsproto.ResponseMessage{ID: req.ID, JSONRPC: req.JSONRPC, Result: lsproto.Null{}}
				}
				return nil
			})
			t.Cleanup(func() { _ = closeClient() })
			msg, initialized, ok := lsptestutil.SendRequest(t, client, lsproto.InitializeInfo, &lsproto.InitializeParams{
				Capabilities: &lsproto.ClientCapabilities{
					General: &lsproto.GeneralClientCapabilities{PositionEncodings: &[]lsproto.PositionEncodingKind{encoding}},
					TextDocument: &lsproto.TextDocumentClientCapabilities{SemanticTokens: &lsproto.SemanticTokensClientCapabilities{
						Requests: &lsproto.ClientSemanticTokensRequestOptions{
							Full:  &lsproto.BooleanOrClientSemanticTokensRequestFullDelta{Boolean: new(true)},
							Range: &lsproto.BooleanOrEmptyObject{Boolean: new(true)},
						},
						TokenTypes: []string{"operator"},
					}},
				},
			})
			assert.Assert(t, ok && msg.AsResponse().Error == nil)
			assert.Assert(t, slices.Contains(*initialized.Capabilities.SignatureHelpProvider.TriggerCharacters, ">"))
			lsptestutil.SendNotification(t, client, lsproto.InitializedInfo, &lsproto.InitializedParams{})
			<-client.Server.InitComplete()
			uri := lsproto.DocumentUri("file:///home/projects/pipe.ts")
			lsptestutil.SendNotification(t, client, lsproto.TextDocumentDidOpenInfo, &lsproto.DidOpenTextDocumentParams{
				TextDocument: &lsproto.TextDocumentItem{Uri: uri, LanguageId: "typescript", Text: source},
			})
			msg, full, ok := lsptestutil.SendRequest(t, client, lsproto.TextDocumentSemanticTokensFullInfo, &lsproto.SemanticTokensParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: uri}})
			assert.Assert(t, ok && msg.AsResponse().Error == nil && full.SemanticTokens != nil)
			column := uint32(5)
			if encoding == lsproto.PositionEncodingKindUTF8 {
				column = 7
			}
			assert.DeepEqual(t, full.SemanticTokens.Data, []uint32{0, column, 2, 0, 0, 1, 9, 2, 0, 0})
			msg, ranged, ok := lsptestutil.SendRequest(t, client, lsproto.TextDocumentSemanticTokensRangeInfo, &lsproto.SemanticTokensRangeParams{
				TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
				Range:        lsproto.Range{Start: lsproto.Position{Line: 1}, End: lsproto.Position{Line: 1, Character: 15}},
			})
			assert.Assert(t, ok && msg.AsResponse().Error == nil && ranged.SemanticTokens != nil)
			assert.DeepEqual(t, ranged.SemanticTokens.Data, []uint32{1, 9, 2, 0, 0})
		})
	}
}
