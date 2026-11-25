package engine

import (
	"errors"
	"fmt"
	"math/big"
)

// ValidIdentifier indica si un nombre puede ser una variable del lexer.
func ValidIdentifier(s string) bool {
	if s == "" || !isLetter(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isLetter(s[i]) && !isDigit(s[i]) {
			return false
		}
	}
	return true
}

// VariablesOf devuelve las variables de una declaración, en orden de
// aparición. Para puntos no hay variables propias.
func VariablesOf(d Decl) []string {
	switch d.Tipo {
	case DeclExpresion, DeclFuncion:
		e, err := ParseExpr(d.Raw)
		if err != nil {
			return nil
		}
		return varList(e)
	case DeclEcuacion:
		lhs, rhs, err := ParseEquation(d.Raw)
		if err != nil {
			return nil
		}
		seen := map[string]bool{}
		var out []string
		for _, v := range append(varList(lhs), varList(rhs)...) {
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
		return out
	}
	return nil
}

// ExprToPolynomial convierte el AST simbólico al álgebra exacta de
// monomios. Solo admite operaciones racionales y sqrt/raiz; las funciones
// trascendentes y las constantes irracionales quedan fuera.
func ExprToPolynomial(e Expr) (Polynomial, error) {
	switch n := e.(type) {
	case Num:
		return polynomialFromNum(n)
	case Var:
		return polynomialFromMonomial(monomialVar(n.Name)), nil
	case Const:
		return nil, fmt.Errorf("la constante %q no pertenece al motor exacto", n.Name)
	case Neg:
		return polynomialFromNeg(n)
	case Add:
		return polynomialFromAdd(n)
	case Mul:
		return polynomialFromMul(n)
	case Div:
		return polynomialFromDiv(n)
	case Pow:
		return polynomialFromPow(n)
	case Call:
		return polynomialFromCall(n)
	}
	return nil, fmt.Errorf("expresión no soportada por el motor exacto: %T", e)
}

func polynomialFromNum(n Num) (Polynomial, error) {
	r, err := ratParse(n.Text)
	if err != nil {
		return nil, err
	}
	return polynomialFromMonomial(monomialFromRat(r)), nil
}

func polynomialFromNeg(n Neg) (Polynomial, error) {
	p, err := ExprToPolynomial(n.X)
	if err != nil {
		return nil, err
	}
	return p.neg().canonical()
}

func polynomialFromAdd(n Add) (Polynomial, error) {
	var out Polynomial
	for _, t := range n.Terms {
		p, err := ExprToPolynomial(t)
		if err != nil {
			return nil, err
		}
		out, err = out.add(p)
		if err != nil {
			return nil, err
		}
	}
	return out.canonical()
}

func polynomialFromMul(n Mul) (Polynomial, error) {
	out := onePolynomial()
	for _, f := range n.Factors {
		p, err := ExprToPolynomial(f)
		if err != nil {
			return nil, err
		}
		out, err = out.mul(p)
		if err != nil {
			return nil, err
		}
	}
	return out.canonical()
}

func polynomialFromDiv(n Div) (Polynomial, error) {
	num, err := ExprToPolynomial(n.Num)
	if err != nil {
		return nil, err
	}
	den, err := ExprToPolynomial(n.Den)
	if err != nil {
		return nil, err
	}
	if len(den) == 0 {
		return nil, errors.New("división por cero")
	}
	if len(den) != 1 {
		return nil, errors.New("división por una suma: no soportada por el motor exacto")
	}
	return num.divMonomial(den[0])
}

func polynomialFromPow(n Pow) (Polynomial, error) {
	base, err := ExprToPolynomial(n.Base)
	if err != nil {
		return nil, err
	}
	exp, err := exactExponent(n.Exp)
	if err != nil {
		return nil, err
	}
	if len(base) == 0 {
		if exp.Sign() > 0 {
			return nil, nil
		}
		return nil, errors.New("0 elevado a un exponente no positivo no está definido")
	}
	if len(base) == 1 {
		m, err := base[0].pow(exp)
		if err != nil {
			return nil, err
		}
		return polynomialFromMonomial(m), nil
	}
	if !exp.IsInt() {
		return nil, errors.New("una suma elevada a un exponente fraccionario no es un polinomio")
	}
	k, ok := ratInt64(exp)
	if !ok {
		return nil, errors.New("exponente demasiado grande")
	}
	return base.powInt(k)
}

func polynomialFromCall(n Call) (Polynomial, error) {
	if n.Name != fnSqrt {
		return nil, fmt.Errorf("la función %s() no pertenece al motor exacto", n.Name)
	}
	arg, err := ExprToPolynomial(n.Arg)
	if err != nil {
		return nil, err
	}
	if len(arg) == 0 {
		return nil, nil
	}
	if len(arg) != 1 {
		return nil, errors.New("raíz de una suma: no simplificable como monomio")
	}
	m, err := arg[0].pow(big.NewRat(1, 2))
	if err != nil {
		return nil, err
	}
	return polynomialFromMonomial(m), nil
}

// exactExponent evalúa un exponente numérico sin pasar por float64.
func exactExponent(e Expr) (*big.Rat, error) {
	switch n := e.(type) {
	case Num:
		return ratParse(n.Text)
	case Neg:
		return exactNegExponent(n)
	case Add:
		return exactAddExponent(n)
	case Mul:
		return exactMulExponent(n)
	case Div:
		return exactDivExponent(n)
	case Pow:
		return exactPowExponent(n)
	case Var, Const, Call:
		return nil, errors.New("exponente simbólico no soportado")
	}
	return nil, fmt.Errorf("exponente no soportado: %T", e)
}

func exactNegExponent(n Neg) (*big.Rat, error) {
	r, err := exactExponent(n.X)
	if err != nil {
		return nil, err
	}
	return ratNeg(r), nil
}

func exactAddExponent(n Add) (*big.Rat, error) {
	out := ratZero()
	for _, t := range n.Terms {
		r, err := exactExponent(t)
		if err != nil {
			return nil, err
		}
		out = ratAdd(out, r)
	}
	return out, nil
}

func exactMulExponent(n Mul) (*big.Rat, error) {
	out := ratOne()
	for _, f := range n.Factors {
		r, err := exactExponent(f)
		if err != nil {
			return nil, err
		}
		out = ratMul(out, r)
	}
	return out, nil
}

func exactDivExponent(n Div) (*big.Rat, error) {
	a, err := exactExponent(n.Num)
	if err != nil {
		return nil, err
	}
	b, err := exactExponent(n.Den)
	if err != nil {
		return nil, err
	}
	return ratQuo(a, b)
}

func exactPowExponent(n Pow) (*big.Rat, error) {
	a, err := exactExponent(n.Base)
	if err != nil {
		return nil, err
	}
	b, err := exactExponent(n.Exp)
	if err != nil {
		return nil, err
	}
	if !b.IsInt() {
		return nil, errors.New("exponente irracional: no soportado")
	}
	k, ok := ratInt64(b)
	if !ok {
		return nil, errors.New("exponente demasiado grande")
	}
	return ratPowInt(a, k)
}
