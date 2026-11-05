package checker

import "github.com/microsoft/typescript-go/internal/ast"

// A pipe captures its inputs before evaluating the target. If evaluating the
// target writes an input reference, a predicate describes the captured value,
// rather than the reference's new value. Do not narrow the latter.
func (c *Checker) pipeTargetWritesArgument(pipe *ast.Node, argument *ast.Node) bool {
	argument = c.getReferenceCandidate(argument)
	var writesTarget func(*ast.Node) bool
	writesTarget = func(target *ast.Node) bool {
		target = ast.SkipParentheses(target)
		if c.isMatchingReference(argument, target) || c.containsMatchingReference(argument, target) || c.containsMatchingReference(target, argument) {
			return true
		}
		if ast.IsArrayLiteralExpression(target) || ast.IsObjectLiteralExpression(target) || ast.IsPropertyAssignment(target) || ast.IsSpreadElement(target) || ast.IsSpreadAssignment(target) {
			return target.ForEachChild(writesTarget)
		}
		return false
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) && ast.GetImmediatelyInvokedFunctionExpression(node) == nil {
			return false
		}
		switch node.Kind {
		case ast.KindBinaryExpression:
			if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) && writesTarget(node.AsBinaryExpression().Left) {
				return true
			}
		case ast.KindPrefixUnaryExpression:
			expr := node.AsPrefixUnaryExpression()
			if (expr.Operator == ast.KindPlusPlusToken || expr.Operator == ast.KindMinusMinusToken) && writesTarget(expr.Operand) {
				return true
			}
		case ast.KindPostfixUnaryExpression:
			if writesTarget(node.AsPostfixUnaryExpression().Operand) {
				return true
			}
		case ast.KindDeleteExpression:
			if writesTarget(node.Expression()) {
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	return visit(pipe.Expression())
}
