package engine

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strings"
	"tuniversalcalc/internal/render"
)

const (
	maxRadicalTerms = 256
	maxRadicalPower = 16
)

// radFactor es un factor no numérico de un término: una base elevada a un
// exponente racional exacto.
type radFactor struct {
	base Expr
	exp  *big.Rat
}

// radTerm es un término exacto coeff × producto de factores. Los factores
// pueden incluir llamadas sqrt(...), incluso con radicandos que son sumas.
type radTerm struct {
	coeff   *big.Rat
	factors []radFactor
}

// radPoly es una suma de términos radicales normalizada.
type radPoly []radTerm

// radicalizer acumula los pasos que se muestran al usuario.
type radicalizer struct {
	steps []string
}

// radicalizarExpr racionaliza fracciones cuyo numerador o denominador es un
// binomio con raíces cuadradas. Multiplica por el conjugado, simplifica y
// cancela factores comunes. Si no encuentra un conjugado aplicable,
// simplifica la expresión como /radical.
func radicalizarExpr(e Expr) (ExactResult, error) {
	r := &radicalizer{}
	out, changed, err := r.rationalize(e)
	if err != nil {
		return ExactResult{}, err
	}
	if !changed {
		res, err := Radical(e)
		if err != nil {
			return ExactResult{}, fmt.Errorf("no se encontró un binomio con raíces cuadradas que racionalizar: %w", err)
		}
		steps := []string{
			"Operación: radicalización por conjugado",
			"Expresión: " + e.Ascii(),
			"No se encontró un binomio radical que requiera conjugado; se simplifica la expresión.",
		}
		steps = append(steps, exactStepsTail(res.Steps)...)
		return ExactResult{
			Poly:    res.Poly,
			Ascii:   res.Ascii,
			Unicode: res.Unicode,
			Steps:   steps,
		}, nil
	}

	ascii, unicode, poly := formatRadicalResult(out)
	steps := []string{
		"Operación: radicalización por conjugado",
		"Expresión: " + e.Ascii(),
	}
	steps = append(steps, r.steps...)
	steps = append(steps,
		"Nota: la igualdad es válida donde el denominador original y el resultante no son cero.",
		"Resultado: "+ascii,
		"Unicode: "+unicode,
	)
	return ExactResult{Poly: poly, Ascii: ascii, Unicode: unicode, Steps: steps}, nil
}

// exactStepsTail elimina la cabecera repetida de los pasos de /radical.
func exactStepsTail(steps []string) []string {
	if len(steps) >= 2 && strings.HasPrefix(steps[0], "Operación:") &&
		strings.HasPrefix(steps[1], "Expresión:") {
		return steps[2:]
	}
	return steps
}

// rationalize solo actúa sobre una división en la raíz del árbol. Para el
// resto de expresiones se deja que /radical haga la simplificación habitual.
func (r *radicalizer) rationalize(e Expr) (Expr, bool, error) {
	switch n := e.(type) {
	case Div:
		return r.rationalizeDiv(n)
	case Neg:
		out, changed, err := r.rationalize(n.X)
		if err != nil || !changed {
			return e, changed, err
		}
		return Neg{X: out}, true, nil
	}
	return e, false, nil
}

func (r *radicalizer) rationalizeDiv(n Div) (Expr, bool, error) {
	num, den := n.Num, n.Den
	conj, target, ok := conjugateFor(num, den)
	if !ok {
		return n, false, nil
	}

	r.steps = append(r.steps,
		"Conjugado del "+target+": "+conj.Ascii(),
		"Multiplicando numerador y denominador por el conjugado...",
	)

	numPoly, err := multiplyRadicalExprs(num, conj)
	if err != nil {
		return n, false, err
	}
	denPoly, err := multiplyRadicalExprs(den, conj)
	if err != nil {
		return n, false, err
	}
	if denPoly.isZero() {
		return nil, false, errors.New("división por cero")
	}

	r.steps = append(r.steps,
		"Numerador: "+formatRadicalPolyFactored(numPoly),
		"Denominador: "+formatRadicalPolyFactored(denPoly),
	)
	numPoly, denPoly, canceled := cancelRadicalPolys(numPoly, denPoly)
	if len(canceled) > 0 {
		r.steps = append(r.steps, "Factores cancelados: "+formatRadicalFactors(canceled))
	} else {
		r.steps = append(r.steps, "Factores cancelados: ninguno")
	}

	if denPoly.isOne() {
		return numPoly.toExpr(), true, nil
	}
	return Div{Num: numPoly.toExpr(), Den: denPoly.toExpr()}, true, nil
}

