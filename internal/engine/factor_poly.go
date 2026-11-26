package engine

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

const (
	maxFactorDepth = 32
	maxFactorDivs  = 100000
)

// FactorResult contiene los factores de un polinomio y los pasos de la
// factorización para el panel de salida.
type FactorResult struct {
	Factors []Polynomial
	Steps   []string
	Ascii   string
	Unicode string
}

// FactorizarPolinomio factoriza una expresión polinómica sobre los
// racionales. Los factores que no encajan en un patrón o no tienen raíces
// racionales se dejan tal cual.
func FactorizarPolinomio(e Expr) (FactorResult, error) {
	p, err := ExprToPolynomial(e)
	if err != nil {
		return FactorResult{}, err
	}
	p, err = p.canonical()
	if err != nil {
		return FactorResult{}, err
	}
	if len(p) == 0 {
		steps := []string{
			"Operación: factorización de polinomios",
			"Expresión: " + e.Ascii(),
			"0 no tiene factores",
			"Factorización: 0",
			"Unicode: 0",
		}
		return FactorResult{Ascii: "0", Unicode: "0", Steps: steps}, nil
	}
	factors, steps, err := factorPolynomial(p, 0)
	if err != nil {
		return FactorResult{}, err
	}
	ascii := formatFactorProduct(factors, false)
	unicode := formatFactorProduct(factors, true)
	all := []string{
		"Operación: factorización de polinomios",
		"Expresión: " + e.Ascii(),
	}
	all = append(all, steps...)
	all = append(all,
		"Factorización: "+ascii,
		"Unicode: "+unicode,
	)
	return FactorResult{Factors: factors, Steps: all, Ascii: ascii, Unicode: unicode}, nil
}

type factorAttempt func(Polynomial) ([]Polynomial, []string, bool, error)

var factorAttempts = []factorAttempt{
	factorBinomial,
	factorPerfectSquareTrinomial,
	factorQuadratic,
	factorByGrouping,
	factorRationalRoots,
}

// factorPolynomial factoriza recursivamente un polinomio ya canónico.
func factorPolynomial(p Polynomial, depth int) ([]Polynomial, []string, error) {
	if depth > maxFactorDepth {
		return []Polynomial{p}, nil, nil
	}
	p, err := p.canonical()
	if err != nil {
		return nil, nil, err
	}
	if len(p) == 0 {
		return nil, nil, nil
	}
	if len(p) == 1 {
		return []Polynomial{p}, nil, nil
	}

	if factors, steps, ok, err := factorCommon(p, depth); err != nil || ok {
		return factors, steps, err
	}
	for _, attempt := range factorAttempts {
		if factors, steps, ok, err := tryFactor(attempt, p, depth); err != nil || ok {
			return factors, steps, err
		}
	}
	return []Polynomial{p}, []string{"No se puede factorizar más sobre los racionales."}, nil
}

func factorCommon(p Polynomial, depth int) ([]Polynomial, []string, bool, error) {
	common, err := commonMonomialFactor(p)
	if err != nil {
		return nil, nil, false, err
	}
	if monomialIsOne(common) {
		return nil, nil, false, nil
	}
	q, err := p.divMonomial(common)
	if err != nil {
		return nil, nil, false, err
	}
	sub, subSteps, err := factorPolynomial(q, depth+1)
	if err != nil {
		return nil, nil, false, err
	}
	factors := []Polynomial{polynomialFromMonomial(common)}
	factors = append(factors, sub...)
	steps := append([]string{
		"Factor común: " + formatFactorMonomial(common, false),
	}, subSteps...)
	return factors, steps, true, nil
}

func tryFactor(attempt factorAttempt, p Polynomial, depth int) ([]Polynomial, []string, bool, error) {
	fs, st, ok, err := attempt(p)
	if err != nil || !ok {
		return nil, nil, ok, err
	}
	factors, subSteps, err := factorEach(fs, depth+1)
	if err != nil {
		return nil, nil, false, err
	}
	return factors, append(st, subSteps...), true, nil
}

