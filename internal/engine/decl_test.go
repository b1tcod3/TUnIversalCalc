package engine

import "testing"

func TestParseDecl(t *testing.T) {
	cases := []struct {
		in  string
		err bool
	}{
		{"", true},
		{"2x+2", true},
		{"120", false},
		{"-12", false},
		{"12.5", true},
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
		{"/puntos (1,2) (3,4)", false},
		{"/puntos (1,2)", true},
		{"/puntos 1,2 3,4 5,6", false},
		{"/puntos (1,2)(3,4)", false},
		{"/puntos", true},
		{"/puntos (a,b)", true},
		{"/puntos (1,2) (1,5)", true},
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

func TestParseDeclPuntos(t *testing.T) {
	d, err := ParseDecl("/puntos (1,2) (3,4) (5,6)")
	if err != nil {
		t.Fatalf("puntos válidos rechazados: %v", err)
	}
	if d.Tipo != DeclPuntos || len(d.Points) != 3 {
		t.Fatalf("puntos mal clasificados: %+v", d)
	}
	want := []Point{{1, 2}, {3, 4}, {5, 6}}
	for i, p := range d.Points {
		if p != want[i] {
			t.Fatalf("punto %d = %+v, se esperaba %+v", i, p, want[i])
		}
	}

	// Formato sin paréntesis también se acepta
	d2, err := ParseDecl("/puntos 0,1 1,3")
	if err != nil {
		t.Fatalf("puntos sin paréntesis rechazados: %v", err)
	}
	if d2.Tipo != DeclPuntos || len(d2.Points) != 2 || d2.Points[1].X != 1 {
		t.Fatalf("puntos sin paréntesis mal parseados: %+v", d2.Points)
	}
}

func TestDeclType(t *testing.T) {
	if DeclEcuacion.String() != "ecuación" || DeclFuncion.Prefix() != "f" || DeclExpresion.Prefix() != "p" {
		t.Fatalf("DeclType String/Prefix incorrectos")
	}
}

func TestParseDeclEnteroSuelto(t *testing.T) {
	d, err := ParseDecl("120")
	if err != nil {
		t.Fatalf("entero suelto rechazado: %v", err)
	}
	if d.Tipo != DeclExpresion || d.Raw != "120" || d.Latex != "120" {
		t.Fatalf("entero mal clasificado: %+v", d)
	}
	d, err = ParseDecl("-12")
	if err != nil || d.Raw != "-12" || d.Pretty != "−12" {
		t.Fatalf("negativo mal clasificado: %+v, %v", d, err)
	}
}
