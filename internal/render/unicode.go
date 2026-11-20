package render

import "strings"

const (
	IntegralSymbol = "∫"
	SqrtSymbol     = "√"
	PiSymbol       = "π"
	InfinitySymbol = "∞"
	DotSymbol      = "·"
	MinusSymbol    = "−"
)

var superscripts = map[rune]string{
	'0': "⁰", '1': "¹", '2': "²", '3': "³", '4': "⁴",
	'5': "⁵", '6': "⁶", '7': "⁷", '8': "⁸", '9': "⁹",
	'+': "⁺", '-': "⁻", '=': "⁼", '(': "⁽", ')': "⁾",
	'a': "ᵃ", 'b': "ᵇ", 'c': "ᶜ", 'd': "ᵈ", 'e': "ᵉ",
	'g': "ᵍ", 'h': "ʰ", 'i': "ⁱ", 'j': "ʲ", 'k': "ᵏ",
	'l': "ˡ", 'm': "ᵐ", 'n': "ⁿ", 'o': "ᵒ", 'p': "ᵖ",
	'r': "ʳ", 's': "ˢ", 't': "ᵗ", 'u': "ᵘ", 'v': "ᵛ",
	'w': "ʷ", 'x': "ˣ", 'y': "ʸ", 'z': "ᶻ",
}

var subscripts = map[rune]string{
	'0': "₀", '1': "₁", '2': "₂", '3': "₃", '4': "₄",
	'5': "₅", '6': "₆", '7': "₇", '8': "₈", '9': "₉",
	'+': "₊", '-': "₋", '=': "₌", '(': "₍", ')': "₎",
	'a': "ₐ", 'e': "ₑ", 'o': "ₒ", 'x': "ₓ",
}

// ToSuperscript convierte una cadena a su forma superíndice Unicode.
// Los caracteres sin mapeo se conservan tal cual.
func ToSuperscript(s string) string {
	var b strings.Builder
	for _, r := range s {
		if sup, ok := superscripts[r]; ok {
			b.WriteString(sup)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ToSubscript convierte una cadena a su forma subíndice Unicode.
func ToSubscript(s string) string {
	var b strings.Builder
	for _, r := range s {
		if sub, ok := subscripts[r]; ok {
			b.WriteString(sub)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Pretty convierte una expresión ASCII a su presentación Unicode:
// x^2 -> x², x^(2y) -> x²ʸ, sqrt -> √, pi -> π, * -> ·, - -> −
func Pretty(s string) string {
	s = prettyRaiz(s)
	s = replaceIdent(s, "sqrt", SqrtSymbol)
	s = replaceIdent(s, "pi", PiSymbol)
	s = replaceIdent(s, "inf", InfinitySymbol)
	s = expandPowers(s)
	s = strings.ReplaceAll(s, "*", DotSymbol)
	s = strings.ReplaceAll(s, "-", MinusSymbol)
	return s
}

// prettyRaiz convierte raiz(n, expr) en su símbolo Unicode: √, ∛, ∜ o
// ⁿ√. El radicando se procesa recursivamente para admitir anidamientos.
func prettyRaiz(s string) string {
	if !strings.Contains(s, "raiz(") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if !strings.HasPrefix(s[i:], "raiz(") {
			b.WriteByte(s[i])
			i++
			continue
		}
		inner, next := groupLatex(s, i+4)
		idx, rad := splitRootInner(inner)
		idx = strings.TrimSpace(idx)
		rad = Pretty(strings.TrimSpace(rad))
		sym := rootSymbol(idx)
		if needsRootParens(rad) {
			b.WriteString(sym + "(" + rad + ")")
		} else {
			b.WriteString(sym + rad)
		}
		i = next
	}
	return b.String()
}

// rootSymbol elige el símbolo de raíz para un índice.
func rootSymbol(idx string) string {
	switch idx {
	case "2":
		return SqrtSymbol
	case "3":
		return "∛"
	case "4":
		return "∜"
	default:
		return ToSuperscript(idx) + SqrtSymbol
	}
}

// needsRootParens indica si el radicando necesita paréntesis para no
// cambiar el significado al quitar la llamada raiz(...).
func needsRootParens(s string) bool {
	return strings.ContainsAny(s, "+−·/ ()⁰¹²³⁴⁵⁶⁷⁸⁹")
}

// expandPowers reemplaza el operador ^ por superíndices Unicode.
func expandPowers(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '^' {
			b.WriteByte(s[i])
			i++
			continue
		}
		exp, next := nextExponent(s, i+1)
		b.WriteString(exp)
		i = next
	}
	return b.String()
}

// nextExponent extrae el exponente tras '^': un token simple o (grupo).
func nextExponent(s string, start int) (string, int) {
	if start >= len(s) {
		return "", start
	}
	if s[start] == '(' {
		depth := 1
		k := start + 1
		for k < len(s) && depth > 0 {
			switch s[k] {
			case '(':
				depth++
			case ')':
				depth--
			}
			k++
		}
		return ToSuperscript(s[start+1 : k-1]), k
	}
	k := start
	if s[k] == '+' || s[k] == '-' {
		k++ // signo del exponente: x^-2
	}
	for k < len(s) && isAlnum(s[k]) {
		k++
	}
	return ToSuperscript(s[start:k]), k
}

func isAlnum(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
