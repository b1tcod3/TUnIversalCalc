package engine

import (
	"fmt"
	"strings"
)

// TokKind clasifica los tokens del lexer simbólico.
type TokKind int

const (
	TokEOF TokKind = iota
	TokNum
	TokVar
	TokConst // pi, inf
	TokFunc  // sqrt, sin, cos, abs
	TokOp    // + - * / ^
	TokLParen
	TokRParen
	TokEq
)

// Token es una unidad léxica con su posición (índice en la cadena).
type Token struct {
	Kind TokKind
	Text string
	Pos  int
}

// nombres de las funciones reconocidas por el léxico y por el motor simbólico.
const (
	fnSqrt = "sqrt"
	fnSin  = "sin"
	fnCos  = "cos"
	fnAbs  = "abs"
)

// funciones y constantes reconocidas (todas en minúsculas).
var funcs = map[string]bool{
	fnSqrt: true, fnSin: true, fnCos: true, fnAbs: true,
}

var consts = map[string]bool{
	"pi": true, "inf": true,
}

// lex convierte la expresión en tokens. Las comas y espacios se ignoran.
func lex(src string) ([]Token, error) {
	var toks []Token
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case c == ' ' || c == ',':
			i++
		case isDigit(c) || c == '.':
			tok, next, err := lexNumber(src, i)
			if err != nil {
				return nil, err
			}
			toks = append(toks, tok)
			i = next
		case isLetter(c):
			tok, next, err := lexIdent(src, i)
			if err != nil {
				return nil, err
			}
			toks = append(toks, tok)
			i = next
		default:
			tok, ok := lexSymbol(c, i)
			if !ok {
				return nil, fmt.Errorf("carácter no permitido: %q (posición %d)", c, i)
			}
			toks = append(toks, tok)
			i++
		}
	}
	toks = append(toks, Token{Kind: TokEOF, Pos: len(src)})
	return toks, nil
}

// lexNumber escanea un literal numérico (con como máximo un punto y sin
// quedar en un punto solitario).
func lexNumber(src string, start int) (Token, int, error) {
	i := start
	dots := 0
	for i < len(src) && (isDigit(src[i]) || src[i] == '.') {
		if src[i] == '.' {
			dots++
		}
		i++
	}
	text := src[start:i]
	if dots > 1 || (src[start] == '.' && start+1 == i) {
		return Token{}, 0, fmt.Errorf("número mal formado %q (posición %d)", text, start)
	}
	return Token{Kind: TokNum, Text: text, Pos: start}, i, nil
}

// lexIdent escanea un identificador y lo clasifica: constante, función o
// variable de una letra.
func lexIdent(src string, start int) (Token, int, error) {
	i := start + 1
	for i < len(src) && (isLetter(src[i]) || isDigit(src[i])) {
		i++
	}
	ident := strings.ToLower(src[start:i])
	switch {
	case consts[ident]:
		return Token{Kind: TokConst, Text: ident, Pos: start}, i, nil
	case funcs[ident]:
		return Token{Kind: TokFunc, Text: ident, Pos: start}, i, nil
	case len(ident) == 1:
		return Token{Kind: TokVar, Text: ident, Pos: start}, i, nil
	default:
		return Token{}, 0, fmt.Errorf(
			"identificador %q no reconocido (posición %d) — usa variables de una letra, pi, inf o sqrt/sin/cos/abs",
			ident, start)
	}
}

// lexSymbol clasifica un carácter de puntuación u operador.
func lexSymbol(c byte, pos int) (Token, bool) {
	switch c {
	case '+', '-', '*', '/', '^':
		return Token{Kind: TokOp, Text: string(c), Pos: pos}, true
	case '(':
		return Token{Kind: TokLParen, Text: "(", Pos: pos}, true
	case ')':
		return Token{Kind: TokRParen, Text: ")", Pos: pos}, true
	case '=':
		return Token{Kind: TokEq, Text: "=", Pos: pos}, true
	}
	return Token{}, false
}
