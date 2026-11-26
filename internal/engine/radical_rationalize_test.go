package engine

import (
	"strings"
	"testing"
)

func TestRadicalizarEjemplos(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"sqrt(10)/(sqrt(5)-2)", "5*sqrt(2) + 2*sqrt(10)"},
		{"(sqrt(4+h)-2)/h", "1/(sqrt(4 + h) + 2)"},
		{"1/(sqrt(2)+1)", "-1 + sqrt(2)"},
		{"1/(sqrt(2)-1)", "1 + sqrt(2)"},
		{"1/(2^(1/2)-1)", "1 + sqrt(2)"},
		{"((4+h)^(1/2)-2)/h", "1/(sqrt(4 + h) + 2)"},
		{"1/(2-sqrt(5))", "-2 - sqrt(5)"},
		{"1/(sqrt(2)-sqrt(3))", "-sqrt(2) - sqrt(3)"},
		{"(2+sqrt(3))/(2-sqrt(3))", "7 + 4*sqrt(3)"},
		{"(sqrt(5)-2)/(sqrt(5)+2)", "9 - 4*sqrt(5)"},
		{"1/(2*sqrt(5)-4)", "1 + sqrt(5)/2"},
		{"-(sqrt(10)/(sqrt(5)-2))", "-5*sqrt(2) - 2*sqrt(10)"},
	}
	for _, c := range cases {
		res, err := Radicalizar(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("Radicalizar(%q): %v", c.in, err)
		}
		if res.Ascii != c.want {
			t.Errorf("Radicalizar(%q) = %q, se esperaba %q", c.in, res.Ascii, c.want)
		}
	}
}

func TestRadicalizarUnicode(t *testing.T) {
	cases := []struct{ in, want string }{
		{"sqrt(10)/(sqrt(5)-2)", "5·√2 + 2·√10"},
		{"(sqrt(4+h)-2)/h", "1/(√(4 + h) + 2)"},
	}
	for _, c := range cases {
		res, err := Radicalizar(exactParse(t, c.in))
		if err != nil {
			t.Fatalf("Radicalizar(%q): %v", c.in, err)
		}
		if res.Unicode != c.want {
			t.Errorf("Unicode de %q = %q, se esperaba %q", c.in, res.Unicode, c.want)
		}
	}
}

func TestRadicalizarExponenteMedio(t *testing.T) {
	res, err := Radicalizar(exactParse(t, "10^(1/2)/(sqrt(5)-2)"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Ascii != "5*sqrt(2) + 2*sqrt(10)" {
		t.Fatalf("resultado = %q", res.Ascii)
	}
}

func TestRadicalizarPasos(t *testing.T) {
	res, err := Radicalizar(exactParse(t, "sqrt(10)/(sqrt(5)-2)"))
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(res.Steps, "|")
	for _, frag := range []string{
		"radicalización por conjugado",
		"Conjugado del denominador: sqrt(5) + 2",
		"Numerador: 5*sqrt(2) + 2*sqrt(10)",
		"Denominador: 1",
		"Resultado: 5*sqrt(2) + 2*sqrt(10)",
	} {
		if !strings.Contains(all, frag) {
			t.Errorf("falta %q en los pasos: %v", frag, res.Steps)
		}
	}
}

func TestRadicalizarCancelaFactor(t *testing.T) {
	res, err := Radicalizar(exactParse(t, "(sqrt(4+h)-2)/h"))
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(res.Steps, "|")
	for _, frag := range []string{
		"Conjugado del numerador: sqrt(4 + h) + 2",
		"Numerador: h",
		"Factores cancelados: h",
		"Resultado: 1/(sqrt(4 + h) + 2)",
	} {
		if !strings.Contains(all, frag) {
			t.Errorf("falta %q en los pasos: %v", frag, res.Steps)
		}
	}
}

func TestRadicalizarSinConjugado(t *testing.T) {
	res, err := Radicalizar(exactParse(t, "sqrt(200)-sqrt(32)"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Ascii != "6*sqrt(2)" {
		t.Fatalf("resultado = %q", res.Ascii)
	}
	all := strings.Join(res.Steps, "|")
	if !strings.Contains(all, "No se encontró un binomio radical") {
		t.Fatalf("falta la nota de simplificación: %v", res.Steps)
	}
}

func TestRadicalizarErrores(t *testing.T) {
	for _, in := range []string{
		"sqrt(alpha+beta)",
		"1/(alpha+beta)",
		"sqrt(-4)/(sqrt(5)-2)",
		"0^0/(sqrt(5)-2)",
	} {
		if _, err := Radicalizar(exactParse(t, in)); err == nil {
			t.Errorf("Radicalizar(%q) debe fallar", in)
		}
	}
}

func TestEvaluateRadicalizar(t *testing.T) {
	d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "sqrt(10)/(sqrt(5)-2)"}
	steps, err := Evaluate(d, "/radicalizar", "x")
	if err != nil {
		t.Fatalf("Evaluate /radicalizar: %v", err)
	}
	all := strings.Join(steps, "|")
	for _, frag := range []string{
		"radicalización por conjugado",
		"Conjugado del denominador: sqrt(5) + 2",
		"Resultado: 5*sqrt(2) + 2*sqrt(10)",
	} {
		if !strings.Contains(all, frag) {
			t.Fatalf("falta %q en: %v", frag, steps)
		}
	}

	// /radical debe seguir usando la simplificación ordinaria.
	steps, err = Evaluate(d, "/radical", "x")
	if err == nil {
		t.Fatalf("/radical no debe aceptar una división por una suma: %v", steps)
	}

	d = Decl{Nombre: "p2", Tipo: DeclExpresion, Raw: "sqrt(200)-sqrt(32)"}
	steps, err = Evaluate(d, "/radical", "x")
	if err != nil {
		t.Fatalf("Evaluate /radical: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "6*sqrt(2)") {
		t.Fatalf("/radical perdió su resultado: %v", steps)
	}
	steps, err = Evaluate(d, "/radicalizar", "x")
	if err != nil {
		t.Fatalf("Evaluate /radicalizar sin conjugado: %v", err)
	}
	all = strings.Join(steps, "|")
	if !strings.Contains(all, "6*sqrt(2)") || !strings.Contains(all, "No se encontró un binomio radical") {
		t.Fatalf("fallback de /radicalizar incorrecto: %v", steps)
	}
}