func factorEach(fs []Polynomial, depth int) ([]Polynomial, []string, error) {
	var out []Polynomial
	var steps []string
	for _, f := range fs {
		if polynomialIsOne(f) {
			continue
		}
		sub, subSteps, err := factorPolynomial(f, depth+1)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, sub...)
		steps = append(steps, subSteps...)
	}
	return out, steps, nil
}

// commonMonomialFactor devuelve el gcd numérico y la menor potencia de cada
// base presente en todos los términos.
func commonMonomialFactor(p Polynomial) (*Monomial, error) {
	if len(p) == 0 {
		return newMonomial(), nil
	}
	gcd := ratAbs(p[0].Coeff)
	for _, m := range p[1:] {
		gcd = ratGCD(gcd, ratAbs(m.Coeff))
	}
	out := monomialFromRat(gcd)
	for b, firstExp := range p[0].Factors {
		minExp := ratCopy(firstExp)
		present := true
		for _, m := range p[1:] {
			e, ok := m.Factors[b]
			if !ok {
				present = false
				break
			}
			if ratCmp(e, minExp) < 0 {
				minExp = ratCopy(e)
			}
		}
		if present && minExp.Sign() != 0 {
			out.Factors[b] = minExp
		}
	}
	return out.canonical()
}

func monomialIsOne(m *Monomial) bool {
	return m != nil && m.Coeff.Cmp(ratOne()) == 0 && len(m.Factors) == 0
}

func polynomialIsOne(p Polynomial) bool {
	return len(p) == 1 && monomialIsOne(p[0])
}

func monomialEqual(a, b *Monomial) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Coeff.Cmp(b.Coeff) == 0 && a.key() == b.key()
}

func polynomialEqual(a, b Polynomial) bool {
	ca, err := a.canonical()
	if err != nil {
		return false
	}
	cb, err := b.canonical()
	if err != nil {
		return false
	}
	if len(ca) != len(cb) {
		return false
	}
	for i := range ca {
		if !monomialEqual(ca[i], cb[i]) {
			return false
		}
	}
	return true
}

func polyAddMonomials(a, b *Monomial) (Polynomial, error) {
	return Polynomial{a.clone(), b.clone()}.canonical()
}

func polySubMonomials(a, b *Monomial) (Polynomial, error) {
	return Polynomial{a.clone(), b.neg()}.canonical()
}

func polyMulMonomials(a, b *Monomial) (*Monomial, error) {
	return a.mul(b)
}

// perfectMonomialRoot comprueba si un monomio positivo es una potencia
// n-ésima exacta de otro monomio con exponentes enteros.
func perfectMonomialRoot(m *Monomial, n int64) (*Monomial, bool) {
	if m == nil || m.Coeff.Sign() < 0 {
		return nil, false
	}
	rootCoeff, err := monomialPowRat(m.Coeff, big.NewRat(1, n))
	if err != nil || len(rootCoeff.Factors) != 0 {
		return nil, false
	}
	out := monomialFromRat(rootCoeff.Coeff)
	for b, e := range m.Factors {
		r := new(big.Rat).Quo(e, big.NewRat(n, 1))
		if !r.IsInt() {
			return nil, false
		}
		out.Factors[b] = r
	}
	canon, err := out.canonical()
	if err != nil {
		return nil, false
	}
	return canon, true
}

func absMonomial(m *Monomial) *Monomial {
	out := m.clone()
	out.Coeff = ratAbs(out.Coeff)
	return out
}

func polyFromMonomials(ms ...*Monomial) (Polynomial, error) {
	p := make(Polynomial, 0, len(ms))
	for _, m := range ms {
		p = append(p, m.clone())
	}
	return p.canonical()
}

