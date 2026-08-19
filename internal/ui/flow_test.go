package ui

import (
	"strings"
	"testing"

	"tuniversalcalc/internal/engine"
)

// TestViewFitsHeight garantiza que la View nunca desborde la terminal:
// si lo hiciera, el terminal hace scroll y corta la parte superior
// (banner incluido). También verifica el ancho de cada línea.
func TestViewFitsHeight(t *testing.T) {
	for _, size := range [][2]int{{100, 28}, {80, 24}, {60, 20}, {120, 34}} {
		m := NewModel()
		m.width, m.height = size[0], size[1]
		m.updateLayout()
		m.exprInput.SetValue("2x+2")
		m.exprValid = true
		lines := strings.Count(m.View(), "\n") + 1
		if lines > m.height {
			t.Errorf("%dx%d: View mide %d líneas, la terminal tiene %d → se corta arriba",
				size[0], size[1], lines, m.height)
		}
		for i, l := range strings.Split(m.View(), "\n") {
			if w := len([]rune(l)); w > size[0] {
				t.Errorf("%dx%d: línea %d mide %d columnas (> %d) → se corta a la derecha",
					size[0], size[1], i+1, w, size[0])
			}
		}
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"2x+2", true},
		{"x^2 + 3x - 5", true},
		{"sin(x) + cos(x)", true},
		{"x^2+", false},
		{"+x", false},
		{"(x^2 + 1", false},
		{"x^2) + 1", false},
		{"3**2", false},
		{"3*-2", true},
		{"-2x^2", true},
		{"x^ -2", false},
		{"x^-2", false},
		{"2x + y/c", true},
		{"3.5x", true},
		{"a@b", false},
	}
	for _, c := range cases {
		err := engine.Validate(c.in)
		if (err == nil) != c.want {
			t.Errorf("Validate(%q) = %v, se esperaba válida=%v", c.in, err, c.want)
		}
	}
}

func TestFlowGuardarYUsarF(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 28
	m.updateLayout()

	// Estado inicial: panel 2 bloqueado
	if m.exprValid {
		t.Fatal("sin expresión, exprValid debe ser false")
	}
	if v := m.View(); !strings.Contains(v, "⛔") {
		t.Fatal("panel 2 debe mostrar el hint de bloqueado")
	}

	// Escribir expresión válida y Enter -> guarda f, foco al panel 2
	m.exprInput.SetValue("2x+2")
	m.refreshExpr()
	if !m.exprValid {
		t.Fatal("2x+2 debe ser válida")
	}
	nm, _ := m.handleEnter()
	m = nm.(Model)
	if m.vars["f"] != "2x+2" {
		t.Fatalf("f no guardada: %q", m.vars["f"])
	}
	if m.focus != FocusCmd {
		t.Fatal("foco debe pasar al panel 2 tras guardar f")
	}

	// Tab desde el panel 2 vuelve al 1; sin expresión, Tab desde 1 se bloquea
	m = m.switchPanel()
	if m.focus != FocusExpr {
		t.Fatal("Tab debe devolver el foco al panel 1")
	}
	m2 := NewModel()
	m2 = m2.switchPanel()
	if m2.focus != FocusExpr {
		t.Fatal("Tab no debe habilitar el panel 2 sin expresión válida")
	}

	// Ejecutar /integral f sobre la variable guardada
	m = m.switchPanel()
	m.cmdInput.SetValue("/integral f")
	nm, cmd := m.handleEnter()
	m = nm.(Model)
	if cmd == nil {
		t.Fatal("debe lanzarse el comando async")
	}
	msg := cmd()
	res, ok := msg.(ResultMsg)
	if !ok || res.Err != nil {
		t.Fatalf("ResultMsg inesperado: %+v", msg)
	}
	var all string
	for _, s := range res.Steps {
		all += " " + s
	}
	if !strings.Contains(all, "∫") || !strings.Contains(all, "2x+2") {
		t.Fatalf("pasos sin ∫ ni 2x+2: %q", all)
	}

	// Romper la expresión estando en panel 2 -> volver al 1
	m.exprInput.SetValue("x^2+")
	m.refreshExpr()
	if m.exprValid {
		t.Fatal("x^2+ debe ser inválida")
	}
	if m.focus != FocusExpr {
		t.Fatal("foco debe volver al panel 1 si la expresión se rompe")
	}
}
