package engine

import (
	"strings"
	"testing"
)

func TestFactorizarRacionalEjemplos(t *testing.T) {
	cases := []struct{ in, want string }{
		{"(x^2+3x+2)/(x^2-x-2)", "(x + 2)/(x - 2)"},
		{"((2x^2-x-1)/(x^2-9))*((x+3)/(2x+1))", "(x - 1)/(x - 3)"},
		{"x^2/(x^2-4)-(x+1)/(x+2)", "1/(x - 2)"},
		{"(y/x-x/y)/(1/y-1/x)", "-x - y"},
	}
	for _, c := range cases {
		res, err := FactorizarRacional(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("FactorizarRacional(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("FactorizarRacional(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestFactorizarRacionalSignos(t *testing.T) {
	cases := []struct{ in, want string }{
		{"(x-1)/(1-x)", "-1"},
		{"(2-x)/(x-2)", "-1"},
		{"(y-x)/(x-y)", "-1"},
		{"(x^2-1)/(x-1)", "x + 1"},
		{"(x^2+2x+1)/(x+1)", "x + 1"},
		{"(2x+2)/(4x+4)", "1/2"},
		{"(x+1)/(x+2)", "(x + 1)/(x + 2)"},
		{"(x^3)/(x^3+x^2)", "x/(x + 1)"},
		{"(x^2-1)/(x^2+2x+1)", "(x - 1)/(x + 1)"},
		{"(x+1)/(x+1)", "1"},
		{"0/(x+1)", "0"},
		{"x/x-1", "0"},
		{"1/x-1/x", "0"},
		{"(x+1)^-1", "1/(x + 1)"},
		{"1/(x+1)+1/(x+1)", "2/(x + 1)"},
		{"1/(x-1)-1/(x+1)", "2/(x^2 - 1)"},
		{"((x+1)/(x-1))/((x+1)/(x-1))", "1"},
		{"(x/2+1)/(x+2)", "1/2"},
		{"(x^2-y^2)/(x-y)", "x + y"},
		{"(x^2+2xy+y^2)/(x+y)", "x + y"},
		{"(x^2+3x+2)/x", "(x^2 + 3*x + 2)/x"},
		{"x^2/x", "x"},
		{"(x^2+3x+2)/(x+1)", "x + 2"},
	}
	for _, c := range cases {
		res, err := FactorizarRacional(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("FactorizarRacional(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("FactorizarRacional(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestFactorizarRacionalPasos(t *testing.T) {
	res, err := FactorizarRacional(exactParse(t, "(x^2+3x+2)/(x^2-x-2)"))
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(res.Steps, "|")
	for _, fragment := range []string{
		"Operación: factorización de una expresión racional",
		"Factorización del numerador: (x + 1)*(x + 2)",
		"Factorización del denominador:",
		"Factores cancelados: (x + 1)",
		"Factorización: (x + 2)/(x - 2)",
		"denominador original no es cero",
	} {
		if !strings.Contains(all, fragment) {
			t.Errorf("pasos sin %q: %v", fragment, res.Steps)
		}
	}
}

func TestFactorizarRacionalSignoEnPasos(t *testing.T) {
	res, err := FactorizarRacional(exactParse(t, "(y/x-x/y)/(1/y-1/x)"))
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(res.Steps, "|")
	if !strings.Contains(all, "Cambio de signo:") {
		t.Fatalf("falta la nota de cambio de signo: %v", res.Steps)
	}
	if !strings.Contains(all, "Factores cancelados:") {
		t.Fatalf("faltan los factores cancelados: %v", res.Steps)
	}
}

func TestComandoFactorizarRacional(t *testing.T) {
	cases := []struct{ in, want string }{
		{"(x^2+3x+2)/(x^2-x-2)", "(x + 2)/(x - 2)"},
		{"((2x^2-x-1)/(x^2-9))*((x+3)/(2x+1))", "(x - 1)/(x - 3)"},
		{"x^2/(x^2-4)-(x+1)/(x+2)", "1/(x - 2)"},
		{"(y/x-x/y)/(1/y-1/x)", "-x - y"},
		{"(x^2+3x+2)/x", "(x^2 + 3*x + 2)/x"},
		{"x^2/x", "x"},
	}
	for _, c := range cases {
		d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: c.in}
		steps, err := Evaluate(d, "/factorizar", "x")
		if err != nil {
			t.Fatalf("Evaluate /factorizar %q: %v", c.in, err)
		}
		all := strings.Join(steps, "|")
		if !strings.Contains(all, "Factorización: "+c.want) {
			t.Errorf("%q sin resultado %q: %v", c.in, c.want, steps)
		}
		if !strings.Contains(all, "Factorización del denominador:") {
			t.Errorf("%q sin factorización del denominador: %v", c.in, steps)
		}
	}
}

func TestComandoFactorizarFuncionRacional(t *testing.T) {
	d := Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "x^2/(x^2-4)-(x+1)/(x+2)"}
	steps, err := Evaluate(d, "/factorizar", "x")
	if err != nil {
		t.Fatalf("Evaluate /factorizar: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "Factorización: 1/(x - 2)") {
		t.Fatalf("resultado racional ausente: %v", steps)
	}
}

func TestFactorizarRacionalErrores(t *testing.T) {
	for _, in := range []string{
		"1/(x-x)",
		"sqrt(x+1)/(x+1)",
		"(x+1)^(1/2)/(x+2)",
		"sin(x)/(x+1)",
	} {
		if _, err := FactorizarRacional(exactParse(t, in)); err == nil {
			t.Errorf("FactorizarRacional(%q) debe fallar", in)
		}
	}
}

func TestFactorizarRacionalNoCancelaDeMas(t *testing.T) {
	res, err := FactorizarRacional(exactParse(t, "(x+1)/(x+2)"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Ascii != "(x + 1)/(x + 2)" {
		t.Fatalf("resultado inesperado: %q", res.Ascii)
	}
}
