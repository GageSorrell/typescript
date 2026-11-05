package fourslash_test

import (
	"testing"

	"github.com/microsoft/typescript-go/internal/fourslash"
	util "github.com/microsoft/typescript-go/internal/fourslash/tests/util"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
)

func TestPipeEditorSignatureHelp(t *testing.T) {
	t.Parallel()
	const content = `
declare function twice(value: number): number;
declare function join(count: number, text: string): string;
declare function tail(a: number, b: number, text: string): string;
declare function identity<T>(value: T): T;
declare function factory(): (value: number) => string;
declare function use(fn: (value: string) => number, text: string): number;
const tuple = [1, 2] as const;
const obj = { method(value: number) { return value + 1; } };
1 |> /*unary*/twice;
...[/*first*/1, /*second*/"x"] |> join;
...[...tuple, /*tupleTail*/"x"] |> tail;
1 |> twice |> /*chain*/twice;
1 |> /*generic*/identity;
1 |> obj./*member*/method;
1 |> /*returned*/factory();
...[1, "x"] |> (/*arrow*/(a, b) => a + b);
twice(/*call*/1) |> twice;
...[twice(/*nested*/1), "x"] |> join;
...[/*contextual*/(value => value.length), "x"] |> use;
`
	f, done := fourslash.NewFourslash(t, nil, content)
	defer done()
	for _, marker := range []string{"unary", "chain", "call", "nested"} {
		f.GoToMarker(t, marker)
		f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: "twice(value: number): number", ParameterName: "value"})
	}
	f.GoToMarker(t, "first")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: "join(count: number, text: string): string", ParameterName: "count"})
	f.GoToMarker(t, "second")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{ParameterName: "text"})
	f.GoToMarker(t, "tupleTail")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: "tail(a: number, b: number, text: string): string", ParameterName: "text"})
	f.GoToMarker(t, "generic")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: "identity(value: 1): 1"})
	f.GoToMarker(t, "member")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: "method(value: number): number"})
	f.GoToMarker(t, "returned")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: "(value: number): string"})
	f.GoToMarker(t, "arrow")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: "(a: number, b: string): string"})
}

func TestPipeEditorHover(t *testing.T) {
	t.Parallel()
	const content = `
declare function identity<T>(value: T): T;
declare function select(value: number): number;
declare function select(value: string): string;
const /*result*/result = 1 |> /*generic*/identity;
"x" |> /*overload*/select;
const obj = { method<T>(value: T) { return value; } };
1 |> (obj./*method*/method);
...[(/*callback*/value => value.length), "x"] |> ((fn: (value: string) => number, value: string) => fn(value));
`
	f, done := fourslash.NewFourslash(t, nil, content)
	defer done()
	f.VerifyQuickInfoAt(t, "result", "const result: 1", "")
	f.VerifyQuickInfoAt(t, "generic", "function identity<1>(value: 1): 1", "")
	f.VerifyQuickInfoAt(t, "overload", "function select(value: string): string", "")
	f.VerifyQuickInfoAt(t, "method", "(method) method<1>(value: 1): 1", "")
	f.VerifyQuickInfoAt(t, "callback", "(parameter) value: string", "")
	f.VerifyNoErrors(t)
}

func TestPipeEditorCompletions(t *testing.T) {
	t.Parallel()
	const content = `
declare function twice(value: number): number;
declare function use(fn: (value: string) => number, text: string): number;
1 |> /*target*/;
1 |> ({ method(value: number) { return value; } })./*member*/;
...[value => value./*callback*/, "x"] |> use;
`
	f, done := fourslash.NewFourslash(t, nil, content)
	defer done()
	for marker, label := range map[string]string{"target": "twice", "member": "method", "callback": "length"} {
		f.VerifyCompletions(t, marker, &fourslash.CompletionsExpectedList{ItemDefaults: &fourslash.CompletionsExpectedItemDefaults{CommitCharacters: &util.DefaultCommitCharacters}, Items: &fourslash.CompletionsExpectedItems{Includes: []fourslash.CompletionsExpectedItem{label}}})
	}
}

func TestPipeEditorFormatting(t *testing.T) {
	t.Parallel()
	f, done := fourslash.NewFourslash(t, nil, "const result=...[1,2]|>add|>(value=>value+1);")
	defer done()
	f.FormatDocument(t, "")
	f.VerifyCurrentFileContent(t, "const result = ...[1, 2] |> add |> (value => value + 1);")
}

