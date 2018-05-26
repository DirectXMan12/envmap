package generate

import (
	"strconv"
)

func NormalString(str string) Expr {
	return LiteralString(strconv.Quote(str))
}
