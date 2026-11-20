package ui

import (
	"strings"
	"testing"
	"tuniversalcalc/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

// TestViewFitsHeight garantiza que la View nunca desborde la terminal:
// si lo hiciera, el terminal hace scroll y corta la parte superior
// (banner incluido). También verifica el ancho de cada línea, ignorando
// los códigos ANSI. Se comprueba con la paleta abierta en ambos paneles.
func TestViewFitsHeight(t *testing.T) {
	for _, size := range [][2]int{{40, 12}, {50, 16}, {60, 20}, {80, 24}, {100, 28}, {120, 40}} {
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
		if w := visualWidth(l); w > size[0] {
			t.Errorf("%dx%d [%s]: línea %d mide %d columnas (> %d) → se corta a la derecha",
				size[0], size[1], label, i+1, w, size[0])
		}
	}
}

// TestViewLlenaLaTerminal comprueba que la cuadrícula ocupa todo el ancho y
// que la vista usa exactamente las filas de la terminal.
func TestViewLlenaLaTerminal(t *testing.T) {
	for _, size := range [][2]int{{40, 12}, {60, 20}, {80, 24}, {100, 28}, {120, 40}} {
		m := NewModel()
		m.width, m.height = size[0], size[1]
		m.updateLayout()
		m.declInput.SetValue("/funcion 2x+2")
		m.declValid = true

		lines := strings.Split(m.View(), "\n")
		if len(lines) != size[1] {
			t.Errorf("%dx%d: View mide %d líneas, se esperaban %d", size[0], size[1], len(lines), size[1])
		}
		maxW := 0
		for _, l := range lines {
			maxW = max(maxW, visualWidth(l))
		}
		if maxW != size[0] {
			t.Errorf("%dx%d: ancho máximo %d, se esperaba llenar %d", size[0], size[1], maxW, size[0])
		}
	}
}

// TestViewConCaracteresAnchos garantiza que la medición ANSI-aware no se
// desborda con emoji, CJK ni texto largo en la salida.
func TestViewConCaracteresAnchos(t *testing.T) {
	m := NewModel()
	m.width, m.height = 40, 12
	m.updateLayout()
	m.declValid = true
	m.addLog(LevelInfo, "😀漢字🙂‍↔️ "+strings.Repeat("ancho ", 20))
	checkView(t, &m, [2]int{40, 12}, "emoji")
}

// TestViewConSesionLarga comprueba que la barra de estado se trunca aunque el
// nombre de sesión no quepa.
func TestViewConSesionLarga(t *testing.T) {
	m := NewModel()
	m.width, m.height = 40, 12
	m.updateLayout()
	m.declValid = true
	m.sessionEnabled = true
	m.sessionName = strings.Repeat("s", 80)
	m.dirty = true
	checkView(t, &m, [2]int{40, 12}, "sesion-larga")
}

// TestViewNoMutaDimensiones protege el ciclo Elm: View no debe cambiar el
// tamaño del modelo al abrir la ayuda o el selector en terminales pequeñas.
func TestViewNoMutaDimensiones(t *testing.T) {
	m := NewModel()
	m.width, m.height = 30, 10
	m.updateLayout()

	m.toggleHelp()
	_ = m.View()
	if m.width != 30 || m.height != 10 {
		t.Fatalf("View mutó las dimensiones con la ayuda abierta: %dx%d", m.width, m.height)
	}

	m.toggleHelp()
	m.picker.open = true
	_ = m.View()
	if m.width != 30 || m.height != 10 {
		t.Fatalf("View mutó las dimensiones con el selector abierto: %dx%d", m.width, m.height)
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

// mustResultMsg ejecuta el comando y exige que produzca un ResultMsg.
func mustResultMsg(t *testing.T, cmd tea.Cmd) ResultMsg {
	t.Helper()
	msg, ok := cmd().(ResultMsg)
	if !ok {
		t.Fatalf("el comando no produjo ResultMsg: %T", msg)
	}
	return msg
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

// declararPuntos registra t1 = (1,2) (3,4) (5,6) via el panel 1.
func declararPuntos(t *testing.T, m *Model) {
	t.Helper()
	m.switchPanel() // volver al panel 1
	m.declInput.SetValue("/puntos (1,2) (3,4) (5,6)")
	m.handleEnter()
	d, ok := m.objs["t1"]
	if !ok || d.Tipo != engine.DeclPuntos {
		t.Fatalf("t1 no registrada correctamente: %+v", m.objs)
	}
	if len(d.Points) != 3 || d.Points[0].X != 1 || d.Points[2].Y != 6 {
		t.Fatalf("puntos mal guardados: %+v", d.Points)
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

func TestLagrangeUltimoObjeto(t *testing.T) {
	m := newFlowModel()
	declararEcuacion(t, &m)
	declararPuntos(t, &m) // t1 pasa a ser el último objeto

	// /lagrange x -> el token de un carácter es variable: opera sobre el último (t1)
	m.cmdInput.SetValue("/lagrange x")
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
	if !strings.Contains(all, "t1") || !strings.Contains(all, "P(x) = x + 1") {
		t.Fatalf("pasos sin objeto t1 ni polinomio: %q", all)
	}
}

func TestNewtonObjetoExplicito(t *testing.T) {
	m := newFlowModel()
	declararEcuacion(t, &m)
	declararPuntos(t, &m)

	// /newton t1 -> objeto explícito
	m.cmdInput.SetValue("/newton t1")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("debe lanzarse el comando async")
	}
	res := mustResultMsg(t, cmd)
	if res.Err != nil {
		t.Fatalf("newton falló: %v", res.Err)
	}
	if !strings.Contains(strings.Join(res.Steps, " "), "t1") {
		t.Fatalf("pasos sin objeto t1: %q", res.Steps)
	}
}

func TestPotenciaYRadicalFlujo(t *testing.T) {
	m := newFlowModel()

	m.declInput.SetValue("/expresion (3a^3b^3)(4ab^2)^2")
	m.handleEnter()
	m.cmdInput.SetValue("/potencia")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("potencia debe lanzar un comando")
	}
	res := mustResultMsg(t, cmd)
	if res.Err != nil {
		t.Fatalf("potencia falló: %v", res.Err)
	}
	if !strings.Contains(strings.Join(res.Steps, " "), "48*a^5*b^7") {
		t.Fatalf("resultado de potencia ausente: %v", res.Steps)
	}

	m.switchPanel()
	m.declInput.SetValue("/expresion raiz(3, 16)")
	m.handleEnter()
	if d, ok := m.objs["p2"]; !ok || !strings.Contains(d.Latex, `\sqrt[3]{16}`) || d.Pretty != "∛16" {
		t.Fatalf("raiz no renderizada correctamente: %+v", d)
	}
	m.cmdInput.SetValue("/radical")
	cmd = m.handleEnter()
	if cmd == nil {
		t.Fatal("radical debe lanzar un comando")
	}
	res = mustResultMsg(t, cmd)
	if res.Err != nil {
		t.Fatalf("radical falló: %v", res.Err)
	}
	if !strings.Contains(strings.Join(res.Steps, " "), "2*raiz(3, 2)") {
		t.Fatalf("resultado de radical ausente: %v", res.Steps)
	}

	m.switchPanel()
	m.declInput.SetValue("/expresion alpha^2+2alpha")
	m.handleEnter()
	m.cmdInput.SetValue("/potencia")
	cmd = m.handleEnter()
	if cmd == nil {
		t.Fatal("potencia con variable larga debe lanzar un comando")
	}
	res = mustResultMsg(t, cmd)
	if res.Err != nil {
		t.Fatalf("potencia con alpha falló: %v", res.Err)
	}
	if !strings.Contains(strings.Join(res.Steps, " "), "alpha^2 + 2*alpha") {
		t.Fatalf("resultado con variable larga ausente: %v", res.Steps)
	}
}

func TestDesarrollarFlujo(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/expresion (a+b)^2")
	m.handleEnter()
	m.cmdInput.SetValue("/desarrollar")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("desarrollar debe lanzar un comando")
	}
	res := mustResultMsg(t, cmd)
	if res.Err != nil {
		t.Fatalf("desarrollar falló: %v", res.Err)
	}
	all := strings.Join(res.Steps, " ")
	for _, frag := range []string{"cuadrado de un binomio", "a^2 + 2*a*b + b^2"} {
		if !strings.Contains(all, frag) {
			t.Fatalf("falta %q en: %v", frag, res.Steps)
		}
	}
}

func TestVariablesDeVariasLetras(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/funcion alpha^2+2alpha")
	m.handleEnter()
	m.cmdInput.SetValue("/lagrange ALPHA")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("lagrange alpha debe resolverse y llegar al motor")
	}
	res := mustResultMsg(t, cmd)
	if res.Err == nil || !strings.Contains(res.Err.Error(), "puntos") {
		t.Fatalf("se esperaba el error de lagrange sobre función: %+v", res)
	}
}

func TestResolverErrores(t *testing.T) {
	m := newFlowModel()
	declararEcuacionYFuncion(t, &m)

	// Error inteligente: objeto inexistente
	m.cmdInput.SetValue("/lagrange g1")
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

func TestExpresionEnteraAutoFactoriza(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/expresion 120")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("una expresión entera debe lanzar la factorización automática")
	}
	res, ok := cmd().(ResultMsg)
	if !ok || res.Err != nil {
		t.Fatalf("ResultMsg inesperado: %+v", res)
	}
	all := strings.Join(res.Steps, " ")
	if !strings.Contains(all, "Factorización: 2³ · 3 · 5") {
		t.Fatalf("pasos sin factorización: %q", all)
	}
	m.applyResult(res)
	if !strings.Contains(m.plainOutput(), "2³ · 3 · 5") {
		t.Fatal("la salida no contiene la factorización")
	}
	if _, ok := m.objs["p1"]; !ok {
		t.Fatalf("p1 no declarada: %+v", m.objs)
	}
}

func TestEnteroSueltoAutoFactoriza(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("120")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("un entero suelto debe lanzar la factorización automática")
	}
	res, ok := cmd().(ResultMsg)
	if !ok || res.Err != nil {
		t.Fatalf("ResultMsg inesperado: %+v", res)
	}
	if !strings.Contains(strings.Join(res.Steps, " "), "2³ · 3 · 5") {
		t.Fatalf("pasos sin factorización: %q", res.Steps)
	}
	if d, ok := m.objs["p1"]; !ok || d.Raw != "120" {
		t.Fatalf("p1 no declarada como entero: %+v", m.objs)
	}
}

func TestExpresionDecimalNoAutoFactoriza(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/expresion 12.5")
	if cmd := m.handleEnter(); cmd != nil {
		t.Fatal("12.5 no debe disparar factorización automática")
	}
	if _, ok := m.objs["p1"]; !ok {
		t.Fatalf("p1 no declarada: %+v", m.objs)
	}
}

func TestFactorizarPolinomioFlujo(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/expresion x^2-4")
	m.handleEnter()
	m.cmdInput.SetValue("/factorizar p1")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("debe lanzarse la factorización del polinomio")
	}
	res := mustResultMsg(t, cmd)
	if res.Err != nil {
		t.Fatalf("factorizar polinomio falló: %v", res.Err)
	}
	if !strings.Contains(strings.Join(res.Steps, " "), "(x - 2)*(x + 2)") {
		t.Fatalf("factorización ausente: %v", res.Steps)
	}
}

