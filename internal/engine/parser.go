package engine

import (
	"errors"
	"fmt"
	"strings"
)

// Precedencias (binding power) del parser Pratt.
const (
	bpEq      = 1  // '='
	bpAdd     = 10 // + -
	bpMul     = 20 // * / e implícito (2x, x(y+1))
	bpUnary   = 25 // unario -
	bpPow     = 30 // ^ (asociativo a la derecha)
	bpOperand = 21 // operando dentro de un producto implícito
)

// ParseExpr parsea una expresión (o una ecuación completa: devuelve el
// miembro izquierdo y se valida que sea una sola expresión sin '=').
// Se usa para /funcion y /expresion.
func ParseExpr(src string) (Expr, error) {
	if strings.TrimSpace(src) == "" {
		return nil, errors.New("expresión vacía")
	}
	p, err := newParser(src)
	if err != nil {
		return nil, err
	}
	e, err := p.parseExpr(0)
	if err != nil {
		return nil, err
	}
	t := p.peek()
	if t.Kind != TokEOF {
		return nil, fmt.Errorf("sobra %q en la posición %d — usa /ecuacion si quieres '='", t.Text, t.Pos)
	}
	return e, nil
}

// ParseEquation parsea una ecuación lhs=rhs para /ecuacion.
func ParseEquation(src string) (lhs, rhs Expr, err error) {
	p, perr := newParser(src)
	if perr != nil {
		return nil, nil, perr
	}
	lhs, err = p.parseExpr(bpEq - 1)
	if err != nil {
		return nil, nil, err
	}
	t := p.peek()
	if t.Kind != TokEq {
		return nil, nil, fmt.Errorf("se esperaba '=' en la posición %d (una ecuación lleva exactamente un '=')", t.Pos)
	}
	p.next() // consume '='
	rhs, err = p.parseExpr(0)
	if err != nil {
		return nil, nil, err
	}
	if t := p.peek(); t.Kind != TokEOF {
		return nil, nil, fmt.Errorf("sobra %q en la posición %d (una ecuación lleva exactamente un '=')", t.Text, t.Pos)
	}
	return lhs, rhs, nil
}

type parser struct {
	toks []Token
	i    int
}

func newParser(src string) (*parser, error) {
	toks, err := lex(src)
	if err != nil {
		return nil, err
	}
	return &parser{toks: toks}, nil
}

func (p *parser) peek() Token { return p.toks[p.i] }

func (p *parser) next() Token {
	t := p.toks[p.i]
	if p.i < len(p.toks)-1 {
		p.i++
	}
	return t
}

// parseExpr es el núcleo Pratt: nud (operandos) + led (operadores binarios
// y multiplicación implícita).
func (p *parser) parseExpr(minBp int) (Expr, error) {
	left, err := p.parseNud()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		switch t.Kind {
		case TokOp:
			bp, right := opBP(t.Text)
			if bp < minBp {
				return left, nil
			}
			p.next()
			rhs, err := p.parseExpr(right)
			if err != nil {
				return nil, err
			}
			left = combine(left, t.Text, rhs)
		case TokNum, TokVar, TokConst, TokFunc, TokLParen:
			// Multiplicación implícita: 2x, x(y+1), 2sin(x)
			if minBp > bpMul {
				return left, nil
			}
			rhs, err := p.parseExpr(bpOperand)
			if err != nil {
				return nil, err
			}
			left = Mul{Factors: []Expr{left, rhs}}
		case TokRParen, TokEq, TokEOF:
			return left, nil
		default:
			return nil, fmt.Errorf("token inesperado %q (posición %d)", t.Text, t.Pos)
		}
	}
}

// parseNud lee un operando: número, variable, constante, función, grupo o
// signo unario.
func (p *parser) parseNud() (Expr, error) {
	t := p.next()
	switch t.Kind {
	case TokNum:
		return Num{Text: t.Text}, nil
	case TokVar:
		return Var{Name: t.Text[0]}, nil
	case TokConst:
		return Const{Name: t.Text}, nil
	case TokFunc:
		if p.peek().Kind != TokLParen {
			return nil, fmt.Errorf("se esperaba '(' tras %s (posición %d)", t.Text, t.Pos)
		}
		p.next()
		arg, err := p.parseExpr(0)
		if err != nil {
			return nil, err
		}
		close := p.peek()
		if close.Kind != TokRParen {
			return nil, fmt.Errorf("se esperaba ')' tras el argumento de %s (posición %d)", t.Text, close.Pos)
		}
		p.next()
		return Call{Name: t.Text, Arg: arg}, nil
	case TokOp:
		switch t.Text {
		case "-":
			operand, err := p.parseExpr(bpUnary)
			if err != nil {
				return nil, err
			}
			return Neg{X: operand}, nil
		default:
			return nil, fmt.Errorf("operador %q sin operando izquierdo (posición %d)", t.Text, t.Pos)
		}
	case TokLParen:
		inner, err := p.parseExpr(0)
		if err != nil {
			return nil, err
		}
		close := p.peek()
		if close.Kind != TokRParen {
			return nil, fmt.Errorf("paréntesis sin cerrar: se esperaba ')' en la posición %d", close.Pos)
		}
		p.next()
		return inner, nil
	case TokRParen:
		return nil, fmt.Errorf("')' sobrante (posición %d)", t.Pos)
	case TokEq:
		return nil, fmt.Errorf("se esperaba una expresión antes de '=' (posición %d)", t.Pos)
	default:
		return nil, errors.New("expresión vacía o falta un operando")
	}
}

// opBP devuelve la precedencia y cómo se lee el lado derecho:
//   - '^' es asociativo a la derecha (consume la misma precedencia)
//   - el resto son asociativos a la izquierda (bp+1)
func opBP(op string) (bp, rhs int) {
	switch op {
	case "+", "-":
		return bpAdd, bpAdd + 1
	case "*", "/":
		return bpMul, bpMul + 1
	case "^":
		return bpPow, bpPow
	}
	return 0, 0
}

func combine(left Expr, op string, right Expr) Expr {
	switch op {
	case "+":
		return Add{Terms: []Expr{left, right}}
	case "-":
		return Add{Terms: []Expr{left, Neg{X: right}}}
	case "*":
		return Mul{Factors: []Expr{left, right}}
	case "/":
		return Div{Num: left, Den: right}
	case "^":
		return Pow{Base: left, Exp: right}
	}
	return left
}
