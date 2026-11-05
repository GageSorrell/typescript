package checker

import (
	"math"
	"slices"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/diagnostics"
)

func hasPipeUnknownSpread(node *ast.Node, args []*ast.Node) bool {
	return ast.IsPipeExpression(node) && slices.ContainsFunc(args, isSpreadArgument)
}

func (c *Checker) pipeMinimumArgumentCount(args []*ast.Node) int {
	count := 0
	for _, arg := range args {
		if !isSpreadArgument(arg) && !c.pipeOptionalArguments[arg] {
			count++
		}
	}
	return count
}

// Positions are measured both from the start and from the end. End bounds
// preserve the alignment of suffixes in variadic tuple rest parameters.
type pipePositionRange struct {
	first, last       int
	endFirst, endLast int
}

type pipeParameterSlot struct {
	t *Type
	pipePositionRange
}

func addPipeBound(a, b int) int {
	if a == math.MaxInt || b == math.MaxInt {
		return math.MaxInt
	}
	return a + b
}

func pipeRangesOverlap(a, b pipePositionRange) bool {
	return a.first <= b.last && b.first <= a.last && a.endFirst <= b.endLast && b.endFirst <= a.endLast
}

func (c *Checker) pipeArgumentRanges(args []*ast.Node) []pipePositionRange {
	minimum := make([]int, len(args)+1)
	maximum := make([]int, len(args)+1)
	for i, arg := range args {
		minLength, maxLength := 1, 1
		if isSpreadArgument(arg) {
			minLength, maxLength = 0, math.MaxInt
		} else if c.pipeOptionalArguments[arg] {
			minLength = 0
		}
		minimum[i+1] = minimum[i] + minLength
		maximum[i+1] = addPipeBound(maximum[i], maxLength)
	}
	ranges := make([]pipePositionRange, len(args))
	endMin, endMax := 0, 0
	for i := len(args) - 1; i >= 0; i-- {
		arg := args[i]
		ranges[i] = pipePositionRange{minimum[i], maximum[i], endMin, endMax}
		minLength, maxLength := 1, 1
		if isSpreadArgument(arg) {
			ranges[i].last, ranges[i].endLast = math.MaxInt, math.MaxInt
			minLength, maxLength = 0, math.MaxInt
		} else if c.pipeOptionalArguments[arg] {
			minLength = 0
		}
		endMin += minLength
		endMax = addPipeBound(endMax, maxLength)
	}
	return ranges
}

func (c *Checker) pipeParameterSlots(signature *Signature) []pipeParameterSlot {
	count := len(signature.parameters)
	if signatureHasRestParameter(signature) {
		count--
	}
	var slots []pipeParameterSlot
	for i := range count {
		slots = append(slots, pipeParameterSlot{c.getTypeOfParameter(signature.parameters[i]), pipePositionRange{i, i, 0, math.MaxInt}})
	}
	if !signatureHasRestParameter(signature) {
		return slots
	}
	restType := c.getTypeOfSymbol(signature.parameters[count])
	if !isTupleType(restType) {
		return append(slots, pipeParameterSlot{c.getIndexedAccessType(restType, c.numberType), pipePositionRange{count, math.MaxInt, 0, math.MaxInt}})
	}
	infos := restType.TargetTupleType().elementInfos
	types := c.getElementTypes(restType)
	seenVariable := false
	position := count
	for i, t := range types {
		variable := infos[i].flags&ElementFlagsVariable != 0
		bounds := pipePositionRange{position, position, 0, math.MaxInt}
		if variable || seenVariable {
			bounds.last = math.MaxInt
			variableAfter := false
			for _, info := range infos[i+1:] {
				if info.flags&ElementFlagsVariable != 0 {
					variableAfter = true
				} else {
					bounds.endFirst++
				}
			}
			if !variable && !variableAfter {
				bounds.endLast = bounds.endFirst
			}
		}
		if infos[i].flags&ElementFlagsVariadic != 0 {
			t = c.getIndexedAccessType(t, c.numberType)
		} else if infos[i].flags&ElementFlagsOptional != 0 {
			t = c.getUnionType([]*Type{t, c.undefinedType})
		}
		slots = append(slots, pipeParameterSlot{t, bounds})
		if variable {
			seenVariable = true
		} else {
			position++
		}
	}
	return slots
}

