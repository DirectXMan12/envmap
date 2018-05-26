package traverse

import (
	"go/ast"
	"strings"
)

// ExtractCommentGroup returns all lines from the given comment
// group, stripping of comment characters and splitting block
// comments.
func ExtractCommentGroup(group *ast.CommentGroup) []string {
	if group == nil {
		return nil
	}
	var res []string
	for _, comment := range group.List {
		text := comment.Text
		switch text[1] {
		case '/':
			// single-line comment
			res = append(res, text[2:])
		case '*':
			// block comment
			res = append(res, strings.Split(text[2:len(text)-2], "\n")...)
		}
	}
	return res
}

// GenTag is a string of the form `+tagName`, optionally
// followed by `:otherName` and/or `=anything`.
// tagName to the rest of the line.
type GenTag string

// Split splits a tag into its constituent parts.
// `+name:kind=values`
func (t GenTag) Split() (name string, kind string, values string) {
	str := string(t)[1:]
	splitPoint := strings.IndexAny(str, ":=")
	if splitPoint < 0 {
		return str, "", ""
	}
	name = str[:splitPoint]
	if str[splitPoint] == ':' {
		rest := str[splitPoint+1:]
		splitPoint = strings.IndexByte(rest, '=')
		if splitPoint >= 0 {
			kind = rest[:splitPoint]
			values = rest[splitPoint+1:]
		}
	} else {
		values = str[splitPoint+1:]
	}
	return name, kind, values
}

// Label returns either `tagName` or `tagName:tagKind`.
func (t GenTag) Label() string {
	parts := strings.SplitN(string(t)[1:], "=", 2)
	return parts[0]
}


// ExtractTags fetches all "tags" from this comment group,
// The tags may have a single space before the initial plus.
// Returning a map from tag label to the full tag.
func ExtractTags(lines []string) map[string][]GenTag {
	res := make(map[string][]GenTag)
	for _, line := range lines {
		if line[0] != '+' {
			if line[0] == ' ' && line[1] == '+' {
				line = line[1:]
			} else {
				continue
			}
		}
		tag := GenTag(line)
		lbl := tag.Label()
		res[lbl] = append(res[lbl], tag)
	}
	return res
}
