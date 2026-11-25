package engine

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
)

// rationalExpr es una fracción de polinomios exactos. Den nunca es el
// polinomio cero.
type rationalExpr struct {
	Num Polynomial
	Den Polynomial
}

// RationalResult contiene la fracción simplificada y los pasos para el panel
// de salida.
type RationalResult struct {
	Num     Polynomial
	Den     Polynomial
	Ascii   string
	Unicode string
	Steps   []string
}

// factorUnit es un factor polinómico normalizado: coeff * Poly. Poly tiene
// contenido 1 y coeficiente principal positivo, de modo que dos factores
// proporcionales u opuestos se reconocen fácilmente.
type factorUnit struct {
	Poly  Polynomial
	Coeff *big.Rat
}

type rationalReduction struct {
	Result    rationalExpr
	NumUnits  []factorUnit
	DenUnits  []factorUnit
	NumSteps  []string
	DenSteps  []string
	Canceled  []Polynomial
	SignNotes []string
}

// FactorizarRacional combina una expresión racional, factoriza numerador y
// denominador, cancela los factores comunes y devuelve la forma simplificada
// sobre los racionales.
func FactorizarRacional(e Expr) (RationalResult, error) {
	r, err := exprToRational(e)
	if err != nil {
		return RationalResult{}, err
	}
	r, err = r.canonical()
	if err != nil {
		return RationalResult{}, err
	}
	if r.Den.isZero() {
		return RationalResult{}, errors.New("división por cero")
	}

	red, err := reduceRational(r)
	if err != nil {
		return RationalResult{}, err
	}

	ascii := formatRationalExpression(red.Result, false)
	unicode := formatRationalExpression(red.Result, true)
	steps := []string{
		"Operación: factorización de una expresión racional",
		"Expresión: " + e.Ascii(),
		"Fracción combinada: " + formatRationalExpression(r, false),
	}
	if red.Result.Num.isZero() {
		steps = append(steps, "0 no tiene factores", "Factorización: 0", "Unicode: 0")
		return RationalResult{
			Num:     red.Result.Num,
			Den:     red.Result.Den,
			Ascii:   "0",
			Unicode: "0",
			Steps:   steps,
		}, nil
	}

	steps = append(steps,
		"Factorización del numerador: "+formatFactorUnits(red.NumUnits, false),
		"Factorización del denominador: "+formatFactorUnits(red.DenUnits, false),
	)
	if method := factorMethodSteps(red.NumSteps); len(method) > 0 {
		steps = append(steps, "Método del numerador: "+strings.Join(method, "; "))
	}
	if method := factorMethodSteps(red.DenSteps); len(method) > 0 {
		steps = append(steps, "Método del denominador: "+strings.Join(method, "; "))
	}
	if len(red.Canceled) > 0 {
		steps = append(steps, "Factores cancelados: "+formatCanceledFactors(red.Canceled))
	} else {
		steps = append(steps, "Factores cancelados: ninguno")
	}
	steps = append(steps, red.SignNotes...)
	steps = append(steps,
		"Factorización: "+ascii,
		"Unicode: "+unicode,
		"Nota: la simplificación es válida donde el denominador original no es cero.",
	)

	return RationalResult{
		Num:     red.Result.Num,
		Den:     red.Result.Den,
		Ascii:   ascii,
		Unicode: unicode,
		Steps:   steps,
	}, nil
}

// exprHasNonConstantDivision detecta si la expresión contiene una división
// cuyo denominador no es una constante. Esas expresiones se tratan como
// racionales aunque ExprToPolynomial pudiera convertirlas usando exponentes
// negativos (por ejemplo, (x+1)/x).
func exprHasNonConstantDivision(e Expr) bool {
	switch n := e.(type) {
	case Div:
		if !exprIsConstantPolynomial(n.Den) {
			return true
		}
		return exprHasNonConstantDivision(n.Num) || exprHasNonConstantDivision(n.Den)
	case Neg:
		return exprHasNonConstantDivision(n.X)
	case Add:
		return slices.ContainsFunc(n.Terms, exprHasNonConstantDivision)
	case Mul:
		return slices.ContainsFunc(n.Factors, exprHasNonConstantDivision)
	case Pow:
		return exprHasNonConstantDivision(n.Base) || exprHasNonConstantDivision(n.Exp)
	case Call:
		return exprHasNonConstantDivision(n.Arg)
	}
	return false
}

