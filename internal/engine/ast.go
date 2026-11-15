package engine

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// Expr es un nodo del árbol de sintaxis abstracta del parser simbólico.
// El '−' binario se representa como Add con un operando Neg; no existe
// un nodo Sub para simplificar las reglas simbólicas.
type Expr interface {
	// Ascii devuelve la forma canónica ASCII, lista para render (LaTeX/Unicode).
	Ascii() string
	exprNode()
}

// Num es una constante numérica (texto original o construida por el motor).
type Num struct{ Text string }

// Var es una variable de una sola letra.
type Var struct{ Name byte }

// Const es una constante simbólica: pi o inf.
type Const struct{ Name string } // "pi" | "inf"

// Neg es la negación unitaria: -x / -(x+1).
type Neg struct{ X Expr }

// Add es la suma (con operandos Neg para restas).
type Add struct{ Terms []Expr }

// Mul es el producto (los números se agrupan en un solo factor Num).
type Mul struct{ Factors []Expr }

// Div es la división a/b.
type Div struct{ Num, Den Expr }

// Pow es la potencia base^exp.
type Pow struct{ Base, Exp Expr }

// Call es una llamada a función: sqrt(x), sin(x), cos(x), abs(x).
type Call struct {
	Name string
	Arg  Expr
}

func (Num) exprNode()   {}
func (Var) exprNode()   {}
func (Const) exprNode() {}
func (Neg) exprNode()   {}
func (Add) exprNode()   {}
func (Mul) exprNode()   {}
func (Div) exprNode()   {}
func (Pow) exprNode()   {}
func (Call) exprNode()  {}

