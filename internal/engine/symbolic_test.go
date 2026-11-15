package engine

import (
	"strings"
	"testing"
)

// sim simplifies an expression (helper).
func sim(t *testing.T, src string) Expr {
	t.Helper()
	e := mustParse(t, src)
	e = simplifyNormaliza(e)
	return e
}

func TestSimplify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"2x+3x", "5*x"},
		{"x^2+x^2", "2*x^2"},
		{"2x+2-2x+1", "3"},
		{"x*x", "x^2"},
		{"x^2*x^3", "x^5"},
		{"3*4", "12"},
		{"x^0", "1"},
		{"x^1", "x"},
		{"1*x", "x"},
		{"0*x+5", "5"},
		{"-(-x)", "x"},
		{"2x-x", "x"},
		{"x-x", "0"},
		{"4/2", "2"},
		{"2sin(x)+3sin(x)", "5*sin(x)"},
		{"-2x^2", "-2*x^2"},
		{"x+x", "2*x"},
		{"3x*2", "6*x"},
	}
	for _, c := range cases {
		got := sim(t, c.in).Ascii()
		if got != c.want {
			t.Errorf("Simplify(%q) = %q, se esperaba %q", c.in, got, c.want)
		}
	}
}

func TestDerive(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"x", "1"},
		{"5", "0"},
		{"x^2", "2*x"},
		{"x^2+3x-5", "2*x + 3"},
		{"x^3+4x^2-2x+7", "3*x^2 + 8*x - 2"},
		{"2x+2", "2"},
		{"sin(x)", "cos(x)"},
		{"cos(x)", "-sin(x)"},
		{"sqrt(x)", "1/(2*sqrt(x))"},
		{"x^2*sin(x)", "2*x*sin(x) + cos(x)*x^2"},
		{"1/x", "-1/x^2"},
		{"3x^2*2x", "18*x^2"},
		{"x^(1/2)", "1/2*x^-0.5"},
		{"-x", "-1"},
		{"-x^2", "-2*x"},
	}
	for _, c := range cases {
		r, steps, err := Derive(mustParse(t, c.in), 'x')
		if err != nil {
			t.Fatalf("Derive(%q): %v", c.in, err)
		}
		if got := r.Ascii(); got != c.want {
			t.Errorf("Derive(%q) = %q, se esperaba %q", c.in, got, c.want)
		}
		if len(steps) == 0 {
			t.Errorf("Derive(%q): sin pasos", c.in)
		}
	}
}

func TestDeriveChainRuleSteps(t *testing.T) {
	r, steps, err := Derive(mustParse(t, "sin(x^2)"), 'x')
	if err != nil {
		t.Fatalf("Derive: %v", err)
	}
	if got := r.Ascii(); got != "2*cos(x^2)*x" {
		t.Fatalf("Derive(sin(x^2)) = %q", got)
	}
	var sb strings.Builder
	for _, s := range steps {
		sb.WriteString(s + "\n")
	}
	all := sb.String()
	if !strings.Contains(all, "cadena") {
		t.Fatalf("pasos sin regla de la cadena:\n%s", all)
	}
}

func TestDeriveOtherVar(t *testing.T) {
	r, _, err := Derive(mustParse(t, "y^2+5"), 'x')
	if err != nil {
		t.Fatalf("Derive: %v", err)
	}
	if r.Ascii() != "0" {
		t.Fatalf("d/dx(y^2+5) = %q, se esperaba 0", r.Ascii())
	}
}

func TestDeriveUnsupported(t *testing.T) {
	if _, _, err := Derive(mustParse(t, "x^x"), 'x'); err == nil {
		t.Fatal("x^x debe dar error")
	}
}

func TestSolveLinear(t *testing.T) {
	steps, err := Solve(mustParse(t, "2x+2"), nil, 'x')
	if err != nil {
		t.Fatalf("Solve(2x+2): %v", err)
	}
	has := strings.Join(steps, " ")
	if !strings.Contains(has, "x = -1") {
		t.Fatalf("solución esperada x = -1, pasos: %v", steps)
	}

	steps, err = Solve(mustParse(t, "3x-6"), nil, 'x')
	if err != nil {
		t.Fatalf("Solve(3x-6): %v", err)
	}
	if !strings.Contains(strings.Join(steps, " "), "x = 2") {
		t.Fatalf("solución esperada x = 2, pasos: %v", steps)
	}
}