func exprIsConstantPolynomial(e Expr) bool {
	p, err := ExprToPolynomial(e)
	if err != nil {
		return false
	}
	for _, m := range p {
		for base := range m.Factors {
			if base.Kind == BaseVar {
				return false
			}
		}
	}
	return true
}

// exprToRational convierte el AST a una fracción de polinomios. Admite las
// operaciones racionales habituales y potencias con exponente entero.
func exprToRational(e Expr) (rationalExpr, error) {
	switch n := e.(type) {
	case Num, Var:
		return rationalFromAtomic(n)
	case Const:
		return rationalExpr{}, fmt.Errorf("la constante %q no pertenece a una expresión racional", n.Name)
	case Neg:
		return rationalFromNeg(n)
	case Add:
		return rationalFromAdd(n)
	case Mul:
		return rationalFromMul(n)
	case Div:
		return rationalFromDiv(n)
	case Pow:
		return rationalPow(n)
	case Call:
		return rationalExpr{}, fmt.Errorf("la función %s() no pertenece a una expresión racional", n.Name)
	}
	return rationalExpr{}, fmt.Errorf("expresión no soportada como expresión racional: %T", e)
}

func rationalFromAtomic(e Expr) (rationalExpr, error) {
	p, err := ExprToPolynomial(e)
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalExpr{Num: p, Den: onePolynomial()}, nil
}

func rationalFromNeg(n Neg) (rationalExpr, error) {
	r, err := exprToRational(n.X)
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalExpr{Num: r.Num.neg(), Den: r.Den}.canonical()
}

func rationalFromAdd(n Add) (rationalExpr, error) {
	out := rationalExpr{Num: nil, Den: onePolynomial()}
	for _, t := range n.Terms {
		r, err := exprToRational(t)
		if err != nil {
			return rationalExpr{}, err
		}
		out, err = rationalAdd(out, r)
		if err != nil {
			return rationalExpr{}, err
		}
	}
	return out.canonical()
}

func rationalFromMul(n Mul) (rationalExpr, error) {
	out := rationalExpr{Num: onePolynomial(), Den: onePolynomial()}
	for _, f := range n.Factors {
		r, err := exprToRational(f)
		if err != nil {
			return rationalExpr{}, err
		}
		out, err = rationalMul(out, r)
		if err != nil {
			return rationalExpr{}, err
		}
	}
	return out.canonical()
}

func rationalFromDiv(n Div) (rationalExpr, error) {
	a, err := exprToRational(n.Num)
	if err != nil {
		return rationalExpr{}, err
	}
	b, err := exprToRational(n.Den)
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalDiv(a, b)
}

func rationalPow(n Pow) (rationalExpr, error) {
	base, err := exprToRational(n.Base)
	if err != nil {
		return rationalExpr{}, err
	}
	exp, err := exactExponent(n.Exp)
	if err != nil {
		return rationalExpr{}, err
	}
	if !exp.IsInt() {
		return rationalExpr{}, errors.New("los exponentes de una expresión racional deben ser enteros")
	}
	k, ok := ratInt64(exp)
	if !ok {
		return rationalExpr{}, errors.New("exponente demasiado grande")
	}
	if k == 0 {
		if base.Num.isZero() {
			return rationalExpr{}, errors.New("0 elevado a 0 no está definido")
		}
		return rationalExpr{Num: onePolynomial(), Den: onePolynomial()}, nil
	}
	if k > 0 {
		num, err := base.Num.powInt(k)
		if err != nil {
			return rationalExpr{}, err
		}
		den, err := base.Den.powInt(k)
		if err != nil {
			return rationalExpr{}, err
		}
		return rationalExpr{Num: num, Den: den}.canonical()
	}
	if base.Num.isZero() {
		return rationalExpr{}, errors.New("0 elevado a un exponente negativo no está definido")
	}
	if k == -1<<63 {
		return rationalExpr{}, errors.New("exponente demasiado grande")
	}
	num, err := base.Den.powInt(-k)
	if err != nil {
		return rationalExpr{}, err
	}
	den, err := base.Num.powInt(-k)
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalExpr{Num: num, Den: den}.canonical()
}

