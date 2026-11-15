package engine

import (
	"errors"
	"fmt"
	"strings"
)

// Evaluate es la interfaz asíncrona del motor: recibe el objeto declarado
// (sobre el que ya resolvió la UI) y el comando del panel 2, y devuelve los
// pasos del cálculo para el panel de salida. El tercer argumento se conserva
// por compatibilidad con la UI, que hoy lo resuelve pero no lo necesita.
//
// El motor combina un parser simbólico (para funciones/expresiones) con
// operaciones de cálculo numérico (interpolación polinómica). Los métodos
// de interpolación trabajan sobre puntos declarados con /puntos.
func Evaluate(decl Decl, cmd, _ string) ([]string, error) {
	if decl.Nombre == "" {
		return nil, errors.New("objeto vacío: declara uno en el panel 1 (/puntos (1,2) (3,4))")
	}

	op := normalizeOp(cmd)
	if op == "" {
		return nil, errors.New("comando vacío en el panel 2")
	}

	steps := []string{
		"Objeto: " + decl.Nombre + "  —  " + decl.Latex,
	}
	tail, err := evalOp(decl, op)
	if err != nil {
		return nil, err
	}
	return append(steps, tail...), nil
}

// normalizeOp limpia el comando del panel 2: minúsculas, sin "/" inicial
// ni paréntesis o comas.
func normalizeOp(cmd string) string {
	op := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(cmd, "/")))
	op = strings.NewReplacer("(", "", ")", "", ",", "").Replace(op)
	return strings.TrimSpace(op)
}

// evalOp despacha la operación reconocida y devuelve los pasos que siguen
// a la cabecera común del objeto. Las operaciones de interpolación bind
// sobre el objeto declarado.
func evalOp(decl Decl, op string) ([]string, error) {
	word := op
	if f := strings.Fields(op); len(f) > 0 {
		word = f[0]
	}

	switch {
	case strings.HasPrefix(word, "lagrange"):
		return evalLagrange(decl)
	case strings.HasPrefix(word, "newton"):
		return evalNewton(decl)
	case strings.HasPrefix(word, "factor"):
		return evalFactorizar(decl)
	default:
		return nil, fmt.Errorf(
			"operación no reconocida: %q — usa /lagrange, /newton o /factorizar",
			word,
		)
	}
}

// evalLagrange exige un objeto de puntos y delega en el algoritmo.
func evalLagrange(decl Decl) ([]string, error) {
	if decl.Tipo != DeclPuntos {
		return nil, errors.New("lagrange opera sobre puntos: declara /puntos (x,y) (x,y)")
	}
	return Lagrange(decl.Points)
}

// evalNewton exige un objeto de puntos y delega en el algoritmo.
func evalNewton(decl Decl) ([]string, error) {
	if decl.Tipo != DeclPuntos {
		return nil, errors.New("newton opera sobre puntos: declara /puntos (x,y) (x,y)")
	}
	return Newton(decl.Points)
}
