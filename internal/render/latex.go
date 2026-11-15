package render

import "strings"

// ToLatex convierte una expresión ASCII a su forma LaTeX emulada para
// terminal (la cadena se muestra como texto, no se compila):
//   - 2x^2+3x*2  →  2x^{2} + 3x \cdot 2
//   - sqrt(...)  →  \sqrt{...}  (resolve anidados)
//   - pi → \pi, inf → \infty
//   - +, -, =, / se rodean de espacios (el '-' unitario no)
func ToLatex(s string) string {
	s = strings.ReplaceAll(s, "pi", `\pi`)
	s = strings.ReplaceAll(s, "inf", `\infty`)
	s = expandPowersLatex(s)
	s = wrapSqrt(s)
	return spaceOperators(s)
}

// expandPowersLatex convierte '^' en ^{...}: x^2 -> x^{2}, x^(2y) -> x^{2y}.
func expandPowersLatex(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '^' {
			b.WriteByte(s[i])
			i++
			continue
		}
		exp, next := exponentLatex(s, i+1)
		b.WriteString("^{" + exp + "}")
		i = next
	}
	return b.String()
}

// exponentLatex extrae el exponente tras '^': token simple (signo seguido
// de alfanumérico) o grupo (...).
func exponentLatex(s string, start int) (string, int) {
	if start >= len(s) {
		return "", start
	}
	if s[start] == '(' {
		inner, next := groupLatex(s, start)
		return inner, next
	}
	k := start
	hasSign := k < len(s) && (s[k] == '+' || s[k] == '-')
	if hasSign {
		k++
	}
	for k < len(s) && isAlnum(s[k]) {
		k++
	}
	return s[start:k], k
}

// wrapSqrt reemplaza las llamadas sqrt(...) por \sqrt{...}, equilibrando
// paréntesis (incluye llamadas anidadas).
func wrapSqrt(s string) string {
	if !strings.Contains(s, "sqrt") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if !strings.HasPrefix(s[i:], "sqrt(") {
			b.WriteByte(s[i])
			i++
			continue
		}
		inner, next := groupLatex(s, i+4)
		b.WriteString(`\sqrt{` + wrapSqrt(inner) + `}`)
		i = next
	}
	return b.String()
}

// groupLatex devuelve el interior del grupo (...) que empieza en start
// (que debe ser '(') y el índice justo tras el cierre.
func groupLatex(s string, start int) (string, int) {
	if start >= len(s) || s[start] != '(' {
		return "", start
	}
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
	return s[start+1 : k-1], k
}

// spaceOperators rodea de espacios los operadores binarios y convierte '*'
// en '\cdot'. El '-' inicial o tras otro operador o un espacio (número
// negativo) no se espacia.
func spaceOperators(s string) string {
	var b strings.Builder
	var prev byte // último carácter significativo escrito
	for i := range len(s) {
		c := s[i]
		switch c {
		case '+', '=':
			spacedOp(&b, &prev, string(c))
		case '-':
			if wantSpace(prev) {
				spacedOp(&b, &prev, "-")
			} else {
				b.WriteByte(c)
				prev = c
			}
		case '*':
			spacedOp(&b, &prev, `\cdot`)
		case '/':
			spacedOp(&b, &prev, "/")
		case ' ':
			if b.Len() > 0 && prev != ' ' {
				b.WriteByte(' ')
				prev = ' '
			}
		default:
			b.WriteByte(c)
			prev = c
		}
	}
	return strings.TrimSpace(b.String())
}

// wantSpace indica si el operador que sigue a prev debe espaciarse:
// solo cuando prev es un operando real (no inicio, espacio ni operador).
func wantSpace(prev byte) bool {
	return prev != 0 && prev != ' ' &&
		prev != '+' && prev != '-' && prev != '*' && prev != '/' && prev != '='
}

// spacedOp escribe op rodeado de espacios (si no los hay ya).
func spacedOp(b *strings.Builder, prev *byte, op string) {
	if b.Len() > 0 && *prev != ' ' {
		b.WriteByte(' ')
	}
	b.WriteString(op)
	b.WriteByte(' ')
	*prev = ' '
}
