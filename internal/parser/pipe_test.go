package parser_test

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/parser"
	"github.com/microsoft/typescript-go/internal/scanner"
	"github.com/microsoft/typescript-go/internal/testutil/parsetestutil"
	"gotest.tools/v3/assert"
)

func parsePipeSource(source string, kind core.ScriptKind) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/pipe.tsx",
		Path:     "/pipe.tsx",
	}, source, kind)
}

// Print explicit grouping so tests distinguish precedence from mere acceptance.
func pipeExpressionTree(node *ast.Node, source string) string {
	switch node.Kind {
	case ast.KindBinaryExpression:
		expr := node.AsBinaryExpression()
		return "(" + pipeExpressionTree(expr.Left, source) + " " + scanner.TokenToString(expr.OperatorToken.Kind) + " " + pipeExpressionTree(expr.Right, source) + ")"
	case ast.KindConditionalExpression:
		expr := node.AsConditionalExpression()
		return "(" + pipeExpressionTree(expr.Condition, source) + " ? " + pipeExpressionTree(expr.WhenTrue, source) + " : " + pipeExpressionTree(expr.WhenFalse, source) + ")"
	case ast.KindParenthesizedExpression:
		return "(" + pipeExpressionTree(node.Expression(), source) + ")"
	case ast.KindSpreadElement:
		return "..." + pipeExpressionTree(node.Expression(), source)
	case ast.KindArrayLiteralExpression:
		elements := make([]string, len(node.AsArrayLiteralExpression().Elements.Nodes))
		for i, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			elements[i] = pipeExpressionTree(element, source)
		}
		return "[" + strings.Join(elements, ", ") + "]"
	case ast.KindCallExpression:
		expr := node.AsCallExpression()
		arguments := make([]string, len(expr.Arguments.Nodes))
		for i, argument := range expr.Arguments.Nodes {
			arguments[i] = pipeExpressionTree(argument, source)
		}
		return pipeExpressionTree(expr.Expression, source) + "(" + strings.Join(arguments, ", ") + ")"
	default:
		return scanner.GetTextOfNodeFromSourceText(source, node, false /*includeTrivia*/)
	}
}

func TestPipeExpressionGrouping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		tree   string
	}{
		{"input |> f |> g;", "((input |> f) |> g)"},
		{"input + other |> f;", "((input + other) |> f)"},
		{"input |> f + g;", "(input |> (f + g))"},
		{"a < b |> f;", "((a < b) |> f)"},
		{"a | b |> f;", "((a | b) |> f)"},
		{"a && b |> f;", "((a && b) |> f)"},
		{"a || b |> f;", "((a || b) |> f)"},
		{"a ?? b |> f;", "((a ?? b) |> f)"},
		{"a |> f ?? g;", "(a |> (f ?? g))"},
		{"a |> f ? b : c;", "((a |> f) ? b : c)"},
		{"a ? b : c |> f;", "(a ? b : (c |> f))"},
		{"result = input |> f;", "(result = (input |> f))"},
		{"a |> f, b |> g;", "((a |> f) , (b |> g))"},
		{"makeValue() |> getFunction();", "(makeValue() |> getFunction())"},
		{"[a, b] |> f;", "([a, b] |> f)"},
		{"...[a, b] |> f |> g;", "((...[a, b] |> f) |> g)"},
		{"...[...Foo] |> f;", "(...[...Foo] |> f)"},
		{"(...[a, b] |> f);", "((...[a, b] |> f))"},
		{"consume(...[a, b] |> f);", "consume((...[a, b] |> f))"},
		{"[...[a, b] |> f];", "[(...[a, b] |> f)]"},
		{"...[a |> f, b] |> g;", "(...[(a |> f), b] |> g)"},
		{"input\n |> /* stage */ f\n |> g;", "((input |> f) |> g)"},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			file := parsePipeSource(test.source, core.ScriptKindTS)
			parsetestutil.CheckDiagnostics(t, file)
			assert.Equal(t, len(file.Statements.Nodes), 1)
			expression := file.Statements.Nodes[0].AsExpressionStatement().Expression
			assert.Equal(t, pipeExpressionTree(expression, test.source), test.tree)
		})
	}
}

