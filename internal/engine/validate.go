package engine

import (
	"errors"
	"fmt"
	"strings"
)

// Validate hace una comprobación estática ligera de la expresión.
// No es un parser: solo descarta errores obvios de sintaxis. El parser
// real (Pratt) reemplazará esta función en la fase siguiente.
func Validate(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return errors.New("expresión vacía")
	}

	depth := 0
	prev := byte(0) // último carácter significativo (ignora espacios y comas)
	for i := 0; i < len(expr); i++ {
		c := expr[i]

		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return errors.New("paréntesis sobrante: ')' sin apertura")
			}
		case '+', '-', '*', '/', '^':
			if isOp(prev) {
				if c != '-' {
					return fmt.Errorf("operadores consecutivos: %q %q", prev, c)
				}
				if prev == '^' {
					// x^-2 no se soporta aún
					return errors.New("exponente negativo: usa x^(-2)")
				}
				// "-" tras un operador se permite (número negativo)
			}
		case ' ', ',':
			continue // no actualiza prev
		default:
			if !(isLetter(c) || isDigit(c) || c == '.') {
				return fmt.Errorf("carácter no permitido: %q", c)
			}
		}
		prev = c
	}

	if depth != 0 {
		return errors.New("paréntesis sin cerrar")
	}
	if isOp(prev) {
		return fmt.Errorf("la expresión no puede terminar con %q", prev)
	}

	// primer carácter: solo se permite "-" como signo
	if first := expr[0]; isOp(first) && first != '-' {
		return fmt.Errorf("la expresión no puede empezar con %q", first)
	}

	return nil
}

func isOp(c byte) bool {
	switch c {
	case '+', '-', '*', '/', '^':
		return true
	}
	return false
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