// conjugateFor elige el binomio que se va a racionalizar. Se prefiere el
// denominador; si este no es un binomio radical, se prueba el numerador.
func conjugateFor(num, den Expr) (Expr, string, bool) {
	if conj, ok := conjugateBinomial(den); ok {
		return conj, "denominador", true
	}
	if conj, ok := conjugateBinomial(num); ok {
		return conj, "numerador", true
	}
	return nil, "", false
}

// conjugateBinomial devuelve el conjugado de una suma de dos términos con
// al menos una raíz cuadrada. El conjugado cambia el signo del segundo
// término: (A + B) → A − B.
func conjugateBinomial(e Expr) (Expr, bool) {
	add, ok := e.(Add)
	if !ok {
		return nil, false
	}
	flat := flattenAdd(add)
	if len(flat.Terms) != 2 {
		return nil, false
	}
	if !exprHasSquareRoot(flat.Terms[0]) && !exprHasSquareRoot(flat.Terms[1]) {
		return nil, false
	}
	second := flat.Terms[1]
	var flipped Expr
	if neg, ok := second.(Neg); ok {
		flipped = neg.X
	} else {
		flipped = Neg{X: second}
	}
	return Add{Terms: []Expr{flat.Terms[0], flipped}}, true
}

// exprIsNegativeConstant comprueba si una expresión se reduce a un número
// negativo, para rechazar raíces cuadradas fuera de los reales.
func exprIsNegativeConstant(e Expr) bool {
	p, err := ExprToPolynomial(e)
	if err != nil || len(p) == 0 {
		return false
	}
	for _, m := range p {
		if len(m.Factors) != 0 {
			return false
		}
	}
	sum := ratZero()
	for _, m := range p {
		sum = ratAdd(sum, m.Coeff)
	}
	return sum.Sign() < 0
}

func exprHasSquareRoot(e Expr) bool {
	switch n := e.(type) {
	case Call:
		return n.Name == fnSqrt || exprHasSquareRoot(n.Arg)
	case Neg:
		return exprHasSquareRoot(n.X)
	case Add:
		return slices.ContainsFunc(n.Terms, exprHasSquareRoot)
	case Mul:
		return slices.ContainsFunc(n.Factors, exprHasSquareRoot)
	case Div:
		return exprHasSquareRoot(n.Num) || exprHasSquareRoot(n.Den)
	case Pow:
		if exp, err := exactExponent(n.Exp); err == nil && exp.Cmp(big.NewRat(1, 2)) == 0 {
			return true
		}
		return exprHasSquareRoot(n.Base) || exprHasSquareRoot(n.Exp)
	}
	return false
}

