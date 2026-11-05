package ls

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/internal/bundled"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/vfs/vfstest"
)

func TestPipeSignaturePositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, input string
		index       int
	}{
		{"ordinary", "/*here*/1 |> fn", 0},
		{"array stays one value", "[/*here*/1, 2] |> fn", 0},
		{"explicit first", "...[/*here*/1, 2, 3] |> fn", 0},
		{"explicit second", "...[1, /*here*/2, 3] |> fn", 1},
		{"before comma whitespace", "...[1  /*here*/, 2] |> fn", 0},
		{"tuple before comma whitespace", "...[...tuple  /*here*/, 3] |> fn", -1},
		{"next explicit", "...[1, /*here*/] |> fn", 1},
		{"fixed tuple suffix", "...[...tuple, /*here*/3] |> fn", 2},
		{"whole tuple spread", "...[/*here*/...tuple, 3] |> fn", -1},
		{"unknown spread", "...[/*here*/...values] |> fn", -1},
		{"unknown suffix", "...[...values, /*here*/3] |> fn", -1},
		{"unknown stable prefix", "...[/*here*/1, ...values] |> fn", 0},
		{"optional tuple suffix", "...[...optional, /*here*/3, ...values] |> fn", -1},
		{"target", "1 |> /*here*/fn", 0},
		{"nested pipe", "(...[1, /*here*/2] |> fn) |> fn", 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			for _, perSignature := range []bool{false, true} {
				for _, supportsNull := range []bool{false, true} {
					source := `declare function fn(a: number, b?: number, c?: number): number;
const tuple = [1, 2] as const;
declare const optional: [number?];
declare const values: number[];
` + test.input
					position := strings.Index(source, "/*here*/")
					source = strings.Replace(source, "/*here*/", "", 1)
					fs := bundled.WrapFS(vfstest.FromMap(map[string]string{"/pipe.ts": source}, true))
					p := compiler.NewProgram(compiler.ProgramOptions{
						Config: &tsoptions.ParsedCommandLine{ParsedConfig: &core.ParsedOptions{
							FileNames: []string{"/pipe.ts"}, CompilerOptions: &core.CompilerOptions{Strict: core.TSTrue, Target: core.ScriptTargetES2022},
						}},
						Host: compiler.NewCompilerHost("/", fs, bundled.LibPath(), nil, nil),
					})
					p.BindSourceFiles()
					caps := &lsproto.ResolvedClientCapabilities{}
					caps.TextDocument.SignatureHelp.SignatureInformation.ActiveParameterSupport = perSignature
					caps.TextDocument.SignatureHelp.SignatureInformation.NoActiveParameterSupport = supportsNull
					ctx := lsproto.WithClientCapabilities(t.Context(), caps)
					l := &LanguageService{program: p}
					help := l.GetSignatureHelpItems(ctx, position, p, p.GetSourceFile("/pipe.ts"), nil)
					if help == nil || len(help.Signatures) == 0 {
						t.Fatal("missing pipe signature help")
					}
					active := help.ActiveParameter
					if perSignature {
						active = help.Signatures[0].ActiveParameter
					}
					if active == nil {
						t.Fatal("active parameter field omitted")
					}
					if test.index < 0 && supportsNull {
						if active.Uinteger != nil {
							t.Fatalf("ambiguous spread should have null active parameter, got %v", *active.Uinteger)
						}
					} else {
						expected := test.index
						if expected < 0 {
							expected = len(*help.Signatures[0].Parameters)
						}
						if active.Uinteger == nil || int(*active.Uinteger) != expected {
							t.Fatalf("expected active parameter %d, got %+v (perSignature=%v supportsNull=%v)", expected, active, perSignature, supportsNull)
						}
					}
				}
			}
		})
	}
}
