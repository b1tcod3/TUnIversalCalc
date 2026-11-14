package engine

import (
	"math"
	"testing"
)

func TestParseIntegerLiteral(t *testing.T) {
	cases := []struct {
		in   string
		n    int64
		want bool
	}{
		{"120", 120, true},
		{"-12", -12, true},
		{"+7", 7, true},
		{" 42 ", 42, true},
		{"0012", 12, true},
		{"-9223372036854775808", math.MinInt64, true},
		{"", 0, false},
		{"+", 0, false},
		{"-", 0, false},
		{"12.0", 0, false},
		{"1e3", 0, false},
		{"12x", 0, false},
		{"x", 0, false},
		{"9223372036854775808", 0, false},
	}
	for _, c := range cases {
		n, ok := ParseIntegerLiteral(c.in)
		if ok != c.want || (ok && n != c.n) {
			t.Errorf("ParseIntegerLiteral(%q) = %d, %v; se esperaba %d, %v", c.in, n, ok, c.n, c.want)
		}
	}
}

func TestFactorizeFormato(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "−1"},
		{2, "2"},
		{3, "3"},
		{4, "2²"},
		{5, "5"},
		{6, "2 · 3"},
		{8, "2³"},
		{9, "3²"},
		{12, "2² · 3"},
		{27, "3³"},
		{30, "2 · 3 · 5"},
		{97, "97"},
		{100, "2² · 5²"},
		{120, "2³ · 3 · 5"},
		{210, "2 · 3 · 5 · 7"},
		{1024, "2¹⁰"},
		{600851475143, "71 · 839 · 1471 · 6857"},
		{-12, "−1 · 2² · 3"},
		{-7, "−1 · 7"},
	}
	for _, c := range cases {
		if got := FormatFactorization(Factorize(c.n)); got != c.want {
			t.Errorf("Factorize(%d) = %q, se esperaba %q", c.n, got, c.want)
		}
	}
}

func TestFactorizeFormaExpandida(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "−1"},
		{120, "2 · 2 · 2 · 3 · 5"},
		{-12, "−1 · 2 · 2 · 3"},
		{97, "97"},
	}
	for _, c := range cases {
		if got := FormatExpanded(Factorize(c.n)); got != c.want {
			t.Errorf("FormatExpanded(%d) = %q, se esperaba %q", c.n, got, c.want)
		}
	}
}

func TestFactorizeMinInt64(t *testing.T) {
	f := Factorize(math.MinInt64)
	if !f.Negative {
		t.Fatal("MinInt64 debe ser negativo")
	}
	if len(f.Factors) != 63 {
		t.Fatalf("factores = %d, se esperaban 63", len(f.Factors))
	}
	for i, p := range f.Factors {
		if p != 2 {
			t.Fatalf("factor[%d] = %d", i, p)
		}
	}
	if got := FormatFactorization(f); got != "−1 · 2⁶³" {
		t.Fatalf("formato = %q", got)
	}
}

func TestFactorizeSemiprimoGrande(t *testing.T) {
	const p, q = uint64(1000000007), uint64(1000000009)
	n := p * q
	f := Factorize(int64(n))
	if len(f.Factors) != 2 || f.Factors[0] != p || f.Factors[1] != q {
		t.Fatalf("factores = %v, se esperaban [%d %d]", f.Factors, p, q)
	}
}

func TestFactorizePropiedadProducto(t *testing.T) {
	for n := int64(2); n <= 5000; n++ {
		f := Factorize(n)
		var prod uint64 = 1
		for _, p := range f.Factors {
			prod *= p
		}
		if prod != uint64(n) {
			t.Fatalf("producto de factores de %d = %d", n, prod)
		}
		for _, p := range f.Factors {
			if !isPrime64(p) {
				t.Fatalf("factor %d de %d no es primo", p, n)
			}
		}
	}
}

func TestIntegerValue(t *testing.T) {
	if n, ok := (Decl{Tipo: DeclExpresion, Raw: "120"}).IntegerValue(); !ok || n != 120 {
		t.Fatalf("IntegerValue = %d, %v", n, ok)
	}
	if _, ok := (Decl{Tipo: DeclFuncion, Raw: "120"}).IntegerValue(); ok {
		t.Fatal("una función no debe ser un entero")
	}
	d := IntegerDecl(-12)
	if d.Tipo != DeclExpresion || d.Nombre != "-12" || d.Raw != "-12" {
		t.Fatalf("IntegerDecl = %+v", d)
	}
}