func TestSolveEquationLHSRHS(t *testing.T) {
	lhs, rhs := mustParse(t, "x+1"), mustParse(t, "2x-3")
	steps, err := Solve(lhs, rhs, 'x')
	if err != nil {
		t.Fatalf("Solve(x+1=2x-3): %v", err)
	}
	var sb strings.Builder
	for _, s := range steps {
		sb.WriteString(s + "\n")
	}
	all := sb.String()
	if !strings.Contains(all, "x = 4") {
		t.Fatalf("solución esperada x = 4:\n%s", all)
	}
}

func TestSolveQuadratic(t *testing.T) {
	steps, err := Solve(mustParse(t, "x^2-4"), nil, 'x')
	if err != nil {
		t.Fatalf("Solve(x^2-4): %v", err)
	}
	var sb strings.Builder
	for _, s := range steps {
		sb.WriteString(s + "\n")
	}
	all := sb.String()
	if !strings.Contains(all, "x = 2") || !strings.Contains(all, "x = -2") {
		t.Fatalf("raíces esperadas ±2:\n%s", all)
	}

	steps, err = Solve(mustParse(t, "x^2+1"), nil, 'x')
	if err != nil {
		t.Fatalf("Solve(x^2+1): %v", err)
	}
	if !strings.Contains(strings.Join(steps, " "), "sin soluciones reales") {
		t.Fatalf("se esperaba Δ<0: %v", steps)
	}
}

func TestSolveNoVariable(t *testing.T) {
	if _, err := Solve(mustParse(t, "2+2"), nil, 'x'); err == nil {
		t.Fatal("sin la variable x debe dar error")
	}
}

func TestSolveIdentityAndContradiction(t *testing.T) {
	// 0 = 0 (caso simplificado a 0)
	steps, err := Solve(mustParse(t, "x-x"), nil, 'x')
	if err != nil {
		t.Fatalf("Solve(x-x): %v", err)
	}
	str := strings.Join(steps, " ")
	if !strings.Contains(str, "Identidad") {
		t.Fatalf("x=x debe ser identidad: %v", steps)
	}
	// 2 = 0 → contradicción: error
	if _, err := Solve(mustParse(t, "x-x+2"), nil, 'x'); err == nil {
		t.Fatal("2=0 debe ser contradicción (error)")
	}
}

func TestIntegrate(t *testing.T) {
	cases := []struct{ in, want string }{
		{"x^2+3x-5", "1/3*x^3 + 3/2*x^2 - 5*x"},
		{"2x+2", "x^2 + 2*x"},
		{"sin(x)", "-cos(x)"},
		{"cos(x)", "sin(x)"},
		{"1/x", "ln(x)"},
		{"1/x^2", "-1/x"},
		{"x^(1/2)", "2/3*x*sqrt(x)"},
		{"2", "2*x"},
	}
	for _, c := range cases {
		r, _, err := Integrate(mustParse(t, c.in), 'x')
		if err != nil {
			t.Fatalf("Integrate(%q): %v", c.in, err)
		}
		if got := r.Ascii(); got != c.want {
			t.Errorf("Integrate(%q) = %q, se esperaba %q", c.in, got, c.want)
		}
	}
}

func TestIntegrateUnsupported(t *testing.T) {
	if _, _, err := Integrate(mustParse(t, "x*sin(x)"), 'x'); err == nil {
		t.Fatal("x·sin(x) debe dar error (necesita partes)")
	}
	if _, _, err := Integrate(mustParse(t, "abs(x)"), 'x'); err == nil {
		t.Fatal("abs(x) debe dar error")
	}
}

