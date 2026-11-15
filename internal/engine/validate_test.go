package engine

import "testing"

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"expresion simple", "2x+2", true},
		{"polinomio", "x^2 + 3x - 5", true},
		{"funciones", "sin(x) + cos(x)", true},
		{"signo y potencia", "3*-2", true},
		{"negativo", "-2x^2", true},
		{"exponente negativo", "x^-2", true},
		{"exponente negativo con espacio", "x^ -2", true},
		{"varias variables", "2x + y/c", true},
		{"decimal", "3.5x", true},
		{"vacio", "", false},
		{"potencia incompleta", "x^2+", false},
		{"signo inicial", "+x", false},
		{"parentesis sin cerrar", "(x^2 + 1", false},
		{"parentesis sobrante", "x^2) + 1", false},
		{"doble operador", "3**2", false},
		{"caracter invalido", "a@b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.in)
			if (err == nil) != tt.want {
				t.Fatalf("Validate(%q) = %v, se esperaba válida=%v", tt.in, err, tt.want)
			}
		})
	}
}
