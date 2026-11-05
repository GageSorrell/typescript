package estransforms

import (
	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/transformers"
)

// Captures belong to arrow invocations, rather than hoisted variables. Source
// expressions occur only in call arguments, keeping await/yield in their scope.
type pipeTransformer struct {
	transformers.Transformer
}

func NewPipeTransformer(opts *transformers.TransformOptions) *transformers.Transformer {
	tx := &pipeTransformer{}
	return tx.NewTransformer(tx.visit, opts.Context)
}

func (tx *pipeTransformer) visit(node *ast.Node) *ast.Node {
	if ast.IsPipeExpression(node) {
		return tx.pipe(node.AsBinaryExpression())
	}
	return tx.Visitor().VisitEachChild(node)
}

func (tx *pipeTransformer) call(fn *ast.Node, args ...*ast.Node) *ast.Node {
	return tx.Factory().NewCallExpression(fn, nil, nil, tx.Factory().NewNodeList(args), ast.NodeFlagsNone)
}

func (tx *pipeTransformer) arrow(params []*ast.Node, body *ast.Node) *ast.Node {
	f := tx.Factory()
	decls := make([]*ast.Node, len(params))
	for i, param := range params {
		decls[i] = f.NewParameterDeclaration(nil, nil, param, nil, nil, nil)
	}
	return f.NewArrowFunction(nil, nil, f.NewNodeList(decls), nil, nil, f.NewToken(ast.KindEqualsGreaterThanToken), body)
}

func (tx *pipeTransformer) array(values ...*ast.Node) *ast.Node {
	return tx.Factory().NewArrayLiteralExpression(tx.Factory().NewNodeList(values), false)
}

func (tx *pipeTransformer) element(value *ast.Node, index string) *ast.Node {
	f := tx.Factory()
	return f.NewElementAccessExpression(value, nil, f.NewNumericLiteral(index, ast.TokenFlagsNone), ast.NodeFlagsNone)
}

// Read call from a fresh function, so a target's own call/apply properties and
// locally shadowed Function/Reflect bindings cannot change pipe invocation.
func (tx *pipeTransformer) invoke(fn, receiver, args *ast.Node) *ast.Node {
	f := tx.Factory()
	// Apply reads captured values without iterating them again after the RHS.
	intrinsicApply := f.NewPropertyAccessExpression(tx.arrow(nil, f.NewVoidZeroExpression()), nil, f.NewIdentifier("apply"), ast.NodeFlagsNone)
	return f.NewFunctionCallCall(intrinsicApply, fn, []*ast.Node{receiver, args})
}

func (tx *pipeTransformer) pipe(node *ast.BinaryExpression) *ast.Node {
	f := tx.Factory()
	input := node.Left
	var args *ast.Node
	if ast.IsSpreadElement(input) {
		array := input.Expression().AsArrayLiteralExpression()
		values := make([]*ast.Node, len(array.Elements.Nodes))
		for i, element := range array.Elements.Nodes {
			if element.Kind == ast.KindOmittedExpression {
				values[i] = f.NewVoidZeroExpression()
				values[i].Loc = element.Loc
			} else {
				values[i] = tx.Visitor().VisitNode(element)
			}
		}
		args = f.UpdateArrayLiteralExpression(array, f.NewNodeList(values), array.MultiLine)
	} else {
		args = tx.array(tx.Visitor().VisitNode(input))
	}
	values := f.NewUniqueName("pipeArgs")
	ref := f.NewUniqueName("pipeTarget")
	invocation := tx.invoke(tx.element(ref, "1"), tx.element(ref, "0"), values)
	capture := tx.arrow([]*ast.Node{values}, tx.arrow([]*ast.Node{ref}, invocation))
	result := tx.call(tx.call(capture, args), tx.reference(node.Right))
	result.Loc = node.Loc
	tx.EmitContext().SetOriginal(result, node.AsNode())
	return result
}

