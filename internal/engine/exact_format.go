package engine

import (
	"math/big"
	"sort"
	"strconv"
	"strings"
	"tuniversalcalc/internal/render"
)

// FormatPolynomial devuelve la forma ASCII canónica y su versión Unicode.
// El ASCII se puede volver a parsear; el Unicode es solo para mostrar.
func FormatPolynomial(p Polynomial) (ascii, unicode string) {
	return casFormatter{}.polynomial(p), casFormatter{unicode: true}.polynomial(p)
}

type casFormatter struct {
	unicode bool
}

func (f casFormatter) minus() string {
	if f.unicode {
		return "−"
	}
	return "-"
}

func (f casFormatter) dot() string {
	if f.unicode {
		return "·"
	}
	return "*"
}

func (f casFormatter) polynomial(p Polynomial) string {
	if len(p) == 0 {
		return "0"
	}
	var b strings.Builder
	first := true
	for _, m := range p {
		if m.isZero() {
			continue
		}
		abs := m.clone()
		abs.Coeff = ratAbs(abs.Coeff)
		term := f.monomial(abs)
		neg := m.Coeff.Sign() < 0
		switch {
		case first && neg:
			b.WriteString(f.minus())
			b.WriteString(term)
		case first:
			b.WriteString(term)
		case neg:
			b.WriteString(" ")
			b.WriteString(f.minus())
			b.WriteString(" ")
			b.WriteString(term)
		default:
			b.WriteString(" + ")
			b.WriteString(term)
		}
		first = false
	}
	if b.Len() == 0 {
		return "0"
	}
	return b.String()
}

// monomial formatea un monomio con coeficiente no negativo.
func (f casFormatter) monomial(m *Monomial) string {
	numParts, denParts := f.factorParts(m)
	numFactors := append([]string{}, numParts...)
	denFactors := append([]string{}, denParts...)

	coeffNum := new(big.Int).Abs(m.Coeff.Num())
	coeffDen := new(big.Int).Set(m.Coeff.Denom())
	if coeffNum.Cmp(big.NewInt(1)) != 0 || len(numFactors) == 0 {
		numFactors = append([]string{coeffNum.String()}, numFactors...)
	}
	if coeffDen.Cmp(big.NewInt(1)) != 0 {
		denFactors = append([]string{coeffDen.String()}, denFactors...)
	}

	num := f.join(numFactors)
	den := f.join(denFactors)
	if num == "" {
		num = "1"
	}
	if den == "" {
		return num
	}
	if needsDenParens(den) {
		den = "(" + den + ")"
	}
	return num + "/" + den
}

// factorParts separa los factores en el numerador y el denominador. Las
// partes fraccionarias se agrupan por índice de raíz.
//
//nolint:nestif // los dos caminos (variable/primo) comparten la agrupación de radicales
func (f casFormatter) factorParts(m *Monomial) (num, den []string) {
	numRad := map[int][]radicalFactor{}
	denRad := map[int][]radicalFactor{}
	bases := make([]Base, 0, len(m.Factors))
	for b := range m.Factors {
		bases = append(bases, b)
	}
	sort.Slice(bases, func(i, j int) bool { return compareBase(bases[i], bases[j]) < 0 })

	for _, b := range bases {
		e := m.Factors[b]
		if e == nil || e.Sign() == 0 {
			continue
		}
		if b.Kind == BaseVar {
			if e.Sign() > 0 {
				q := ratFloor(e)
				if q.Sign() > 0 {
					num = append(num, f.varPower(b.Var, q))
				}
				if r := ratFrac(e); r.Sign() != 0 {
					addRadical(numRad, r, b)
				}
				continue
			}
			absE := ratNeg(e)
			q := ratFloor(absE)
			if q.Sign() > 0 {
				den = append(den, f.varPower(b.Var, q))
			}
			if r := ratFrac(absE); r.Sign() != 0 {
				addRadical(denRad, r, b)
			}
			continue
		}

		// Los primos canónicos solo conservan su parte fraccionaria; si
		// llega una parte entera se ignora porque ya está en el coeficiente.
		if e.Sign() > 0 {
			if r := ratFrac(e); r.Sign() != 0 {
				addRadical(numRad, r, b)
			}
			continue
		}
		absE := ratNeg(e)
		if r := ratFrac(absE); r.Sign() != 0 {
			addRadical(denRad, r, b)
		}
	}

	num = append(num, f.radicalGroups(numRad)...)
	den = append(den, f.radicalGroups(denRad)...)
	return num, den
}

type radicalFactor struct {
	base  Base
	power *big.Int
}

func addRadical(groups map[int][]radicalFactor, r *big.Rat, b Base) {
	idx := r.Denom()
	if !idx.IsInt64() {
		return
	}
	groups[int(idx.Int64())] = append(groups[int(idx.Int64())], radicalFactor{
		base:  b,
		power: new(big.Int).Set(r.Num()),
	})
}

func (f casFormatter) radicalGroups(groups map[int][]radicalFactor) []string {
	if len(groups) == 0 {
		return nil
	}
	indices := make([]int, 0, len(groups))
	for idx := range groups {
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	out := make([]string, 0, len(indices))
	for _, idx := range indices {
		out = append(out, f.radical(idx, groups[idx]))
	}
	return out
}

func (f casFormatter) radical(index int, factors []radicalFactor) string {
	primeProd := big.NewInt(1)
	var varParts []string
	for _, rf := range factors {
		if rf.base.Kind == BasePrime {
			p := new(big.Int).SetUint64(rf.base.Prime)
			p.Exp(p, rf.power, nil)
			primeProd.Mul(primeProd, p)
			continue
		}
		varParts = append(varParts, f.varPower(rf.base.Var, rf.power))
	}
	var parts []string
	if primeProd.Cmp(big.NewInt(1)) != 0 {
		parts = append(parts, primeProd.String())
	}
	parts = append(parts, varParts...)
	rad := f.join(parts)
	if rad == "" {
		rad = "1"
	}
	if f.unicode {
		sym := "√"
		switch index {
		case 3:
			sym = "∛"
		case 4:
			sym = "∜"
		default:
			if index != 2 {
				sym = render.ToSuperscript(strconv.Itoa(index)) + "√"
			}
		}
		if needsRootParens(rad) {
			return sym + "(" + rad + ")"
		}
		return sym + rad
	}
	if index == 2 {
		return "sqrt(" + rad + ")"
	}
	return "raiz(" + strconv.Itoa(index) + ", " + rad + ")"
}

func (f casFormatter) varPower(name string, exp *big.Int) string {
	if exp == nil || exp.Cmp(big.NewInt(1)) == 0 {
		return name
	}
	if f.unicode {
		return name + render.ToSuperscript(exp.String())
	}
	return name + "^" + exp.String()
}

func (f casFormatter) join(parts []string) string {
	var clean []string
	for _, p := range parts {
		if p != "" {
			clean = append(clean, p)
		}
	}
	return strings.Join(clean, f.dot())
}

func needsRootParens(s string) bool {
	return strings.ContainsAny(s, "+−·/ *()⁰¹²³⁴⁵⁶⁷⁸⁹")
}

func needsDenParens(s string) bool {
	return strings.ContainsAny(s, "*·+/ −")
}