func TestPipeExpressionContexts(t *testing.T) {
	t.Parallel()
	tests := []string{
		"const result = ...[a, b] |> fn;",
		"function f() { return ...[a, b] |> fn; }",
		"const f = () => ...[a, b] |> fn;",
		"class C { value = ...[a, b] |> fn; }",
		"value |> (x => x + 1);",
		"...[a, b] |> ((a, b) => a + b);",
		"value |> object.method; value |> object[key];",
		"const text = `${...[a, b] |> fn}`;",
		"if (value |> test) { ...[a, b] |> fn; }",
		"async function f() { ...[await a(), b] |> (await factory()); }",
		"function* f() { ...[(yield a), b] |> ((yield factory)); }",
		"...[a, ...[b, c] |> fn] |> fn;",
	}
	for _, kind := range []core.ScriptKind{core.ScriptKindTS, core.ScriptKindTSX, core.ScriptKindJS, core.ScriptKindJSX} {
		for _, source := range tests {
			t.Run(source+"/"+kind.String(), func(t *testing.T) {
				t.Parallel()
				file := parsePipeSource(source, kind)
				parsetestutil.CheckDiagnostics(t, file)
				assert.Equal(t, len(file.JSDiagnostics()), 0)
			})
		}
	}
}

func TestPipeJsxExpressions(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"<View value={...[a, b] |> fn} />;",
		"<View>{...[a, b] |> fn}</View>;",
		"<View>{value |> fn}</View>;",
		"<View {...(value |> props)} />;",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			file := parsePipeSource(source, core.ScriptKindTSX)
			parsetestutil.CheckDiagnostics(t, file)
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if ast.IsJsxExpression(node) && node.Expression() != nil {
					assert.Assert(t, node.AsJsxExpression().DotDotDotToken == nil)
				}
				return node.ForEachChild(visit)
			}
			file.AsNode().ForEachChild(visit)
		})
	}
}

func TestPipeSyntaxErrors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"...Foo |> fn;",
		"...([a, b]) |> fn;",
		"...[a, b];",
		"...[a, b] + c |> fn;",
		"...[a, b].map(fn) |> fn;",
		"value |>;",
		"value |> x => x;",
		"value |> ...[a, b] |> fn;",
		"...[a, b] | > fn;",
		"...[a, |> fn;",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			file := parsePipeSource(source, core.ScriptKindTS)
			assert.Assert(t, len(file.Diagnostics()) > 0)
			for _, diagnostic := range file.Diagnostics() {
				assert.Assert(t, diagnostic.Pos() >= 0 && diagnostic.End() <= len(source))
			}
		})
	}
}

func TestPipeSourceRanges(t *testing.T) {
	t.Parallel()
	source := "...[a, ...Foo] |> fn;"
	file := parsePipeSource(source, core.ScriptKindTS)
	parsetestutil.CheckDiagnostics(t, file)
	expression := file.Statements.Nodes[0].AsExpressionStatement().Expression
	pipe := expression.AsBinaryExpression()
	assert.Equal(t, source[expression.Pos():expression.End()], "...[a, ...Foo] |> fn")
	assert.Equal(t, source[pipe.OperatorToken.Pos():pipe.OperatorToken.End()], " |>")
	assert.Equal(t, source[pipe.Left.Pos():pipe.Left.End()], "...[a, ...Foo]")
	assert.Equal(t, pipe.Left.Parent, expression)
	assert.Equal(t, pipe.OperatorToken.Parent, expression)
	assert.Equal(t, pipe.Right.Parent, expression)
	array := pipe.Left.AsSpreadElement().Expression
	assert.Equal(t, array.Parent, pipe.Left)
	assert.Equal(t, source[array.Pos():array.End()], "[a, ...Foo]")
	assert.Equal(t, array.AsArrayLiteralExpression().Elements.Nodes[0].Parent, array)
}

func TestOrdinarySpreadsWithPipes(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"fn(...args);",
		"fn(...[a, b]);",
		"fn(...(value |> fn));",
		"[...args, ...[a, b]];",
		"const fn = (...[a, b]) => a + b;",
		"function fn(...args) { return args; }",
		"<View>{...items}</View>;",
		"<View {...props} />;",
		"type T = A | B; a | b; a || b; a |= b; a ||= b;",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			parsetestutil.CheckDiagnostics(t, parsePipeSource(source, core.ScriptKindTSX))
		})
	}
}

func TestPipeLookaheadNestedSpreads(t *testing.T) {
	t.Parallel()
	// Re-parsing nested spread literals without caching takes exponential time.
	for _, suffix := range []string{";", " |> fn;"} {
		source := strings.Repeat("[...", 40) + "[]" + strings.Repeat("]", 40) + suffix
		parsetestutil.CheckDiagnostics(t, parsePipeSource(source, core.ScriptKindTS))
	}
}