// factorBinomial reconoce diferencias de cuadrados y sumas/diferencias de
// cubos.
//
//nolint:nestif // dos patrones binomiales con varias comprobaciones exactas
func factorBinomial(p Polynomial) ([]Polynomial, []string, bool, error) {
	if len(p) != 2 {
		return nil, nil, false, nil
	}
	t0, t1 := p[0], p[1]

	if t0.Coeff.Sign()*t1.Coeff.Sign() < 0 {
		pos, neg := t0, t1
		if pos.Coeff.Sign() < 0 {
			pos, neg = t1, t0
		}
		r1, ok1 := perfectMonomialRoot(pos, 2)
		r2, ok2 := perfectMonomialRoot(absMonomial(neg), 2)
		if ok1 && ok2 {
			f1, err := polySubMonomials(r1, r2)
			if err != nil {
				return nil, nil, false, err
			}
			f2, err := polyAddMonomials(r1, r2)
			if err != nil {
				return nil, nil, false, err
			}
			step := "Diferencia de cuadrados: (" + formatFactorMonomial(r1, false) +
				")^2 - (" + formatFactorMonomial(r2, false) + ")^2"
			return []Polynomial{f1, f2}, []string{step}, true, nil
		}
	}

	r0, ok0 := perfectMonomialRoot(absMonomial(t0), 3)
	r1, ok1 := perfectMonomialRoot(absMonomial(t1), 3)
	if !ok0 || !ok1 {
		return nil, nil, false, nil
	}
	switch {
	case t0.Coeff.Sign() > 0 && t1.Coeff.Sign() > 0:
		return cubeFactorization(r0, r1, true)
	case t0.Coeff.Sign() > 0 && t1.Coeff.Sign() < 0:
		return cubeFactorization(r0, r1, false)
	}
	return nil, nil, false, nil
}

func cubeFactorization(a, b *Monomial, sum bool) ([]Polynomial, []string, bool, error) {
	var first Polynomial
	var err error
	if sum {
		first, err = polyAddMonomials(a, b)
	} else {
		first, err = polySubMonomials(a, b)
	}
	if err != nil {
		return nil, nil, false, err
	}
	a2, err := polyMulMonomials(a, a)
	if err != nil {
		return nil, nil, false, err
	}
	b2, err := polyMulMonomials(b, b)
	if err != nil {
		return nil, nil, false, err
	}
	ab, err := polyMulMonomials(a, b)
	if err != nil {
		return nil, nil, false, err
	}
	var quad Polynomial
	if sum {
		quad, err = polySubMonomials(a2, ab)
	} else {
		quad, err = polyAddMonomials(a2, ab)
	}
	if err != nil {
		return nil, nil, false, err
	}
	quad = append(quad, b2.clone())
	quad, err = quad.canonical()
	if err != nil {
		return nil, nil, false, err
	}
	op := "+"
	kind := "Suma"
	if !sum {
		op = "−"
		kind = "Diferencia"
	}
	step := kind + " de cubos: (" + formatFactorMonomial(a, false) +
		")^3 " + op + " (" + formatFactorMonomial(b, false) + ")^3"
	return []Polynomial{first, quad}, []string{step}, true, nil
}

// factorPerfectSquareTrinomial reconoce a^2 ± 2ab + b^2.
func factorPerfectSquareTrinomial(p Polynomial) ([]Polynomial, []string, bool, error) {
	if len(p) != 3 {
		return nil, nil, false, nil
	}
	for i := range 3 {
		for j := i + 1; j < 3; j++ {
			k := 3 - i - j
			ti, tj := p[i], p[j]
			if ti.Coeff.Sign() < 0 || tj.Coeff.Sign() < 0 {
				continue
			}
			ri, oki := perfectMonomialRoot(ti, 2)
			rj, okj := perfectMonomialRoot(tj, 2)
			if !oki || !okj {
				continue
			}
			prod, err := polyMulMonomials(ri, rj)
			if err != nil {
				return nil, nil, false, err
			}
			prod.Coeff = ratMul(prod.Coeff, big.NewRat(2, 1))
			if monomialEqual(prod, p[k]) {
				f, err := polyAddMonomials(ri, rj)
				if err != nil {
					return nil, nil, false, err
				}
				step := "Trinomio cuadrado perfecto: (" + formatFactorMonomial(ri, false) +
					" + " + formatFactorMonomial(rj, false) + ")^2"
				return []Polynomial{f, f}, []string{step}, true, nil
			}
			if monomialEqual(prod.neg(), p[k]) {
				f, err := polySubMonomials(ri, rj)
				if err != nil {
					return nil, nil, false, err
				}
				step := "Trinomio cuadrado perfecto: (" + formatFactorMonomial(ri, false) +
					" - " + formatFactorMonomial(rj, false) + ")^2"
				return []Polynomial{f, f}, []string{step}, true, nil
			}
		}
	}
	return nil, nil, false, nil
}

