package engine

import (
	"strings"
	"testing"
)

func TestCompletarCuadrado(t *testing.T) {
	cases := []struct{ in, want string }{
		{"x^2+x+1", "(x + 1/2)^2 + 3/4"},
		{"2x^2-12x+11", "2*(x - 3)^2 - 7"},
		{"x^2-6x+9", "(x - 3)^2"},
		{"-x^2+2x+3", "-(x - 1)^2 + 4"},
		{"x^2+1", "x^2 + 1"},
		{"3t^2+6t", "3*(t + 1)^2 - 3"},
	}
	for _, c := range cases {
		res, err := CompletarCuadrado(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("CompletarCuadrado(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("CompletarCuadrado(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestCompletarCuadradoUnicode(t *testing.T) {
	res, err := CompletarCuadrado(exactParse(t, "2x^2-12x+11"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Unicode != "2·(x − 3)² − 7" {
		t.Fatalf("Unicode = %q", res.Unicode)
	}
}

func TestCompletarCuadradoPasos(t *testing.T) {
	res, err := CompletarCuadrado(exactParse(t, "2x^2-12x+11"))
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(res.Steps, "|")
	for _, frag := range []string{
		"completar el cuadrado",
		"Coeficientes: a = 2, b = -12, c = 11",
		"Forma mónica: x^2 - 6*x + 11/2",
		"Mitad del coeficiente lineal: b/(2a) = -3",
		"Cuadrado que se suma y resta: 9",
		"Ajuste del término independiente: c - a*(b/(2a))^2 = -7",
		"Forma completada: 2*(x - 3)^2 - 7",
		"Vértice (h, k): (3, -7)",
		"Unicode: 2·(x − 3)² − 7",
	} {
		if !strings.Contains(all, frag) {
			t.Errorf("falta %q en los pasos: %v", frag, res.Steps)
		}
	}
}

func TestEvaluateFactorizarCuadrado(t *testing.T) {
	d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "2x^2-12x+11"}
	steps, err := Evaluate(d, "/factorizar -cuadrado", "x")
	if err != nil {
		t.Fatalf("Evaluate /factorizar -cuadrado: %v", err)
	}
	all := strings.Join(steps, "|")
	if !strings.Contains(all, "Resultado: 2*(x - 3)^2 - 7") {
		t.Fatalf("resultado ausente: %v", steps)
	}

	steps, err = Evaluate(d, "/factorizar --cuadrado p1", "x")
	if err != nil {
		t.Fatalf("Evaluate /factorizar --cuadrado: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "Vértice (h, k): (3, -7)") {
		t.Fatalf("vértice ausente: %v", steps)
	}
}

func TestCompletarCuadradoErrores(t *testing.T) {
	cases := []struct {
		in       string
		fragment string
	}{
		{"x+1", "grado 2"},
		{"x^3+1", "grado 2"},
		{"x^2+y^2", "una variable"},
		{"sqrt(2)x^2+1", "racionales"},
		{"5", "variable"},
	}
	for _, c := range cases {
		_, err := CompletarCuadrado(exactParse(t, c.in))
		if err == nil || !strings.Contains(err.Error(), c.fragment) {
			t.Errorf("CompletarCuadrado(%q) error = %v, se esperaba %q", c.in, err, c.fragment)
		}
	}
}

func TestFactorizarBanderaDesconocida(t *testing.T) {
	d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "x^2+x+1"}
	if _, err := Evaluate(d, "/factorizar -otra", "x"); err == nil ||
		!strings.Contains(err.Error(), "bandera no reconocida") {
		t.Fatalf("bandera desconocida debe fallar: %v", err)
	}
}
