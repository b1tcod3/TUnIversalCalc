package engine

import (
	"strings"
	"testing"
)

func exactParse(t *testing.T, src string) Expr {
	t.Helper()
	e, err := ParseExpr(src)
	if err != nil {
		t.Fatalf("ParseExpr(%q): %v", src, err)
	}
	return e
}

func TestPotenciaExacta(t *testing.T) {
	cases := []struct{ in, want string }{
		{"(3a^3b^3)(4ab^2)^2", "48*a^5*b^7"},
		{"((3x^(3/2)y^3)/(x^2y^(-1/2)))^-2", "x/(9*y^7)"},
		{"(2x)^-2", "1/(4*x^2)"},
		{"(a+b)^2", "a^2 + 2*a*b + b^2"},
		{"x^0", "1"},
		{"x^2*x^3", "x^5"},
		{"(x^2)^3", "x^6"},
		{"2^-3", "1/8"},
		{"0^2", "0"},
		{"(1/4)^(1/2)", "1/2"},
		{"(-8)^(2/3)", "4"},
		{"(-8)^(1/3)", "-2"},
		{"(-3)^(2/5)", "raiz(5, 9)"},
	}
	for _, c := range cases {
		res, err := Potencia(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("Potencia(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("Potencia(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestExactosVariados(t *testing.T) {
	cases := []struct{ in, want string }{
		{"2x+3x", "5*x"},
		{"0.5x+0.5x", "x"},
		{"1/3+2/3", "1"},
		{"0.1+0.2", "3/10"},
		{"x^2/x^3", "1/x"},
		{"(2x)^2", "4*x^2"},
		{"(x+1)^2", "x^2 + 2*x + 1"},
		{"(x-y)^2", "x^2 - 2*x*y + y^2"},
		{"(x+1)^3", "x^3 + 3*x^2 + 3*x + 1"},
		{"sqrt(50)", "5*sqrt(2)"},
		{"sqrt(72)", "6*sqrt(2)"},
		{"sqrt(12)", "2*sqrt(3)"},
		{"sqrt(2)+sqrt(8)", "3*sqrt(2)"},
		{"sqrt(2)*sqrt(8)", "4"},
		{"raiz(3, 27)", "3"},
		{"raiz(3, 54)", "3*raiz(3, 2)"},
		{"raiz(4, 16)", "2"},
		{"raiz(4, 32)", "2*raiz(4, 2)"},
		{"raiz(5, x^10)", "x^2"},
		{"raiz(2, x^3)", "x*sqrt(x)"},
		{"raiz(2, x^-2)", "1/x"},
		{"alpha^2*alpha^3", "alpha^5"},
		{"2alpha+3alpha", "5*alpha"},
		{"(alpha+beta)^2", "alpha^2 + 2*alpha*beta + beta^2"},
		{"alphabeta", "alphabeta"},
		{"2ab", "2*a*b"},
		{"alpha*beta", "alpha*beta"},
		{"raiz(3, alpha^3)", "alpha"},
		{"raiz(3, -alpha^3)", "-alpha"},
		{"1/x+1/x", "2/x"},
		{"sqrt(x^4)", "x^2"},
		{"sqrt(x^3)", "x*sqrt(x)"},
	}
	for _, c := range cases {
		res, err := Potencia(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("Potencia(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("Potencia(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestPotenciaErrores(t *testing.T) {
	for _, in := range []string{"0^0", "0^-1", "(a+b)^-1", "(a+b)^x"} {
		if _, err := Potencia(exactParse(t, in)); err == nil {
			t.Errorf("Potencia(%q) debe fallar", in)
		}
	}
}

func TestRadicalExacto(t *testing.T) {
	cases := []struct{ in, want string }{
		{"sqrt(200)-sqrt(32)", "6*sqrt(2)"},
		{"sqrt(8)", "2*sqrt(2)"},
		{"sqrt(2)*sqrt(2)", "2"},
		{"sqrt(2)/2+1/sqrt(2)", "sqrt(2)"},
		{"raiz(3, 16)", "2*raiz(3, 2)"},
		{"raiz(4, x^6)", "x*sqrt(x)"},
		{"raiz(3, -8)", "-2"},
		{"sqrt(1/4)", "1/2"},
		{"sqrt(0)", "0"},
		{"raiz(9)", "3"},
		{"sqrt(x^2)", "x"},
	}
	for _, c := range cases {
		res, err := Radical(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("Radical(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("Radical(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestDesarrollar(t *testing.T) {
	cases := []struct{ in, want string }{
		{"(a+b)^2", "a^2 + 2*a*b + b^2"},
		{"(a-b)^2", "a^2 - 2*a*b + b^2"},
		{"(a+b)(a-b)", "a^2 - b^2"},
		{"(a+b)^3", "a^3 + 3*a^2*b + 3*a*b^2 + b^3"},
		{"(a+b)^4", "a^4 + 4*a^3*b + 6*a^2*b^2 + 4*a*b^3 + b^4"},
		{"(2x+3)(3x-4)", "6*x^2 + x - 12"},
		{"(x+1)(x+2)", "x^2 + 3*x + 2"},
		{"(x/2+1)^2", "x^2/4 + x + 1"},
		{"(a+b+c)^2", "a^2 + 2*a*b + 2*a*c + b^2 + 2*b*c + c^2"},
	}
	for _, c := range cases {
		res, err := Desarrollar(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("Desarrollar(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("Desarrollar(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestDesarrollarPasosNotables(t *testing.T) {
	cases := []struct {
		in, fragment string
	}{
		{"(a+b)^2", "cuadrado de un binomio"},
		{"(a+b)(a-b)", "suma por diferencia"},
		{"(a+b+c)^2", "cuadrado de un trinomio"},
		{"(a+b)^4", "binomio de Newton"},
	}
	for _, c := range cases {
		res, err := Desarrollar(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("Desarrollar(%q): %v", c.in, err)
		}
		if !strings.Contains(strings.Join(res.Steps, "|"), c.fragment) {
			t.Errorf("Desarrollar(%q) sin %q: %v", c.in, c.fragment, res.Steps)
		}
	}
}

func TestDesarrollarNoFalsaDiferencia(t *testing.T) {
	res, err := Desarrollar(exactParse(t, "(2x+3)(3x-4)"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(res.Steps, "|"), "suma por diferencia") {
		t.Fatalf("no es una suma por diferencia: %v", res.Steps)
	}
}

func TestDesarrollarErrores(t *testing.T) {
	for _, in := range []string{"(a+b)^-1", "(a+b)^(1/2)", "sin(a+b)^2"} {
		if _, err := Desarrollar(exactParse(t, in)); err == nil {
			t.Errorf("Desarrollar(%q) debe fallar", in)
		}
	}
}

func TestComandoDesarrollar(t *testing.T) {
	d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "(a+b)^2"}
	steps, err := Evaluate(d, "/desarrollar", "x")
	if err != nil {
		t.Fatalf("Evaluate /desarrollar: %v", err)
	}
	all := strings.Join(steps, "|")
	if !strings.Contains(all, "a^2 + 2*a*b + b^2") {
		t.Fatalf("resultado desarrollado ausente: %v", steps)
	}
	if !strings.Contains(all, "cuadrado de un binomio") {
		t.Fatalf("fórmula notable ausente: %v", steps)
	}
}

func TestRadicalErrores(t *testing.T) {
	for _, in := range []string{"raiz(3, alpha+beta)", "sqrt(alpha+beta)"} {
		if _, err := Radical(exactParse(t, in)); err == nil {
			t.Errorf("Radical(%q) debe fallar", in)
		}
	}
}

func TestExactUnicode(t *testing.T) {
	res, err := Potencia(exactParse(t, "(3a^3b^3)(4ab^2)^2"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Unicode != "48·a⁵·b⁷" {
		t.Fatalf("Unicode = %q", res.Unicode)
	}
	res, err = Radical(exactParse(t, "sqrt(200)-sqrt(32)"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Unicode != "6·√2" {
		t.Fatalf("Unicode radical = %q", res.Unicode)
	}
	res, err = Radical(exactParse(t, "raiz(3, x^2)"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Unicode != "∛(x²)" {
		t.Fatalf("Unicode raíz cúbica = %q", res.Unicode)
	}
	res, err = Potencia(exactParse(t, "2spider"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Unicode != "2·spider" {
		t.Fatalf("Unicode con variable larga = %q", res.Unicode)
	}
}

func TestVariablesOfMultiLetter(t *testing.T) {
	d := Decl{Tipo: DeclExpresion, Raw: "2alpha+beta*gamma"}
	got := strings.Join(VariablesOf(d), ",")
	if got != "alpha,beta,gamma" {
		t.Fatalf("VariablesOf = %q", got)
	}
}

func TestCommandsPotenciaRadical(t *testing.T) {
	d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "sqrt(200)-sqrt(32)"}
	steps, err := Evaluate(d, "/radical", "x")
	if err != nil {
		t.Fatalf("Evaluate /radical: %v", err)
	}
	all := strings.Join(steps, "|")
	if !strings.Contains(all, "6*sqrt(2)") {
		t.Fatalf("resultado radical ausente: %v", steps)
	}

	d = Decl{Nombre: "p2", Tipo: DeclExpresion, Raw: "(3a^3b^3)(4ab^2)^2"}
	steps, err = Evaluate(d, "/potencia", "x")
	if err != nil {
		t.Fatalf("Evaluate /potencia: %v", err)
	}
	all = strings.Join(steps, "|")
	if !strings.Contains(all, "48*a^5*b^7") {
		t.Fatalf("resultado potencia ausente: %v", steps)
	}
}
