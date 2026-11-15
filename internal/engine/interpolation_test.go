package engine

import (
	"strings"
	"testing"
)

func lastStep(steps []string) string {
	if len(steps) == 0 {
		return ""
	}
	return steps[len(steps)-1]
}

func hasStep(steps []string, frag string) bool {
	for _, s := range steps {
		if strings.Contains(s, frag) {
			return true
		}
	}
	return false
}

func TestLagrangeRecta(t *testing.T) {
	steps, err := Lagrange([]Point{{X: 1, Y: 2}, {X: 3, Y: 4}, {X: 5, Y: 6}})
	if err != nil {
		t.Fatalf("Lagrange: %v", err)
	}
	if got := lastStep(steps); !strings.HasSuffix(got, "P(x) = x + 1") {
		t.Fatalf("polinomio final esperado x + 1: %q", got)
	}
	// los tres polinomios base aparecen con su producto
	for _, li := range []string{"L0", "L1", "L2"} {
		if !hasStep(steps, li) {
			t.Fatalf("falta %s: %v", li, steps)
		}
	}
}

func TestLagrangeCuadratica(t *testing.T) {
	steps, err := Lagrange([]Point{{X: 0, Y: 1}, {X: 1, Y: 3}, {X: 2, Y: 7}})
	if err != nil {
		t.Fatalf("Lagrange: %v", err)
	}
	if got := lastStep(steps); got != "Polinomio (expandido): P(x) = x² + x + 1" {
		t.Fatalf("polinomio final esperado x² + x + 1: %q", got)
	}
}

func TestNewtonRecta(t *testing.T) {
	steps, err := Newton([]Point{{X: 1, Y: 2}, {X: 3, Y: 4}, {X: 5, Y: 6}})
	if err != nil {
		t.Fatalf("Newton: %v", err)
	}
	// la tabla triangular completa se muestra con su cabecera
	if !hasStep(steps, "D.2") || !hasStep(steps, "f[x]") {
		t.Fatalf("la tabla de diferencias divididas está incompleta: %v", steps)
	}
	// coeficientes de la diagonal: 2, 1, 0
	if !hasStep(steps, "f[x0] = 2") || !hasStep(steps, "f[x0..x1] = 1") || !hasStep(steps, "f[x0..x2] = 0") {
		t.Fatalf("coeficientes de la tabla esperados: %v", steps)
	}
	// forma de Newton y forma expandida coinciden
	if got := lastStep(steps); !strings.HasSuffix(got, "P(x) = x + 1") {
		t.Fatalf("polinomio final esperado x + 1: %q", got)
	}
}

func TestNewtonCuadratica(t *testing.T) {
	steps, err := Newton([]Point{{X: 0, Y: 1}, {X: 1, Y: 3}, {X: 2, Y: 7}})
	if err != nil {
		t.Fatalf("Newton: %v", err)
	}
	if got := lastStep(steps); got != "Polinomio (expandido): P(x) = x² + x + 1" {
		t.Fatalf("polinomio final esperado x² + x + 1: %q", got)
	}
}

func TestMisMatchMultiplicativo(t *testing.T) {
	// x^2 a través de (-1,1) (0,0) (1,1)
	steps, err := Newton([]Point{{X: -1, Y: 1}, {X: 0, Y: 0}, {X: 1, Y: 1}})
	if err != nil {
		t.Fatalf("Newton: %v", err)
	}
	if got := lastStep(steps); got != "Polinomio (expandido): P(x) = x²" {
		t.Fatalf("polinomio final esperado x²: %q", got)
	}
}

func TestInterpolationErrores(t *testing.T) {
	if _, err := Lagrange([]Point{{X: 1, Y: 2}}); err == nil {
		t.Fatal("un solo punto debe fallar")
	}
	if _, err := Newton([]Point{}); err == nil {
		t.Fatal("lista vacía debe fallar")
	}
	// abscisa repetida
	dup := []Point{{X: 1, Y: 2}, {X: 1, Y: 5}, {X: 3, Y: 4}}
	if _, err := Lagrange(dup); err == nil || !strings.Contains(err.Error(), "repetida") {
		t.Fatalf("abscisa repetida debe fallar: %v", err)
	}
	if _, err := Newton(dup); err == nil {
		t.Fatalf("Newton con abscisa repetida debe fallar: %v", err)
	}
}

func TestFormatPoly(t *testing.T) {
	cases := []struct {
		coeff []float64
		want  string
	}{
		{[]float64{0}, "0"},
		{[]float64{5}, "5"},
		{[]float64{1, 1}, "x + 1"},
		{[]float64{2, 3}, "3x + 2"},
		{[]float64{-1, 0, 1}, "x² − 1"},
		{[]float64{-2, -3}, "−3x − 2"},
		{[]float64{1, 1, 1}, "x² + x + 1"},
		{[]float64{0.5, 0.5}, "1/2x + 1/2"},
	}
	for _, c := range cases {
		if got := formatPoly(c.coeff); got != c.want {
			t.Errorf("formatPoly(%v) = %q, se esperaba %q", c.coeff, got, c.want)
		}
	}
}
