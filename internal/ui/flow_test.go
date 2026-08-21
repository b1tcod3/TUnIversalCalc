package ui

import (
	"regexp"
	"strings"
	"testing"

	"tuniversalcalc/internal/engine"
)

// ansiRe quita los códigos de escape CSI (colores, cursor) para medir el
// ancho real de una línea de la View.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// TestViewFitsHeight garantiza que la View nunca desborde la terminal:
// si lo hiciera, el terminal hace scroll y corta la parte superior
// (banner incluido). También verifica el ancho de cada línea, ignorando
// los códigos ANSI. Se comprueba con la paleta abierta en ambos paneles.
func TestViewFitsHeight(t *testing.T) {
	for _, size := range [][2]int{{100, 28}, {80, 24}, {60, 20}, {120, 34}} {
		m := NewModel()
		m.width, m.height = size[0], size[1]
		m.updateLayout()
		m.declInput.SetValue("/funcion 2x+2")
		m.declValid = true
		checkView(t, m, size, "normal")

		// Paleta abierta en el panel de declaración ("/" lista todo)
		m.declInput.SetValue("/")
		m.refreshPalette()
		checkView(t, m, size, "paleta-declaracion")

		// Paleta abierta en el panel de comandos
		m.focus = FocusCmd
		m.cmdInput.SetValue("/")
		m.refreshPalette()
		checkView(t, m, size, "paleta-comando")

		// Sin matches (fragmento roto): paleta cerrada pero sigue el layout
		m.cmdInput.SetValue("/rslv x")
		m.refreshPalette()
		checkView(t, m, size, "sin-match")
	}
}

func checkView(t *testing.T, m Model, size [2]int, label string) {
	t.Helper()
	lines := strings.Count(m.View(), "\n") + 1
	if lines > m.height {
		t.Errorf("%dx%d [%s]: View mide %d líneas, la terminal tiene %d → se corta arriba",
			size[0], size[1], label, lines, m.height)
	}
	for i, l := range strings.Split(m.View(), "\n") {
		if w := len([]rune(stripANSI(l))); w > size[0] {
			t.Errorf("%dx%d [%s]: línea %d mide %d columnas (> %d) → se corta a la derecha",
				size[0], size[1], label, i+1, w, size[0])
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
		{"x^ -2", true}, // exponente negativo: ahora el parser lo soporta
		{"x^-2", true},  // exponente negativo: ahora el parser lo soporta
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

func TestFlowDeclararYOperar(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 28
	m.updateLayout()

	// Estado inicial: panel 2 bloqueado
	if m.declValid {
		t.Fatal("sin objetos, declValid debe ser false")
	}
	if v := m.View(); !strings.Contains(v, "⛔") {
		t.Fatal("panel 2 debe mostrar el hint de bloqueado")
	}

	// Escribir /ecuacion 2x+2=0 y Enter -> auto-e1 registrada, foco al panel 2
	m.declInput.SetValue("/ecuacion 2x+2=0")
	m.handleEnter()
	if d, ok := m.objs["e1"]; !ok || d.Tipo != engine.DeclEcuacion || d.Latex != "2x + 2 = 0" {
		t.Fatalf("e1 no registrada correctamente: %+v", m.objs)
	}
	if m.lastObj != "e1" || !m.declValid {
		t.Fatalf("lastObj=%q declValid=%v", m.lastObj, m.declValid)
	}
	if m.focus != FocusCmd {
		t.Fatal("foco debe pasar al panel 2 tras declarar")
	}

	// Tab desde el panel 2 vuelve al 1; sin objetos, Tab desde 1 se bloquea
	m.switchPanel()
	if m.focus != FocusDecl {
		t.Fatal("Tab debe devolver el foco al panel 1")
	}
	m2 := NewModel()
	m2.switchPanel()
	if m2.focus != FocusDecl {
		t.Fatal("Tab no debe habilitar el panel 2 sin objetos declarados")
	}

	// Segunda declaración: /funcion 2x+2 -> auto-f1, es el último objeto
	m.declInput.SetValue("/funcion 2x+2")
	m.handleEnter()
	if d, ok := m.objs["f1"]; !ok || d.Tipo != engine.DeclFuncion {
		t.Fatalf("f1 no registrada: %+v", m.objs)
	}
	if m.lastObj != "f1" {
		t.Fatalf("lastObj debe ser f1, es %q", m.lastObj)
	}

	// /resolver x -> opera sobre el último objeto (f1)
	m.cmdInput.SetValue("/resolver x")
	cmd := m.handleEnter()
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
	if !strings.Contains(all, "f1") || !strings.Contains(all, "2x + 2") {
		t.Fatalf("pasos sin objeto f1 ni LaTeX: %q", all)
	}

	// /resolver e1 x -> objeto explícito con variable
	m.cmdInput.SetValue("/resolver e1 x")
	cmd = m.handleEnter()
	if cmd == nil {
		t.Fatal("debe lanzarse el comando async")
	}
	res = cmd().(ResultMsg)
	if !strings.Contains(strings.Join(res.Steps, " "), "e1") {
		t.Fatalf("pasos sin objeto e1: %q", res.Steps)
	}

	// Error inteligente: objeto inexistente
	m.cmdInput.SetValue("/resolver g1")
	cmd = m.handleEnter()
	if cmd != nil {
		t.Fatal("no debe lanzarse comando con objeto inexistente")
	}
	last := m.outputLog[len(m.outputLog)-1]
	if last.Level != LevelError || !strings.Contains(last.Content, "g1") || !strings.Contains(last.Content, "e1") {
		t.Fatalf("error esperado citando g1 y objetos: %+v", last)
	}

	// Error del motor: operación no reconocida
	m.cmdInput.SetValue("/voltear e1")
	cmd = m.handleEnter()
	if cmd == nil {
		t.Fatal("debe lanzarse el comando async")
	}
	msg = cmd()
	res, ok = msg.(ResultMsg)
	if !ok || res.Err == nil {
		t.Fatalf("se esperaba error de operación: %+v", msg)
	}

	// Plantilla de declaración en el panel 1: "/ecu" -> "/ecuacion ?"
	m.switchPanel()
	m.declInput.SetValue("/ecu")
	m.handleEnter()
	if m.declInput.Value() != "/ecuacion " {
		t.Fatalf("plantilla no insertada: %q", m.declInput.Value())
	}

	// Declaración inválida: error en log y sin bloqueo posterior
	m.declInput.SetValue("/funcion x=2")
	m.handleEnter()
	last = m.outputLog[len(m.outputLog)-1]
	if last.Level != LevelError || !strings.Contains(last.Content, "no lleva") {
		t.Fatalf("error esperado por '=' en función: %+v", last)
	}
}
