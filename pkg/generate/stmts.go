package generate

import (
	"go/ast"
	"go/token"
)

//go:generate go run $GOPATH/src/github.com/directxman12/envmap/cmd/buildergen/main.go -- $GOFILE

var Nothing = &statement{&ast.EmptyStmt{}}

//+envmap:render=block,*ast.BlockStmt
//+envmap:render=block,ast.Stmt
type BlockBuilder struct {
	block *ast.BlockStmt
}

func Block() *BlockBuilder {
	return &BlockBuilder{
		block: &ast.BlockStmt{},
	}
}

func (b *BlockBuilder) Then(s ...Statement) *BlockBuilder {
	for _, stmt := range s {
		b.block.List = append(b.block.List, stmt.RenderStmt())
	}
	return b
}

func Declare(d Declaration) Statement {
	return &statement{&ast.DeclStmt{Decl: d.RenderDecl()}}
}

func Labeled(lbl string, s Statement) Statement {
	return &statement{
		&ast.LabeledStmt{
			Label: ast.NewIdent(lbl),
			Stmt: s.RenderStmt(),
		},
	}
}

func SendOnChan(ch Expr, v Expr) Statement {
	return &statement{
		&ast.SendStmt{
			Chan: ch.RenderExpr(),
			Value: v.RenderExpr(),
		},
	}
}

func Increment(e Expr) Statement {
	return &statement{
		&ast.IncDecStmt{X: e.RenderExpr(), Tok: token.INC},
	}
}

func Decrement(e Expr) Statement {
	return &statement{
		&ast.IncDecStmt{X: e.RenderExpr(), Tok: token.DEC},
	}
}

//+envmap:render=stmt,ast.Stmt
type AssignBuilder struct {
	stmt *ast.AssignStmt
}
func AssignTo(targets ...Expr) *AssignBuilder {
	return &AssignBuilder{
		stmt: &ast.AssignStmt{Lhs: exprs(targets), Tok: token.ASSIGN},
	}
}
func Define(targets ...string) *AssignBuilder {
	idents := make([]ast.Expr, len(targets))
	for i, target := range targets {
		idents[i] = ast.NewIdent(target)
	}
	return &AssignBuilder{
		stmt: &ast.AssignStmt{Lhs: idents, Tok: token.DEFINE},
	}
}
func (b *AssignBuilder) Values(vals ...Expr) *AssignBuilder {
	b.stmt.Rhs = exprs(vals)
	return b
}
func (b *AssignBuilder) WithOp(op string) *AssignBuilder {
	// TODO: panic if this was a define, or unknown op?
	switch op {
	case "+":
		b.stmt.Tok = token.ADD_ASSIGN
	case "-":
		b.stmt.Tok = token.SUB_ASSIGN
	case "*":
		b.stmt.Tok = token.MUL_ASSIGN
	case "/":
		b.stmt.Tok = token.QUO_ASSIGN
	case "%":
		b.stmt.Tok = token.REM_ASSIGN
	case "&":
		b.stmt.Tok = token.AND_ASSIGN
	case "|":
		b.stmt.Tok = token.OR_ASSIGN
	case "^":
		b.stmt.Tok = token.XOR_ASSIGN
	case "<<":
		b.stmt.Tok = token.SHL_ASSIGN
	case ">>":
		b.stmt.Tok = token.SHR_ASSIGN
	case "&^":
		b.stmt.Tok = token.AND_NOT_ASSIGN
	default:
		b.stmt.Tok = token.ILLEGAL
	}
	return b
}

func Go(funcToCall CallExpr) Statement {
	return &statement{
		&ast.GoStmt{Call: funcToCall.RenderCallExpr()},
	}
}

func Defer(funcToCall CallExpr) Statement {
	return &statement{
		&ast.DeferStmt{Call: funcToCall.RenderCallExpr()},
	}
}

func Return(results ...Expr) Statement {
	return &statement{
		&ast.ReturnStmt{Results: exprs(results)},
	}
}

func Fallthrough() Statement {
	return &statement{&ast.BranchStmt{Tok: token.FALLTHROUGH}}
}

func Break() Statement {
	return &statement{&ast.BranchStmt{Tok: token.BREAK}}
}
func BreakTo(lbl string) Statement {
	return &statement{&ast.BranchStmt{Tok: token.BREAK, Label: ast.NewIdent(lbl)}}
}
func Continue() Statement {
	return &statement{&ast.BranchStmt{Tok: token.BREAK}}
}
func ContinueTo(lbl string) Statement {
	return &statement{&ast.BranchStmt{Tok: token.CONTINUE, Label: ast.NewIdent(lbl)}}
}
func Goto(lbl string) Statement {
	return &statement{&ast.BranchStmt{Tok: token.GOTO, Label: ast.NewIdent(lbl)}}
}

//+envmap:render=stmt,ast.Stmt
type IfStatementBuilder struct {
	stmt *ast.IfStmt
}
func If(init Statement, cond Expr) *IfStatementBuilder {
	var stmt ast.Stmt
	if init != nil {
		stmt = init.RenderStmt()
	}
	return &IfStatementBuilder{
		stmt: &ast.IfStmt{
			Init: stmt,
			Cond: cond.RenderExpr(),
		},
	}
}
func (b *IfStatementBuilder) Then(body BlockStatement) *IfStatementBuilder {
	b.stmt.Body = body.RenderBlockStmt()
	return b
}

func (b *IfStatementBuilder) Else(stmt Statement) *IfStatementBuilder {
	b.stmt.Else = stmt.RenderStmt()
	return b
}

