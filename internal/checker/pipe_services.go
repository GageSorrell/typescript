package checker

import "github.com/microsoft/typescript-go/internal/ast"

// GetPipeArgumentInfo maps an original input element to its effective call
// position. A negative position means that a spread makes the position ambiguous.
// sourceIndex == len(node.Arguments()) describes the next input position.
func (c *Checker) GetPipeArgumentInfo(node *ast.Node, sourceIndex int) (argumentIndex, argumentCount int) {
	args := c.getEffectiveCallArguments(node)
	ranges := c.pipeArgumentRanges(args)
	if sourceIndex >= len(node.Arguments()) {
		if len(ranges) == 0 {
			return 0, 0
		}
		last := ranges[len(ranges)-1]
		if last.first != last.last || isSpreadArgument(args[len(args)-1]) || c.pipeOptionalArguments[args[len(args)-1]] {
			return -1, len(args)
		}
		return last.first + 1, len(args)
	}
	source := node.Arguments()[sourceIndex]
	index := -1
	for i, arg := range args {
		if arg == source || ast.IsSyntheticExpression(arg) && arg.Parent == source {
			bounds := ranges[i]
			if index >= 0 || bounds.first != bounds.last || isSpreadArgument(arg) {
				return -1, len(args)
			}
			index = bounds.first
		}
	}
	return index, len(args)
}