func (c *Checker) pipeSpreadArgumentTargets(args []*ast.Node, signature *Signature) [][]*Type {
	ranges := c.pipeArgumentRanges(args)
	slots := c.pipeParameterSlots(signature)
	targets := make([][]*Type, len(args))
	for i, bounds := range ranges {
		for _, slot := range slots {
			if pipeRangesOverlap(bounds, slot.pipePositionRange) {
				targets[i] = append(targets[i], slot.t)
			}
		}
	}
	return targets
}

func (c *Checker) pipeSpreadContextualType(targets []*Type) *Type {
	if len(targets) == 0 {
		return nil
	}
	return c.getIntersectionType(targets)
}

func (c *Checker) checkPipeSpreadArgument(arg *ast.Node, targets []*Type, context *InferenceContext, checkMode CheckMode) *Type {
	if t := c.pipeSpreadContextualType(targets); t != nil {
		return c.checkExpressionWithContextualType(arg, t, context, checkMode)
	}
	// Excess values are ignored by the receiver, but their expressions and
	// iterable operands still need checking.
	return c.checkExpressionEx(arg, checkMode)
}

func (c *Checker) checkPipeSpreadArguments(args []*ast.Node, signature *Signature, relation *Relation, checkMode CheckMode, reportErrors bool, diagnosticOutput *[]*ast.Diagnostic) bool {
	targets := c.pipeSpreadArgumentTargets(args, signature)
	for i, arg := range args {
		argType := c.checkPipeSpreadArgument(arg, targets[i], nil, checkMode)
		if checkMode&CheckModeSkipContextSensitive != 0 {
			argType = c.getRegularTypeOfObjectLiteral(argType)
		}
		for _, target := range targets[i] {
			errorNode := c.getEffectiveCheckNode(arg)
			var location *ast.Node
			if reportErrors {
				location = errorNode
			}
			if !c.checkTypeRelatedToAndOptionallyElaborate(argType, target, relation, location, errorNode, diagnostics.Argument_of_type_0_is_not_assignable_to_parameter_of_type_1, diagnosticOutput) {
				c.maybeAddMissingAwaitInfo(arg, argType, target, relation, reportErrors, diagnosticOutput)
				return false
			}
		}
	}
	return true
}

func (c *Checker) inferPipeSpreadArguments(args []*ast.Node, signature *Signature, context *InferenceContext, checkMode CheckMode) {
	targets := c.pipeSpreadArgumentTargets(args, signature)
	for i, arg := range args {
		var argType *Type
		for _, target := range targets[i] {
			if c.couldContainTypeVariables(target) {
				if argType == nil {
					argType = c.checkPipeSpreadArgument(arg, targets[i], context, checkMode)
				}
				c.inferTypes(context.inferences, argType, target, InferencePriorityNone, false)
			}
		}
	}
	if restType := c.getNonArrayRestType(signature); restType != nil && c.couldContainTypeVariables(restType) {
		start := c.getParameterCount(signature) - 1
		restArgs := c.pipeSpreadRestArguments(args, start)
		spreadType := c.getSpreadArgumentType(restArgs, 0, len(restArgs), restType, context, checkMode)
		c.inferTypes(context.inferences, spreadType, restType, InferencePriorityNone, false)
	}
}

func (c *Checker) pipeSpreadRestArguments(args []*ast.Node, start int) []*ast.Node {
	ranges := c.pipeArgumentRanges(args)
	var restArgs []*ast.Node
	for i, arg := range args {
		if ranges[i].last >= start {
			restArgs = append(restArgs, arg)
		}
	}
	return restArgs
}

func (c *Checker) pipeSpreadInputTypeAtPosition(args []*ast.Node, position int) *Type {
	ranges := c.pipeArgumentRanges(args)
	var types []*Type
	for i, arg := range args {
		if ranges[i].first <= position && position <= ranges[i].last {
			types = append(types, c.checkExpression(arg))
		}
	}
	if len(types) == 0 {
		return nil
	}
	return c.getWidenedLiteralType(c.getUnionType(types))
}

// Flow predicates can describe a source reference only when its parameter
// position is certain. An array spread may otherwise have shifted a suffix.
func (c *Checker) pipeSpreadSourceArgumentAtPosition(args []*ast.Node, position int) *ast.Node {
	var result *ast.Node
	for i, bounds := range c.pipeArgumentRanges(args) {
		if bounds.first <= position && position <= bounds.last {
			if result != nil || isSpreadArgument(args[i]) || bounds.first != bounds.last {
				return nil
			}
			result = args[i]
		}
	}
	return result
}
