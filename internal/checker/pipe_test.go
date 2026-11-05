package checker_test

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/bundled"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/diagnosticwriter"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

func newPipeProgram(t *testing.T, source string) *compiler.Program {
	t.Helper()
	fs := bundled.WrapFS(vfstest.FromMap(map[string]string{"/pipe.ts": source}, true))
	host := compiler.NewCompilerHost("/", fs, bundled.LibPath(), nil, nil)
	p := compiler.NewProgram(compiler.ProgramOptions{
		Config: &tsoptions.ParsedCommandLine{ParsedConfig: &core.ParsedOptions{
			FileNames: []string{"/pipe.ts"},
			CompilerOptions: &core.CompilerOptions{
				Strict: core.TSTrue, Target: core.ScriptTargetES2022,
				Declaration: core.TSTrue,
			},
		}},
		Host: host,
	})
	p.BindSourceFiles()
	assert.Equal(t, len(p.GetSourceFile("/pipe.ts").Diagnostics()), 0, "pipe fixture must parse")
	return p
}

func pipeDiagnosticsText(diags []*ast.Diagnostic) string {
	var b strings.Builder
	diagnosticwriter.WriteFormatDiagnostics(&b, diagnosticwriter.FromASTDiagnostics(diags), &diagnosticwriter.FormattingOptions{NewLine: "\n"})
	return b.String()
}

func TestPipeCallChecking(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"unary chain": `declare function twice(n: number): number; declare function text(n: number): string;
            const result: string = 1 |> twice |> text;`,
		"unpacked arguments": `declare function join(n: number, s: string): string;
            const result: string = ...[1, "x"] |> join;`,
		"empty arguments": `declare function now(): number; const result: number = ...[] |> now;`,
		"single array":    `declare function sum(values: number[]): number; const result: number = [1, 2] |> sum;`,
		"tuple spread": `declare function join(n: number, s: string): string;
            const values = [1, "x"] as const; const result: string = ...[...values] |> join;`,
		"optional and rest": `declare function optional(n: number, s?: string): number;
            declare function rest(n: number, ...rest: string[]): string;
            const a: number = ...[1] |> optional;
            const b: string = ...[1, "a", "b"] |> rest;`,
		"overloads": `declare function select(n: number): number; declare function select(s: string): string;
            const a: number = 1 |> select; const b: string = "x" |> select;`,
		"generic chains": `declare function identity<T>(value: T): T; declare function box<T>(value: T): { value: T };
            const a: number = (1 |> identity |> box).value;
            const b: string = "x" |> identity<string>;`,
		"contextual callback": `declare function use(fn: (value: string) => number): number;
            const result: number = (value => value.length) |> use;`,
		"contextual unpacked callback": `declare function use(fn: (value: string) => number, value: string): number;
            const result: number = ...[value => value.length, "x"] |> use;`,
		"inline arrow":               `const result: string = 1 |> (value => value.toFixed());`,
		"inline multiple parameters": `const result: string = ...[1, "x"] |> ((n, s) => n.toFixed() + s.toUpperCase());`,
		"inline rest parameters":     `const result: number = ...[1, 2, 3] |> ((...values) => values[0] + values[1]);`,
		"factory target":             `declare function factory(): (n: number) => string; const result: string = 1 |> factory();`,
		"member receiver": `const object = { value: "x", method(this: { value: string }, n: number) { return this.value + n; } };
            const result: string = 1 |> object.method; const second: string = 2 |> (object.method);`,
		"input before target": `let fn: ((n: number) => number) | undefined;
            const result: number = (fn = n => n + 1, 1) |> fn;`,
		"predicate": `declare function isString(value: unknown): value is string;
            function use(value: unknown) { if (value |> isString) { const s: string = value; } else { const n: unknown = value; } }`,
		"unpacked predicate": `declare function isString(tag: number, value: unknown): value is string;
            function use(value: unknown) { if (...[1, value] |> isString) { const s: string = value; } }`,
		"assertion": `declare function assertString(value: unknown): asserts value is string;
            function use(value: unknown) { value |> assertString; const s: string = value; }`,
		"truthiness assertion": `declare function assert(value: unknown): asserts value;
            function use(value: string | undefined) { value |> assert; const s: string = value; }`,
		"never return": `declare function fail(value: unknown): never;
            function use(value: unknown): string { value |> fail; }`,
		"IIFE effects":              `function use() { let value: string | number = "x"; 1 |> (_ => { value = 1; }); const n: number = value; }`,
		"evolving array":            `const values = []; 1 |> values.push; const result: number[] = values;`,
		"contextual generic result": `declare function make<T>(value: number): T; const result: string = 1 |> make;`,
		"inferred predicate":        `function use(value: string | number) { if (value |> (v => typeof v === "string")) { const s: string = value; } else { const n: number = value; } }`,
		"member this predicate": `interface Box { value: unknown; isString(this: Box, n: number): this is Box & { value: string }; }
            declare const box: Box; if (1 |> box.isString) { const s: string = box.value; }`,
		"unpacked assertion": `declare function assertString(tag: number, value: unknown): asserts value is string;
            function use(value: unknown) { ...[1, value] |> assertString; const s: string = value; }`,
		"predicate alias": `declare function isString(value: unknown): value is string;
            function use(value: string | number) { const text = value |> isString; if (text) { const s: string = value; } }`,
		"negated predicate": `declare function isString(value: unknown): value is string;
            function use(value: string | number) { if (!(value |> isString)) { const n: number = value; } }`,
		"generic member receiver": `const object = { method<T>(this: { method: unknown }, value: T) { return value; } };
            const n: number = 1 |> object.method<number>;`,
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := newPipeProgram(t, source)
			diags := p.GetSemanticDiagnostics(t.Context(), p.GetSourceFile("/pipe.ts"))
			if len(diags) != 0 {
				t.Fatal(pipeDiagnosticsText(diags))
			}
		})
	}
}

func TestPipeCallErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		code   int32
	}{
		{"argument type", `declare function fn(value: string): void; 1 |> fn;`, 2345},
		{"missing argument", `declare function fn(a: number, b: number): void; 1 |> fn;`, 2554},
		{"extra argument", `declare function fn(a: number): void; ...[1, 2] |> fn;`, 2554},
		{"empty argument list", `declare function fn(a: number): void; ...[] |> fn;`, 2554},
		{"array is one argument", `declare function fn(a: number, b?: number): void; [1, 2] |> fn;`, 2345},
		{"not callable", `1 |> 2;`, 2349},
		{"possibly undefined", `declare const fn: ((n: number) => number) | undefined; 1 |> fn;`, 2722},
		{"generic constraint", `declare function fn<T extends string>(value: T): T; 1 |> fn;`, 2345},
		{"overload failure", `declare function fn(value: string): string; declare function fn(value: number): number; true |> fn;`, 2769},
		{"this receiver", `declare function fn(this: { value: string }, n: number): void; 1 |> fn;`, 2684},
		{"ordinary spread call unchanged", `declare function fn(a: number, b: number): void; const values = [1, 2]; fn(...values);`, 2556},
		{"wrong inferred result", `declare function fn<T>(value: T): T; const result: string = 1 |> fn;`, 2322},
		{"input captured before target assignment", `declare function isString(value: unknown): value is string;
            function use(value: string | number) { if (value |> (value = 1, isString)) { const wrong: string = value; } }`, 2322},
		{"input captured before target property assignment", `declare function isString(value: unknown): value is string;
            function use(box: { value: string | number }) { if (box.value |> (box.value = 1, isString)) { const wrong: string = box.value; } }`, 2322},
		{"input captured before target IIFE assignment", `declare function isString(value: unknown): value is string;
            function use(value: string | number) { if (value |> (() => { value = 1; return isString; })()) { const wrong: string = value; } }`, 2322},
		{"assertion requires explicit target annotation", `const assertString = (value: unknown): asserts value is string => {};
            function use(value: unknown) { value |> assertString; }`, 2775},
		{"unique symbols", `const a = undefined |> Symbol; const b = undefined |> Symbol; const wrong: typeof a = b;`, 2322},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			p := newPipeProgram(t, test.source)
			diags := p.GetSemanticDiagnostics(t.Context(), p.GetSourceFile("/pipe.ts"))
			for _, diagnostic := range diags {
				if diagnostic.Code() == test.code {
					return
				}
			}
			t.Fatalf("expected diagnostic %d; got:\n%s", test.code, pipeDiagnosticsText(diags))
		})
	}
}

func TestPipeDeclarationInference(t *testing.T) {
	t.Parallel()
	p := newPipeProgram(t, `declare function size(value: string): number;
        declare function identity<T>(value: T): T;
        declare function pair(a: number, b: string): { count: number; text: string };
        export const count = "abc" |> size;
        export const result = ...[1, "x"] |> pair;
        export const callbackResult = "abc" |> (value => value.length);
        export function generic<T>(value: T) { return value |> identity; }
        export const key = undefined |> Symbol;
        declare function merge<T>(a: T, b: T): T;
        const numbers = [1, 2];
        export const arrayResult = ...[...numbers] |> merge;`)
	file := p.GetSourceFile("/pipe.ts")
	parents := map[*ast.Node]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		parents[node] = node.Parent
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	diags := p.GetSemanticDiagnostics(t.Context(), file)
	assert.Equal(t, len(diags), 0, pipeDiagnosticsText(diags))
	var output string
	emit := p.Emit(t.Context(), compiler.EmitOptions{
		EmitOnly: compiler.EmitOnlyDts,
		WriteFile: func(fileName string, text string, data *compiler.WriteFileData) error {
			if strings.HasSuffix(fileName, ".d.ts") {
				output += text
			}
			return nil
		},
	})
	assert.Equal(t, len(emit.Diagnostics), 0, pipeDiagnosticsText(emit.Diagnostics))
	for _, expected := range []string{"count: number", "count: number;", "text: string;", "callbackResult: number", "generic<T>(value: T): T", "key: unique symbol", "arrayResult: number"} {
		assert.Assert(t, strings.Contains(output, expected), "missing %q in:\n%s", expected, output)
	}
	for node, parent := range parents {
		assert.Equal(t, node.Parent, parent, "checking/emit changed source parent for %s", node.Kind)
	}
}
