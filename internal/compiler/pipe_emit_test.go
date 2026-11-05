package compiler_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/bundled"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/json"
	"github.com/microsoft/typescript-go/internal/parser"
	"github.com/microsoft/typescript-go/internal/sourcemap"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/vfs/vfstest"
)

func emitPipe(t *testing.T, source, extension string, target core.ScriptTarget) map[string]string {
	t.Helper()
	name := "/pipe." + extension
	fs := bundled.WrapFS(vfstest.FromMap(map[string]string{name: source}, true))
	host := compiler.NewCompilerHost("/", fs, bundled.LibPath(), nil, nil)
	options := &core.CompilerOptions{
		Target: target, Module: core.ModuleKindESNext, NoCheck: core.TSTrue,
		AllowJs: core.TSTrue, OutDir: "/out", SourceMap: core.TSTrue,
		Declaration: core.TSTrue, DeclarationMap: core.TSTrue,
		Jsx: core.JsxEmitReact,
	}
	p := compiler.NewProgram(compiler.ProgramOptions{
		Config: &tsoptions.ParsedCommandLine{ParsedConfig: &core.ParsedOptions{FileNames: []string{name}, CompilerOptions: options}},
		Host:   host,
	})
	file := p.GetSourceFile(name)
	if len(file.Diagnostics()) != 0 {
		t.Fatalf("syntax diagnostics: %v", file.Diagnostics())
	}
	parents := map[*ast.Node]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool { parents[node] = node.Parent; return node.ForEachChild(visit) }
	visit(file.AsNode())
	outputs := map[string]string{}
	result := p.Emit(t.Context(), compiler.EmitOptions{WriteFile: func(name, text string, data *compiler.WriteFileData) error {
		outputs[name] = text
		return nil
	}})
	if result.EmitSkipped || len(result.Diagnostics) != 0 {
		t.Fatalf("emit failed: %+v", result)
	}
	for node, parent := range parents {
		if node.Parent != parent {
			t.Fatal("emission changed an original source parent")
		}
	}
	js := outputs["/out/pipe.js"]
	if js == "" {
		t.Fatalf("missing JavaScript: %v", outputs)
	}
	emitted := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/out/pipe.js", Path: "/out/pipe.js"}, js, core.ScriptKindJS)
	if len(emitted.Diagnostics()) != 0 || len(emitted.JSDiagnostics()) != 0 {
		t.Fatalf("invalid emitted JavaScript:\n%s", js)
	}
	visit = func(node *ast.Node) bool {
		if ast.IsPipeExpression(node) {
			t.Fatalf("pipe remains in emitted JavaScript:\n%s", js)
		}
		return node.ForEachChild(visit)
	}
	visit(emitted.AsNode())
	return outputs
}

func TestPipeEmitTargets(t *testing.T) {
	t.Parallel()
	for target := core.ScriptTargetES2015; target <= core.ScriptTargetESNext; target++ {
		if target > core.ScriptTargetES2025 && target != core.ScriptTargetESNext {
			continue
		}
		t.Run(target.String(), func(t *testing.T) {
			t.Parallel()
			for _, extension := range []string{"ts", "tsx", "js", "jsx"} {
				t.Run(extension, func(t *testing.T) {
					t.Parallel()
					emitPipe(t, "const fn = n => n + 1; const result = ...[...[1, 2]] |> fn |> fn;", extension, target)
				})
			}
		})
	}
}

