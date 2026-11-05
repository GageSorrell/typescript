package scanner

import (
	"testing"

	"github.com/microsoft/typescript-go/internal/ast"
	"gotest.tools/v3/assert"
)

func TestScanPipeOperator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text  string
		kinds []ast.Kind
	}{
		{"|>", []ast.Kind{ast.KindBarGreaterThanToken}},
		{"value|>fn", []ast.Kind{ast.KindIdentifier, ast.KindBarGreaterThanToken, ast.KindIdentifier}},
		{"| >", []ast.Kind{ast.KindBarToken, ast.KindGreaterThanToken}},
		{"|/*comment*/>", []ast.Kind{ast.KindBarToken, ast.KindGreaterThanToken}},
		{"| || |= ||= |>", []ast.Kind{ast.KindBarToken, ast.KindBarBarToken, ast.KindBarEqualsToken, ast.KindBarBarEqualsToken, ast.KindBarGreaterThanToken}},
		{"|>>=", []ast.Kind{ast.KindBarGreaterThanToken, ast.KindGreaterThanToken, ast.KindEqualsToken}},
		{`"|>"`, []ast.Kind{ast.KindStringLiteral}},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			s := NewScanner()
			s.SetText(test.text)
			for _, kind := range test.kinds {
				assert.Equal(t, s.Scan(), kind)
				if kind == ast.KindBarGreaterThanToken {
					assert.Equal(t, s.TokenText(), "|>")
					assert.Equal(t, s.TokenEnd()-s.TokenStart(), 2)
				}
			}
			assert.Equal(t, s.Scan(), ast.KindEndOfFile)
		})
	}
	assert.Equal(t, TokenToString(ast.KindBarGreaterThanToken), "|>")
}