// factorQuadratic factoriza ax^2+bx+c con raíces racionales.
func factorQuadratic(p Polynomial) ([]Polynomial, []string, bool, error) {
	name, a, b, c, ok := quadraticCoefficients(p)
	if !ok {
		return nil, nil, false, nil
	}
	disc, r1, r2, ok, err := quadraticRoots(a, b, c)
	if err != nil || !ok {
		return nil, nil, false, err
	}
	factors, err := quadraticFactors(name, a, r1, r2)
	if err != nil {
		return nil, nil, false, err
	}
	steps := []string{
		"Trinomio: a=" + ratString(a) + ", b=" + ratString(b) + ", c=" + ratString(c),
		"Discriminante: Δ = " + ratString(disc),
		"Raíces racionales: " + ratString(r1) + " y " + ratString(r2),
	}
	return factors, steps, true, nil
}

func quadraticCoefficients(p Polynomial) (name string, a, b, c *big.Rat, ok bool) {
	name, coeffs, ok := univariateCoeffs(p)
	if !ok || name == "" {
		return "", nil, nil, nil, false
	}
	maxDeg := 0
	for d := range coeffs {
		if d > maxDeg {
			maxDeg = d
		}
	}
	if maxDeg != 2 {
		return "", nil, nil, nil, false
	}
	a = coeffs[2]
	if a == nil || a.Sign() == 0 {
		return "", nil, nil, nil, false
	}
	b = coeffs[1]
	if b == nil {
		b = ratZero()
	}
	c = coeffs[0]
	if c == nil {
		c = ratZero()
	}
	return name, a, b, c, true
}

func quadraticRoots(a, b, c *big.Rat) (disc, r1, r2 *big.Rat, ok bool, err error) {
	disc = ratSub(ratMul(b, b), ratMul(ratFromInt64(4), ratMul(a, c)))
	sqrtD, ok := ratSqrt(disc)
	if !ok {
		return nil, nil, nil, false, nil
	}
	twoA := ratMul(ratFromInt64(2), a)
	r1, err = ratQuo(ratAdd(ratNeg(b), sqrtD), twoA)
	if err != nil {
		return nil, nil, nil, false, err
	}
	r2, err = ratQuo(ratSub(ratNeg(b), sqrtD), twoA)
	if err != nil {
		return nil, nil, nil, false, err
	}
	return disc, r1, r2, true, nil
}

func quadraticFactors(name string, a, r1, r2 *big.Rat) ([]Polynomial, error) {
	f1, q1, err := linearFactor(name, r1)
	if err != nil {
		return nil, err
	}
	f2, q2, err := linearFactor(name, r2)
	if err != nil {
		return nil, err
	}
	qProd := ratMul(ratFromInt(q1), ratFromInt(q2))
	content, err := ratQuo(a, qProd)
	if err != nil {
		return nil, err
	}

	var factors []Polynomial
	if content.Cmp(ratOne()) != 0 {
		factors = append(factors, polynomialFromMonomial(monomialFromRat(content)))
	}
	factors = append(factors, f1)
	if ratCmp(r1, r2) != 0 {
		factors = append(factors, f2)
	} else {
		factors = append(factors, f1)
	}
	return factors, nil
}

func linearFactor(name string, root *big.Rat) (Polynomial, *big.Int, error) {
	pNum := new(big.Int).Set(root.Num())
	qDen := new(big.Int).Set(root.Denom())
	x := monomialVar(name)
	x.Coeff = ratFromInt(qDen)
	c := monomialFromRat(ratFromInt(pNum)).neg()
	p, err := polyFromMonomials(x, c)
	if err != nil {
		return nil, nil, err
	}
	return p, qDen, nil
}