func TestPipeEmitRuntime(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("node"); err != nil {
		t.Fatal("Node.js is required for pipe runtime tests")
	}
	tests := map[string]string{
		"values and chains":   `let calls = 0; const first = n => { calls++; return n + 1; }; eq(2 |> first |> first, 4); eq(calls, 2); eq([1, 2] |> (a => a.length), 2); eq(...[] |> (() => 3), 3);`,
		"left first":          `const log = []; const input = () => { log.push("input"); return 3; }; const factory = () => { log.push("target"); return n => { log.push("call"); return n * 2; }; }; eq(input() |> factory() |> (n => { log.push("next"); return n; }), 6); eq(log.join(), "input,target,call,next");`,
		"array iteration":     `const log = []; const xs = { *[Symbol.iterator]() { log.push("start"); yield 2; log.push("end"); yield 3; } }; const factory = () => { log.push("target"); return (...args) => args.join(); }; eq(...[1, ...xs, 4] |> factory(), "1,2,3,4"); eq(log.join(), "start,end,target");`,
		"runtime lengths":     `const fn = (a = 7, b = 8) => a + b; const empty = [], extra = [1, 2, 3]; eq(...[...empty] |> fn, 15); eq(...[...extra] |> fn, 3);`,
		"member receiver":     `const obj = { n: 10, fn(x) { return this.n + x; } }; eq(2 |> obj.fn, 12); eq(2 |> (obj.fn), 12); eq(2 |> (obj.fn as any), 12); let keys = 0; eq(2 |> obj[(keys++, "fn")], 12); eq(keys, 1);`,
		"captured receiver":   `let obj = { n: 10, fn(x) { return this.n + x; } }; const key = () => { obj = { n: 99 }; return "fn"; }; eq(2 |> obj[key()], 12);`,
		"getter reentry":      `let nested = false; const obj = { n: 10, get fn() { if (!nested) { nested = true; eq(20 |> obj.fn, 30); } return function(x) { return this.n + x; }; } }; eq(2 |> obj.fn, 12);`,
		"recursive default":   `let entered = false; function run(value = (entered ? 10 : (entered = true, run() + 1)) |> (n => n * 2)) { return value; } eq(run(), 42);`,
		"class initializers":  `class C { n = 3; value = this.n |> (n => n + 1); static value = 2 |> (n => n * 3); } eq(new C().value, 4); eq(C.value, 6);`,
		"private and super":   `class A { fn(x) { return this.n + x; } } class B extends A { n = 10; #fn(x) { return this.n + x; } run() { return [2 |> super.fn, 3 |> this.#fn, 4 |> super["fn"]]; } } eq(new B().run().join(), "12,13,14");`,
		"lexical context":     `function run(x) { return x |> (n => this.n + arguments[0] + n); } eq(run.call({ n: 5 }, 2), 9);`,
		"exceptions":          `const log = []; try { (() => { throw "input"; })() |> (() => { log.push("target"); })(); } catch (e) { eq(e, "input"); } eq(log.length, 0); try { 1 |> (() => { throw "target"; })(); } catch (e) { eq(e, "target"); } try { 1 |> (() => { throw "call"; }) |> (n => { log.push("next"); }); } catch (e) { eq(e, "call"); } eq(log.length, 0);`,
		"shadowed intrinsics": `const Function = 1, Reflect = 2; function fn(x) { return this.n + x; } fn.call = () => { throw "own call"; }; fn.apply = fn.call; const obj = { n: 10, fn }; eq(2 |> obj.fn, 12);`,
		"arguments captured before iterator mutation": `const saved = Array.prototype[Symbol.iterator]; const factory = () => { Array.prototype[Symbol.iterator] = function*() { yield 99; }; return (a, b) => a + b; }; try { eq(...[1, 2] |> factory(), 3); } finally { Array.prototype[Symbol.iterator] = saved; }`,
		"elisions captured as undefined":              `const factory = () => { Array.prototype[0] = 99; return (a, b) => { eq(a, undefined); return b; }; }; try { eq(...[, 2] |> factory(), 2); } finally { delete Array.prototype[0]; }`,
		"await input target key":                      `const log = []; const fn = n => { log.push("call"); return n + 1; }; const obj = { n: 10, fn(x) { return this.n + x; } }; eq((await Promise.resolve(2)) |> (await Promise.resolve(fn)), 3); eq(2 |> obj[await Promise.resolve("fn")], 12); eq(log.join(), "call");`,
		"concurrent await":                            `const fn = x => x; let release; const pending = new Promise(resolve => release = resolve); async function run(x, wait) { return x |> (await wait); } const a = run(1, pending), b = run(2, Promise.resolve(fn)); eq(await b, 2); release(fn); eq(await a, 1);`,
		"yield":                                       `function* run() { return (yield "input") |> (yield "target"); } const it = run(); eq(it.next().value, "input"); eq(it.next(3).value, "target"); eq(it.next(n => n * 2).value, 6); function* keyed() { return 2 |> ({ n: 10, fn(x) { return this.n + x; } })[yield "key"]; } const k = keyed(); eq(k.next().value, "key"); eq(k.next("fn").value, 12);`,
		"optional reference":                          `const obj = { n: 10, a: { n: 20, fn(x) { return this.n + x; } } }; eq(2 |> obj?.a.fn, 22); eq(2 |> obj.a?.["fn"], 22); eq(2 |> (obj?.a.fn), 22); eq(2 |> obj?.a?.fn, 22);`,
		"optional skip":                               `let keys = 0, calls = 0; const obj = null; let threw = false; try { 2 |> obj?.a[(keys++, "fn")]; } catch (e) { threw = e instanceof TypeError; } eq(threw, true); eq(keys, 0); try { 2 |> obj?.make((calls++, 1)).fn; } catch (e) {} eq(calls, 0);`,
		"optional does not hide errors":               `let threw = false; try { 2 |> ({ a: undefined })?.a.fn; } catch (e) { threw = e instanceof TypeError; } eq(threw, true);`,
		"optional calls":                              `const obj = { n: 10, make(x) { eq(this.n, 10); return { n: x, fn(y) { return this.n + y; } }; } }; eq(2 |> obj?.make(20).fn, 22); eq(2 |> obj.make?.(30).fn, 32);`,
		"optional await and yield":                    `const obj = { n: 10, fn(x) { return this.n + x; } }; eq(2 |> obj?.[await Promise.resolve("fn")], 12); function* run() { return 2 |> obj?.[yield "key"]; } const it = run(); eq(it.next().value, "key"); eq(it.next("fn").value, 12);`,
		"optional default and field":                  `const obj = { n: 10, fn(x) { return this.n + x; } }; function run(value = 2 |> obj?.["fn"]) { return value; } class C { value = 3 |> obj?.["fn"]; } eq(run(), 12); eq(new C().value, 13);`,
		"optional recursive key":                      `let entered = false; const obj = { n: 10, fn(x) { return this.n + x; } }; function run(x) { return x |> obj?.[(entered ? "fn" : (entered = true, eq(run(20), 30), "fn"))]; } eq(run(2), 12);`,
		"nested pipes and spread errors":              `let called = false; const xs = { *[Symbol.iterator]() { yield 1; throw "iterator"; } }; try { ...[...xs] |> (() => { called = true; return x => x; })(); } catch (e) { eq(e, "iterator"); } eq(called, false); eq(...[1 |> (n => n + 1), ...[2, 3]] |> ((...ns) => ns.join()), "2,2,3");`,
		"no implicit await":                           `const promise = Promise.resolve(2); eq(promise |> (value => value), promise);`,
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, target := range []core.ScriptTarget{core.ScriptTargetES2015, core.ScriptTargetESNext} {
				t.Run(target.String(), func(t *testing.T) {
					t.Parallel()
					program := `function eq(actual, expected) { if (!Object.is(actual, expected)) throw new Error(String(actual) + " != " + String(expected)); } async function main() { ` + source + ` } main().then(() => console.log("pipe runtime ok"), e => { console.error(e); process.exitCode = 1; });`
					outputs := emitPipe(t, program, "ts", target)
					output, err := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", outputs["/out/pipe.js"]).CombinedOutput()
					if err != nil || strings.TrimSpace(string(output)) != "pipe runtime ok" {
						t.Fatalf("runtime failed: %v\n%s\n%s", err, output, outputs["/out/pipe.js"])
					}
				})
			}
		})
	}
}

