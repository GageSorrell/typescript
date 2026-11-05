package ls

import (
	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/checker"
	"github.com/microsoft/typescript-go/internal/core"
)

// Only transparent wrappers preserve a pipe target's receiver and invocation.
func getPipeInvocationOfTarget(node *ast.Node) *ast.Node {
	for node.Parent != nil && ast.IsOuterExpression(node.Parent, ast.OEKAll) && node.Parent.Expression() == node {
		node = node.Parent
	}
	if ast.IsPipeExpression(node.Parent) && node.Parent.Expression() == node {
		return node.Parent
	}
	return nil
}

func getPipeArgumentListInfo(node *ast.Node, position int, c *checker.Checker) *argumentListInfo {
	var pipe *ast.Node
	if ast.IsPipeExpression(node.Parent) {
		pipe = node.Parent
	} else if ast.IsArrayLiteralExpression(node.Parent) && ast.IsSpreadElement(node.Parent.Parent) && ast.IsPipeExpression(node.Parent.Parent.Parent) && node.Parent.Parent.Parent.AsBinaryExpression().Left == node.Parent.Parent {
		pipe = node.Parent.Parent.Parent
	}
	if pipe == nil || position < pipe.Pos() || position > pipe.End() {
		return nil
	}
	sourceIndex := 0
	if ast.IsSpreadElement(pipe.AsBinaryExpression().Left) && position <= pipe.AsBinaryExpression().Left.End() {
		args := pipe.Arguments()
		sourceIndex = ast.IndexOfNode(args, node)
		if sourceIndex < 0 {
			sourceIndex = 0
			for sourceIndex < len(args) && position > args[sourceIndex].End() {
				sourceIndex++
			}
		}
	}
	index, count := c.GetPipeArgumentInfo(pipe, sourceIndex)
	return &argumentListInfo{
		invocation:    &invocation{callInvocation: &callInvocation{node: pipe}},
		argumentsSpan: core.NewTextRange(pipe.Pos(), pipe.End()),
		argumentIndex: index, argumentCount: count,
	}
}