func TestPipeEditorSemanticTokens(t *testing.T) {
	t.Parallel()
	f, done := fourslash.NewFourslash(t, nil, "declare function twice(value: number): number;\n1 |> twice;")
	defer done()
	f.VerifySemanticTokens(t, []fourslash.SemanticToken{
		{Type: "function.declaration", Text: "twice"},
		{Type: "parameter.declaration", Text: "value"},
		{Type: "operator", Text: "|>"},
		{Type: "function", Text: "twice"},
	})
}

func TestPipeEditorNavigation(t *testing.T) {
	t.Parallel()
	f, done := fourslash.NewFourslash(t, nil, `declare function select(value: number): number;
declare function select(value: string): string;
const obj = { method(value: number) { return value; } };
"x" |> /*target*/select;
1 |> (obj./*member*/method);`)
	defer done()
	f.VerifyBaselineGoToDefinition(t, true, "target", "member")
	f.VerifyBaselineFindAllReferences(t, "target", "member")
}

func TestPipeEditorCallableSemanticTokens(t *testing.T) {
	t.Parallel()
	f, done := fourslash.NewFourslash(t, nil, "declare const fn: { (value: number): number; tag: number };\n1 |> (fn);")
	defer done()
	f.VerifySemanticTokens(t, []fourslash.SemanticToken{
		{Type: "variable.declaration.readonly", Text: "fn"},
		{Type: "parameter.declaration", Text: "value"},
		{Type: "property.declaration", Text: "tag"},
		{Type: "operator", Text: "|>"},
		{Type: "function.readonly", Text: "fn"},
	})
}

func TestPipeEditorDiagnostics(t *testing.T) {
	t.Parallel()
	f, done := fourslash.NewFourslash(t, nil, `// @strict: true
declare function twice(value: number): number;
declare function add(a: number, b: number): number;
const values = [1, 2];
const valid = ...[...values] |> add;
"bad" |> twice;
...[1, "bad"] |> add;
1 |>;
...[1, 2];`)
	defer done()
	f.VerifyBaselineNonSuggestionDiagnostics(t)
}

func TestPipeEditorSignatureTrigger(t *testing.T) {
	t.Parallel()
	f, done := fourslash.NewFourslash(t, nil, "declare function twice(value: number): number;\n1 |>/*target*/ twice;")
	defer done()
	f.GoToMarker(t, "target")
	f.VerifySignatureHelpPresent(t, &lsproto.SignatureHelpContext{TriggerKind: lsproto.SignatureHelpTriggerKindTriggerCharacter, TriggerCharacter: new(">")})
}

func TestPipeEditorInputCompletions(t *testing.T) {
	t.Parallel()
	const content = `
declare function mode(value: "red" | "blue"): void;
declare function use(value: { count: number }, text: "red" | "blue"): void;
"/*literal*/" |> mode;
...[{ /*object*/ }, "red"] |> use;
...[{ count: 1 }, "/*unpacked*/"] |> use;
`
	f, done := fourslash.NewFourslash(t, nil, content)
	defer done()
	for marker, label := range map[string]string{"literal": "red", "unpacked": "blue", "object": "count"} {
		f.VerifyCompletions(t, marker, &fourslash.CompletionsExpectedList{ItemDefaults: &fourslash.CompletionsExpectedItemDefaults{CommitCharacters: &util.DefaultCommitCharacters}, Items: &fourslash.CompletionsExpectedItems{Includes: []fourslash.CompletionsExpectedItem{label}}})
	}
}

func TestPipeEditorFormattingContexts(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, input, expected string }{
		{"unpack spacing", "const x=... [1,... values]|>fn;", "const x = ...[1, ...values] |> fn;"},
		{"multiline", "const x=1\n|>first\n|>second;", "const x = 1\n    |> first\n    |> second;"},
		{"tsx", "// @filename: pipe.tsx\nconst x=<View value={... [1,2]|>fn}>{1|>fn}</View>;", "const x = <View value={...[1, 2] |> fn}>{1 |> fn}</View>;"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f, done := fourslash.NewFourslash(t, nil, test.input)
			defer done()
			f.FormatDocument(t, "")
			f.VerifyCurrentFileContent(t, test.expected)
		})
	}
}

func TestPipeEditorIncomplete(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"1 |>", "1 |> obj.", "...[1,] |>", "...[...values] |> obj.", "1 |> (", "1 |> fn |>"} {
		t.Run(expression, func(t *testing.T) {
			t.Parallel()
			f, done := fourslash.NewFourslash(t, nil, "declare const values: number[];\n"+expression+"/*end*/")
			defer done()
			f.GoToMarker(t, "end")
			f.VerifyNoSignatureHelp(t)
			f.FormatDocument(t, "")
		})
	}
}
