package render

import "testing"

func TestToSuperscript(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"digitos", "0123456789", "⁰¹²³⁴⁵⁶⁷⁸⁹"},
		{"signos", "+-=()", "⁺⁻⁼⁽⁾"},
		{"letras con mapeo", "abcxyz", "ᵃᵇᶜˣʸᶻ"},
		{"sin mapeo se conserva", "fqF?", "fqF?"},
		{"mapea solo lo conocido", "a²", "ᵃ²"},
		{"vacio", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ToSuperscript(tt.in); got != tt.want {
				t.Fatalf("ToSuperscript(%q) = %q, se esperaba %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestToSubscript(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"digitos", "0123456789", "₀₁₂₃₄₅₆₇₈₉"},
		{"signos", "+-=()", "₊₋₌₍₎"},
		{"letras con mapeo", "aeox", "ₐₑₒₓ"},
		{"sin mapeo se conserva", "bcf", "bcf"},
		{"vacio", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ToSubscript(tt.in); got != tt.want {
				t.Fatalf("ToSubscript(%q) = %q, se esperaba %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPretty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"potencia simple", "x^2", "x²"},
		{"potencia agrupada", "x^(2y)", "x²ʸ"},
		{"potencia de dos digitos", "x^10", "x¹⁰"},
		{"exponente negativo", "x^-2", "x⁻²"},
		{"exponente con suma", "x^(2+1)+y", "x²⁺¹+y"},
		{"potencia sin parentesis", "x^2y", "x²ʸ"},
		{"potencia mas suma", "x^2+1", "x²+1"},
		{"potencia menos resta", "x^2-1", "x²−1"},
		{"potencia incompleta", "x^", "x"},
		{"raiz", "sqrt(x^2+1)", "√(x²+1)"},
		{"raiz anidada", "sqrt(sqrt(x))", "√(√(x))"},
		{"raiz cuadrada nueva", "raiz(9)", "√9"},
		{"raiz cubica", "raiz(3, 8)", "∛8"},
		{"raiz cuarta", "raiz(4, x+1)", "∜(x+1)"},
		{"raiz quinta", "raiz(5, x)", "⁵√x"},
		{"raiz anidada nueva", "raiz(3, sqrt(x))", "∛(√(x))"},
		{"pi y multiplicacion", "pi*r^2", "π·r²"},
		{"infinito", "inf+1", "∞+1"},
		{"2pi", "2pi", "2π"},
		{"variable con pi", "spider", "spider"},
		{"variable con sqrt", "sqrtx", "sqrtx"},
		{"resta y multiplicacion", "2*3-1", "2·3−1"},
		{"sin cambios", "sin(x)", "sin(x)"},
		{"vacio", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Pretty(tt.in); got != tt.want {
				t.Fatalf("Pretty(%q) = %q, se esperaba %q", tt.in, got, tt.want)
			}
		})
	}
}
