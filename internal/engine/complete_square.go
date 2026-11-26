package engine

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"tuniversalcalc/internal/render"
)

// completeSquareCoeffs guarda los coeficientes de ax² + bx + c en una sola
// variable.
type completeSquareCoeffs struct {
	Var string
	A   *big.Rat
	B   *big.Rat
	C   *big.Rat
}

// CompletarCuadrado reescribe una expresión cuadrática ax² + bx + c como
// a(x + b/(2a))² + k y muestra el proceso y el vértice.
func CompletarCuadrado(e Expr) (ExactResult, error) {
	p, err := ExprToPolynomial(e)
	if err != nil {
		return ExactResult{}, err
	}
	p, err = p.canonical()
	if err != nil {
		return ExactResult{}, err
	}
	q, err := extractCompleteSquareCoeffs(p)
	if err != nil {
		return ExactResult{}, err
	}

	twoA := ratMul(ratFromInt64(2), q.A)
	half, err := ratQuo(q.B, twoA)
	if err != nil {
		return ExactResult{}, err
	}
	square := ratMul(half, half)
	k := ratSub(q.C, ratMul(q.A, square))
	result := completedSquareExpr(q, half, k)

	ascii := result.Ascii()
	unicode := render.Pretty(ascii)
	steps := []string{
		"Operación: completar el cuadrado",
		"Expresión: " + e.Ascii(),
		fmt.Sprintf("Coeficientes: a = %s, b = %s, c = %s",
			ratString(q.A), ratString(q.B), ratString(q.C)),
	}
	if q.A.Cmp(ratOne()) != 0 {
		if monic, err := p.divMonomial(monomialFromRat(q.A)); err == nil {
			monicAscii, _ := FormatPolynomial(monic)
			steps = append(steps, "Forma mónica: "+monicAscii)
		}
	}
	steps = append(steps,
		"Mitad del coeficiente lineal: b/(2a) = "+ratString(half),
		"Cuadrado que se suma y resta: "+ratString(square),
		"Ajuste del término independiente: c - a*(b/(2a))^2 = "+ratString(k),
		"Forma completada: "+ascii,
		fmt.Sprintf("Vértice (h, k): (%s, %s)", ratString(ratNeg(half)), ratString(k)),
		"Resultado: "+ascii,
		"Unicode: "+unicode,
	)
	return ExactResult{Ascii: ascii, Unicode: unicode, Steps: steps}, nil
}

// extractCompleteSquareCoeffs extrae a, b y c de un polinomio univariado de
// grado 2. Rechaza varias variables, exponentes no enteros y coeficientes
// irracionales.
func extractCompleteSquareCoeffs(p Polynomial) (completeSquareCoeffs, error) {
	q := completeSquareCoeffs{A: ratZero(), B: ratZero(), C: ratZero()}
	for _, m := range p {
		if m.Coeff.Sign() == 0 {
			continue
		}
		var (
			varName string
			varExp  *big.Rat
			hasVar  bool
		)
		for base, exp := range m.Factors {
			if base.Kind != BaseVar {
				return completeSquareCoeffs{}, errors.New("completar el cuadrado solo admite coeficientes racionales")
			}
			if hasVar {
				return completeSquareCoeffs{}, errors.New("completar el cuadrado solo admite una variable")
			}
			varName = base.Var
			varExp = exp
			hasVar = true
		}
		if !hasVar {
			q.C = ratAdd(q.C, m.Coeff)
			continue
		}
		if q.Var != "" && q.Var != varName {
			return completeSquareCoeffs{}, errors.New("completar el cuadrado solo admite una variable")
		}
		q.Var = varName
		if !varExp.IsInt() {
			return completeSquareCoeffs{}, errors.New("el exponente de la variable debe ser entero")
		}
		exp, ok := ratInt64(varExp)
		if !ok {
			return completeSquareCoeffs{}, errors.New("exponente de la variable demasiado grande")
		}
		switch exp {
		case 2:
			q.A = ratAdd(q.A, m.Coeff)
		case 1:
			q.B = ratAdd(q.B, m.Coeff)
		default:
			return completeSquareCoeffs{}, errors.New("completar el cuadrado requiere un polinomio de grado 2")
		}
	}
	if q.Var == "" {
		return completeSquareCoeffs{}, errors.New("completar el cuadrado requiere una variable")
	}
	if q.A.Sign() == 0 {
		return completeSquareCoeffs{}, errors.New("completar el cuadrado requiere un término de grado 2")
	}
	return q, nil
}

// completedSquareExpr construye a(x + b/(2a))² + k.
func completedSquareExpr(q completeSquareCoeffs, half, k *big.Rat) Expr {
	base := Expr(Var{Name: q.Var})
	if half.Sign() != 0 {
		base = Add{Terms: []Expr{Var{Name: q.Var}, Num{Text: ratString(half)}}}
	}
	square := Expr(Pow{Base: base, Exp: Num{Text: "2"}})

	term := square
	switch {
	case q.A.Cmp(ratOne()) == 0:
	case q.A.Cmp(ratNeg(ratOne())) == 0:
		term = Neg{X: square}
	default:
		term = Mul{Factors: []Expr{Num{Text: ratString(q.A)}, square}}
	}
	if k.Sign() == 0 {
		return term
	}
	return Add{Terms: []Expr{term, Num{Text: ratString(k)}}}
}

// parseFactorFlags interpreta las banderas de /factorizar.
func parseFactorFlags(op string) (bool, error) {
	fields := strings.Fields(op)
	cuadrado := false
	for _, field := range fields[1:] {
		if !strings.HasPrefix(field, "-") {
			continue
		}
		switch strings.ToLower(field) {
		case "-cuadrado", "--cuadrado":
			cuadrado = true
		default:
			return false, fmt.Errorf("bandera no reconocida: %q", field)
		}
	}
	return cuadrado, nil
}

// evalCompletarCuadrado ejecuta la bandera -cuadrado sobre una expresión.
func evalCompletarCuadrado(decl Decl) ([]string, error) {
	if decl.Tipo != DeclExpresion && decl.Tipo != DeclFuncion {
		return nil, errors.New("factorizar -cuadrado opera sobre una expresión o función cuadrática")
	}
	e, err := ParseExpr(decl.Raw)
	if err != nil {
		return nil, err
	}
	res, err := CompletarCuadrado(e)
	if err != nil {
		return nil, fmt.Errorf("factorizar -cuadrado: %w", err)
	}
	return res.Steps, nil
}
