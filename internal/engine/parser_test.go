package engine

import "testing"

func mustParse(t *testing.T, src string) Expr {
	t.Helper()
	e, err := ParseExpr(src)
	if err != nil {
		t.Fatalf("ParseExpr(%q): %v", src, err)
	}
	return e
}

func TestParsePrecedence(t *testing.T) {
	cases := []struct{ in, want string }{
		{"2+3*4", "2 + 3*4"},
		{"(2+3)*4", "(2 + 3)*4"},
		{"2^3^2", "2^3^2"},
		{"-2^2", "-2^2"}, // = -(2^2)
		{"2x+2", "2*x + 2"},
		{"x^2+3x*2", "x^2 + 3*x*2"},
		{"x^(2y)", "x^(2*y)"},
		{"2(x+1)", "2*(x + 1)"},
		{"x(x+1)", "x*(x + 1)"},
		{"sin(x)cos(x)", "sin(x)*cos(x)"},
		{"3.5x", "3.5*x"},
		{"2/-3", "2/(-3)"}, // el signo se normaliza al simplificar
		{"x^2 - x + 1", "x^2 - x + 1"},
		{"sqrt(x)+1", "sqrt(x) + 1"},
	}
	for _, c := range cases {
		got := mustParse(t, c.in).Ascii()
		if got != c.want {
			t.Errorf("ParseExpr(%q) = %q, se esperaba %q", c.in, got, c.want)
		}
	}
}

func TestParseEquation(t *testing.T) {
	cases := []struct {
		in, lhs, rhs string
	}{
		{"2x+2=0", "2*x + 2", "0"},
		{"(x+1)=(2-x)", "x + 1", "2 - x"},
		{"x^2-4=0", "x^2 - 4", "0"},
	}
	for _, c := range cases {
		l, r, err := ParseEquation(c.in)
		if err != nil {
			t.Fatalf("ParseEquation(%q): %v", c.in, err)
		}
		if l.Ascii() != c.lhs || r.Ascii() != c.rhs {
			t.Errorf("ParseEquation(%q) = %q y %q, se esperaba %q y %q", c.in, l.Ascii(), r.Ascii(), c.lhs, c.rhs)
		}
	}
}

func TestParseErrors(t *testing.T) {
	cases := []string{
		"", "x^2+", "+x", "3**2", "(x^2+1", "x^2)+1",
		"2x=0=1", "x/", "/x", "x^",
	}
	for _, in := range cases {
		if _, err := ParseExpr(in); err == nil {
			t.Errorf("ParseExpr(%q) debe fallar", in)
		}
	}

	// '=' sin el comando /ecuacion no cae en esta función
	if _, err := ParseExpr("x=2"); err == nil {
		t.Error("ParseExpr('x=2') debe fallar: sobra '='")
	}

	// mal formadas
	for _, in := range []string{"2x+2=", "2x=0=1"} {
		if _, _, err := ParseEquation(in); err == nil {
			t.Errorf("ParseEquation(%q) debe fallar", in)
		}
	}
}

func TestParseFuncCallVsImplicitMul(t *testing.T) {
	// sin(x) es llamada; x(y+1) es multiplicación
	if s := mustParse(t, "sin(x)").Ascii(); s != "sin(x)" {
		t.Fatalf("sin(x) → %q", s)
	}
	if s := mustParse(t, "x(y+1)").Ascii(); s != "x*(y + 1)" {
		t.Fatalf("x(y+1) → %q", s)
	}
	if s := mustParse(t, "2sin(x)").Ascii(); s != "2*sin(x)" {
		t.Fatalf("2sin(x) → %q", s)
	}
}