// univariateCoeffs devuelve los coeficientes de un polinomio en una sola
// variable, con exponentes enteros no negativos.
//
//nolint:nestif // validación de una sola variable y sus exponentes
func univariateCoeffs(p Polynomial) (string, map[int]*big.Rat, bool) {
	name := ""
	coeffs := map[int]*big.Rat{}
	for _, m := range p {
		deg := 0
		if len(m.Factors) > 0 {
			if len(m.Factors) != 1 {
				return "", nil, false
			}
			for b, e := range m.Factors {
				if b.Kind != BaseVar {
					return "", nil, false
				}
				if name == "" {
					name = b.Var
				} else if name != b.Var {
					return "", nil, false
				}
				if !e.IsInt() || e.Sign() < 0 {
					return "", nil, false
				}
				n := e.Num()
				if !n.IsInt64() {
					return "", nil, false
				}
				deg = int(n.Int64())
			}
		}
		if old, ok := coeffs[deg]; ok {
			coeffs[deg] = ratAdd(old, m.Coeff)
		} else {
			coeffs[deg] = ratCopy(m.Coeff)
		}
	}
	for d, c := range coeffs {
		if c.Sign() == 0 {
			delete(coeffs, d)
		}
	}
	return name, coeffs, true
}

func ratSqrt(r *big.Rat) (*big.Rat, bool) {
	if r.Sign() < 0 {
		return nil, false
	}
	n := new(big.Int).Sqrt(r.Num())
	if new(big.Int).Mul(n, n).Cmp(r.Num()) != 0 {
		return nil, false
	}
	d := new(big.Int).Sqrt(r.Denom())
	if new(big.Int).Mul(d, d).Cmp(r.Denom()) != 0 {
		return nil, false
	}
	return new(big.Rat).SetFrac(n, d), true
}

func ratGCD(a, b *big.Rat) *big.Rat {
	a, b = ratAbs(a), ratAbs(b)
	if a.Sign() == 0 {
		return ratCopy(b)
	}
	if b.Sign() == 0 {
		return ratCopy(a)
	}
	num := new(big.Int).GCD(nil, nil, a.Num(), b.Num())
	den := new(big.Int).Mul(a.Denom(), b.Denom())
	g := new(big.Int).GCD(nil, nil, a.Denom(), b.Denom())
	den.Div(den, g)
	return new(big.Rat).SetFrac(num, den)
}

// factorByGrouping intenta agrupar los cuatro términos en dos pares con un
// binomio común.
func factorByGrouping(p Polynomial) ([]Polynomial, []string, bool, error) {
	if len(p) != 4 {
		return nil, nil, false, nil
	}
	pairs := [][2][2]int{
		{{0, 1}, {2, 3}},
		{{0, 2}, {1, 3}},
		{{0, 3}, {1, 2}},
	}
	for _, pair := range pairs {
		a := Polynomial{p[pair[0][0]], p[pair[0][1]]}
		b := Polynomial{p[pair[1][0]], p[pair[1][1]]}
		ga, qa, err := groupPair(a)
		if err != nil {
			return nil, nil, false, err
		}
		gb, qb, err := groupPair(b)
		if err != nil {
			return nil, nil, false, err
		}
		var sum Polynomial
		switch {
		case polynomialEqual(qa, qb):
			sum, err = polyAddMonomials(ga, gb)
		case polynomialEqual(qa, qb.neg()):
			sum, err = polyAddMonomials(ga, gb.neg())
		default:
			continue
		}
		if err != nil {
			return nil, nil, false, err
		}
		gbSign := "+"
		gbText := formatFactorMonomial(gb, false)
		if gb.Coeff.Sign() < 0 {
			gbSign = "-"
			gbText = formatFactorMonomial(absMonomial(gb), false)
		}
		steps := []string{
			"Agrupación: " + formatFactorMonomial(ga, false) + "*" +
				formatFactorPolynomial(qa, false) + " " + gbSign + " " +
				gbText + "*" + formatFactorPolynomial(qb, false),
		}
		return []Polynomial{qa, sum}, steps, true, nil
	}
	return nil, nil, false, nil
}

func groupPair(p Polynomial) (*Monomial, Polynomial, error) {
	g, err := commonMonomialFactor(p)
	if err != nil {
		return nil, nil, err
	}
	q, err := p.divMonomial(g)
	if err != nil {
		return nil, nil, err
	}
	if len(q) > 0 && q[0].Coeff.Sign() < 0 {
		g = g.neg()
		q = q.neg()
	}
	return g, q, nil
}

