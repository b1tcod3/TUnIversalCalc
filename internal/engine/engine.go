package engine

import (
	"errors"
	"fmt"
	"strings"

	"tuniversalcalc/internal/render"
)

// Evaluate es la interfaz asíncrona del motor: recibe el objeto declarado
// (sobre el que ya resolvió la UI), el comando del panel 2 y la variable,
// y devuelve los pasos del cálculo para el panel de salida.
//
// Por ahora es un stub que simula los pasos, pero ya es "inteligente":
// devuelve errores claros para operaciones no reconocidas y contextualiza
// cada paso con el nombre y el LaTeX del objeto. El parser real (Pratt)
// llegará en la siguiente fase con la misma firma.
func Evaluate(decl Decl, cmd, varName string) ([]string, error) {
	if decl.Nombre == "" {
		return nil, errors.New("objeto vacío: declara uno en el panel 1 (/ecuacion 2x+2=0)")
	}
	if varName == "" {
		varName = "x"
	}

	op := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(cmd, "/")))
	op = strings.NewReplacer("(", "", ")", "", ",", "").Replace(op)
	op = strings.TrimSpace(op)
	if op == "" {
		return nil, errors.New("comando vacío en el panel 2")
	}

	pretty, latex := decl.Pretty, decl.Latex
	steps := []string{
		"Objeto: " + decl.Nombre + "  —  " + latex,
	}

	switch {
	case strings.HasPrefix(op, "integral"):
		steps = append(steps,
			"Operación: integral",
			"Se busca una antiderivada F tal que F′ = "+decl.Nombre,
			"Aplicando la regla del poder y la linealidad...",
			render.IntegralSymbol+" ("+pretty+") d"+varName+"  →  F("+varName+") + C",
		)

	case strings.HasPrefix(op, "deriv"):
		steps = append(steps,
			"Operación: derivada",
			"Aplicando la regla del poder y de la cadena...",
			"d/d"+varName+" ("+pretty+")  →  "+decl.Nombre+"′("+varName+")",
		)

	case strings.HasPrefix(op, "limit"):
		steps = append(steps,
			"Operación: límite",
			"Sustituyendo "+varName+" → valor indicado...",
			"lim ("+pretty+")  →  L",
		)

	case strings.HasPrefix(op, "resolver"):
		steps = append(steps,
			"Operación: resolver ecuación",
			"Despejando "+varName+" en "+decl.Nombre+"...",
			"Resultado: "+pretty+"  →  "+varName+" = x₀ (simulado)",
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
		name := op
		if f := strings.Fields(op); len(f) > 0 {
			name = f[0]
		}
		return nil, fmt.Errorf(
			"operación no reconocida: %q — usa /resolver, /integral, /derivar, /limite, /simplificar o /matriz",
			name,
		)
	}

	steps = append(steps, "Motor: stub simbólico — parser real en la próxima fase")
	return steps, nil
}