func (r rationalExpr) canonical() (rationalExpr, error) {
	num, err := r.Num.canonical()
	if err != nil {
		return rationalExpr{}, err
	}
	den, err := r.Den.canonical()
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalExpr{Num: num, Den: den}, nil
}

func rationalAdd(a, b rationalExpr) (rationalExpr, error) {
	if a.Num.isZero() {
		return b.canonical()
	}
	if b.Num.isZero() {
		return a.canonical()
	}
	n1, err := a.Num.mul(b.Den)
	if err != nil {
		return rationalExpr{}, err
	}
	n2, err := b.Num.mul(a.Den)
	if err != nil {
		return rationalExpr{}, err
	}
	num, err := n1.add(n2)
	if err != nil {
		return rationalExpr{}, err
	}
	if num.isZero() {
		return rationalExpr{Num: nil, Den: onePolynomial()}, nil
	}
	den, err := a.Den.mul(b.Den)
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalExpr{Num: num, Den: den}.canonical()
}

func rationalMul(a, b rationalExpr) (rationalExpr, error) {
	if a.Num.isZero() || b.Num.isZero() {
		return rationalExpr{Num: nil, Den: onePolynomial()}, nil
	}
	num, err := a.Num.mul(b.Num)
	if err != nil {
		return rationalExpr{}, err
	}
	den, err := a.Den.mul(b.Den)
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalExpr{Num: num, Den: den}.canonical()
}

func rationalDiv(a, b rationalExpr) (rationalExpr, error) {
	if b.Num.isZero() {
		return rationalExpr{}, errors.New("división por cero")
	}
	if a.Num.isZero() {
		return rationalExpr{Num: nil, Den: onePolynomial()}, nil
	}
	num, err := a.Num.mul(b.Den)
	if err != nil {
		return rationalExpr{}, err
	}
	den, err := a.Den.mul(b.Num)
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalExpr{Num: num, Den: den}.canonical()
}

func reduceRational(r rationalExpr) (rationalReduction, error) {
	num, err := r.Num.canonical()
	if err != nil {
		return rationalReduction{}, err
	}
	den, err := r.Den.canonical()
	if err != nil {
		return rationalReduction{}, err
	}
	if den.isZero() {
		return rationalReduction{}, errors.New("división por cero")
	}
	if num.isZero() {
		return rationalReduction{
			Result: rationalExpr{Num: nil, Den: onePolynomial()},
		}, nil
	}

	var canceled []Polynomial
	num, den, common, err := cancelCommonMonomial(num, den)
	if err != nil {
		return rationalReduction{}, err
	}
	if common != nil && !monomialIsOne(common) {
		canceled = append(canceled, polynomialFromMonomial(common))
	}

	numUnits, numSteps, numNotes, err := factorAndNormalize(num)
	if err != nil {
		return rationalReduction{}, err
	}
	denUnits, denSteps, denNotes, err := factorAndNormalize(den)
	if err != nil {
		return rationalReduction{}, err
	}

	numLeft, denLeft, numScale, denScale, canceledFactors := cancelFactorUnits(numUnits, denUnits)
	canceled = append(canceled, canceledFactors...)
	result, err := rebuildRational(numLeft, denLeft, numScale, denScale)
	if err != nil {
		return rationalReduction{}, err
	}

	return rationalReduction{
		Result:    result,
		NumUnits:  numUnits,
		DenUnits:  denUnits,
		NumSteps:  numSteps,
		DenSteps:  denSteps,
		Canceled:  canceled,
		SignNotes: dedupeStrings(append(numNotes, denNotes...)),
	}, nil
}