// factorRationalRoots extrae raíces racionales de un polinomio univariado
// de grado >= 2.
func factorRationalRoots(p Polynomial) ([]Polynomial, []string, bool, error) {
	name, coeffs, ok := univariateCoeffs(p)
	if !ok || name == "" {
		return nil, nil, false, nil
	}
	deg := 0
	for d := range coeffs {
		if d > deg {
			deg = d
		}
	}
	if deg < 2 {
		return nil, nil, false, nil
	}
	root, ok := findRationalRoot(coeffs, deg)
	if !ok {
		return nil, nil, false, nil
	}
	f, qDen, err := linearFactor(name, root)
	if err != nil {
		return nil, nil, false, err
	}
	q, ok, err := divideByLinear(coeffs, name, root)
	if err != nil || !ok {
		return nil, nil, false, err
	}
	// divideByLinear divide por (x - r); el factor devuelto es q*x - p.
	// Hay que dividir el cociente por q para que el producto sea exacto.
	q, err = scalePolynomial(q, new(big.Rat).SetFrac(big.NewInt(1), qDen))
	if err != nil {
		return nil, nil, false, err
	}
	step := "Raíz racional: " + name + " = " + ratString(root)
	return []Polynomial{f, q}, []string{step}, true, nil
}

func scalePolynomial(p Polynomial, factor *big.Rat) (Polynomial, error) {
	out := p.clone()
	for _, m := range out {
		m.Coeff = ratMul(m.Coeff, factor)
	}
	return out.canonical()
}

func findRationalRoot(coeffs map[int]*big.Rat, deg int) (*big.Rat, bool) {
	lead := coeffs[deg]
	cons := coeffs[0]
	if lead == nil || cons == nil || lead.Sign() == 0 || cons.Sign() == 0 {
		return nil, false
	}
	lcm := big.NewInt(1)
	for _, c := range coeffs {
		g := new(big.Int).GCD(nil, nil, lcm, c.Denom())
		lcm.Mul(lcm, c.Denom())
		lcm.Div(lcm, g)
	}
	leadInt := new(big.Int).Mul(lead.Num(), new(big.Int).Div(lcm, lead.Denom()))
	consInt := new(big.Int).Mul(cons.Num(), new(big.Int).Div(lcm, cons.Denom()))
	pDivs := divisors(consInt)
	qDivs := divisors(leadInt)
	for _, p := range pDivs {
		for _, q := range qDivs {
			for _, sign := range []int64{1, -1} {
				cand := new(big.Rat).SetFrac(new(big.Int).Mul(p, big.NewInt(sign)), q)
				if evalRatCoeffs(coeffs, cand).Sign() == 0 {
					return cand, true
				}
			}
		}
	}
	return nil, false
}

func divisors(n *big.Int) []*big.Int {
	abs := new(big.Int).Abs(n)
	if abs.Sign() == 0 {
		return nil
	}
	one := big.NewInt(1)
	limit := new(big.Int).Sqrt(abs)
	maxIter := big.NewInt(maxFactorDivs)
	var out []*big.Int
	for i := big.NewInt(1); i.Cmp(limit) <= 0; i.Add(i, one) {
		if len(out) > maxFactorDivs || i.Cmp(maxIter) > 0 {
			break
		}
		if new(big.Int).Mod(abs, i).Sign() == 0 {
			out = append(out, new(big.Int).Set(i))
			other := new(big.Int).Div(abs, i)
			if other.Cmp(i) != 0 {
				out = append(out, other)
			}
		}
	}
	return out
}

func evalRatCoeffs(coeffs map[int]*big.Rat, x *big.Rat) *big.Rat {
	deg := 0
	for d := range coeffs {
		if d > deg {
			deg = d
		}
	}
	out := ratZero()
	for i := deg; i >= 0; i-- {
		out = ratMul(out, x)
		if c, ok := coeffs[i]; ok {
			out = ratAdd(out, c)
		}
	}
	return out
}

