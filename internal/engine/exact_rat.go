package engine

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Helpers de aritmética exacta sobre *big.Rat. Todas las operaciones
// devuelven un valor nuevo: los *big.Rat son mutables y no deben compartirse
// entre monomios.

func ratZero() *big.Rat { return new(big.Rat) }

func ratOne() *big.Rat { return new(big.Rat).SetInt64(1) }

func ratFromInt64(n int64) *big.Rat { return new(big.Rat).SetInt64(n) }

func ratFromUint64(n uint64) *big.Rat { return new(big.Rat).SetUint64(n) }

func ratFromInt(n *big.Int) *big.Rat { return new(big.Rat).SetInt(n) }

func ratCopy(r *big.Rat) *big.Rat { return new(big.Rat).Set(r) }

func ratAdd(a, b *big.Rat) *big.Rat { return new(big.Rat).Add(a, b) }

func ratSub(a, b *big.Rat) *big.Rat { return new(big.Rat).Sub(a, b) }

func ratMul(a, b *big.Rat) *big.Rat { return new(big.Rat).Mul(a, b) }

func ratNeg(a *big.Rat) *big.Rat { return new(big.Rat).Neg(a) }

func ratAbs(a *big.Rat) *big.Rat { return new(big.Rat).Abs(a) }

func ratQuo(a, b *big.Rat) (*big.Rat, error) {
	if b.Sign() == 0 {
		return nil, errors.New("división por cero")
	}
	return new(big.Rat).Quo(a, b), nil
}

func ratCmp(a, b *big.Rat) int { return a.Cmp(b) }

func ratString(r *big.Rat) string {
	if r == nil {
		return "0"
	}
	return r.RatString()
}

// ratParse convierte un literal decimal o una fracción "a/b" en exacto.
func ratParse(s string) (*big.Rat, error) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok {
		return nil, fmt.Errorf("número no exacto: %q", s)
	}
	return r, nil
}

// ratFloor devuelve el mayor entero <= r. big.Int.Div es división euclídea
// y el denominador de un big.Rat siempre es positivo.
func ratFloor(r *big.Rat) *big.Int {
	return new(big.Int).Div(r.Num(), r.Denom())
}

// ratFrac devuelve r - floor(r), siempre en [0, 1).
func ratFrac(r *big.Rat) *big.Rat {
	q := ratFromInt(ratFloor(r))
	return ratSub(r, q)
}

func ratInt64(r *big.Rat) (int64, bool) {
	if !r.IsInt() || !r.Num().IsInt64() {
		return 0, false
	}
	return r.Num().Int64(), true
}

// ratPowInt eleva r a un exponente entero, incluyendo exponentes negativos.
func ratPowInt(r *big.Rat, n int64) (*big.Rat, error) {
	if n == 0 {
		return ratOne(), nil
	}
	if r.Sign() == 0 {
		if n < 0 {
			return nil, errors.New("0 elevado a un exponente negativo no está definido")
		}
		return ratZero(), nil
	}
	neg := n < 0
	if neg {
		if n == -1<<63 {
			return nil, errors.New("exponente demasiado grande")
		}
		n = -n
	}
	num := new(big.Int).Exp(r.Num(), big.NewInt(n), nil)
	den := new(big.Int).Exp(r.Denom(), big.NewInt(n), nil)
	z := new(big.Rat).SetFrac(num, den)
	if neg {
		z.Inv(z)
	}
	return z, nil
}
