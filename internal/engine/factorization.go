package engine

import (
	"errors"
	"fmt"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"tuniversalcalc/internal/render"
)

// PrimePower es un factor primo con su exponente.
type PrimePower struct {
	Prime uint64
	Exp   int
}

// Factorization es la descomposición de un entero en factores primos.
type Factorization struct {
	Original int64
	Negative bool
	Factors  []uint64     // con repetición: 2, 2, 2, 3, 5
	Groups   []PrimePower // agrupado: 2³, 3¹, 5¹
}

// ParseIntegerLiteral reconoce un entero decimal con signo opcional. No
// acepta decimales, exponentes, espacios internos ni variables.
func ParseIntegerLiteral(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	body := s
	if body[0] == '+' || body[0] == '-' {
		body = body[1:]
	}
	if body == "" {
		return 0, false
	}
	for _, r := range body {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// IntegerValue devuelve el valor entero de una declaración /expresion.
func (d Decl) IntegerValue() (int64, bool) {
	if d.Tipo != DeclExpresion {
		return 0, false
	}
	return ParseIntegerLiteral(d.Raw)
}

// IntegerDecl construye una declaración de expresión a partir de un entero.
// La usa la UI para aceptar /factorizar 120 sin obligar a declarar antes.
func IntegerDecl(n int64) Decl {
	raw := strconv.FormatInt(n, 10)
	return Decl{
		Tipo:   DeclExpresion,
		Nombre: raw,
		Raw:    raw,
		Latex:  raw,
		Pretty: render.Pretty(raw),
	}
}

// absInt64 calcula |n| sin desbordar con math.MinInt64.
func absInt64(n int64) uint64 {
	if n >= 0 {
		return uint64(n)
	}
	// -(n+1) es seguro incluso para MinInt64; sumar 1 después evita el
	// overflow de -MinInt64.
	return uint64(-(n + 1)) + 1 //nolint:gosec // n < 0 garantiza que -(n+1) no es negativo
}

// Factorize descompone n en factores primos. Para 0 y ±1 devuelve una
// factorización vacía con el signo correspondiente.
func Factorize(n int64) Factorization {
	f := Factorization{Original: n, Negative: n < 0}
	if n == 0 || n == 1 || n == -1 {
		return f
	}
	f.Factors = factorizeUint64(absInt64(n))
	f.Groups = groupFactors(f.Factors)
	return f
}

// factorizeUint64 combina la rueda 6k±1 (para factores pequeños) con
// Miller-Rabin y Pollard Rho (para restos grandes).
func factorizeUint64(n uint64) []uint64 {
	if n < 2 {
		return nil
	}
	out := make([]uint64, 0, 8)
	for n%2 == 0 {
		out = append(out, 2)
		n /= 2
	}
	for n%3 == 0 {
		out = append(out, 3)
		n /= 3
	}
	if n == 1 {
		return out
	}
	if isPrime64(n) {
		out = append(out, n)
		return out
	}

	// Rueda 6k±1 para reducir el número con factores pequeños.
	const smallLimit = 1 << 20
	for i := uint64(5); i <= n/i && i <= smallLimit; i += 6 {
		for n%i == 0 {
			out = append(out, i)
			n /= i
		}
		j := i + 2
		for n%j == 0 {
			out = append(out, j)
			n /= j
		}
		if n == 1 {
			break
		}
	}
	if n > 1 {
		factorRec(n, &out)
	}
	slices.Sort(out)
	return out
}

// factorRec factoriza recursivamente un resto compuesto usando Pollard Rho.
func factorRec(n uint64, out *[]uint64) {
	if n == 1 {
		return
	}
	if isPrime64(n) {
		*out = append(*out, n)
		return
	}
	d := pollardRho(n)
	factorRec(d, out)
	factorRec(n/d, out)
}

func groupFactors(factors []uint64) []PrimePower {
	groups := make([]PrimePower, 0, len(factors))
	for _, p := range factors {
		if len(groups) > 0 && groups[len(groups)-1].Prime == p {
			groups[len(groups)-1].Exp++
			continue
		}
		groups = append(groups, PrimePower{Prime: p, Exp: 1})
	}
	return groups
}

// --- Miller-Rabin determinista para 64 bits ---

var mrBases = [...]uint64{2, 325, 9375, 28178, 450775, 9780504, 1795265022}

func isPrime64(n uint64) bool {
	if n < 2 {
		return false
	}
	for _, p := range [...]uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37} {
		if n%p == 0 {
			return n == p
		}
	}

	d := n - 1
	s := 0
	for d&1 == 0 {
		d >>= 1
		s++
	}
	for _, a := range mrBases {
		if a%n == 0 {
			continue
		}
		x := powMod(a%n, d, n)
		if x == 1 || x == n-1 {
			continue
		}
		composite := true
		for r := 1; r < s; r++ {
			x = mulMod(x, x, n)
			if x == n-1 {
				composite = false
				break
			}
		}
		if composite {
			return false
		}
	}
	return true
}

// mulMod calcula (a*b) % m sin desbordar usando aritmética de 128 bits.
func mulMod(a, b, m uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	if hi == 0 {
		return lo % m
	}
	// Div64 exige hi < m; reducirlo antes es válido porque
	// (hi·2^64 + lo) mod m = ((hi mod m)·2^64 + lo) mod m.
	_, rem := bits.Div64(hi%m, lo, m)
	return rem
}

func powMod(base, exp, mod uint64) uint64 {
	res := uint64(1)
	base %= mod
	for exp > 0 {
		if exp&1 == 1 {
			res = mulMod(res, base, mod)
		}
		base = mulMod(base, base, mod)
		exp >>= 1
	}
	return res
}

func gcd64(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// addMod calcula (a+b) % m sin desbordar.
func addMod(a, b, m uint64) uint64 {
	a %= m
	b %= m
	if a >= m-b {
		return a - (m - b)
	}
	return a + b
}

// pollardRho encuentra un divisor no trivial de un compuesto n.
func pollardRho(n uint64) uint64 {
	if n%2 == 0 {
		return 2
	}
	if n%3 == 0 {
		return 3
	}
	for c := uint64(1); ; c++ {
		x, y := uint64(2), uint64(2)
		d := uint64(1)
		for d == 1 {
			x = addMod(mulMod(x, x, n), c, n)
			y = addMod(mulMod(y, y, n), c, n)
			y = addMod(mulMod(y, y, n), c, n)
			diff := x
			if y > diff {
				diff = y - diff
			} else {
				diff -= y
			}
			d = gcd64(diff, n)
		}
		if d != n {
			return d
		}
	}
}

// --- presentación ---

// FormatFactorization devuelve la forma agrupada, por ejemplo "2³ · 3 · 5".
func FormatFactorization(f Factorization) string {
	if f.Original == 0 {
		return "0"
	}
	if len(f.Groups) == 0 {
		if f.Negative {
			return "−1"
		}
		return "1"
	}
	parts := make([]string, 0, len(f.Groups)+1)
	if f.Negative {
		parts = append(parts, "−1")
	}
	for _, g := range f.Groups {
		part := strconv.FormatUint(g.Prime, 10)
		if g.Exp > 1 {
			part += render.ToSuperscript(strconv.Itoa(g.Exp))
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, " · ")
}

// FormatExpanded devuelve la forma con repetición, por ejemplo
// "2 · 2 · 2 · 3 · 5".
func FormatExpanded(f Factorization) string {
	if f.Original == 0 {
		return "0"
	}
	if len(f.Factors) == 0 {
		if f.Negative {
			return "−1"
		}
		return "1"
	}
	parts := make([]string, 0, len(f.Factors)+1)
	if f.Negative {
		parts = append(parts, "−1")
	}
	for _, p := range f.Factors {
		parts = append(parts, strconv.FormatUint(p, 10))
	}
	return strings.Join(parts, " · ")
}

// evalFactorizar produce los pasos de la descomposición para el panel de
// salida.
func evalFactorizar(decl Decl) ([]string, error) {
	if decl.Tipo != DeclExpresion {
		return nil, errors.New("factorizar opera sobre una expresión entera: usa /expresion 120")
	}
	n, ok := decl.IntegerValue()
	if !ok {
		return nil, fmt.Errorf("factorizar requiere un entero literal (sin decimales ni variables): %q", decl.Raw)
	}

	f := Factorize(n)
	steps := []string{"Operación: descomposición en factores primos"}
	switch n {
	case 0:
		return append(steps, "0 no tiene factores primos"), nil
	case 1:
		return append(steps, "1 no tiene factores primos"), nil
	case -1:
		return append(steps, "−1 no tiene factores primos"), nil
	}

	if f.Negative {
		steps = append(steps, "Signo: −1 (se factoriza el valor absoluto)")
	}
	mag := absInt64(n)
	steps = append(steps, "N = "+strconv.FormatUint(mag, 10))

	if len(f.Groups) == 1 && f.Groups[0].Exp == 1 && f.Groups[0].Prime == mag {
		steps = append(steps, fmt.Sprintf("%d es primo", mag))
	} else {
		cur := mag
		for _, g := range f.Groups {
			values := []string{strconv.FormatUint(cur, 10)}
			for range g.Exp {
				cur /= g.Prime
				values = append(values, strconv.FormatUint(cur, 10))
			}
			veces := "vez"
			if g.Exp != 1 {
				veces = "veces"
			}
			steps = append(steps, fmt.Sprintf("÷%d (%d %s): %s",
				g.Prime, g.Exp, veces, strings.Join(values, " → ")))
		}
	}

	steps = append(steps, "Factorización: "+FormatFactorization(f))
	if len(f.Factors) > 1 || f.Negative {
		steps = append(steps, "Forma expandida: "+FormatExpanded(f))
	}
	return steps, nil
}