func factorAndNormalize(p Polynomial) ([]factorUnit, []string, []string, error) {
	factors, steps, err := factorPolynomial(p, 0)
	if err != nil {
		return nil, nil, nil, err
	}
	units, notes, err := normalizeFactorUnits(factors)
	if err != nil {
		return nil, nil, nil, err
	}
	return units, steps, notes, nil
}

func rebuildRational(numLeft, denLeft []factorUnit, numScale, denScale *big.Rat) (rationalExpr, error) {
	numProd, err := productPolynomials(numLeft)
	if err != nil {
		return rationalExpr{}, err
	}
	denProd, err := productPolynomials(denLeft)
	if err != nil {
		return rationalExpr{}, err
	}
	scale, err := ratQuo(numScale, denScale)
	if err != nil {
		return rationalExpr{}, err
	}
	numProd, err = scalePolynomial(numProd, ratFromInt(scale.Num()))
	if err != nil {
		return rationalExpr{}, err
	}
	denProd, err = scalePolynomial(denProd, ratFromInt(scale.Denom()))
	if err != nil {
		return rationalExpr{}, err
	}
	return rationalExpr{Num: numProd, Den: denProd}.canonical()
}

// cancelCommonMonomial extrae el máximo factor monomial común entre
// numerador y denominador. Así x^3/(x^3+x^2) se reduce a x/(x+1) aunque el
// factorizador devuelva x^3 como un único factor.
func cancelCommonMonomial(num, den Polynomial) (Polynomial, Polynomial, *Monomial, error) {
	numCommon, err := commonMonomialFactor(num)
	if err != nil {
		return nil, nil, nil, err
	}
	denCommon, err := commonMonomialFactor(den)
	if err != nil {
		return nil, nil, nil, err
	}
	common, err := monomialGCD(numCommon, denCommon)
	if err != nil {
		return nil, nil, nil, err
	}
	if monomialIsOne(common) {
		return num, den, nil, nil
	}
	num, err = num.divMonomial(common)
	if err != nil {
		return nil, nil, nil, err
	}
	den, err = den.divMonomial(common)
	if err != nil {
		return nil, nil, nil, err
	}
	return num, den, common, nil
}

func monomialGCD(a, b *Monomial) (*Monomial, error) {
	if a == nil || b == nil || a.isZero() || b.isZero() {
		return newMonomial(), nil
	}
	out := monomialFromRat(ratGCD(ratAbs(a.Coeff), ratAbs(b.Coeff)))
	for base, expA := range a.Factors {
		expB, ok := b.Factors[base]
		if !ok {
			continue
		}
		exp := expA
		if ratCmp(expB, exp) < 0 {
			exp = expB
		}
		if exp.Sign() != 0 {
			out.Factors[base] = ratCopy(exp)
		}
	}
	return out.canonical()
}

func normalizeFactorUnits(factors []Polynomial) ([]factorUnit, []string, error) {
	units := make([]factorUnit, 0, len(factors))
	var notes []string
	for _, p := range factors {
		p, err := p.canonical()
		if err != nil {
			return nil, nil, err
		}
		if len(p) == 0 {
			continue
		}
		content, err := polynomialContent(p)
		if err != nil {
			return nil, nil, err
		}
		inv, err := ratQuo(ratOne(), content)
		if err != nil {
			return nil, nil, err
		}
		norm, err := scalePolynomial(p, inv)
		if err != nil {
			return nil, nil, err
		}
		units = append(units, factorUnit{Poly: norm, Coeff: content})
		if content.Sign() < 0 && len(norm) > 1 {
			notes = append(notes, "Cambio de signo: "+formatFactorPolynomial(p, false)+
				" = "+formatFactorUnits([]factorUnit{{Poly: norm, Coeff: content}}, false))
		}
	}
	return units, notes, nil
}

