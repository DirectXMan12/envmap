package generate

import (
	"go/ast"
)

//go:generate go run $GOPATH/src/github.com/directxman12/envmap/cmd/buildergen/main.go -- $GOFILE

//+envmap:gen-wrapper
type declaration struct{ ast.Decl }

//+envmap:gen-wrapper
type RawImport ast.ImportSpec

//+envmap:gen-wrapper
type typeDefinition struct{ ast.Expr }

//+envmap:gen-wrapper=pointer
type RawTypeDeclaration ast.TypeSpec

//+envmap:gen-wrapper=pointer
type RawFieldGroup ast.Field

//+envmap:gen-wrapper=pointer,ast.Decl
type RawFunctionDeclaration ast.FuncDecl

//+envmap:gen-wrapper=pointer,ast.Stmt
type RawBlockStatement ast.BlockStmt

//+envmap:gen-wrapper=pointer,ast.Expr
type RawFunctionDefinition ast.FuncType

//+envmap:gen-wrapper=pointer
type RawValueDeclaration ast.ValueSpec

//+envmap:gen-wrapper=extra:ast.Stmt
type expr struct{ ast.Expr }

func (e expr) RenderStmt() ast.Stmt {
	return &ast.ExprStmt{X: e.Expr}
}

//+envmap:gen-wrapper
type statement struct{ ast.Stmt }

//+envmap:gen-wrapper=pointer,ast.Expr,extra:ast.Stmt
type RawCallExpr ast.CallExpr

func (e *RawCallExpr) RenderStmt() ast.Stmt {
	inner := ast.CallExpr(*e)
	return &ast.ExprStmt{X: &inner}
}

//+envmap:gen-wrapper=pointer
type RawCaseClause ast.CaseClause

//+envmap:gen-wrapper=pointer
type RawSelectClause ast.CommClause