func divideByLinear(coeffs map[int]*big.Rat, name string, root *big.Rat) (Polynomial, bool, error) {
	deg := 0
	for d := range coeffs {
		if d > deg {
			deg = d
		}
	}
	rem := make([]*big.Rat, deg+1)
	for i := range rem {
		rem[i] = ratZero()
	}
	for d, c := range coeffs {
		rem[d] = ratCopy(c)
	}
	qcoeff := map[int]*big.Rat{}
	for i := deg; i >= 1; i-- {
		c := rem[i]
		qcoeff[i-1] = c
		rem[i] = ratZero()
		rem[i-1] = ratAdd(rem[i-1], ratMul(c, root))
	}
	if rem[0].Sign() != 0 {
		return nil, false, nil
	}
	q := make(Polynomial, 0, len(qcoeff))
	for d, c := range qcoeff {
		if c.Sign() == 0 {
			continue
		}
		m := monomialFromRat(c)
		if d > 0 {
			m.Factors[variableBase(name)] = ratFromInt64(int64(d))
		}
		q = append(q, m)
	}
	out, err := q.canonical()
	return out, err == nil, err
}

func formatFactorProduct(factors []Polynomial, unicode bool) string {
	var order []string
	counts := map[string]int{}
	for _, f := range factors {
		if polynomialIsOne(f) {
			continue
		}
		text := formatFactorPolynomial(f, unicode)
		if text == "" {
			continue
		}
		if _, ok := counts[text]; !ok {
			order = append(order, text)
		}
		counts[text]++
	}
	if len(order) == 0 {
		return "1"
	}
	if len(order) == 1 && counts[order[0]] == 1 {
		s := order[0]
		if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
			s = s[1 : len(s)-1]
		}
		return s
	}
	var parts []string
	for _, text := range order {
		n := counts[text]
		if n == 1 {
			parts = append(parts, text)
			continue
		}
		if needsFactorPowerParens(text) {
			text = "(" + text + ")"
		}
		if unicode {
			parts = append(parts, text+renderSuperscriptInt(n))
		} else {
			parts = append(parts, text+"^"+strconv.Itoa(n))
		}
	}
	sep := "*"
	if unicode {
		sep = "·"
	}
	return strings.Join(parts, sep)
}

// needsFactorPowerParens evita duplicar paréntesis en factores como
// (x - 1), pero protege x^2, x^(-1/2) y 2*x antes de elevar a una potencia.
func needsFactorPowerParens(s string) bool {
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return false
	}
	return strings.ContainsAny(s, "*·/ ^")
}

func formatFactorPolynomial(p Polynomial, unicode bool) string {
	if len(p) == 0 {
		return "0"
	}
	if len(p) == 1 {
		return formatFactorMonomial(p[0], unicode)
	}
	ascii, uni := FormatPolynomial(p)
	if unicode {
		return "(" + uni + ")"
	}
	return "(" + ascii + ")"
}

func formatFactorMonomial(m *Monomial, unicode bool) string {
	if m == nil || m.isZero() {
		return "0"
	}
	bases := sortedMonomialBases(m.Factors)
	sep := "*"
	if unicode {
		sep = "·"
	}
	parts := make([]string, 0, len(bases)+1)
	if coeff := formatMonomialCoeff(m.Coeff, len(bases) > 0); coeff != "" {
		parts = append(parts, coeff)
	}
	for _, b := range bases {
		parts = append(parts, formatBasePower(b, m.Factors[b], unicode))
	}
	return strings.Join(parts, sep)
}

func sortedMonomialBases(factors map[Base]*big.Rat) []Base {
	bases := make([]Base, 0, len(factors))
	for b := range factors {
		bases = append(bases, b)
	}
	sort.Slice(bases, func(i, j int) bool { return compareBase(bases[i], bases[j]) < 0 })
	return bases
}

func formatMonomialCoeff(coeff *big.Rat, hasBases bool) string {
	if hasBases && coeff.Cmp(ratOne()) == 0 {
		return ""
	}
	// el signo se conserva en el primer factor
	if hasBases && coeff.Cmp(ratFromInt64(-1)) == 0 {
		return "-1"
	}
	return ratString(coeff)
}