// polynomialContent devuelve el gcd de los coeficientes con el signo del
// término principal, para normalizar factores proporcionales.
func polynomialContent(p Polynomial) (*big.Rat, error) {
	if len(p) == 0 {
		return nil, errors.New("polinomio cero")
	}
	g := ratAbs(p[0].Coeff)
	for _, m := range p[1:] {
		g = ratGCD(g, ratAbs(m.Coeff))
	}
	if g.Sign() == 0 {
		return nil, errors.New("polinomio cero")
	}
	if p[0].Coeff.Sign() < 0 {
		g = ratNeg(g)
	}
	return g, nil
}

func cancelFactorUnits(numUnits, denUnits []factorUnit) (
	numLeft, denLeft []factorUnit,
	numScale, denScale *big.Rat,
	canceled []Polynomial,
) {
	numScale, denScale = ratOne(), ratOne()
	used := make([]bool, len(numUnits))
	for _, d := range denUnits {
		denScale = ratMul(denScale, d.Coeff)
		if polynomialIsOne(d.Poly) {
			continue
		}
		matched := -1
		for i, n := range numUnits {
			if used[i] {
				continue
			}
			if polynomialEqual(n.Poly, d.Poly) {
				matched = i
				break
			}
		}
		if matched >= 0 {
			used[matched] = true
			canceled = append(canceled, d.Poly)
			continue
		}
		denLeft = append(denLeft, d)
	}
	for i, n := range numUnits {
		numScale = ratMul(numScale, n.Coeff)
		if used[i] || polynomialIsOne(n.Poly) {
			continue
		}
		numLeft = append(numLeft, n)
	}
	return numLeft, denLeft, numScale, denScale, canceled
}

func productPolynomials(units []factorUnit) (Polynomial, error) {
	out := onePolynomial()
	for _, u := range units {
		var err error
		out, err = out.mul(u.Poly)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// formatCanceledFactors mantiene los paréntesis de un único factor
// polinómico: "Factores cancelados: (x + 1)".
func formatCanceledFactors(factors []Polynomial) string {
	if len(factors) == 1 && len(factors[0]) > 1 {
		return formatFactorPolynomial(factors[0], false)
	}
	return formatFactorProduct(factors, false)
}

func factorMethodSteps(steps []string) []string {
	out := make([]string, 0, len(steps))
	for _, s := range steps {
		if strings.Contains(s, "No se puede factorizar más") {
			continue
		}
		out = append(out, s)
	}
	return out
}

func formatRationalExpression(r rationalExpr, unicode bool) string {
	num := formatPolynomialDisplay(r.Num, unicode)
	if polynomialIsOne(r.Den) {
		return num
	}
	den := formatPolynomialDisplay(r.Den, unicode)
	if needsNumeratorParens(r.Num, num) {
		num = "(" + num + ")"
	}
	if needsDenParens(den) {
		den = "(" + den + ")"
	}
	return num + "/" + den
}

func formatPolynomialDisplay(p Polynomial, unicode bool) string {
	ascii, uni := FormatPolynomial(p)
	if unicode {
		return uni
	}
	return ascii
}

func needsNumeratorParens(p Polynomial, text string) bool {
	return len(p) > 1 || strings.Contains(text, "/")
}

func formatFactorUnits(units []factorUnit, unicode bool) string {
	scale := ratOne()
	polys := make([]Polynomial, 0, len(units))
	for _, u := range units {
		scale = ratMul(scale, u.Coeff)
		if !polynomialIsOne(u.Poly) {
			polys = append(polys, u.Poly)
		}
	}
	if len(polys) == 0 {
		return ratString(scale)
	}
	product := formatFactorProduct(polys, unicode)
	one := ratOne()
	if scale.Cmp(one) == 0 {
		return product
	}
	sep := "*"
	if unicode {
		sep = "·"
	}
	if scale.Cmp(ratFromInt64(-1)) == 0 {
		if needsFactorPowerParens(product) {
			product = "(" + product + ")"
		}
		return "-" + product
	}
	if needsFactorPowerParens(product) {
		product = "(" + product + ")"
	}
	return ratString(scale) + sep + product
}
