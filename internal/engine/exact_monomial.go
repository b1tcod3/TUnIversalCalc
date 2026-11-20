package engine

import (
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

// BaseKind distingue una variable de un factor primo numérico.
type BaseKind int

const (
	BasePrime BaseKind = iota
	BaseVar
)

// Base es una base del monomio: una variable simbólica o un primo.
type Base struct {
	Kind  BaseKind
	Var   string
	Prime uint64
}

func primeBase(p uint64) Base { return Base{Kind: BasePrime, Prime: p} }

func variableBase(name string) Base { return Base{Kind: BaseVar, Var: name} }

func compareBase(a, b Base) int {
	if a.Kind != b.Kind {
		if a.Kind < b.Kind {
			return -1
		}
		return 1
	}
	if a.Kind == BasePrime {
		switch {
		case a.Prime < b.Prime:
			return -1
		case a.Prime > b.Prime:
			return 1
		}
		return 0
	}
	return strings.Compare(a.Var, b.Var)
}

func baseKey(b Base) string {
	if b.Kind == BasePrime {
		return "p:" + strconv.FormatUint(b.Prime, 10)
	}
	return "v:" + b.Var
}

// Monomial es coeficiente exacto × producto de bases elevadas a exponentes
// racionales exactos.
type Monomial struct {
	Coeff   *big.Rat
	Factors map[Base]*big.Rat
}

func newMonomial() *Monomial {
	return &Monomial{Coeff: ratOne(), Factors: map[Base]*big.Rat{}}
}

func zeroMonomial() *Monomial {
	return &Monomial{Coeff: ratZero(), Factors: map[Base]*big.Rat{}}
}

func monomialFromRat(c *big.Rat) *Monomial {
	m := newMonomial()
	m.Coeff = ratCopy(c)
	if m.Coeff.Sign() == 0 {
		return zeroMonomial()
	}
	return m
}

func monomialVar(name string) *Monomial {
	m := newMonomial()
	m.Factors[variableBase(name)] = ratOne()
	return m
}

func (m *Monomial) clone() *Monomial {
	if m == nil {
		return zeroMonomial()
	}
	out := &Monomial{Coeff: ratCopy(m.Coeff), Factors: make(map[Base]*big.Rat, len(m.Factors))}
	for b, e := range m.Factors {
		out.Factors[b] = ratCopy(e)
	}
	return out
}

func (m *Monomial) isZero() bool {
	return m == nil || m.Coeff == nil || m.Coeff.Sign() == 0
}

func (m *Monomial) neg() *Monomial {
	out := m.clone()
	out.Coeff = ratNeg(out.Coeff)
	return out
}

// canonical normaliza exponentes: los primos con parte entera pasan al
// coeficiente y solo queda su parte fraccionaria en el radical.
func (m *Monomial) canonical() (*Monomial, error) {
	if m == nil {
		return zeroMonomial(), nil
	}
	out := &Monomial{Coeff: ratOne(), Factors: map[Base]*big.Rat{}}
	if m.Coeff != nil {
		out.Coeff = ratCopy(m.Coeff)
	}
	if out.Coeff.Sign() == 0 {
		return zeroMonomial(), nil
	}
	for b, e := range m.Factors {
		if e == nil || e.Sign() == 0 {
			continue
		}
		if b.Kind == BaseVar {
			out.Factors[b] = ratCopy(e)
			continue
		}
		q := ratFloor(e)
		r := ratFrac(e)
		if q.Sign() != 0 {
			if !q.IsInt64() {
				return nil, errors.New("exponente primo demasiado grande")
			}
			pPow, err := ratPowInt(ratFromUint64(b.Prime), q.Int64())
			if err != nil {
				return nil, err
			}
			out.Coeff = ratMul(out.Coeff, pPow)
		}
		if r.Sign() != 0 {
			out.Factors[b] = r
		}
	}
	if out.Coeff.Sign() == 0 {
		return zeroMonomial(), nil
	}
	return out, nil
}

func (m *Monomial) mul(o *Monomial) (*Monomial, error) {
	if m.isZero() || o.isZero() {
		return zeroMonomial(), nil
	}
	out := m.clone()
	out.Coeff = ratMul(m.Coeff, o.Coeff)
	for b, e := range o.Factors {
		if old, ok := out.Factors[b]; ok {
			out.Factors[b] = ratAdd(old, e)
		} else {
			out.Factors[b] = ratCopy(e)
		}
	}
	return out.canonical()
}

func (m *Monomial) div(o *Monomial) (*Monomial, error) {
	if o.isZero() {
		return nil, errors.New("división por un monomio cero")
	}
	out := m.clone()
	c, err := ratQuo(m.Coeff, o.Coeff)
	if err != nil {
		return nil, err
	}
	out.Coeff = c
	for b, e := range o.Factors {
		if old, ok := out.Factors[b]; ok {
			out.Factors[b] = ratSub(old, e)
		} else {
			out.Factors[b] = ratNeg(e)
		}
	}
	return out.canonical()
}

func (m *Monomial) pow(r *big.Rat) (*Monomial, error) {
	if m.isZero() {
		if r.Sign() <= 0 {
			return nil, errors.New("0 elevado a un exponente no positivo no está definido")
		}
		return zeroMonomial(), nil
	}
	if r.Sign() == 0 {
		return newMonomial(), nil
	}
	out, err := monomialPowRat(m.Coeff, r)
	if err != nil {
		return nil, err
	}
	for b, e := range m.Factors {
		ne := ratMul(e, r)
		if old, ok := out.Factors[b]; ok {
			out.Factors[b] = ratAdd(old, ne)
		} else {
			out.Factors[b] = ne
		}
	}
	return out.canonical()
}

// monomialPowRat eleva un coeficiente racional a un exponente racional.
// Para exponentes fraccionarios factoriza numerador y denominador y extrae
// las potencias perfectas.
func monomialPowRat(coeff, r *big.Rat) (*Monomial, error) {
	if r.IsInt() {
		n, ok := ratInt64(r)
		if !ok {
			return nil, errors.New("exponente demasiado grande")
		}
		c, err := ratPowInt(coeff, n)
		if err != nil {
			return nil, err
		}
		return monomialFromRat(c), nil
	}
	if coeff.Sign() == 0 {
		if r.Sign() <= 0 {
			return nil, errors.New("0 elevado a un exponente no positivo no está definido")
		}
		return zeroMonomial(), nil
	}

	sign := coeff.Sign()
	abs := ratAbs(coeff)
	numFactors, err := factorRatPart(abs.Num())
	if err != nil {
		return nil, err
	}
	denFactors, err := factorRatPart(abs.Denom())
	if err != nil {
		return nil, err
	}

	out := newMonomial()
	for _, g := range numFactors {
		addPrimeExponent(out, g.Prime, ratMul(ratFromInt64(int64(g.Exp)), r))
	}
	for _, g := range denFactors {
		addPrimeExponent(out, g.Prime, ratMul(ratFromInt64(-int64(g.Exp)), r))
	}
	if sign < 0 {
		if r.Denom().Bit(0) == 0 {
			return nil, errors.New("raíz de índice par de una cantidad negativa: los complejos no están soportados")
		}
		// (-a)^(p/q) es negativo solo si p es impar.
		if new(big.Int).Abs(r.Num()).Bit(0) == 1 {
			out.Coeff = ratNeg(out.Coeff)
		}
	}
	return out.canonical()
}

func addPrimeExponent(m *Monomial, p uint64, e *big.Rat) {
	b := primeBase(p)
	if old, ok := m.Factors[b]; ok {
		m.Factors[b] = ratAdd(old, e)
		return
	}
	m.Factors[b] = ratCopy(e)
}

// factorRatPart factoriza un entero que quepa en uint64.
func factorRatPart(n *big.Int) ([]PrimePower, error) {
	if n == nil || n.Sign() == 0 {
		return nil, nil
	}
	abs := new(big.Int).Abs(n)
	if abs.Cmp(big.NewInt(1)) == 0 {
		return nil, nil
	}
	if !abs.IsUint64() {
		return nil, errors.New("coeficiente demasiado grande para simplificar raíces exactas")
	}
	return groupFactors(factorizeUint64(abs.Uint64())), nil
}

// key identifica la parte no numérica de un monomio; dos monomios con la
// misma clave pueden sumarse como términos semejantes.
func (m *Monomial) key() string {
	if m == nil {
		return ""
	}
	bases := make([]Base, 0, len(m.Factors))
	for b, e := range m.Factors {
		if e != nil && e.Sign() != 0 {
			bases = append(bases, b)
		}
	}
	sort.Slice(bases, func(i, j int) bool { return compareBase(bases[i], bases[j]) < 0 })
	var b strings.Builder
	for _, base := range bases {
		b.WriteString(baseKey(base))
		b.WriteByte('=')
		b.WriteString(ratString(m.Factors[base]))
		b.WriteByte(';')
	}
	return b.String()
}

// variableDegree suma los exponentes de las variables; se usa para ordenar
// los términos del polinomio.
func (m *Monomial) variableDegree() *big.Rat {
	deg := ratZero()
	for b, e := range m.Factors {
		if b.Kind == BaseVar {
			deg = ratAdd(deg, e)
		}
	}
	return deg
}