// A reference is [receiver, callable]. Undefined represents a short-circuited
// optional chain; the enclosing pipe still attempts its unconditional call.
func (tx *pipeTransformer) reference(node *ast.Node) *ast.Node {
	f := tx.Factory()
	unwrapped := ast.SkipOuterExpressions(node, ast.OEKAll)
	if unwrapped.Flags&ast.NodeFlagsOptionalChain != 0 {
		return tx.chainReference(unwrapped)
	}
	if ast.IsPropertyAccessExpression(unwrapped) || ast.IsElementAccessExpression(unwrapped) {
		base := tx.Visitor().VisitNode(unwrapped.Expression())
		if base.Kind == ast.KindSuperKeyword {
			return tx.array(f.NewKeywordExpression(ast.KindThisKeyword), tx.Visitor().VisitNode(node))
		}
		obj := f.NewUniqueName("pipeReceiver")
		params, args := []*ast.Node{obj}, []*ast.Node{base}
		var member *ast.Node
		if ast.IsPropertyAccessExpression(unwrapped) {
			member = f.NewPropertyAccessExpression(obj, nil, unwrapped.Name(), ast.NodeFlagsNone)
		} else {
			key := f.NewUniqueName("pipeKey")
			params = append(params, key)
			args = append(args, tx.Visitor().VisitNode(unwrapped.AsElementAccessExpression().ArgumentExpression))
			member = f.NewElementAccessExpression(obj, nil, key, ast.NodeFlagsNone)
		}
		member.Loc = unwrapped.Loc
		tx.EmitContext().SetOriginal(member, unwrapped)
		return tx.call(tx.arrow(params, tx.array(obj, member)), args...)
	}
	return tx.array(f.NewVoidZeroExpression(), tx.Visitor().VisitNode(node))
}

func (tx *pipeTransformer) chainReference(node *ast.Node) *ast.Node {
	var links []*ast.Node
	root := node
	for {
		root = ast.SkipPartiallyEmittedExpressions(root)
		if root.Flags&ast.NodeFlagsOptionalChain == 0 || !ast.IsPropertyAccessExpression(root) && !ast.IsElementAccessExpression(root) && !ast.IsCallExpression(root) {
			break
		}
		links = append(links, root)
		root = root.Expression()
	}
	ref := tx.reference(root)
	for i := len(links) - 1; i >= 0; i-- {
		ref = tx.chainLink(ref, links[i])
	}
	return ref
}

func (tx *pipeTransformer) chainLink(previous, link *ast.Node) *ast.Node {
	f := tx.Factory()
	ref := f.NewUniqueName("pipeReference")
	value := tx.element(ref, "1")
	guard := f.NewBinaryExpression(nil, ref, nil, f.NewToken(ast.KindEqualsEqualsEqualsToken), f.NewVoidZeroExpression())
	if link.QuestionDotToken() != nil {
		isNull := f.NewBinaryExpression(nil, value, nil, f.NewToken(ast.KindEqualsEqualsEqualsToken), f.NewKeywordExpression(ast.KindNullKeyword))
		isUndefined := f.NewBinaryExpression(nil, value, nil, f.NewToken(ast.KindEqualsEqualsEqualsToken), f.NewVoidZeroExpression())
		nullish := f.NewBinaryExpression(nil, isNull, nil, f.NewToken(ast.KindBarBarToken), isUndefined)
		guard = f.NewBinaryExpression(nil, guard, nil, f.NewToken(ast.KindBarBarToken), nullish)
	}
	var body *ast.Node
	var arguments []*ast.Node
	var params []*ast.Node
	switch link.Kind {
	case ast.KindPropertyAccessExpression:
		member := f.NewPropertyAccessExpression(value, nil, link.Name(), ast.NodeFlagsNone)
		member.Loc = link.Loc
		tx.EmitContext().SetOriginal(member, link)
		body = tx.array(value, member)
	case ast.KindElementAccessExpression:
		key := f.NewUniqueName("pipeKey")
		params = []*ast.Node{key}
		arguments = []*ast.Node{tx.Visitor().VisitNode(link.AsElementAccessExpression().ArgumentExpression)}
		body = tx.array(value, f.NewElementAccessExpression(value, nil, key, ast.NodeFlagsNone))
	case ast.KindCallExpression:
		args := f.NewUniqueName("pipeCallArgs")
		params = []*ast.Node{args}
		arguments = []*ast.Node{f.NewArrayLiteralExpression(tx.Visitor().VisitNodes(link.AsCallExpression().Arguments), false)}
		body = tx.array(f.NewVoidZeroExpression(), tx.invoke(value, tx.element(ref, "0"), args))
	}
	if params != nil {
		body = tx.arrow(params, body)
	}
	stage := tx.call(tx.arrow([]*ast.Node{ref}, f.NewConditionalExpression(guard, f.NewToken(ast.KindQuestionToken), f.NewVoidZeroExpression(), f.NewToken(ast.KindColonToken), body)), previous)
	if params != nil {
		// Optional call keeps keys and call arguments outside generated functions,
		// while skipping them when the original chain short-circuits.
		stage = f.NewCallExpression(stage, f.NewToken(ast.KindQuestionDotToken), nil, f.NewNodeList(arguments), ast.NodeFlagsOptionalChain)
	}
	return stage
}