func TestPipeEmitJSX(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{"tsx", "jsx"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			emitPipe(t, `const React = { createElement(tag, props, child) { return child; } }; const fn = n => n + 1; export const view = <View value={...[1, 2] |> fn}>{3 |> fn}</View>;`, extension, core.ScriptTargetESNext)
		})
	}
}

func TestPipeEmitMapsAndDeclarations(t *testing.T) {
	t.Parallel()
	source := "// pipe leading comment\nconst twice = (n: number) => n * 2;\nexport const result = /* input */ 3 |> /* target */ twice;\n"
	outputs := emitPipe(t, source, "ts", core.ScriptTargetESNext)
	for _, comment := range []string{"pipe leading comment", "input", "target"} {
		if !strings.Contains(outputs["/out/pipe.js"], comment) {
			t.Fatalf("lost comment %q:\n%s", comment, outputs["/out/pipe.js"])
		}
	}
	if !strings.Contains(outputs["/out/pipe.d.ts"], "result: number") {
		t.Fatalf("incorrect inferred declaration: %s", outputs["/out/pipe.d.ts"])
	}
	var sourceMap struct {
		Sources  []string `json:"sources"`
		Mappings string   `json:"mappings"`
	}
	if err := json.Unmarshal([]byte(outputs["/out/pipe.js.map"]), &sourceMap); err != nil || len(sourceMap.Sources) != 1 || sourceMap.Mappings == "" {
		t.Fatalf("invalid source map: %v %+v", err, sourceMap)
	}
	decoder := sourcemap.DecodeMappings(sourceMap.Mappings)
	positions := map[core.UTF16Offset]bool{}
	for mapping := range decoder.Values() {
		if mapping.IsSourceMapping() && mapping.SourceLine == 2 {
			positions[mapping.SourceCharacter] = true
		}
	}
	line := strings.Split(source, "\n")[2]
	for _, token := range []string{"3", "twice"} {
		if !positions[core.UTF16Offset(strings.Index(line, token))] {
			t.Fatalf("missing mapping for pipe operand %q: %v", token, positions)
		}
	}
	if decoder.Error() != nil {
		t.Fatal(decoder.Error())
	}
}
