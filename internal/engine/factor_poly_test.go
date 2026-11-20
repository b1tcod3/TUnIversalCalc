package engine

import (
	"strings"
	"testing"
)

func TestFactorizarPolinomios(t *testing.T) {
	cases := []struct{ in, want string }{
		{"4x^2-25", "(2*x - 5)*(2*x + 5)"},
		{"2x^2+5x-12", "(2*x - 3)*(x + 4)"},
		{"x^3-3x^2-4x+12", "(x - 3)*(x - 2)*(x + 2)"},
		{"x^4+27x", "x*(x + 3)*(x^2 - 3*x + 9)"},
		{"3x^(3/2)-9x^(1/2)+6x^(-1/2)", "3*x^(-1/2)*(x - 2)*(x - 1)"},
		{"x^3y-4xy", "x*y*(x - 2)*(x + 2)"},
		{"x^2+2x+1", "(x + 1)^2"},
		{"x^2+5x+6", "(x + 2)*(x + 3)"},
		{"6x^2+5x-6", "(3*x - 2)*(2*x + 3)"},
		{"x^2-4", "(x - 2)*(x + 2)"},
		{"x^2+1", "x^2 + 1"},
		{"x^3-6x^2+11x-6", "(x - 1)*(x - 3)*(x - 2)"},
		{"x^2+2xy+y^2", "(x + y)^2"},
		{"x^2-2xy+y^2", "(x - y)^2"},
		{"4x^2+12x+9", "(2*x + 3)^2"},
		{"x^4-5x^2+4", "(x - 1)*(x + 1)*(x - 2)*(x + 2)"},
		{"x^2-5", "x^2 - 5"},
		{"x^4+4", "x^4 + 4"},
		{"x^3+y^3", "(x + y)*(x^2 - x*y + y^2)"},
		{"2x^3-x^2-x-3", "(2*x - 3)*(x^2 + x + 1)"},
		{"2x^3+x^2+x-1", "(2*x - 1)*(x^2 + x + 1)"},
	}
	for _, c := range cases {
		res, err := FactorizarPolinomio(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("FactorizarPolinomio(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("FactorizarPolinomio(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestFactorizarPasos(t *testing.T) {
	cases := []struct{ in, fragment string }{
		{"4x^2-25", "Diferencia de cuadrados"},
		{"2x^2+5x-12", "Trinomio: a=2, b=5, c=-12"},
		{"x^3-3x^2-4x+12", "Agrupación"},
		{"x^4+27x", "Factor común: x"},
		{"x^4+27x", "Suma de cubos"},
		{"3x^(3/2)-9x^(1/2)+6x^(-1/2)", "Factor común: 3*x^(-1/2)"},
		{"x^3y-4xy", "Factor común: x*y"},
	}
	for _, c := range cases {
		res, err := FactorizarPolinomio(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("FactorizarPolinomio(%q): %v", c.in, err)
		}
		if !strings.Contains(strings.Join(res.Steps, "|"), c.fragment) {
			t.Errorf("FactorizarPolinomio(%q) sin %q: %v", c.in, c.fragment, res.Steps)
		}
	}
}

func TestComandoFactorizarPolinomio(t *testing.T) {
	d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "x^2-4"}
	steps, err := Evaluate(d, "/factorizar", "x")
	if err != nil {
		t.Fatalf("Evaluate /factorizar: %v", err)
	}
	all := strings.Join(steps, "|")
	if !strings.Contains(all, "(x - 2)*(x + 2)") {
		t.Fatalf("factorización ausente: %v", steps)
	}
}

func TestFactorizarFuncionConstanteEntera(t *testing.T) {
	d := Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "120"}
	steps, err := Evaluate(d, "/factorizar", "x")
	if err != nil {
		t.Fatalf("Evaluate /factorizar: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "2³ · 3 · 5") {
		t.Fatalf("factorización entera ausente: %v", steps)
	}
}

func TestFactorizarConstanteNoEntera(t *testing.T) {
	d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "12.5"}
	if _, err := Evaluate(d, "/factorizar", "x"); err == nil {
		t.Fatal("12.5 debe fallar")
	}
}
