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
	TokComma
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
	fnRaiz = "raiz"
)

// funciones y constantes reconocidas (todas en minúsculas).
var funcs = map[string]bool{
	fnSqrt: true, fnSin: true, fnCos: true, fnAbs: true, fnRaiz: true,
}

var consts = map[string]bool{
	"pi": true, "inf": true,
}

// Funciones que el usuario puede intentar usar pero que el motor todavía no
// soporta. Evitan que ln(x), log(x), etc. se interpreten como productos.
var unsupportedFuncs = map[string]bool{
	"ln": true, "log": true, "tan": true, "exp": true,
	"sec": true, "csc": true, "cot": true,
}

// lex convierte la expresión en tokens. Los espacios se ignoran; las comas
// se conservan para separar los argumentos de raiz(n, expr).
func lex(src string) ([]Token, error) {
	var toks []Token
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case c == ' ':
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
			// Dos letras sueltas siguen siendo multiplicación implícita
			// (ab = a*b). Los nombres de tres o más letras son variables
			// completas (alpha, beta, tasa).
			if tok.Kind == TokVar && len(tok.Text) == 2 &&
				isLetter(tok.Text[0]) && isLetter(tok.Text[1]) {
				toks = append(toks,
					Token{Kind: TokVar, Text: tok.Text[:1], Pos: i},
					Token{Kind: TokVar, Text: tok.Text[1:], Pos: i + 1},
				)
			} else {
				toks = append(toks, tok)
			}
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
	leadingDot := src[start] == '.' && start+1 == i
	if dots > 1 || leadingDot {
		return Token{}, 0, fmt.Errorf("número mal formado %q (posición %d)", text, start)
	}
	return Token{Kind: TokNum, Text: text, Pos: start}, i, nil
}

// lexIdent escanea un identificador y lo clasifica: constante, función,
// variable o nombre de variable de varias letras.
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
	default:
		// Un identificador de tres o más letras seguido de '(' se interpreta
		// como una función desconocida (foo(x)). Los de dos letras se parten
		// después en dos variables (ab(x) = a*b*(x)), salvo funciones
		// conocidas pero no soportadas como ln(x).
		if i < len(src) && src[i] == '(' {
			if len(ident) > 2 || unsupportedFuncs[ident] {
				return Token{}, 0, fmt.Errorf(
					"función %q no reconocida (posición %d) — usa sqrt, sin, cos, abs o raiz",
					ident, start)
			}
		}
		return Token{Kind: TokVar, Text: ident, Pos: start}, i, nil
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
	case ',':
		return Token{Kind: TokComma, Text: ",", Pos: pos}, true
	case '=':
		return Token{Kind: TokEq, Text: "=", Pos: pos}, true
	}
	return Token{}, false
}
