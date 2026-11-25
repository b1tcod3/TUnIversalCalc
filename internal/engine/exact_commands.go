package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ExactResult es el resultado del motor exacto, con sus dos
// representaciones y los pasos para el panel de salida.
type ExactResult struct {
	Poly    Polynomial
	Ascii   string
	Unicode string
	Steps   []string
}

// Potencia simplifica productos, cocientes y potencias con exponentes
// exactos.
func Potencia(e Expr) (ExactResult, error) {
	return simplifyExact(e, "potencia")
}

// Radical simplifica raíces y exponentes fraccionarios exactos.
func Radical(e Expr) (ExactResult, error) {
	return simplifyExact(e, "radical")
}

// Desarrollar expande productos y potencias enteras de sumas, mostrando la
// fórmula del producto notable cuando se reconoce.
func Desarrollar(e Expr) (ExactResult, error) {
	p, err := ExprToPolynomial(e)
	if err != nil {
		return ExactResult{}, err
	}
	p, err = p.canonical()
	if err != nil {
		return ExactResult{}, err
	}
	ascii, unicode := FormatPolynomial(p)
	steps := []string{
		"Operación: desarrollo de productos notables",
		"Expresión: " + e.Ascii(),
	}
	steps = append(steps, notableProductSteps(e)...)
	steps = append(steps,
		"Expandiendo productos y agrupando términos semejantes...",
		"Resultado: "+ascii,
		"Unicode: "+unicode,
	)
	return ExactResult{Poly: p, Ascii: ascii, Unicode: unicode, Steps: steps}, nil
}

// notableProductSteps reconoce las formas notables más habituales y
// devuelve una o dos líneas explicativas. Es solo presentación: la
// expansión real la hace ExprToPolynomial.
func notableProductSteps(e Expr) []string {
	switch n := e.(type) {
	case Pow:
		return notablePowerSteps(n)
	case Mul:
		return notableProductMulSteps(n)
	}
	return nil
}

func notablePowerSteps(n Pow) []string {
	exp, ok := exprInt(n.Exp)
	if !ok {
		return nil
	}
	base, ok := n.Base.(Add)
	if !ok {
		return nil
	}
	add := flattenAdd(base)
	switch {
	case len(add.Terms) == 2 && exp == 2:
		return []string{
			"Producto notable: cuadrado de un binomio",
			"(u ± v)^2 = u^2 ± 2*u*v + v^2",
		}
	case len(add.Terms) == 2 && exp == 3:
		return []string{
			"Producto notable: cubo de un binomio",
			"(u ± v)^3 = u^3 ± 3*u^2*v + 3*u*v^2 ± v^3",
		}
	case len(add.Terms) == 2 && exp > 3:
		return []string{
			"Producto notable: binomio de Newton",
			"(u + v)^n = Σ C(n,k)·u^(n−k)·v^k",
		}
	case len(add.Terms) == 3 && exp == 2:
		return []string{
			"Producto notable: cuadrado de un trinomio",
			"(u + v + w)^2 = u^2 + v^2 + w^2 + 2*u*v + 2*u*w + 2*v*w",
		}
	}
	return nil
}

func notableProductMulSteps(n Mul) []string {
	if len(n.Factors) != 2 {
		return nil
	}
	a0, okA := n.Factors[0].(Add)
	b0, okB := n.Factors[1].(Add)
	if !okA || !okB {
		return nil
	}
	a, b := flattenAdd(a0), flattenAdd(b0)
	if isDifferenceOfSquares(a, b) {
		return []string{
			"Producto notable: suma por diferencia",
			"(u + v)(u − v) = u^2 − v^2",
		}
	}
	return []string{
		"Producto de binomios: se expande término a término",
		"(u + a)(u + b) = u^2 + (a + b)u + a*b",
	}
}

// flattenAdd convierte la representación binaria del parser (a+b)+c en una
// lista plana de términos para reconocer trinomios y polinomios.
func flattenAdd(a Add) Add {
	var terms []Expr
	for _, t := range a.Terms {
		if sub, ok := t.(Add); ok {
			terms = append(terms, flattenAdd(sub).Terms...)
			continue
		}
		terms = append(terms, t)
	}
	return Add{Terms: terms}
}

