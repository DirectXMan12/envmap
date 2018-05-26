package generate

import (
	"go/ast"
	"go/token"
	"strings"
)

type passthroughPositioner struct {}
func (p *passthroughPositioner) Position(pos OrderingPosition, _ ast.Node) (token.Pos, token.Pos) {
	return token.Pos(pos), token.Pos(0)
}

var (
	// DefaultPositioner is a ConcretePositioner that simply uses the ordering
	// position as the actual position.  This is "good enough" for basic layout.
	DefaultPositioner ConcretePositioner = &passthroughPositioner{}
)

// OrderingPosition defines a general ordering
// of declarations, comments, and statements within
// a package.
type OrderingPosition int

// ConcretePositioner knows how to convert orderings
// to actually positioned nodes.
type ConcretePositioner interface {
	// Position converts the ordering position for a node into an actual
	// start and end position for the node.
	Position(orderPos OrderingPosition, node ast.Node) (pos, end token.Pos)
}

type tokenTracker struct {
	// nextPos tracks the next "position" in
	// the output.  On output, it can be mapped
	// to an actual position.
	nextPos OrderingPosition
	positioner ConcretePositioner
}

func (t *tokenTracker) positionNode(node ast.Node) (pos, end token.Pos) {
	pos, end = t.positioner.Position(t.nextPos, node)
	t.nextPos++
	return pos, end
}

// renderable represents builders that can be converted into a go AST node.
type renderable interface {
	// Render produces an AST node from this object.
	Render() ast.Node
}

type positionable interface {
	// TODO: make these public?
	position(*tokenTracker)
}

type Builder struct {
	tokenTracker *tokenTracker
}

func createDocComment(lines []string) *ast.CommentGroup {
	comments := make([]*ast.Comment, len(lines))
	for i, line := range lines {
		if strings.ContainsRune(line, '\n') {
			line = "/*"+line+"*/"
		} else {
			line = "//"+line
		}
		node := &ast.Comment{Text: line}
		comments[i] = node
	}
	return &ast.CommentGroup{List: comments}
}

func nameIdents(names []string) []*ast.Ident {
	if len(names) == 0 {
		return nil
	}
	nameIdents := make([]*ast.Ident, len(names))
	for i, name := range names {
		nameIdents[i] = ast.NewIdent(name)
	}
	return nameIdents
}

func exprs(exprs []Expr) []ast.Expr {
	if len(exprs) == 0 {
		return nil
	}
	astExprs := make([]ast.Expr, len(exprs))
	for i, expr := range exprs {
		astExprs[i] = expr.RenderExpr()
	}
	return astExprs
}
