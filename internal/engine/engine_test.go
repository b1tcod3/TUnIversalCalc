package engine

import (
	"strings"
	"testing"
)

func evalDecl(t *testing.T, decl Decl, cmd string) ([]string, error) {
	t.Helper()
	return Evaluate(decl, cmd, "x")
}

// puntosColineales: (1,2) (3,4) (5,6) están en y = x + 1.
func puntosColineales() Decl {
	return Decl{
		Nombre: "t1",
		Tipo:   DeclPuntos,
		Raw:    "(1,2) (3,4) (5,6)",
		Latex:  "(1,2) (3,4) (5,6)",
		Pretty: "(1,2) (3,4) (5,6)",
		Points: []Point{{X: 1, Y: 2}, {X: 3, Y: 4}, {X: 5, Y: 6}},
	}
}

// puntosCuadratica: (0,1) (1,3) (2,7) están en y = x² + x + 1.
func puntosCuadratica() Decl {
	return Decl{
		Nombre: "t1",
		Tipo:   DeclPuntos,
		Raw:    "(0,1) (1,3) (2,7)",
		Latex:  "(0,1) (1,3) (2,7)",
		Pretty: "(0,1) (1,3) (2,7)",
		Points: []Point{{X: 0, Y: 1}, {X: 1, Y: 3}, {X: 2, Y: 7}},
	}
}

func TestEvaluateLagrange(t *testing.T) {
	steps, err := evalDecl(t, puntosColineales(), "/lagrange t1")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	all := strings.Join(steps, "|")
	if !strings.Contains(all, "t1") || !strings.Contains(all, "interpolación") {
		t.Fatalf("cabecera de la operación faltante: %s", all)
	}
	// tres puntos colineales → recta x + 1
	if !strings.Contains(all, "L0") || !strings.Contains(all, "L1") || !strings.Contains(all, "L2") {
		t.Fatalf("deben aparecer los polinomios base L0, L1, L2: %s", all)
	}
	if !strings.Contains(all, "P(x) = x + 1") {
		t.Fatalf("polinomio esperado x + 1: %s", all)
	}
}

func TestEvaluateNewton(t *testing.T) {
	steps, err := evalDecl(t, puntosColineales(), "/newton t1")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	all := strings.Join(steps, "|")
	if !strings.Contains(all, "diferencias divididas") || !strings.Contains(all, "Tabla") {
		t.Fatalf("deben aparecer la tabla de diferencias divididas: %s", all)
	}
	if !strings.Contains(all, "P(x) = x + 1") {
		t.Fatalf("polinomio esperado x + 1: %s", all)
	}
}

func TestEvaluateLagrangeCuadratica(t *testing.T) {
	steps, err := evalDecl(t, puntosCuadratica(), "/lagrange")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "P(x) = x² + x + 1") {
		t.Fatalf("polinomio esperado x² + x + 1: %v", steps)
	}
}

func TestEvaluateOperacionesRequierenPuntos(t *testing.T) {
	// lagrange sobre una función no declarada como puntos falla
	funcDecl := Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "2x+2", Latex: "2x + 2"}
	if _, err := evalDecl(t, funcDecl, "/lagrange f1"); err == nil ||
		!strings.Contains(err.Error(), "puntos") {
		t.Fatalf("lagrange sobre función debe fallar: %v", err)
	}
	if _, err := evalDecl(t, funcDecl, "/newton f1"); err == nil {
		t.Fatal("newton sobre función debe fallar")
	}
}

func TestEvaluateFactorizar(t *testing.T) {
	d := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "120", Latex: "120", Pretty: "120"}
	steps, err := evalDecl(t, d, "/factorizar")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	all := strings.Join(steps, "|")
	fragments := []string{
		"N = 120",
		"÷2 (3 veces)",
		"÷3 (1 vez)",
		"Factorización: 2³ · 3 · 5",
		"Forma expandida: 2 · 2 · 2 · 3 · 5",
	}
	for _, frag := range fragments {
		if !strings.Contains(all, frag) {
			t.Fatalf("falta %q en: %s", frag, all)
		}
	}
}

func TestEvaluateFactorizarErrores(t *testing.T) {
	// Los puntos no son expresiones factorizables.
	puntos := Decl{Nombre: "t1", Tipo: DeclPuntos, Raw: "(1,2) (3,4)"}
	if _, err := evalDecl(t, puntos, "/factorizar"); err == nil ||
		!strings.Contains(err.Error(), "expresión") {
		t.Fatalf("puntos deben fallar: %v", err)
	}

	// Decimales no son enteros ni polinomios no constantes.
	dec := Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "12.5", Latex: "12.5"}
	if _, err := evalDecl(t, dec, "/factorizar"); err == nil ||
		!strings.Contains(err.Error(), "entero o un polinomio") {
		t.Fatalf("12.5 debe fallar: %v", err)
	}

	// Una función constante entera sí se factoriza como entero.
	fn := Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "120", Latex: "120"}
	steps, err := evalDecl(t, fn, "/factorizar")
	if err != nil {
		t.Fatalf("función constante 120: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "2³ · 3 · 5") {
		t.Fatalf("factorización entera ausente: %v", steps)
	}
}

func TestEvaluateErrors(t *testing.T) {
	// operación desconocida
	if _, err := evalDecl(t, puntosColineales(), "/voltear t1"); err == nil ||
		!strings.Contains(err.Error(), "no reconocida") {
		t.Fatalf("operación desconocida debe fallar: %v", err)
	}

	// objeto vacío
	if _, err := Evaluate(Decl{}, "/lagrange x", "x"); err == nil {
		t.Fatal("objeto vacío debe fallar")
	}

	// comando vacío
	if _, err := evalDecl(t, puntosColineales(), ""); err == nil {
		t.Fatal("comando vacío debe fallar")
	}
}
