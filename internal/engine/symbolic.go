package engine

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// --- recolección de términos: coeff × resto (producto no numérico) ---

type symTerm struct {
	coeff float64
	rest  Expr // nil cuando no hay parte no numérica (término constante)
}

// collectTerms descompone una suma (canonicalizada) en términos coeff·rest.
// Solo soporta Add con hijos Num/Neg/Mul (el resto del árbol llega ya
// simplificado con esta forma); cualquier otra forma se trata como un
// término unitario.
func collectTerms(e Expr) []symTerm {
	var terms []symTerm
	var walk func(x Expr, sign float64)
	walk = func(x Expr, sign float64) {
		switch n := x.(type) {
		case Num:
			f, err := n.Float()
			if err == nil {
				terms = append(terms, symTerm{coeff: sign * f})
			}
		case Neg:
			walk(n.X, -sign)
		case Add:
			for _, t := range n.Terms {
				walk(t, sign)
			}
		case Mul:
			coeff, rest, ok := mulCoeffRest(n.Factors, sign)
			if !ok {
				terms = append(terms, symTerm{coeff: 1, rest: x})
				return
			}
			terms = append(terms, symTerm{coeff: coeff, rest: restToExpr(rest)})
		default:
			terms = append(terms, symTerm{coeff: sign, rest: x})
		}
	}
	walk(e, 1)
	return terms
}

// mulCoeffRest separa los factores de un producto en coeficiente numérico
// y resto no numérico. ok=false si algún Num está mal formado.
func mulCoeffRest(factors []Expr, sign float64) (coeff float64, rest []Expr, ok bool) {
	coeff = sign
	for _, f := range factors {
		switch ff := f.(type) {
		case Num:
			v, err := ff.Float()
			if err != nil {
				return 0, nil, false
			}
			coeff *= v
		case Neg:
			inner := ff.X
			if num, isn := inner.(Num); isn {
				v, err := num.Float()
				if err == nil {
					coeff *= -v
					continue
				}
			}
			coeff = -coeff
			rest = append(rest, inner)
		default:
			rest = append(rest, f)
		}
	}
	return coeff, rest, true
}

// restToExpr reconstruye el resto de un término: nil si vacío, el único
// factor si hay uno, o un producto con todos.
func restToExpr(rest []Expr) Expr {
	switch len(rest) {
	case 0:
		return nil
	case 1:
		return rest[0]
	default:
		return Mul{Factors: rest}
	}
}

// termKey identifica un término: el resto en forma canónica.
func termKey(t symTerm) string {
	if t.rest == nil {
		return ""
	}
	return t.rest.Ascii()
}

// mergeTerms agrupa términos con el mismo resto y reordena las constantes
// primero (detección de polinomios). Devuelve la lista agrupada.
func mergeTerms(terms []symTerm) []symTerm {
	var order []string
	byKey := map[string]float64{}
	rests := map[string]Expr{}
	for _, t := range terms {
		k := termKey(t)
		if _, seen := byKey[k]; !seen {
			order = append(order, k)
			rests[k] = t.rest
		}
		byKey[k] += t.coeff
	}
	// las constantes (rest "") van primero
	var out []symTerm
	for _, k := range order {
		if k == "" {
			continue
		}
		out = append(out, symTerm{coeff: byKey[k], rest: rests[k]})
	}
	if c, ok := byKey[""]; ok && c != 0 {
		out = append(out, symTerm{coeff: c})
	}
	return out
}