func formatBasePower(b Base, e *big.Rat, unicode bool) string {
	name := b.Var
	if b.Kind == BasePrime {
		name = strconv.FormatUint(b.Prime, 10)
	}
	if e.Cmp(ratOne()) == 0 {
		return name
	}
	if e.IsInt() && e.Num().IsInt64() {
		n := e.Num().Int64()
		if n > 0 && !unicode {
			return name + "^" + strconv.FormatInt(n, 10)
		}
		if n > 0 {
			return name + renderSuperscriptInt64(n)
		}
	}
	return name + "^(" + ratString(e) + ")"
}

func renderSuperscriptInt(n int) string {
	return renderSuperscriptInt64(int64(n))
}

func renderSuperscriptInt64(n int64) string {
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '0':
			b.WriteRune('⁰')
		case '1':
			b.WriteRune('¹')
		case '2':
			b.WriteRune('²')
		case '3':
			b.WriteRune('³')
		case '4':
			b.WriteRune('⁴')
		case '5':
			b.WriteRune('⁵')
		case '6':
			b.WriteRune('⁶')
		case '7':
			b.WriteRune('⁷')
		case '8':
			b.WriteRune('⁸')
		case '9':
			b.WriteRune('⁹')
		case '-':
			b.WriteRune('⁻')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// evalFactorizar es el dispatcher: enteros al algoritmo actual y
// expresiones/funciones al factorizador simbólico. La bandera -cuadrado
// desvía a la forma canónica de completar el cuadrado.
func evalFactorizar(decl Decl, op string) ([]string, error) {
	cuadrado, err := parseFactorFlags(op)
	if err != nil {
		return nil, err
	}
	if cuadrado {
		return evalCompletarCuadrado(decl)
	}
	if decl.Tipo != DeclExpresion && decl.Tipo != DeclFuncion {
		return nil, errors.New("factorizar opera sobre una expresión, función o entero literal")
	}
	if n, ok := decl.IntegerValue(); ok {
		return evalFactorizarEntero(IntegerDecl(n))
	}
	e, err := ParseExpr(decl.Raw)
	if err != nil {
		return nil, err
	}
	if steps, ok := factorizarRacionalDirecto(e); ok {
		return steps, nil
	}
	p, err := ExprToPolynomial(e)
	if err != nil {
		return factorizarRacionalFallback(e, err)
	}
	return factorizarPolinomioConvertido(e, p)
}

func factorizarPolinomioConvertido(e Expr, p Polynomial) ([]string, error) {
	p, err := p.canonical()
	if err != nil {
		return nil, err
	}
	if len(p) == 0 {
		return []string{"Operación: factorización de polinomios", "0 no tiene factores"}, nil
	}
	if len(p) == 1 && len(p[0].Factors) == 0 && p[0].Coeff.IsInt() && p[0].Coeff.Num().IsInt64() {
		return evalFactorizarEntero(IntegerDecl(p[0].Coeff.Num().Int64()))
	}
	if len(p) == 1 && len(p[0].Factors) == 0 {
		return nil, errors.New("factorizar requiere un entero o un polinomio no constante")
	}
	res, err := FactorizarPolinomio(e)
	if err != nil {
		return nil, fmt.Errorf("factorizar: %w", err)
	}
	return res.Steps, nil
}

// factorizarRacionalDirecto intenta la ruta racional cuando hay una división
// por un denominador no constante, aunque ExprToPolynomial pudiera convertirla
// usando exponentes negativos.
func factorizarRacionalDirecto(e Expr) ([]string, bool) {
	if !exprHasNonConstantDivision(e) {
		return nil, false
	}
	res, err := FactorizarRacional(e)
	if err != nil {
		return nil, false
	}
	return res.Steps, true
}

// factorizarRacionalFallback intenta la ruta racional cuando la expresión no
// se puede convertir directamente en un polinomio (por ejemplo, una división
// por una suma). Si tampoco es racional, conserva el error polinómico.
func factorizarRacionalFallback(e Expr, polyErr error) ([]string, error) {
	if res, err := FactorizarRacional(e); err == nil {
		return res.Steps, nil
	}
	return nil, fmt.Errorf("factorizar: %w", polyErr)
}
