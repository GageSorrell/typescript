package format_test

import (
	"testing"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/format"
	"github.com/microsoft/typescript-go/internal/ls/lsutil"
	"github.com/microsoft/typescript-go/internal/parser"
)

func TestPipeFormattingOperations(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"document", "range", "semicolon"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			const input = "const result=... [1,2]|>fn;"
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/pipe.ts", Path: "/pipe.ts"}, input, core.ScriptKindTS)
			ctx := format.WithFormatCodeSettings(t.Context(), lsutil.GetDefaultFormatCodeSettings(), "\n")
			var edits []core.TextChange
			switch operation {
			case "document":
				edits = format.FormatDocument(ctx, file)
			case "range":
				edits = format.FormatSelection(ctx, file, 0, len(input))
			case "semicolon":
				edits = format.FormatOnSemicolon(ctx, file, len(input))
			}
			if actual := applyBulkEdits(input, edits); actual != "const result = ...[1, 2] |> fn;" {
				t.Fatalf("unexpected pipe formatting: %q", actual)
			}
		})
	}
}

func TestPipeFormattingWithoutOperatorSpaces(t *testing.T) {
	t.Parallel()
	const input = "input  |>  fn;"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/pipe.ts", Path: "/pipe.ts"}, input, core.ScriptKindTS)
	options := lsutil.GetDefaultFormatCodeSettings()
	options.InsertSpaceBeforeAndAfterBinaryOperators = core.TSFalse
	ctx := format.WithFormatCodeSettings(t.Context(), options, "\n")
	if actual := applyBulkEdits(input, format.FormatDocument(ctx, file)); actual != "input|>fn;" {
		t.Fatalf("unexpected compact pipe formatting: %q", actual)
	}
}