// radFromExpr convierte una expresión en una suma de términos radicales.
// Solo admite operaciones racionales; las divisiones anidadas se rechazan.
//
//nolint:gocyclo // despacho recursivo por cada tipo de nodo del AST
func radFromExpr(e Expr) (radPoly, error) {
	switch n := e.(type) {
	case Num:
		r, err := ratParse(n.Text)
		if err != nil {
			return nil, err
		}
		return radPoly{{coeff: r}}, nil
	case Var:
		return radPoly{radSingleTerm(ratOne(), n, ratOne())}, nil
	case Const:
		return nil, fmt.Errorf("la constante %q no pertenece a una expresión radical", n.Name)
	case Neg:
		p, err := radFromExpr(n.X)
		if err != nil {
			return nil, err
		}
		return p.neg(), nil
	case Add:
		var out radPoly
		for _, t := range n.Terms {
			p, err := radFromExpr(t)
			if err != nil {
				return nil, err
			}
			out, err = out.add(p)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case Mul:
		out := radPoly{{coeff: ratOne()}}
		for _, f := range n.Factors {
			p, err := radFromExpr(f)
			if err != nil {
				return nil, err
			}
			out, err = out.mul(p)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case Div:
		return nil, errors.New("división anidada: no soportada en una racionalización")
	case Pow:
		return radFromPow(n)
	case Call:
		if n.Name != fnSqrt {
			return nil, fmt.Errorf("la función %s() no pertenece al motor radical", n.Name)
		}
		if exprIsNegativeConstant(n.Arg) {
			return nil, errors.New("raíz cuadrada de una cantidad negativa: los complejos no están soportados")
		}
		return radPoly{radSingleTerm(ratOne(), n, ratOne())}, nil
	}
	return nil, fmt.Errorf("expresión no soportada como expresión radical: %T", e)
}

func radFromPow(n Pow) (radPoly, error) {
	exp, err := exactExponent(n.Exp)
	if err != nil {
		return nil, err
	}
	if !exp.IsInt() {
		if exp.Cmp(big.NewRat(1, 2)) == 0 {
			if exprIsNegativeConstant(n.Base) {
				return nil, errors.New("raíz cuadrada de una cantidad negativa: los complejos no están soportados")
			}
			return radPoly{radSingleTerm(ratOne(), Call{Name: fnSqrt, Arg: n.Base}, ratOne())}, nil
		}
		return nil, errors.New("solo se admiten exponentes enteros o raíces cuadradas en la racionalización")
	}
	k, ok := ratInt64(exp)
	if !ok {
		return nil, errors.New("exponente demasiado grande")
	}
	if k < 0 {
		return nil, errors.New("exponente negativo: no soportado en una racionalización")
	}
	base, err := radFromExpr(n.Base)
	if err != nil {
		return nil, err
	}
	if k == 0 && base.isZero() {
		return nil, errors.New("0 elevado a 0 no está definido")
	}
	return base.powInt(k)
}

func radSingleTerm(coeff *big.Rat, base Expr, exp *big.Rat) radTerm {
	return radTerm{coeff: coeff, factors: []radFactor{{base: base, exp: exp}}}
}

func (p radPoly) clone() radPoly {
	out := make(radPoly, 0, len(p))
	for _, t := range p {
		out = append(out, t.clone())
	}
	return out
}

func (t radTerm) clone() radTerm {
	out := radTerm{coeff: ratCopy(t.coeff), factors: make([]radFactor, 0, len(t.factors))}
	for _, f := range t.factors {
		out.factors = append(out.factors, radFactor{base: f.base, exp: ratCopy(f.exp)})
	}
	return out
}

func (p radPoly) isZero() bool {
	for _, t := range p {
		if t.coeff != nil && t.coeff.Sign() != 0 {
			return false
		}
	}
	return true
}

func (p radPoly) isOne() bool {
	return len(p) == 1 && len(p[0].factors) == 0 && p[0].coeff != nil && p[0].coeff.Cmp(ratOne()) == 0
}

func (p radPoly) neg() radPoly {
	out := p.clone()
	for i := range out {
		out[i].coeff = ratNeg(out[i].coeff)
	}
	return out
}

func (p radPoly) add(q radPoly) (radPoly, error) {
	terms := make(radPoly, 0, len(p)+len(q))
	terms = append(terms, p.clone()...)
	terms = append(terms, q.clone()...)
	return normalizeRadPoly(terms)
}

func (p radPoly) mul(q radPoly) (radPoly, error) {
	if p.isZero() || q.isZero() {
		return nil, nil
	}
	if len(p)*len(q) > maxRadicalTerms*4 {
		return nil, fmt.Errorf("la operación radical genera demasiados términos; límite %d", maxRadicalTerms)
	}
	terms := make(radPoly, 0, len(p)*len(q))
	for _, a := range p {
		for _, b := range q {
			t, err := multiplyRadTerms(a, b)
			if err != nil {
				return nil, err
			}
			terms = append(terms, t...)
			if len(terms) > maxRadicalTerms*4 {
				return nil, fmt.Errorf("la operación radical genera demasiados términos; límite %d", maxRadicalTerms)
			}
		}
	}
	return normalizeRadPoly(terms)
}

func (p radPoly) powInt(k int64) (radPoly, error) {
	if k < 0 {
		return nil, errors.New("exponente negativo: no soportado")
	}
	if k == 0 {
		return radPoly{{coeff: ratOne()}}, nil
	}
	if k > maxRadicalPower {
		return nil, fmt.Errorf("exponente %d demasiado grande para radicales; límite %d", k, maxRadicalPower)
	}
	out := radPoly{{coeff: ratOne()}}
	for range k {
		var err error
		out, err = out.mul(p)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func multiplyRadTerms(a, b radTerm) (radPoly, error) {
	t := radTerm{coeff: ratMul(a.coeff, b.coeff)}
	t.factors = append(t.factors, a.clone().factors...)
	t.factors = append(t.factors, b.clone().factors...)
	return normalizeRadTerm(t)
}

// normalizeRadPoly agrupa términos semejantes y aplica las reducciones de
// raíces cuadradas exactas.
func normalizeRadPoly(terms radPoly) (radPoly, error) {
	var expanded radPoly
	for _, t := range terms {
		parts, err := normalizeRadTerm(t)
		if err != nil {
			return nil, err
		}
		expanded = append(expanded, parts...)
		if len(expanded) > maxRadicalTerms*4 {
			return nil, fmt.Errorf("la expresión radical genera demasiados términos; límite %d", maxRadicalTerms)
		}
	}
	byKey := map[string]*big.Rat{}
	bases := map[string][]radFactor{}
	for _, t := range expanded {
		if t.coeff == nil || t.coeff.Sign() == 0 {
			continue
		}
		k := radTermKey(t)
		if old, ok := byKey[k]; ok {
			old.Add(old, t.coeff)
			if old.Sign() == 0 {
				delete(byKey, k)
			}
			continue
		}
		byKey[k] = ratCopy(t.coeff)
		bases[k] = t.factors
	}
	if len(byKey) > maxRadicalTerms {
		return nil, fmt.Errorf("la expresión radical genera demasiados términos; límite %d", maxRadicalTerms)
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return radicalKeyOrder(keys[i]) < radicalKeyOrder(keys[j])
	})
	out := make(radPoly, 0, len(keys))
	for _, k := range keys {
		out = append(out, radTerm{coeff: byKey[k], factors: bases[k]})
	}
	return out, nil
}

// radicalKeyOrder coloca los términos con factores antes que las
// constantes, para que sqrt(4+h) + 2 se lea en el orden habitual.
func radicalKeyOrder(key string) string {
	if key == "" {
		return "~"
	}
	return key
}

//nolint:gocyclo // reducción de raíces y factores en una sola pasada
func normalizeRadTerm(t radTerm) (radPoly, error) {
	if t.coeff == nil || t.coeff.Sign() == 0 {
		return nil, nil
	}
	factors, err := mergeRadFactors(t.factors)
	if err != nil {
		return nil, err
	}

	remaining := make([]radFactor, 0, len(factors))
	var multipliers []radPoly
	for _, f := range factors {
		if f.exp == nil || f.exp.Sign() == 0 {
			continue
		}
		call, isSqrt := f.base.(Call)
		if !isSqrt || call.Name != fnSqrt || !f.exp.IsInt() {
			remaining = append(remaining, f)
			continue
		}
		k, ok := ratInt64(f.exp)
		if !ok {
			return nil, errors.New("exponente de raíz demasiado grande")
		}
		arg, err := radFromExpr(call.Arg)
		if err != nil {
			return nil, err
		}
		if k%2 == 0 {
			q := k / 2
			if q < 0 {
				return nil, errors.New("exponente negativo de raíz: no soportado")
			}
			powered, err := arg.powInt(q)
			if err != nil {
				return nil, err
			}
			multipliers = append(multipliers, powered)
			continue
		}
		if k > 1 {
			q := (k - 1) / 2
			powered, err := arg.powInt(q)
			if err != nil {
				return nil, err
			}
			multipliers = append(multipliers, powered)
			remaining = append(remaining, radFactor{base: f.base, exp: ratOne()})
			continue
		}
		remaining = append(remaining, f)
	}

	out := radPoly{{coeff: ratCopy(t.coeff), factors: remaining}}
	for _, m := range multipliers {
		out, err = out.mul(m)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func mergeRadFactors(factors []radFactor) ([]radFactor, error) {
	merged := map[string]radFactor{}
	order := make([]string, 0, len(factors))
	for _, f := range factors {
		if f.exp == nil || f.exp.Sign() == 0 {
			continue
		}
		base := f.base
		exp := ratCopy(f.exp)
		if pow, ok := base.(Pow); ok {
			inner, err := exactExponent(pow.Exp)
			if err != nil {
				return nil, err
			}
			base = pow.Base
			exp = ratMul(exp, inner)
		}
		if _, ok := base.(Num); ok {
			return nil, errors.New("base numérica inesperada en un factor radical")
		}
		key := base.Ascii()
		if old, ok := merged[key]; ok {
			old.exp = ratAdd(old.exp, exp)
			merged[key] = old
			continue
		}
		merged[key] = radFactor{base: base, exp: exp}
		order = append(order, key)
	}
	out := make([]radFactor, 0, len(order))
	for _, key := range order {
		f := merged[key]
		if f.exp.Sign() != 0 {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].base.Ascii() < out[j].base.Ascii() })
	return out, nil
}

func radTermKey(t radTerm) string {
	var b strings.Builder
	for _, f := range t.factors {
		b.WriteString(f.base.Ascii())
		b.WriteByte('^')
		b.WriteString(ratString(f.exp))
		b.WriteByte(';')
	}
	return b.String()
}

func (p radPoly) toExpr() Expr {
	if len(p) == 0 {
		return Num{Text: "0"}
	}
	ordered := p.clone()
	sort.SliceStable(ordered, func(i, j int) bool {
		return radicalKeyOrder(radTermKey(ordered[i])) < radicalKeyOrder(radTermKey(ordered[j]))
	})
	var terms []Expr
	for _, t := range ordered {
		terms = append(terms, radTermToExpr(t))
	}
	if len(terms) == 1 {
		return terms[0]
	}
	return Add{Terms: terms}
}

func radTermToExpr(t radTerm) Expr {
	if t.coeff == nil || t.coeff.Sign() == 0 {
		return Num{Text: "0"}
	}
	neg := t.coeff.Sign() < 0
	abs := ratAbs(t.coeff)
	var core Expr
	if len(t.factors) == 0 {
		core = Num{Text: ratString(abs)}
	} else {
		core = radFactorsToExpr(t.factors)
		if abs.Cmp(ratOne()) != 0 {
			core = Mul{Factors: []Expr{Num{Text: ratString(abs)}, core}}
		}
	}
	if neg {
		return Neg{X: core}
	}
	return core
}

// multiplyRadicalExprs multiplica dos AST y devuelve el resultado radical.
func multiplyRadicalExprs(a, b Expr) (radPoly, error) {
	pa, err := radFromExpr(a)
	if err != nil {
		return nil, err
	}
	pb, err := radFromExpr(b)
	if err != nil {
		return nil, err
	}
	return pa.mul(pb)
}

// cancelRadicalPolys cancela los factores monomiales comunes entre el
// numerador y el denominador. También normaliza el signo del denominador.
func cancelRadicalPolys(num, den radPoly) (radPoly, radPoly, []radFactor) {
	if den.isZero() {
		return num, den, nil
	}
	nf := commonRadicalFactors(num)
	df := commonRadicalFactors(den)
	common := intersectRadicalFactors(nf, df)
	if len(common) > 0 {
		num = divideRadicalPoly(num, common)
		den = divideRadicalPoly(den, common)
	}
	// Si el denominador es una constante negativa, se pasa el signo al
	// numerador.
	if len(den) == 1 && len(den[0].factors) == 0 && den[0].coeff.Sign() < 0 {
		num = num.neg()
		den = den.neg()
	}
	// Un denominador constante distinto de 1 se incorpora al numerador.
	if len(den) == 1 && len(den[0].factors) == 0 {
		c := den[0].coeff
		if c.Sign() != 0 && c.Cmp(ratOne()) != 0 {
			inv := new(big.Rat).Inv(c)
			num = scaleRadicalPoly(num, inv)
			den = radPoly{{coeff: ratOne()}}
			common = append(common, radFactor{base: Num{Text: ratString(c)}, exp: ratOne()})
		}
	}
	return num, den, common
}

func scaleRadicalPoly(p radPoly, factor *big.Rat) radPoly {
	out := p.clone()
	for i := range out {
		out[i].coeff = ratMul(out[i].coeff, factor)
	}
	return out
}

func commonRadicalFactors(p radPoly) []radFactor {
	if len(p) == 0 {
		return nil
	}
	common := map[string]radFactor{}
	for _, f := range p[0].factors {
		if f.exp.Sign() <= 0 {
			continue
		}
		common[f.base.Ascii()] = f
	}
	for _, t := range p[1:] {
		current := map[string]radFactor{}
		for _, f := range t.factors {
			if f.exp.Sign() > 0 {
				current[f.base.Ascii()] = f
			}
		}
		for key, f := range common {
			other, ok := current[key]
			if !ok {
				delete(common, key)
				continue
			}
			if other.exp.Cmp(f.exp) < 0 {
				common[key] = other
			}
		}
	}
	out := make([]radFactor, 0, len(common))
	for _, f := range common {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].base.Ascii() < out[j].base.Ascii() })
	return out
}

func intersectRadicalFactors(a, b []radFactor) []radFactor {
	byKey := map[string]radFactor{}
	for _, f := range a {
		byKey[f.base.Ascii()] = f
	}
	var out []radFactor
	for _, f := range b {
		other, ok := byKey[f.base.Ascii()]
		if !ok {
			continue
		}
		exp := f.exp
		if other.exp.Cmp(exp) < 0 {
			exp = other.exp
		}
		if exp.Sign() > 0 {
			out = append(out, radFactor{base: f.base, exp: exp})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].base.Ascii() < out[j].base.Ascii() })
	return out
}

func divideRadicalPoly(p radPoly, common []radFactor) radPoly {
	out := make(radPoly, 0, len(p))
	for _, t := range p {
		nt := t.clone()
		for _, c := range common {
			for i := range nt.factors {
				if nt.factors[i].base.Ascii() != c.base.Ascii() {
					continue
				}
				nt.factors[i].exp = ratSub(nt.factors[i].exp, c.exp)
				break
			}
		}
		filtered := nt.factors[:0]
		for _, f := range nt.factors {
			if f.exp.Sign() != 0 {
				filtered = append(filtered, f)
			}
		}
		nt.factors = filtered
		out = append(out, nt)
	}
	return out
}

func formatRadicalPolyFactored(p radPoly) string {
	return formatRadicalExpr(radPolyFactoredExpr(p))
}

func radPolyFactoredExpr(p radPoly) Expr {
	if p.isZero() {
		return Num{Text: "0"}
	}
	common := commonRadicalFactors(p)
	if len(common) == 0 {
		return p.toExpr()
	}
	res := divideRadicalPoly(p, common)
	commonExpr := radFactorsToExpr(common)
	if res.isOne() {
		return commonExpr
	}
	return Mul{Factors: []Expr{commonExpr, res.toExpr()}}
}

func radFactorsToExpr(factors []radFactor) Expr {
	if len(factors) == 0 {
		return Num{Text: "1"}
	}
	if len(factors) == 1 && factors[0].exp.Cmp(ratOne()) == 0 {
		return factors[0].base
	}
	var out []Expr
	for _, f := range factors {
		if f.exp.Cmp(ratOne()) == 0 {
			out = append(out, f.base)
			continue
		}
		out = append(out, Pow{Base: f.base, Exp: Num{Text: ratString(f.exp)}})
	}
	if len(out) == 1 {
		return out[0]
	}
	return Mul{Factors: out}
}

func formatRadicalFactors(factors []radFactor) string {
	return formatRadicalExpr(radFactorsToExpr(factors))
}

func formatRadicalExpr(e Expr) string {
	ascii, _, _ := formatRadicalResult(e)
	return ascii
}

// formatRadicalResult intenta la forma polinómica exacta y, si no es
// posible, conserva la expresión radical en ASCII y Unicode.
func formatRadicalResult(e Expr) (string, string, Polynomial) {
	if p, err := ExprToPolynomial(e); err == nil {
		p, err = p.canonical()
		if err == nil {
			ascii, unicode := FormatPolynomial(p)
			return ascii, unicode, p
		}
	}
	ascii := e.Ascii()
	return ascii, render.Pretty(ascii), nil
}
