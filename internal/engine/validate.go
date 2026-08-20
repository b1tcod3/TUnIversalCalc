package engine

// Validate es una comprobación ligera de sintaxis que usa el parser real
// (Pratt). Se mantiene como API estable para la UI: devuelve un error si la
// expresión no es parseable.
func Validate(expr string) error {
	_, err := ParseExpr(expr)
	return err
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
