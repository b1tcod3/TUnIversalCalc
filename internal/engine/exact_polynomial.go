package engine

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	maxCasTerms = 512
	maxCasPower = 32
)

// Polynomial es una suma de monomios; se mantiene canónico tras cada
// operación: sin términos nulos y con los semejantes agrupados.
type Polynomial []*Monomial

func onePolynomial() Polynomial { return Polynomial{newMonomial()} }

func polynomialFromMonomial(m *Monomial) Polynomial {
	if m.isZero() {
		return nil
	}
	return Polynomial{m}
}

func (p Polynomial) isZero() bool {
	for _, m := range p {
		if !m.isZero() {
			return false
		}
	}
	return true
}

func (p Polynomial) clone() Polynomial {
	out := make(Polynomial, 0, len(p))
	for _, m := range p {
		out = append(out, m.clone())
	}
	return out
}

func (p Polynomial) canonical() (Polynomial, error) {
	byKey := map[string]*Monomial{}
	for _, m := range p {
		cm, err := m.canonical()
		if err != nil {
			return nil, err
		}
		if cm.isZero() {
			continue
		}
		k := cm.key()
		if old, ok := byKey[k]; ok {
			old.Coeff = ratAdd(old.Coeff, cm.Coeff)
			if old.Coeff.Sign() == 0 {
				delete(byKey, k)
			}
			continue
		}
		byKey[k] = cm
	}
	if len(byKey) > maxCasTerms {
		return nil, fmt.Errorf("la expresión genera demasiados términos (%d); límite %d", len(byKey), maxCasTerms)
	}
	out := make(Polynomial, 0, len(byKey))
	for _, m := range byKey {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return termLess(out[i], out[j]) })
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func termLess(a, b *Monomial) bool {
	da, db := a.variableDegree(), b.variableDegree()
	if c := ratCmp(da, db); c != 0 {
		return c > 0
	}
	return compareTermVars(a, b) < 0
}

// compareTermVars ordena por la primera variable y, si coincide, por su
// exponente descendente. Así a^2 va antes que a*b y b^2.
func compareTermVars(a, b *Monomial) int {
	av, bv := variableBases(a), variableBases(b)
	for i := 0; i < len(av) && i < len(bv); i++ {
		if c := strings.Compare(av[i].Var, bv[i].Var); c != 0 {
			return c
		}
		if c := ratCmp(a.Factors[av[i]], b.Factors[bv[i]]); c != 0 {
			return -c
		}
	}
	switch {
	case len(av) < len(bv):
		return -1
	case len(av) > len(bv):
		return 1
	}
	return strings.Compare(a.key(), b.key())
}

func variableBases(m *Monomial) []Base {
	var out []Base
	for base := range m.Factors {
		if base.Kind == BaseVar {
			out = append(out, base)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Var < out[j].Var })
	return out
}

func (p Polynomial) add(q Polynomial) (Polynomial, error) {
	terms := make(Polynomial, 0, len(p)+len(q))
	terms = append(terms, p.clone()...)
	terms = append(terms, q.clone()...)
	return terms.canonical()
}

func (p Polynomial) neg() Polynomial {
	out := make(Polynomial, 0, len(p))
	for _, m := range p {
		out = append(out, m.neg())
	}
	return out
}

func (p Polynomial) mul(q Polynomial) (Polynomial, error) {
	if p.isZero() || q.isZero() {
		return nil, nil
	}
	terms := make(Polynomial, 0, len(p)*len(q))
	for _, a := range p {
		for _, b := range q {
			prod, err := a.mul(b)
			if err != nil {
				return nil, err
			}
			terms = append(terms, prod)
			if len(terms) > maxCasTerms*4 {
				return nil, fmt.Errorf("la multiplicación genera demasiados términos; límite %d", maxCasTerms)
			}
		}
	}
	return terms.canonical()
}

// divMonomial divide cada término por un monomio. La división por una suma
// no se soporta en este motor.
func (p Polynomial) divMonomial(m *Monomial) (Polynomial, error) {
	if m.isZero() {
		return nil, errors.New("división por cero")
	}
	out := make(Polynomial, 0, len(p))
	for _, t := range p {
		q, err := t.div(m)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out.canonical()
}

func (p Polynomial) powInt(n int64) (Polynomial, error) {
	if n == 0 {
		return onePolynomial(), nil
	}
	if n < 0 {
		if len(p) != 1 {
			return nil, errors.New("una suma elevada a un exponente negativo no es un polinomio")
		}
		r := ratFromInt64(-n)
		m, err := p[0].pow(r)
		if err != nil {
			return nil, err
		}
		return polynomialFromMonomial(m), nil
	}
	if len(p) == 1 {
		m, err := p[0].pow(ratFromInt64(n))
		if err != nil {
			return nil, err
		}
		return polynomialFromMonomial(m), nil
	}
	if n > maxCasPower {
		return nil, fmt.Errorf("exponente %d demasiado grande para expandir una suma; límite %d", n, maxCasPower)
	}
	result := onePolynomial()
	for range n {
		var err error
		result, err = result.mul(p)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