func TestLimitSubstitution(t *testing.T) {
	cases := []struct {
		name string
		in   string
		val  Value
		want string
	}{
		{"sustitucion directa", "x^2+1", Value{Val: 2}, "5"},
		{"lhopital", "sin(x)/x", Value{Val: 0}, "1"},
		{"polinomio en 1", "x^2-1", Value{Val: 1}, "0"},
		{"potencia en infinito", "x^2", Value{IsInf: true, Sign: 1}, "inf"},
		{"inverso en infinito", "1/x^2", Value{IsInf: true, Sign: 1}, "0"},
		{"suma en infinito", "x^2+3x+1", Value{IsInf: true, Sign: 1}, "inf"},
		{"negativo en infinito", "-x^2+5", Value{IsInf: true, Sign: 1}, "-inf"},
		{"producto en infinito", "(x+1)*(x+2)", Value{IsInf: true, Sign: 1}, "inf"},
		{"cociente en infinito", "(x+1)/(x+2)", Value{IsInf: true, Sign: 1}, "1"},
		{"cubo en menos infinito", "x^3", Value{IsInf: true, Sign: -1}, "-inf"},
		{"cuadrado en menos infinito", "x^2", Value{IsInf: true, Sign: -1}, "inf"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r, _, err := Limit(mustParse(t, c.in), 'x', c.val)
			if err != nil {
				t.Fatalf("Limit(%q, %v): %v", c.in, c.val, err)
			}
			if got := r.Ascii(); got != c.want {
				t.Fatalf("Limit(%q, %v) = %q, se esperaba %q", c.in, c.val, got, c.want)
			}
		})
	}
}

func TestSimplifyPublicAPI(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		want  string
		steps int
	}{
		{"agrupa terminos", "2x+3x", "5*x", 2},
		{"ya simplificada", "x", "x", 1},
		{"potencia numerica", "2^3", "8", 2},
		{"potencia indefinida", "0^-1", "0^-1", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, steps, err := Simplify(mustParse(t, c.in))
			if err != nil {
				t.Fatalf("Simplify(%q): %v", c.in, err)
			}
			if got.Ascii() != c.want {
				t.Fatalf("Simplify(%q) = %q, se esperaba %q", c.in, got.Ascii(), c.want)
			}
			if len(steps) != c.steps {
				t.Fatalf("Simplify(%q): %d pasos, se esperaban %d", c.in, len(steps), c.steps)
			}
		})
	}
}

func TestMergeMulBasesNumInvalido(t *testing.T) {
	coeff, rest := mergeMulBases([]Expr{Num{Text: "no-num"}, Num{Text: "2"}})
	if coeff != 2 {
		t.Fatalf("coeficiente = %v, se esperaba 2", coeff)
	}
	if len(rest) != 1 || rest[0].Ascii() != "no-num" {
		t.Fatalf("el Num inválido debe conservarse como factor: %v", rest)
	}
}

func TestLHopitalSinDerivada(t *testing.T) {
	e := Call{Name: "tan", Arg: Var{Name: 'x'}}
	got := lHopital(e, 'x')
	if got == nil {
		t.Fatal("lHopital no debe devolver nil si la derivada falla")
	}
	if got.Ascii() != e.Ascii() {
		t.Fatalf("lHopital(%q) = %q, se esperaba conservar la expresión", e.Ascii(), got.Ascii())
	}
}

func TestLimitUnsupported(t *testing.T) {
	if _, _, err := Limit(mustParse(t, "sin(x)/x"), 'x', Value{IsInf: true, Sign: 1}); err == nil {
		t.Fatal("sin(x)/x en ∞ debe dar error (no polinómico)")
	}
}

func TestNumEval(t *testing.T) {
	v, err := numEval(mustParse(t, "2x^2+1"), 'x', 3)
	if err != nil {
		t.Fatalf("numEval: %v", err)
	}
	if v != 19 {
		t.Fatalf("2·3²+1 = %v, se esperaba 19", v)
	}
	if _, err := numEval(mustParse(t, "y^2"), 'x', 1); err == nil {
		t.Fatal("y sin sustituir debe dar error")
	}
}
