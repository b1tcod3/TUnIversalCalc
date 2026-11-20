package render

import "strings"

// replaceIdent sustituye una palabra completa, no una subcadena. Así
// "pi" se convierte en π pero "spider" o "sqrtx" se conservan.
func replaceIdent(s, word, repl string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		c := s[i]
		if c >= '0' && c <= '9' || c == '.' {
			b.WriteByte(c)
			i++
			continue
		}
		if !isIdentStart(c) {
			b.WriteByte(c)
			i++
			continue
		}
		j := i + 1
		for j < len(s) && isIdentPart(s[j]) {
			j++
		}
		ident := s[i:j]
		if ident == word {
			b.WriteString(repl)
		} else {
			b.WriteString(ident)
		}
		i = j
	}
	return b.String()
}

func isIdentStart(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || c >= '0' && c <= '9'
}