// isDifferenceOfSquares comprueba que dos binomios sean conjugados:
// (u + v)(u - v), en cualquier orden de factores o de términos.
func isDifferenceOfSquares(a, b Add) bool {
	apos, aneg := splitAddTerms(a)
	bpos, bneg := splitAddTerms(b)
	if len(apos) == 2 && len(aneg) == 0 && len(bpos) == 1 && len(bneg) == 1 {
		return isConjugatePair(apos, bpos[0], bneg[0])
	}
	if len(bpos) == 2 && len(bneg) == 0 && len(apos) == 1 && len(aneg) == 1 {
		return isConjugatePair(bpos, apos[0], aneg[0])
	}
	return false
}

func isConjugatePair(pos []Expr, first, second Expr) bool {
	return sameExpr(pos[0], first) && sameExpr(pos[1], second) ||
		sameExpr(pos[1], first) && sameExpr(pos[0], second)
}

func splitAddTerms(a Add) (pos, neg []Expr) {
	for _, t := range a.Terms {
		if n, ok := t.(Neg); ok {
			neg = append(neg, n.X)
			continue
		}
		pos = append(pos, t)
	}
	return pos, neg
}

func sameExpr(a, b Expr) bool { return a.Ascii() == b.Ascii() }

// exprInt extrae un entero literal de un exponente.
func exprInt(e Expr) (int64, bool) {
	switch n := e.(type) {
	case Num:
		if strings.ContainsAny(n.Text, "./") {
			return 0, false
		}
		v, err := strconv.ParseInt(n.Text, 10, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	case Neg:
		v, ok := exprInt(n.X)
		return -v, ok
	}
	return 0, false
}

func simplifyExact(e Expr, op string) (ExactResult, error) {
	p, err := ExprToPolynomial(e)
	if err != nil {
		return ExactResult{}, err
	}
	p, err = p.canonical()
	if err != nil {
		return ExactResult{}, err
	}
	ascii, unicode := FormatPolynomial(p)
	label := "potencias"
	if op == "radical" {
		label = "radicales"
	}
	steps := []string{
		"Operación: simplificación exacta de " + label,
		"Expresión: " + e.Ascii(),
	}
	if op == "radical" {
		steps = append(steps, "Convirtiendo raíces a exponentes racionales y extrayendo potencias perfectas...")
	} else {
		steps = append(steps, "Aplicando producto, cociente y potencia de potencias con exponentes exactos...")
	}
	steps = append(steps,
		"Resultado: "+ascii,
		"Unicode: "+unicode,
	)
	return ExactResult{Poly: p, Ascii: ascii, Unicode: unicode, Steps: steps}, nil
}

func evalPotencia(decl Decl) ([]string, error) {
	e, err := exactDeclExpr(decl, "potencia")
	if err != nil {
		return nil, err
	}
	res, err := Potencia(e)
	if err != nil {
		return nil, fmt.Errorf("potencia: %w", err)
	}
	return res.Steps, nil
}

func evalRadical(decl Decl) ([]string, error) {
	e, err := exactDeclExpr(decl, "radical")
	if err != nil {
		return nil, err
	}
	res, err := Radical(e)
	if err != nil {
		return nil, fmt.Errorf("radical: %w", err)
	}
	return res.Steps, nil
}

func evalDesarrollar(decl Decl) ([]string, error) {
	e, err := exactDeclExpr(decl, "desarrollar")
	if err != nil {
		return nil, err
	}
	res, err := Desarrollar(e)
	if err != nil {
		return nil, fmt.Errorf("desarrollar: %w", err)
	}
	return res.Steps, nil
}

func exactDeclExpr(decl Decl, op string) (Expr, error) {
	if decl.Tipo != DeclExpresion && decl.Tipo != DeclFuncion {
		return nil, errors.New(op + " opera sobre una expresión o función: declara /expresion ...")
	}
	e, err := ParseExpr(decl.Raw)
	if err != nil {
		return nil, err
	}
	return e, nil
}
