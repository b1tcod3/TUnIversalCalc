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
		checkView(t, &m, size, "normal")

		// Paleta abierta en el panel de declaración ("/" lista todo)
		m.declInput.SetValue("/")
		m.refreshPalette()
		checkView(t, &m, size, "paleta-declaracion")

		// Paleta abierta en el panel de comandos
		m.focus = FocusCmd
		m.cmdInput.SetValue("/")
		m.refreshPalette()
		checkView(t, &m, size, "paleta-comando")

		// Sin matches (fragmento roto): paleta cerrada pero sigue el layout
		m.cmdInput.SetValue("/rslv x")
		m.refreshPalette()
		checkView(t, &m, size, "sin-match")
	}
}

func checkView(t *testing.T, m *Model, size [2]int, label string) {
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

// newFlowModel devuelve un modelo con layout calculado para las pruebas
// de flujo (terminal 100x28).
func newFlowModel() Model {
	m := NewModel()
	m.width, m.height = 100, 28
	m.updateLayout()
	return m
}

// declararEcuacion registra e1 = 2x + 2 = 0 vía el panel 1.
func declararEcuacion(t *testing.T, m *Model) {
	t.Helper()
	m.declInput.SetValue("/ecuacion 2x+2=0")
	m.handleEnter()
	d, ok := m.objs["e1"]
	if !ok || d.Tipo != engine.DeclEcuacion || d.Latex != "2x + 2 = 0" {
		t.Fatalf("e1 no registrada correctamente: %+v", m.objs)
	}
}

// declararEcuacionYFuncion registra e1 y f1 y deja el foco en el panel 2.
func declararEcuacionYFuncion(t *testing.T, m *Model) {
	t.Helper()
	declararEcuacion(t, m)
	m.switchPanel() // volver al panel 1 para declarar la función
	m.declInput.SetValue("/funcion 2x+2")
	m.handleEnter()
	if _, ok := m.objs["f1"]; !ok {
		t.Fatalf("f1 no registrada: %+v", m.objs)
	}
}

func TestEstadoInicialPanel2Bloqueado(t *testing.T) {
	m := newFlowModel()

	if m.declValid {
		t.Fatal("sin objetos, declValid debe ser false")
	}
	if v := m.View(); !strings.Contains(v, "⛔") {
		t.Fatal("panel 2 debe mostrar el hint de bloqueado")
	}
}

func TestDeclararEcuacionAutoNombre(t *testing.T) {
	m := newFlowModel()
	declararEcuacion(t, &m)

	if m.lastObj != "e1" || !m.declValid {
		t.Fatalf("lastObj=%q declValid=%v", m.lastObj, m.declValid)
	}
	if m.focus != FocusCmd {
		t.Fatal("foco debe pasar al panel 2 tras declarar")
	}
}

func TestTabNavegacionPaneles(t *testing.T) {
	m := newFlowModel()
	declararEcuacion(t, &m)

	// Tab desde el panel 2 vuelve al 1
	m.switchPanel()
	if m.focus != FocusDecl {
		t.Fatal("Tab debe devolver el foco al panel 1")
	}

	// Sin objetos, Tab desde el panel 1 se bloquea
	m2 := newFlowModel()
	m2.switchPanel()
	if m2.focus != FocusDecl {
		t.Fatal("Tab no debe habilitar el panel 2 sin objetos declarados")
	}
}

func TestDeclararFuncionAutoNombre(t *testing.T) {
	m := newFlowModel()
	declararEcuacion(t, &m)
	m.switchPanel()

	// Segunda declaración: /funcion 2x+2 -> auto-f1, es el último objeto
	m.declInput.SetValue("/funcion 2x+2")
	m.handleEnter()
	if _, ok := m.objs["f1"]; !ok {
		t.Fatalf("f1 no registrada: %+v", m.objs)
	}
	if m.lastObj != "f1" {
		t.Fatalf("lastObj debe ser f1, es %q", m.lastObj)
	}
}

func TestResolverUltimoObjeto(t *testing.T) {
	m := newFlowModel()
	declararEcuacionYFuncion(t, &m)

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
	all := strings.Join(res.Steps, " ")
	if !strings.Contains(all, "f1") || !strings.Contains(all, "2x + 2") {
		t.Fatalf("pasos sin objeto f1 ni LaTeX: %q", all)
	}
}

func TestResolverObjetoExplicito(t *testing.T) {
	m := newFlowModel()
	declararEcuacionYFuncion(t, &m)

	// /resolver e1 x -> objeto explícito con variable
	m.cmdInput.SetValue("/resolver e1 x")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("debe lanzarse el comando async")
	}
	res := cmd().(ResultMsg)
	if !strings.Contains(strings.Join(res.Steps, " "), "e1") {
		t.Fatalf("pasos sin objeto e1: %q", res.Steps)
	}
}

func TestResolverErrores(t *testing.T) {
	m := newFlowModel()
	declararEcuacionYFuncion(t, &m)

	// Error inteligente: objeto inexistente
	m.cmdInput.SetValue("/resolver g1")
	if cmd := m.handleEnter(); cmd != nil {
		t.Fatal("no debe lanzarse comando con objeto inexistente")
	}
	last := m.outputLog[len(m.outputLog)-1]
	if last.Level != LevelError || !strings.Contains(last.Content, "g1") || !strings.Contains(last.Content, "e1") {
		t.Fatalf("error esperado citando g1 y objetos: %+v", last)
	}

	// Error del motor: operación no reconocida
	m.cmdInput.SetValue("/voltear e1")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("debe lanzarse el comando async")
	}
	msg := cmd()
	res, ok := msg.(ResultMsg)
	if !ok || res.Err == nil {
		t.Fatalf("se esperaba error de operación: %+v", msg)
	}
}

func TestPlantillaYDeclaracionInvalida(t *testing.T) {
	m := newFlowModel()
	declararEcuacionYFuncion(t, &m)

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
	last := m.outputLog[len(m.outputLog)-1]
	if last.Level != LevelError || !strings.Contains(last.Content, "no lleva") {
		t.Fatalf("error esperado por '=' en función: %+v", last)
	}
}