// FormatFloatText da el texto decimal simple de un float (para exponentes).
func FormatFloatText(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// NumFromFloat construye un Num a partir de un float64 con formato bonito
// (fracciones simples: 0.5 → 1/2, 1.5 → 3/2). Los exponentes NO deben usar
// esto: una fracción en el exponente cambiaría el significado al re-parsear
// (x^3/2 ≠ x^(3/2)); usa FormatFloatText para ellos.
func NumFromFloat(f float64) Num {
	return Num{Text: formatNum(f)}
}

// Float convierte el texto del número; falla si no es un número válido.
// Soporta tanto decimales ("3.5") como fracciones construidas por el motor
// ("3/2", "-1/3") — los números escritos por el usuario nunca llevan "/".
func (n Num) Float() (float64, error) {
	if strings.Contains(n.Text, "/") {
		parts := strings.Split(n.Text, "/")
		if len(parts) == 2 {
			a, errA := strconv.ParseFloat(parts[0], 64)
			b, errB := strconv.ParseFloat(parts[1], 64)
			if errA == nil && errB == nil && b != 0 {
				return a / b, nil
			}
		}
		return 0, errors.New("número mal formado: " + n.Text)
	}
	return strconv.ParseFloat(n.Text, 64)
}

// IsZero/Negative ayudan a limpiar operaciones constantes.
func (n Num) isZero() bool {
	f, err := n.Float()
	return err == nil && f == 0
}

func (n Num) isNeg() bool {
	return strings.HasPrefix(n.Text, "-")
}

// --- Ascii canónico ---

func (e Num) Ascii() string { return e.Text }

func (e Var) Ascii() string { return string(e.Name) }

func (e Const) Ascii() string { return e.Name }

func (e Neg) Ascii() string {
	s := e.X.Ascii()
	if _, ok := e.X.(Add); ok {
		return "-(" + s + ")"
	}
	return "-" + s
}

func (e Add) Ascii() string {
	var b strings.Builder
	for i, t := range e.Terms {
		s := t.Ascii()
		// un término negativo se imprime como resta
		if strings.HasPrefix(s, "-") {
			if i > 0 {
				b.WriteString(" - ")
			} else {
				b.WriteString("-")
			}
			b.WriteString(strings.TrimPrefix(s, "-"))
			continue
		}
		if i > 0 {
			b.WriteString(" + ")
		}
		b.WriteString(s)
	}
	return b.String()
}

func (e Mul) Ascii() string {
	var b strings.Builder
	for i, f := range e.Factors {
		if i > 0 {
			b.WriteString("*")
		}
		b.WriteString(mulFactorAscii(f))
	}
	return b.String()
}

// mulFactorAscii pone paréntesis si el factor lo exige en un producto.
func mulFactorAscii(f Expr) string {
	switch f.(type) {
	case Add, Div:
		return "(" + f.Ascii() + ")"
	}
	return f.Ascii()
}

func (e Div) Ascii() string {
	num, den := e.Num.Ascii(), e.Den.Ascii()
	if _, ok := e.Num.(Add); ok {
		num = "(" + num + ")"
	}
	switch e.Den.(type) {
	case Add, Mul, Neg, Div:
		den = "(" + den + ")"
	}
	return num + "/" + den
}

func (e Pow) Ascii() string {
	base := powBaseAscii(e.Base)
	exp := powExpAscii(e.Exp)
	return base + "^" + exp
}

func powBaseAscii(b Expr) string {
	switch b.(type) {
	case Add, Mul, Div, Neg:
		return "(" + b.Ascii() + ")"
	}
	return b.Ascii()
}

func powExpAscii(e Expr) string {
	switch e.(type) {
	case Add, Mul, Div:
		return "(" + e.Ascii() + ")"
	case Num:
		if strings.Contains(e.Ascii(), "/") {
			return "(" + e.Ascii() + ")"
		}
	}
	return e.Ascii()
}

func (e Call) Ascii() string {
	return e.Name + "(" + e.Arg.Ascii() + ")"
}

// --- utilidades compartidas ---

// isVarIn informa si la variable v aparece en el árbol.
func isVarIn(e Expr, v byte) bool {
	switch n := e.(type) {
	case Var:
		return n.Name == v
	case Num, Const:
		return false
	case Neg:
		return isVarIn(n.X, v)
	case Add:
		for _, t := range n.Terms {
			if isVarIn(t, v) {
				return true
			}
		}
	case Mul:
		for _, f := range n.Factors {
			if isVarIn(f, v) {
				return true
			}
		}
	case Div:
		return isVarIn(n.Num, v) || isVarIn(n.Den, v)
	case Pow:
		return isVarIn(n.Base, v) || isVarIn(n.Exp, v)
	case Call:
		return isVarIn(n.Arg, v)
	}
	return false
}

// varList devuelve las variables presentes, en orden de aparición.
func varList(e Expr) []byte {
	seen := map[byte]bool{}
	var out []byte
	var walk func(x Expr)
	walk = func(x Expr) {
		switch n := x.(type) {
		case Var:
			if !seen[n.Name] {
				seen[n.Name] = true
				out = append(out, n.Name)
			}
		case Neg:
			walk(n.X)
		case Add:
			for _, t := range n.Terms {
				walk(t)
			}
		case Mul:
			for _, f := range n.Factors {
				walk(f)
			}
		case Div:
			walk(n.Num)
			walk(n.Den)
		case Pow:
			walk(n.Base)
			walk(n.Exp)
		case Call:
			walk(n.Arg)
		}
	}
	walk(e)
	return out
}

// matchNum devuelve el float de un nodo Num o false.
func matchNum(e Expr) (float64, bool) {
	n, ok := e.(Num)
	if !ok {
		return 0, false
	}
	f, err := n.Float()
	return f, err == nil
}

// powFloat calcula base^exp evitando NaN para 0^negativo.
func powFloat(base, exp float64) (float64, bool) {
	if base == 0 && exp < 0 {
		return 0, false // indefinido
	}
	return math.Pow(base, exp), true
}

// formatNum da una representación compacta de un número para mensajes.
func formatNum(f float64) string {
	if math.IsInf(f, 1) {
		return "∞"
	}
	if math.IsInf(f, -1) {
		return "-∞"
	}
	if math.IsNaN(f) {
		return "indefinido"
	}
	// Fracciones simples y bonitas (medio, tercios, cuartos...)
	neg := f < 0
	if neg {
		f = -f
	}
	for d := 2.0; d <= 12; d++ {
		r := math.Round(f * d)
		if math.Abs(f*d-r) < 1e-9 {
			n := r
			g := gcd(int64(n), int64(d))
			n /= float64(g)
			dd := d / float64(g)
			s := strconv.FormatInt(int64(n), 10)
			if dd > 1 {
				s += "/" + strconv.FormatInt(int64(dd), 10)
			}
			if neg {
				return "-" + s
			}
			return s
		}
	}
	s := strconv.FormatFloat(f, 'g', 6, 64)
	if neg {
		return "-" + s
	}
	return s
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
