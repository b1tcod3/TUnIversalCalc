package engine

import "errors"

// Evaluate toma una cadena de texto (ej. "3+5") y devuelve el resultado numérico.
// Por ahora es un cascarón básico listo para recibir tu lógica matemática.
func Evaluate(expr string) (float64, error) {
	if expr == "" {
		return 0, errors.New("expresión vacía")
	}

	// TODO: Aquí irá el analizador (lexer/parser) para tus cálculos
	return 0, nil
}
