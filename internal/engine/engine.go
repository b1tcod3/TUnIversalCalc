package engine

import (
	"errors"
	"strings"

	"tuniversalcalc/internal/render"
)

// Evaluate es la interfaz asíncrona del motor: recibe una expresión y un
// comando (ej. "/integral") y devuelve los pasos del cálculo en formato
// Unicode para mostrar en el panel de salida.
//
// Por ahora es un stub que simula los pasos; el parser real (Pratt) y el
// cálculo simbólico llegarán en la siguiente fase con la misma firma.
func Evaluate(expr, cmd string) ([]string, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, errors.New("expresión vacía: escribe algo en el panel 1")
	}

	op := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(cmd, "/")))
	pretty := render.Pretty(expr)

	steps := []string{
		"Expresión: " + pretty,
	}

	switch {
	case strings.HasPrefix(op, "integral"):
		steps = append(steps,
			"Operación: integral",
			"Se busca una antiderivada F tal que F′ = f",
			"Aplicando la regla del poder y linealidad...",
			"Resultado: "+render.IntegralSymbol+" ("+pretty+") dx  →  F(x) + C",
		)

	case strings.HasPrefix(op, "deriv"):
		steps = append(steps,
			"Operación: derivada",
			"Aplicando la regla de la cadena y del poder...",
			"Resultado: d/dx ("+pretty+")  →  f′(x)",
		)

	case strings.HasPrefix(op, "limit"):
		steps = append(steps,
			"Operación: límite",
			"Sustituyendo el valor de la variable...",
			"Resultado: lim ("+pretty+")  →  L",
		)

	case strings.HasPrefix(op, "resolver"):
		steps = append(steps,
			"Operación: resolver ecuación",
			"Despejando la variable...",
			"Resultado: solución de "+pretty+"  →  x = x₀",
		)

	case strings.HasPrefix(op, "simplificar"):
		steps = append(steps,
			"Operación: simplificar",
			"Agrupando términos semejantes...",
			"Resultado: "+pretty+"  →  forma simplificada",
		)

	case strings.HasPrefix(op, "matriz"):
		steps = append(steps,
			"Operación: matriz",
			"Reduciendo filas (Gauss-Jordan)...",
			"Resultado: forma escalonada",
		)

	default:
		steps = append(steps,
			"Evaluación numérica (motor en desarrollo)",
			"Resultado: "+pretty+"  ≈  42.0 (simulado)",
		)
	}

	steps = append(steps, "Motor: stub simbólico — parser real en la próxima fase")
	return steps, nil
}