func TestFactorizarRacionalFlujo(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/expresion (x^2+3x+2)/(x^2-x-2)")
	m.handleEnter()
	m.cmdInput.SetValue("/factorizar p1")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("debe lanzarse la factorización de la expresión racional")
	}
	res := mustResultMsg(t, cmd)
	if res.Err != nil {
		t.Fatalf("factorizar expresión racional falló: %v", res.Err)
	}
	all := strings.Join(res.Steps, " ")
	if !strings.Contains(all, "Factorización: (x + 2)/(x - 2)") {
		t.Fatalf("resultado racional ausente: %v", res.Steps)
	}
	if !strings.Contains(all, "Factorización del denominador:") {
		t.Fatalf("factorización del denominador ausente: %v", res.Steps)
	}
}

func TestFactorizarLiteral(t *testing.T) {
	m := newFlowModel()
	declararEcuacionYFuncion(t, &m)

	m.cmdInput.SetValue("/factorizar 120")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("debe lanzarse la factorización")
	}
	res, ok := cmd().(ResultMsg)
	if !ok || res.Err != nil {
		t.Fatalf("ResultMsg inesperado: %+v", res)
	}
	all := strings.Join(res.Steps, " ")
	if !strings.Contains(all, "Factorización: 2³ · 3 · 5") {
		t.Fatalf("pasos sin factorización: %q", all)
	}
}