// rebuildTerms reconstruye el Expr de una lista de términos agrupada,
// descartando coeficientes nulos.
func rebuildTerms(terms []symTerm) Expr {
	var parts []Expr
	for _, t := range terms {
		if t.coeff == 0 {
			continue
		}
		parts = append(parts, termToExpr(t))
	}
	if len(parts) == 0 {
		return Num{Text: "0"}
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return Add{Terms: parts}
}

func termToExpr(t symTerm) Expr {
	c, r := t.coeff, t.rest
	if r == nil {
		return Num{Text: NumFromFloat(c).Text}
	}
	if c == 1 {
		return r
	}
	if c == -1 {
		return Neg{X: r}
	}
	return Mul{Factors: []Expr{Num{Text: NumFromFloat(c).Text}, r}}
}

// simplifyNormaliza es el simplificador recursivo: pliega constantes,
// fusiona potencias y agrupa términos semejantes.
func simplifyNormaliza(e Expr) Expr {
	switch n := e.(type) {
	case Num, Var, Const:
		return e
	case Neg:
		return simplifyNeg(n)
	case Add:
		var children []Expr
		for _, t := range n.Terms {
			children = append(children, simplifyNormaliza(t))
		}
		return rebuildTerms(mergeTerms(collectTerms(Add{Terms: children})))
	case Mul:
		return simplifyMulForzado(n)
	case Div:
		return simplifyDiv(n)
	case Pow:
		return simplifyPow(n)
	case Call:
		return Call{Name: n.Name, Arg: simplifyNormaliza(n.Arg)}
	}
	return e
}

// simplifyNeg pliega la negación de constantes y el doble signo.
func simplifyNeg(n Neg) Expr {
	x := simplifyNormaliza(n.X)
	switch xx := x.(type) {
	case Num:
		f, err := xx.Float()
		if err == nil {
			return Num{Text: NumFromFloat(-f).Text}
		}
	case Neg:
		return simplifyNormaliza(xx.X)
	}
	return Neg{X: x}
}

// simplifyDiv pliega cocientes de constantes, elimina el denominador 1
// y sube los signos del denominador.
func simplifyDiv(n Div) Expr {
	num, den := simplifyNormaliza(n.Num), simplifyNormaliza(n.Den)
	if nf, ok := matchNum(num); ok {
		if nf == 0 {
			return Num{Text: "0"}
		}
		if df, ok := matchNum(den); ok && df != 0 {
			return Num{Text: NumFromFloat(nf / df).Text}
		}
	}
	if df, ok := matchNum(den); ok && df == 1 {
		return num
	}
	if neg, ok := den.(Neg); ok {
		// el signo va arriba: a/(-b) → -(a/b)
		return Neg{X: simplifyNormaliza(Div{Num: num, Den: neg.X})}
	}
	return Div{Num: num, Den: den}
}

// simplifyPow pliega potencias triviales (exponente 0/1/−1, base 1) y
// potencias de constantes.
func simplifyPow(n Pow) Expr {
	base, exp := simplifyNormaliza(n.Base), simplifyNormaliza(n.Exp)
	if ef, ok := matchNum(exp); ok {
		switch ef {
		case 0:
			return Num{Text: "1"}
		case 1:
			return base
		case -1:
			// x^(-1) se muestra como 1/x
			return Div{Num: Num{Text: "1"}, Den: base}
		}
		if bf, ok := matchNum(base); ok {
			if v, valid := powFloat(bf, ef); valid {
				return Num{Text: NumFromFloat(v).Text}
			}
		}
	}
	if bf, ok := matchNum(base); ok && bf == 1 {
		return Num{Text: "1"}
	}
	return Pow{Base: base, Exp: exp}
}

// simplifyMulForzado aplana productos, fusiona potencias de la misma base
// y agrupa los coeficientes numéricos en un solo factor.
func simplifyMulForzado(n Mul) Expr {
	facts, zero := flattenMul(n)
	if zero {
		return Num{Text: "0"}
	}

	coeff, rest := mergeMulBases(facts)

	if coeff == 0 {
		return Num{Text: "0"}
	}
	if len(rest) == 0 {
		return Num{Text: NumFromFloat(coeff).Text}
	}
	switch coeff {
	case 1:
		if len(rest) == 1 {
			return rest[0]
		}
		return Mul{Factors: rest}
	case -1:
		if len(rest) == 1 {
			return Neg{X: rest[0]}
		}
		return Neg{X: Mul{Factors: rest}}
	default:
		all := append([]Expr{Num{Text: NumFromFloat(coeff).Text}}, rest...)
		return Mul{Factors: all}
	}
}

// flattenMul simplifica cada factor y aplana los productos anidados.
// El segundo valor indica si apareció un factor cero (anula el producto).
func flattenMul(n Mul) ([]Expr, bool) {
	var facts []Expr
	for _, f := range n.Factors {
		s := simplifyNormaliza(f)
		if m, ok := s.(Mul); ok {
			facts = append(facts, m.Factors...)
			continue
		}
		if z, ok := matchNum(s); ok && z == 0 {
			return nil, true
		}
		facts = append(facts, s)
	}
	return facts, false
}

// mergeMulBases agrupa factores con la misma base de potencia sumando
// exponentes y acumula el coeficiente numérico en un solo valor.
func mergeMulBases(facts []Expr) (float64, []Expr) {
	coeff := 1.0
	var rest []Expr
	baseExp := map[string]float64{}
	rests := map[string]Expr{}
	var order []string

	for _, f := range facts {
		switch ff := f.(type) {
		case Num:
			v, _ := ff.Float()
			coeff *= v
		default:
			base, expo := ff, Expr(Num{Text: "1"})
			if pw, ok := ff.(Pow); ok {
				base, expo = pw.Base, pw.Exp
			}
			k := base.Ascii()
			if ef, ok := matchNum(expo); ok {
				if _, seen := baseExp[k]; !seen {
					order = append(order, k)
					rests[k] = base
				}
				baseExp[k] += ef
				continue
			}
			rest = append(rest, ff)
		}
	}
	for _, k := range order {
		e := baseExp[k]
		base := rests[k]
		switch e {
		case 1:
			rest = append(rest, base)
		case 0:
			// se cancela
		default:
			rest = append(rest, Pow{Base: base, Exp: Num{Text: FormatFloatText(e)}})
		}
	}
	return coeff, rest
}

// --- derivada simbólica ---

// Derive calcula la derivada de e respecto a v con la regla de la cadena
// y devuelve los pasos aplicados. El resultado está simplificado.
func Derive(e Expr, v byte) (Expr, []string, error) {
	d := &deriver{v: v}
	res, err := d.derive(simplifyNormaliza(e))
	if err != nil {
		return nil, nil, err
	}
	steps := append([]string{}, d.steps...)
	steps = append(steps, "Regla de la cadena aplicada donde hace falta; simplificando el resultado...")
	return simplifyNormaliza(res), steps, nil
}

type deriver struct {
	v     byte
	steps []string
}

func (d *deriver) rule(msg string) {
	d.steps = append(d.steps, msg)
}

func (d *deriver) derive(e Expr) (Expr, error) {
	switch n := e.(type) {
	case Num:
		d.rule("d/d" + string(d.v) + " (constante " + n.Text + ") = 0")
		return Num{Text: "0"}, nil
	case Var:
		if n.Name == d.v {
			d.rule("d/d" + string(d.v) + " " + string(d.v) + " = 1")
			return Num{Text: "1"}, nil
		}
		return Num{Text: "0"}, nil
	case Const:
		return Num{Text: "0"}, nil
	case Neg:
		return d.deriveNeg(n)
	case Add:
		return d.deriveAdd(n)
	case Mul:
		return d.deriveMul(n)
	case Div:
		return d.deriveDiv(n)
	case Pow:
		return d.derivePow(n)
	case Call:
		return d.deriveCall(n)
	}
	return nil, errors.New("derivada no soportada para este nodo")
}

func (d *deriver) deriveNeg(n Neg) (Expr, error) {
	inner, err := d.derive(n.X)
	if err != nil {
		return nil, err
	}
	return Neg{X: inner}, nil
}

func (d *deriver) deriveAdd(n Add) (Expr, error) {
	d.rule("regla de la suma sobre " + n.Ascii())
	var parts []Expr
	for _, t := range n.Terms {
		dd, err := d.derive(t)
		if err != nil {
			return nil, err
		}
		parts = append(parts, dd)
	}
	return simplifyNormaliza(Add{Terms: parts}), nil
}

func (d *deriver) deriveDiv(n Div) (Expr, error) {
	d.rule("regla del cociente: (u/v)′ = (u′v − uv′)/v²")
	du, err := d.derive(n.Num)
	if err != nil {
		return nil, err
	}
	dv, err := d.derive(n.Den)
	if err != nil {
		return nil, err
	}
	num := Add{Terms: []Expr{
		Mul{Factors: []Expr{du, n.Den}},
		Neg{X: Mul{Factors: []Expr{n.Num, dv}}},
	}}
	den := Pow{Base: n.Den, Exp: Num{Text: "2"}}
	return simplifyNormaliza(Div{Num: num, Den: den}), nil
}

func (d *deriver) deriveMul(n Mul) (Expr, error) {
	d.rule("regla del producto sobre " + n.Ascii())
	// producto de factores: se va aplicando (u·v)' = u'v + uv'
	var parts []Expr
	for _, f := range n.Factors {
		df, err := d.derive(f)
		if err != nil {
			return nil, err
		}
		var others []Expr
		for _, g := range n.Factors {
			if g == f {
				continue
			}
			others = append(others, g)
		}
		term := Mul{Factors: append([]Expr{df}, others...)}
		parts = append(parts, simplifyNormaliza(term))
	}
	return simplifyNormaliza(Add{Terms: parts}), nil
}

func (d *deriver) derivePow(n Pow) (Expr, error) {
	if !isVarIn(n.Exp, d.v) {
		d.rule("regla de la potencia: (u^n)′ = n·u^(n−1)·u′  sobre " + n.Ascii())
		du, err := d.derive(n.Base)
		if err != nil {
			return nil, err
		}
		return Mul{Factors: []Expr{
			n.Exp,
			Pow{Base: n.Base, Exp: Add{Terms: []Expr{n.Exp, Neg{X: Num{Text: "1"}}}}},
			du,
		}}, nil
	}
	if !isVarIn(n.Base, d.v) {
		// a^u → a^u · ln(a) · u'
		d.rule("regla exponencial: (a^u)′ = a^u·ln(a)·u′  sobre " + n.Ascii())
		du, err := d.derive(n.Exp)
		if err != nil {
			return nil, err
		}
		return Mul{Factors: []Expr{n, Call{Name: "ln", Arg: n.Base}, du}}, nil
	}
	return nil, errors.New("derivada de v^v no soportada (funciones como x^x)")
}

func (d *deriver) deriveCall(n Call) (Expr, error) {
	du, err := d.derive(n.Arg)
	if err != nil {
		return nil, err
	}
	d.rule("regla de la cadena sobre " + n.Ascii())
	var out Expr
	switch n.Name {
	case fnSqrt:
		out = Div{Num: du, Den: Mul{Factors: []Expr{Num{Text: "2"}, Call{Name: fnSqrt, Arg: n.Arg}}}}
	case fnSin:
		out = Mul{Factors: []Expr{Call{Name: fnCos, Arg: n.Arg}, du}}
	case fnCos:
		out = Neg{X: Mul{Factors: []Expr{Call{Name: fnSin, Arg: n.Arg}, du}}}
	case "abs":
		// |u|' = u·u' / |u|
		out = Div{Num: Mul{Factors: []Expr{n.Arg, du}}, Den: n}
	default:
		return nil, fmt.Errorf("derivada de %s() no soportada", n.Name)
	}
	return simplifyNormaliza(out), nil
}

// --- simplificación (API pública) ---

// Simplify devuelve la forma canónica: suma agrupada por términos
// semejantes con el coeficiente numérico aislado.
func Simplify(e Expr) (Expr, []string, error) {
	orig := e.Ascii()
	res := simplifyNormaliza(e)
	var steps []string
	if res.Ascii() != orig {
		steps = append(steps,
			"Agrupando términos semejantes y plegando constantes...",
			orig+"  →  "+res.Ascii())
	} else {
		steps = append(steps, "La expresión ya está en forma simplificada")
	}
	return res, steps, nil
}

// --- resolución de ecuaciones ---

// Solve resuelve la ecuación lhs=rhs (o expr=0 con rhs nil) para la
// variable v. Soporta polinomios de grado ≤ 2 con coeficientes numéricos
// (la solución lineal admite constantes simbólicas).
func Solve(lhs, rhs Expr, v byte) ([]string, error) {
	whole := lhs
	if rhs != nil {
		whole = Add{Terms: []Expr{lhs, Neg{X: rhs}}}
	}
	if !isVarIn(whole, v) {
		return nil, fmt.Errorf("la expresión no contiene la variable %q", string(v))
	}
	whole = simplifyNormaliza(whole)

	var steps []string
	steps = append(steps,
		"Igualando a cero: "+whole.Ascii(),
		"Extrayendo los coeficientes de "+string(v)+" ...")

	coeffs := map[int]float64{}
	for _, t := range collectTerms(whole) {
		deg := termDegree(t.rest, v)
		if deg < 0 {
			return nil, fmt.Errorf("término %q depende de %q de forma no polinómica — resolución analítica no soportada aún",
				termKey(t), string(v))
		}
		coeffs[deg] += t.coeff
	}

	a, b, c := coeffs[2], coeffs[1], coeffs[0]

	switch {
	case a != 0:
		steps = append(steps, "Ecuación cuadrática: "+formatNum(a)+"·"+string(v)+"² + "+formatNum(b)+"·"+string(v)+" + "+formatNum(c)+" = 0")
		disc := b*b - 4*a*c
		steps = append(steps, "Discriminante Δ = b² − 4ac = "+formatNum(disc))
		if disc < 0 {
			steps = append(steps, "Δ < 0: sin soluciones reales (raíces complejas)")
			return steps, nil
		}
		sd := math.Sqrt(disc)
		x1 := (-b + sd) / (2 * a)
		x2 := (-b - sd) / (2 * a)
		if math.Abs(x1-x2) < 1e-12 {
			steps = append(steps, "Raíz doble: "+string(v)+" = "+formatNum(x1))
			return steps, nil
		}
		steps = append(steps,
			string(v)+" = "+formatNum(x1),
			string(v)+" = "+formatNum(x2))
		return steps, nil

	case b == 0:
		if c == 0 {
			steps = append(steps, "Identidad: cualquier "+string(v)+" es solución")
			return steps, nil
		}
		return nil, fmt.Errorf("ecuación sin solución: %g = 0 es falso", c)

	default:
		cExpr := constTermExpr(whole, v)
		steps = append(steps, "Ecuación lineal: "+formatNum(b)+"·"+string(v)+" + "+cExprAscii(cExpr)+" = 0")
		if ce, ok := matchNum(cExpr); ok {
			steps = append(steps, "Despejando: "+string(v)+" = "+formatNum(-ce/b))
			return steps, nil
		}
		sol := simplifyNormaliza(Div{Num: Neg{X: cExpr}, Den: Num{Text: NumFromFloat(b).Text}})
		steps = append(steps, "Despejando: "+string(v)+" = "+sol.Ascii())
		return steps, nil
	}
}

// termDegree devuelve el grado del término respecto a v:
//
//	0 para constantes, 1 para v, 2 para v^2 y -1 si no es polinómico.
func termDegree(rest Expr, v byte) int {
	if rest == nil {
		return 0
	}
	switch n := rest.(type) {
	case Var:
		if n.Name == v {
			return 1
		}
		return -1
	case Pow:
		if vn, ok := n.Base.(Var); ok && vn.Name == v {
			if ef, ok := matchNum(n.Exp); ok {
				return int(ef)
			}
		}
		return -1
	case Call:
		if isVarIn(n, v) {
			return -1
		}
		return 0
	case Neg:
		return termDegree(n.X, v)
	case Mul:
		d := 0
		for _, f := range n.Factors {
			fd := termDegree(f, v)
			if fd < 0 {
				return -1
			}
			d += fd
		}
		return d
	case Div:
		return -1 // cociente: no polinómico para resolver
	}
	return -1
}

// constTermExpr devuelve la expresión de los términos constantes respecto
// a v (útil para soluciones lineales con constantes simbólicas).
func constTermExpr(e Expr, v byte) Expr {
	var consts []Expr
	for _, t := range collectTerms(e) {
		if termDegree(t.rest, v) == 0 {
			consts = append(consts, termToExpr(symTerm{coeff: t.coeff, rest: t.rest}))
		}
	}
	if len(consts) == 0 {
		return Num{Text: "0"}
	}
	if len(consts) == 1 {
		return consts[0]
	}
	return Add{Terms: consts}
}

func cExprAscii(e Expr) string {
	if n, ok := e.(Num); ok && !n.isNeg() && !n.isZero() {
		return n.Text
	}
	if n, ok := e.(Num); ok && n.isNeg() {
		return "(" + n.Text + ")"
	}
	return "(" + e.Ascii() + ")"
}

// --- integral indefinida ---

// Integrate calcula la antiderivada término a término para polinomios y
// las formas básicas sin(x), cos(x) y sqrt(x). El resto devuelve error.
func Integrate(e Expr, v byte) (Expr, []string, error) {
	e = simplifyNormaliza(e)
	vars := varList(e)
	other := ""
	var otherSb strings.Builder
	for _, vv := range vars {
		if vv != v {
			otherSb.WriteString(string(vv))
		}
	}
	other += otherSb.String()
	if other != "" {
		return nil, nil, fmt.Errorf("integral con otras variables (%s): no soportada aún", other)
	}

	var steps []string
	var parts []symTerm
	for _, t := range collectTerms(e) {
		steps = append(steps, "Regla de la potencia: ∫xⁿ dx = xⁿ⁺¹/(n+1)  (n ≠ −1)")
		ant, ok := integrateTerm(t, v)
		if !ok {
			return nil, nil, fmt.Errorf("término %q: integral no soportada (necesita partes o sustitución)", termKey(t))
		}
		parts = append(parts, ant)
	}
	if len(parts) == 0 {
		return Num{Text: "0"}, append(steps, "Integral de 0: C"), nil
	}
	res := simplifyNormaliza(rebuildTerms(mergeTerms(parts)))
	return res, steps, nil
}

// integrateTerm integra un término coeff·rest respecto a v.
func integrateTerm(t symTerm, v byte) (symTerm, bool) {
	if t.rest == nil {
		return symTerm{coeff: t.coeff, rest: Var{Name: v}}, true
	}
	switch n := t.rest.(type) {
	case Var:
		return integrateVar(t.coeff, n, v)
	case Div:
		return integrateDiv(t, n, v)
	case Pow:
		return integratePow(t, n, v)
	case Call:
		return integrateCall(t.coeff, n, v)
	}
	return symTerm{}, false
}

// integrateVar integra c·x o c (variable distinta): ∫c·y dv = c·y·v.
func integrateVar(c float64, n Var, v byte) (symTerm, bool) {
	if n.Name == v {
		// ∫ c·v dv = c·v²/2
		return symTerm{coeff: c / 2, rest: Pow{Base: Var{Name: v}, Exp: Num{Text: "2"}}}, true
	}
	return symTerm{coeff: c, rest: Mul{Factors: []Expr{Var{Name: n.Name}, Var{Name: v}}}}, true
}

// integrateDiv integra cocientes: 1/(polinomio en v) se reescribe como
// potencia negativa y ∫ c/v dv = c·ln|v|.
func integrateDiv(t symTerm, n Div, v byte) (symTerm, bool) {
	c := t.coeff
	if nf, ok := matchNum(n.Num); ok {
		lead, deg, okd := asyDegree(n.Den, v)
		if okd && deg > 0 {
			// c/(lead·v^d) → (c/lead)·v^(−d) y se reaplica la regla
			t2 := symTerm{
				coeff: c * nf / lead,
				rest:  Pow{Base: Var{Name: v}, Exp: Num{Text: FormatFloatText(-float64(deg))}},
			}
			return integrateTerm(t2, v)
		}
	}
	if d, ok := n.Den.(Var); ok && d.Name == v {
		// ∫ c/v dv = c·ln|v|
		return symTerm{coeff: c, rest: Call{Name: "ln", Arg: Var{Name: v}}}, true
	}
	return symTerm{}, false
}

// integratePotencia aplica la regla de la potencia a c·v^e, con casos
// especiales para e=−1 (logaritmo) y e=0.5 (raíz).
func integratePow(t symTerm, n Pow, v byte) (symTerm, bool) {
	vn, ok := n.Base.(Var)
	if !ok || vn.Name != v {
		return symTerm{}, false
	}
	ef, err := numEval(n.Exp, v, 0)
	if err != nil {
		return symTerm{}, false
	}
	switch ef {
	case -1:
		// ∫ v⁻¹ = ln|v|
		return symTerm{coeff: t.coeff, rest: Call{Name: "ln", Arg: Var{Name: v}}}, true
	case 0.5:
		// ∫√v = (2/3)·v·√v
		return symTerm{
			coeff: t.coeff * 2 / 3,
			rest:  Mul{Factors: []Expr{Var{Name: v}, Call{Name: fnSqrt, Arg: Var{Name: v}}}},
		}, true
	}
	// ∫ vᵉ dv = v^(e+1)/(e+1)
	return symTerm{
		coeff: t.coeff / (ef + 1),
		rest:  Pow{Base: Var{Name: v}, Exp: Num{Text: FormatFloatText(ef + 1)}},
	}, true
}

// integrateCall integra sin(v) y cos(v) con argumento exacto v.
func integrateCall(c float64, n Call, v byte) (symTerm, bool) {
	vn, ok := n.Arg.(Var)
	if !ok || vn.Name != v {
		return symTerm{}, false
	}
	switch n.Name {
	case fnSin:
		// ∫ sin(v) dv = −cos(v)
		return symTerm{coeff: -c, rest: Call{Name: fnCos, Arg: Var{Name: v}}}, true
	case fnCos:
		// ∫ cos(v) dv = sin(v)
		return symTerm{coeff: c, rest: Call{Name: fnSin, Arg: Var{Name: v}}}, true
	}
	return symTerm{}, false
}

// --- límites ---

// Limit calcula el límite de e cuando v → a por sustitución directa;
// si resulta indeterminado intenta L'Hôpital (hasta 2 derivadas).
func Limit(e Expr, v byte, a Value) (Expr, []string, error) {
	e = simplifyNormaliza(e)
	if !isVarIn(e, v) {
		return e, []string{"La expresión no depende de " + string(v) + ": el límite es la propia expresión"}, nil
	}
	steps := []string{"Sustituyendo " + string(v) + " → " + a.String() + " ..."}

	if a.IsInf {
		return limitInf(e, v, a.Sign, steps)
	}

	f := e
	for i := range 3 {
		res, err := numEval(f, v, a.Val)
		if err != nil {
			return nil, nil, err
		}
		if !math.IsNaN(res) && !math.IsInf(res, 0) {
			return Num{Text: NumFromFloat(res).Text}, steps, nil
		}
		if i < 2 {
			steps = append(steps, "Forma indeterminada → aplicando L'Hôpital...")
			f = lHopital(f, v)
		}
	}
	return nil, nil, errors.New("límite indeterminado (0/0 o ∞/∞) que no se resuelve con dos derivadas: sin soportar aún")
}

// lHopital aplica un paso de la regla de L'Hôpital: si f es un cociente
// deriva numerador y denominador por separado (más fiable que la regla del
// cociente completa); en cualquier otro caso deriva la expresión entera.
func lHopital(f Expr, v byte) Expr {
	if d, ok := f.(Div); ok {
		dn, _, errn := Derive(d.Num, v)
		if errn == nil {
			dd, _, errd := Derive(d.Den, v)
			if errd == nil {
				return Div{Num: simplifyNormaliza(dn), Den: simplifyNormaliza(dd)}
			}
		}
	}
	f2, _, _ := Derive(f, v)
	return simplifyNormaliza(f2)
}

// Value es el punto donde se evalúa el límite.
type Value struct {
	Val   float64
	IsInf bool
	Sign  int // +1 o -1 cuando IsInf
}

func (v Value) String() string {
	if v.IsInf {
		if v.Sign < 0 {
			return "-∞"
		}
		return "∞"
	}
	return formatNum(v.Val)
}

// limitInf evalúa el límite en ±∞ con la asintótica del término dominante.
func limitInf(e Expr, v byte, sign int, steps []string) (Expr, []string, error) {
	steps = append(steps, "Dominando por el término de mayor grado en "+string(v)+" ...")
	lead, deg, ok := asyDegree(e, v)
	if !ok {
		return nil, nil, errors.New("límite en ∞ con funciones no polinómicas: sin soportar aún")
	}
	switch {
	case deg < 0:
		return Num{Text: "0"}, steps, nil
	case deg == 0:
		return Num{Text: NumFromFloat(lead).Text}, steps, nil
	}
	trend := lead * float64(sign)
	if trend > 0 {
		return Const{Name: "inf"}, steps, nil
	}
	return Neg{X: Const{Name: "inf"}}, steps, nil
}

// asyDegree estima el comportamiento asintótico de e en ±∞: f ~ lead·v^deg.
func asyDegree(e Expr, v byte) (lead float64, deg int, ok bool) {
	switch n := e.(type) {
	case Num:
		f, err := n.Float()
		return f, 0, err == nil
	case Var:
		if n.Name == v {
			return 1, 1, true
		}
		return 1, 0, true
	case Const:
		return 1, 0, true
	case Neg:
		l, d, o := asyDegree(n.X, v)
		return -l, d, o
	case Add:
		return asyDegreeAdd(n, v)
	case Mul:
		return asyDegreeMul(n, v)
	case Div:
		ln, dn, lo := asyDegree(n.Num, v)
		ld, dd, do := asyDegree(n.Den, v)
		if !lo || !do {
			return 0, 0, false
		}
		return ln / ld, dn - dd, true
	case Pow:
		return asyDegreePow(n, v)
	case Call:
		if isVarIn(n, v) {
			return 0, 0, false
		}
		return 1, 0, true
	}
	return 0, 0, false
}

// asyDegreeAdd toma el término dominante (mayor grado) de la suma.
func asyDegreeAdd(n Add, v byte) (float64, int, bool) {
	bestL, bestD := 0.0, math.MinInt
	for _, t := range n.Terms {
		l, d, o := asyDegree(t, v)
		if !o {
			return 0, 0, false
		}
		if d > bestD {
			bestL, bestD = l, d
		}
	}
	return bestL, bestD, true
}

// asyDegreeMul multiplica coeficientes y suma grados de los factores.
func asyDegreeMul(n Mul, v byte) (float64, int, bool) {
	lead, deg := 1.0, 0
	for _, f := range n.Factors {
		l, d, o := asyDegree(f, v)
		if !o {
			return 0, 0, false
		}
		lead *= l
		deg += d
	}
	return lead, deg, true
}

// asyDegreePow estima v^e con exponente numérico; el resto no es
// polinómico y no se soporta.
func asyDegreePow(n Pow, v byte) (float64, int, bool) {
	if b, ok := n.Base.(Var); ok && b.Name == v {
		if ef, err := numEval(n.Exp, v, 0); err == nil {
			return 1, int(ef), true
		}
	}
	if _, ok := n.Base.(Num); ok {
		if _, err := numEval(n.Base, v, 0); err == nil {
			return 1.0, 0, true
		}
	}
	return 0, 0, false
}

// numEval evalúa numéricamente la expresión sustituyendo v por a.
func numEval(e Expr, v byte, a float64) (float64, error) {
	switch n := e.(type) {
	case Num:
		return n.Float()
	case Var:
		if n.Name == v {
			return a, nil
		}
		return 0, fmt.Errorf("variable %q sin sustituir", string(n.Name))
	case Const:
		if n.Name == "pi" {
			return math.Pi, nil
		}
		return 0, errors.New("∞ no se puede sustituir numéricamente")
	case Neg:
		x, err := numEval(n.X, v, a)
		return -x, err
	case Add:
		return numEvalAdd(n, v, a)
	case Mul:
		return numEvalMul(n, v, a)
	case Div:
		return numEvalDiv(n, v, a)
	case Pow:
		return numEvalPow(n, v, a)
	case Call:
		return numEvalCall(n, v, a)
	}
	return 0, errors.New("evaluación numérica no soportada")
}

func numEvalAdd(n Add, v byte, a float64) (float64, error) {
	sum := 0.0
	for _, t := range n.Terms {
		x, err := numEval(t, v, a)
		if err != nil {
			return 0, err
		}
		sum += x
	}
	return sum, nil
}

func numEvalMul(n Mul, v byte, a float64) (float64, error) {
	prod := 1.0
	for _, f := range n.Factors {
		x, err := numEval(f, v, a)
		if err != nil {
			return 0, err
		}
		prod *= x
	}
	return prod, nil
}

func numEvalDiv(n Div, v byte, a float64) (float64, error) {
	num, err := numEval(n.Num, v, a)
	if err != nil {
		return 0, err
	}
	den, err := numEval(n.Den, v, a)
	if err != nil {
		return 0, err
	}
	return num / den, nil
}

func numEvalPow(n Pow, v byte, a float64) (float64, error) {
	b, err := numEval(n.Base, v, a)
	if err != nil {
		return 0, err
	}
	xp, err := numEval(n.Exp, v, a)
	if err != nil {
		return 0, err
	}
	return math.Pow(b, xp), nil
}

func numEvalCall(n Call, v byte, a float64) (float64, error) {
	x, err := numEval(n.Arg, v, a)
	if err != nil {
		return 0, err
	}
	switch n.Name {
	case fnSqrt:
		return math.Sqrt(x), nil
	case fnSin:
		return math.Sin(x), nil
	case fnCos:
		return math.Cos(x), nil
	case "abs":
		return math.Abs(x), nil
	}
	return 0, fmt.Errorf("%s() sin evaluación numérica", n.Name)
}
