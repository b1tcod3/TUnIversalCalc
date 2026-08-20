package engine

import "testing"

func TestParseDecl(t *testing.T) {
	cases := []struct {
		in   string
		err  bool
	}{
		{"", true},
		{"2x+2", true},
		{"/funcion 2x+2", false},
		{"/funcion", true},
		{"/funcion   ", true},
		{"/funcion x=2", true},
		{"/funcion x^2+", true},
		{"/ecuacion 2x+2=0", false},
		{"/ecuacion 2x+2", true},
		{"/ecuacion 2x+2=", true},
		{"/ecuacion (2x+2)=0", false},
		{"/ecuacion 2x=0=1", true},
		{"/ecuacion x^2- =1", true},
		{"/expresion x^2+1", false},
		{"/expresion x=1", true},
		{"/expresion ", true},
		{"/resolver x", true},
		{"/zz 2x", true},
	}
	for _, c := range cases {
		_, err := ParseDecl(c.in)
		if (err != nil) != c.err {
			t.Errorf("ParseDecl(%q): error=%v, se esperaba err=%v", c.in, err, c.err)
		}
	}
}

func TestParseDeclContent(t *testing.T) {
	d, err := ParseDecl("/ecuacion 2x+2=0")
	if err != nil {
		t.Fatalf("ecuación válida rechazada: %v", err)
	}
	if d.Tipo != DeclEcuacion || d.Raw != "2x+2=0" {
		t.Fatalf("ecuación mal clasificada: %+v", d)
	}
	if d.Latex != "2x + 2 = 0" {
		t.Fatalf("LaTeX = %q", d.Latex)
	}

	f, err := ParseDecl("/funcion x^2+3x*2")
	if err != nil {
		t.Fatalf("función válida rechazada: %v", err)
	}
	if f.Tipo != DeclFuncion {
		t.Fatalf("tipo = %v", f.Tipo)
	}
	if f.Latex != "x^{2} + 3x \\cdot 2" {
		t.Fatalf("LaTeX = %q", f.Latex)
	}

	p, err := ParseDecl("/expresion x^(2y)+sqrt(x)")
	if err != nil {
		t.Fatalf("expresión válida rechazada: %v", err)
	}
	if p.Tipo != DeclExpresion || p.Pretty != "x²ʸ+√(x)" {
		t.Fatalf("expresión mal transformada: %+v", p)
	}
}

func TestDeclType(t *testing.T) {
	if DeclEcuacion.String() != "ecuación" || DeclFuncion.Prefix() != "f" || DeclExpresion.Prefix() != "p" {
		t.Fatalf("DeclType String/Prefix incorrectos")
	}
}