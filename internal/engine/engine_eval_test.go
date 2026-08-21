package engine

import (
	"strings"
	"testing"
)

func evalDecl(t *testing.T, decl Decl, cmd string) ([]string, error) {
	t.Helper()
	return Evaluate(decl, cmd, "x")
}

func TestEvaluateDerive(t *testing.T) {
	steps, err := evalDecl(t, Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "x^2+3x-5", Latex: "x^{2} + 3x - 5"}, "/derivar f1")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	var all string
	for _, s := range steps {
		all += s + "|"
	}
	// cabecera con el objeto y el resultado real simplificado
	if !strings.Contains(all, "f1") || !strings.Contains(all, "x^{2} + 3x - 5") {
		t.Fatalf("cabecera del objeto faltante: %s", all)
	}
	if !strings.Contains(all, "2*x + 3") {
		t.Fatalf("derivada esperada 2x+3: %s", all)
	}
}

func TestEvaluateResolve(t *testing.T) {
	steps, err := evalDecl(t, Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "2x+2", Latex: "2x + 2"}, "/resolver f1 x")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "x = -1") {
		t.Fatalf("solución esperada x = -1: %v", steps)
	}
}

func TestEvaluateIntegral(t *testing.T) {
	steps, err := evalDecl(t, Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "x^2+1", Latex: "x^{2} + 1"}, "/integral p1")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	var all string
	for _, s := range steps {
		all += s + "|"
	}
	if !strings.Contains(all, "x^3/3") && !strings.Contains(all, "1/3*x^3") {
		t.Fatalf("antiderivada esperada: %s", all)
	}
	if !strings.Contains(all, "C") {
		t.Fatalf("la integral debe llevar + C: %s", all)
	}
}

func TestEvaluateSimplify(t *testing.T) {
	steps, err := evalDecl(t, Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "2x+3x", Latex: "2x + 3x"}, "/simplificar p1")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "5*x") {
		t.Fatalf("simplificación esperada 5x: %v", steps)
	}
}

func TestEvaluateLimit(t *testing.T) {
	steps, err := evalDecl(t, Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "x^2-1", Latex: "x^{2} - 1"}, "/limite f1")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	// por defecto x→0 y el límite de x²−1 es −1
	if !strings.Contains(strings.Join(steps, "|"), "-1") {
		t.Fatalf("límite esperado -1: %v", steps)
	}
}

func TestEvaluateMatrixStub(t *testing.T) {
	steps, err := evalDecl(t, Decl{Nombre: "p1", Tipo: DeclExpresion, Raw: "x+1", Latex: "x + 1"}, "/matriz p1")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "no soporta matrices") {
		t.Fatalf("matriz debe ser stub explícito: %v", steps)
	}
}

func TestEvaluateErrors(t *testing.T) {
	// operación desconocida
	if _, err := evalDecl(t, Decl{Nombre: "e1", Tipo: DeclEcuacion, Raw: "x=2"}, "/voltear e1"); err == nil ||
		!strings.Contains(err.Error(), "no reconocida") {
		t.Fatalf("operación desconocida debe fallar: %v", err)
	}

	// sintaxis inválida con posición
	if _, err := evalDecl(t, Decl{Nombre: "e1", Tipo: DeclEcuacion, Raw: "x^2+", Latex: "x^{2} +"}, "/resolver e1 x"); err == nil {
		t.Fatal("sintaxis inválida debe fallar")
	}

	// integral sobre una ecuación
	if _, err := evalDecl(t, Decl{Nombre: "e1", Tipo: DeclEcuacion, Raw: "x=2", Latex: "x = 2"}, "/integral e1"); err == nil {
		t.Fatal("integral sobre ecuación debe fallar")
	}

	// derivada no soportada
	if _, err := evalDecl(t, Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "x^x", Latex: "x^x"}, "/derivar f1"); err == nil {
		t.Fatal("derivada de x^x debe fallar")
	}
}

func TestEvaluateEmptyObjeto(t *testing.T) {
	if _, err := Evaluate(Decl{}, "/resolver x", "x"); err == nil {
		t.Fatal("objeto vacío debe fallar")
	}
	if _, err := evalDecl(t, Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "x+1", Latex: "x + 1"}, ""); err == nil {
		t.Fatal("comando vacío debe fallar")
	}
}

func TestEvaluateVarNameWeird(t *testing.T) {
	// /limite f1 2 → el "2" no es variable: se cae a x por defecto
	steps, err := Evaluate(Decl{Nombre: "f1", Tipo: DeclFuncion, Raw: "x^2-1", Latex: "x^{2} - 1"}, "/limite f1 x 2", "x")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !strings.Contains(strings.Join(steps, "|"), "3") {
		t.Fatalf("límite x→2 de x²−1 = 3: %v", steps)
	}
}
