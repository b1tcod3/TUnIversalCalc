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

// funciones y constantes reconocidas (todas en minúsculas).
var funcs = map[string]bool{
	"sqrt": true, "sin": true, "cos": true, "abs": true,
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
			start := i
			dots := 0
			for i < len(src) && (isDigit(src[i]) || src[i] == '.') {
				if src[i] == '.' {
					dots++
				}
				i++
			}
			if dots > 1 {
				return nil, fmt.Errorf("número mal formado %q (posición %d)", src[start:i], start)
			}
			if src[start] == '.' && start+1 == i {
				return nil, fmt.Errorf("número mal formado %q (posición %d)", src[start:i], start)
			}
			toks = append(toks, Token{Kind: TokNum, Text: src[start:i], Pos: start})
		case isLetter(c):
			start := i
			i++
			for i < len(src) && (isLetter(src[i]) || isDigit(src[i])) {
				i++
			}
			ident := strings.ToLower(src[start:i])
			switch {
			case consts[ident]:
				toks = append(toks, Token{Kind: TokConst, Text: ident, Pos: start})
			case funcs[ident]:
				toks = append(toks, Token{Kind: TokFunc, Text: ident, Pos: start})
			case len(ident) == 1:
				toks = append(toks, Token{Kind: TokVar, Text: ident, Pos: start})
			default:
				return nil, fmt.Errorf(
					"identificador %q no reconocido (posición %d) — usa variables de una letra, pi, inf o sqrt/sin/cos/abs",
					ident, start)
			}
		case c == '+' || c == '-' || c == '*' || c == '/' || c == '^':
			toks = append(toks, Token{Kind: TokOp, Text: string(c), Pos: i})
			i++
		case c == '(':
			toks = append(toks, Token{Kind: TokLParen, Text: "(", Pos: i})
			i++
		case c == ')':
			toks = append(toks, Token{Kind: TokRParen, Text: ")", Pos: i})
			i++
		case c == '=':
			toks = append(toks, Token{Kind: TokEq, Text: "=", Pos: i})
			i++
		default:
			return nil, fmt.Errorf("carácter no permitido: %q (posición %d)", c, i)
		}
	}
	toks = append(toks, Token{Kind: TokEOF, Pos: len(src)})
	return toks, nil
}