//+envmap:render=cl,*ast.CaseClause
type CaseClauseBuilder struct {
	cl *ast.CaseClause
}
func When(conds ...Expr) *CaseClauseBuilder {
	return &CaseClauseBuilder{
		cl: &ast.CaseClause{
			List: exprs(conds),
		},
	}
}
func WhenType(typ TypeDefinition) *CaseClauseBuilder {
	return &CaseClauseBuilder{
		cl: &ast.CaseClause{
			List: []ast.Expr{typ.RenderExpr()},
		},
	}
}
func WhenDefault() *CaseClauseBuilder {
	return &CaseClauseBuilder{
		cl: &ast.CaseClause{},
	}
}
func (b *CaseClauseBuilder) Then(stmts ...Statement) *CaseClauseBuilder {
	for _, stmt := range stmts {
		b.cl.Body = append(b.cl.Body, stmt.RenderStmt())
	}
	return b
}

//+envmap:render=stmt,ast.Stmt
type SwitchStatementBuilder struct {
	stmt *ast.SwitchStmt
}
func Switch(init Statement, tag Expr) *SwitchStatementBuilder {
	var initStmt ast.Stmt
	var tagExpr ast.Expr
	if init != nil {
		initStmt = init.RenderStmt()
	}
	if tag != nil {
		tagExpr = tag.RenderExpr()
	}
	return &SwitchStatementBuilder{
		stmt: &ast.SwitchStmt{
			Init: initStmt,
			Tag: tagExpr,
			Body: &ast.BlockStmt{},
		},
	}
}
func (b *SwitchStatementBuilder) Case(cl CaseClause) *SwitchStatementBuilder {
	b.stmt.Body.List = append(b.stmt.Body.List, cl.RenderCaseClause())
	return b
}

//+envmap:render=stmt,ast.Stmt
type TypeSwitchStatementBuilder struct {
	stmt *ast.TypeSwitchStmt
}
func TypeSwitch(init Statement, tagStmt Statement) *TypeSwitchStatementBuilder {
	// TODO: this doesn't handle assign vs define
	var initStmt ast.Stmt
	if init != nil {
		initStmt = init.RenderStmt()
	}
	return &TypeSwitchStatementBuilder{
		stmt: &ast.TypeSwitchStmt{
			Init: initStmt,
			Assign: tagStmt.RenderStmt(),
			Body: &ast.BlockStmt{},
		},
	}
}
func (b *TypeSwitchStatementBuilder) Case(cl CaseClause) *TypeSwitchStatementBuilder {
	b.stmt.Body.List = append(b.stmt.Body.List, cl.RenderCaseClause())
	return b
}

//+envmap:render=cl,*ast.CommClause
type SelectClauseBuilder struct {
	cl *ast.CommClause
}
func WhenNoComms() *SelectClauseBuilder {
	return &SelectClauseBuilder{&ast.CommClause{}}
}

func WhenSend(ch Expr, v Expr) *SelectClauseBuilder {
	return &SelectClauseBuilder{
		&ast.CommClause{
			Comm: SendOnChan(ch, v).RenderStmt(),
		},
	}
}
func WhenRecv(val, okVal Expr, ch Expr) *SelectClauseBuilder {
	var targetExprs []Expr
	if val != nil {
		targetExprs = append(targetExprs, val)
		if okVal != nil {
			targetExprs = append(targetExprs, okVal)
		}
	} else if okVal != nil {
		targetExprs = []Expr{Ident("_"), okVal}
	}
	return &SelectClauseBuilder{
		&ast.CommClause{
			Comm: AssignTo(targetExprs...).Values(ReceiveFromChan(ch)).RenderStmt(),
		},
	}
}
func (b *SelectClauseBuilder) Then(stmts ...Statement) *SelectClauseBuilder {
	for _, stmt := range stmts {
		b.cl.Body = append(b.cl.Body, stmt.RenderStmt())
	}
	return b
}

//+envmap:render=stmt,ast.Stmt
type SelectStatementBuilder struct {
	stmt *ast.SelectStmt
}
func Select() *SelectStatementBuilder {
	return &SelectStatementBuilder{
		&ast.SelectStmt{
			Body: &ast.BlockStmt{},
		},
	}
}
func (b *SelectStatementBuilder) Case(cl SelectClause) *SelectStatementBuilder {
	b.stmt.Body.List = append(b.stmt.Body.List, cl.RenderCommClause())
	return b
}

func For(init Statement, cond Expr, post Statement, body BlockStatement) Statement {
	var initStmt, postStmt ast.Stmt
	var condExpr ast.Expr
	if init != nil {
		initStmt = init.RenderStmt()
	}
	if post != nil {
		postStmt = post.RenderStmt()
	}
	if cond != nil {
		condExpr = cond.RenderExpr()
	}
	return &statement{
		&ast.ForStmt{
			Init: initStmt,
			Cond: condExpr,
			Post: postStmt,
			Body: body.RenderBlockStmt(),
		},
	}
}

func ForRange(key, val, iterable Expr, define bool, body BlockStatement) Statement {
	// TODO: better handling of assign
	var keyExpr, valExpr ast.Expr
	var tok token.Token
	if key != nil {
		keyExpr = key.RenderExpr()
	}
	if val != nil {
		valExpr = val.RenderExpr()
		if key == nil {
			keyExpr = ast.NewIdent("_")
		}
	}
	if keyExpr != nil || valExpr != nil {
		if define {
			tok = token.DEFINE
		} else {
			tok = token.ASSIGN
		}
	}
	return &statement{
		&ast.RangeStmt{
			Key: keyExpr,
			Value: valExpr,
			Tok: tok,
			X: iterable.RenderExpr(),
			Body: body.RenderBlockStmt(),
		},
	}
}
