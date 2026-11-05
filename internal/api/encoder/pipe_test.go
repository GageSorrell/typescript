package encoder_test

import (
	"testing"

	"github.com/microsoft/typescript-go/internal/api/encoder"
	"github.com/microsoft/typescript-go/internal/ast"
	"gotest.tools/v3/assert"
)

func TestPipeEncodingRoundTrip(t *testing.T) {
	t.Parallel()
	source := "...[a, ...values] |> fn |> next;"
	file := parseSourceFile(source)
	assert.Equal(t, len(file.Diagnostics()), 0)
	encoded, _, err := encoder.EncodeSourceFile(file)
	assert.NilError(t, err)
	decoded, err := encoder.DecodeSourceFile(encoded)
	assert.NilError(t, err)
	assert.Equal(t, decoded.Text(), source)
	expression := decoded.Statements.Nodes[0].AsExpressionStatement().Expression
	outer := expression.AsBinaryExpression()
	assert.Equal(t, outer.OperatorToken.Kind, ast.KindBarGreaterThanToken)
	inner := outer.Left.AsBinaryExpression()
	assert.Equal(t, inner.OperatorToken.Kind, ast.KindBarGreaterThanToken)
	assert.Equal(t, inner.Left.Kind, ast.KindSpreadElement)
	array := inner.Left.AsSpreadElement().Expression.AsArrayLiteralExpression()
	assert.Equal(t, len(array.Elements.Nodes), 2)
	assert.Equal(t, array.Elements.Nodes[1].Kind, ast.KindSpreadElement)
	original := file.Statements.Nodes[0].AsExpressionStatement().Expression.AsBinaryExpression().Left.AsBinaryExpression()
	assert.Equal(t, inner.OperatorToken.Loc, original.OperatorToken.Loc)
	assert.Equal(t, inner.Left.Loc, original.Left.Loc)
}

func TestPipeEncodingRejectsUpstreamProtocol(t *testing.T) {
	t.Parallel()
	encoded, _, err := encoder.EncodeSourceFile(parseSourceFile("value |> fn;"))
	assert.NilError(t, err)
	encoded[encoder.HeaderOffsetMetadata+3] = 5
	_, err = encoder.DecodeSourceFile(encoded)
	assert.ErrorContains(t, err, "unsupported protocol version 5")
}
