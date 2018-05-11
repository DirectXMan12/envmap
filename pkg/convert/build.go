package convert

import (
	"go/ast"
)

func stringsToCommentGroup(strs ...string) *ast.CommentGroup {
	comments := make([]*ast.Comment, len(strs))
	for i, str := range strs {
		comments[i] = &ast.Comment{
			Text: str,
		}
	}
	return &ast.CommentGroup{
		List: comments,
	}
}
