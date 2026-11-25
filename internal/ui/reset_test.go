package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIsClearCommand(t *testing.T) {
	for _, in := range []string{"/limpiar", "/limpiar todo", "LIMPIAR", "/limpiar()", "  /limpiar  "} {
		if !isClearCommand(in) {
			t.Errorf("isClearCommand(%q) = false", in)
		}
	}
	for _, in := range []string{"", "/limpiar algo", "limpiarx", "/potencia"} {
		if isClearCommand(in) {
			t.Errorf("isClearCommand(%q) = true", in)
		}
	}
}

func TestLimpiarPideConfirmacionYCancela(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/expresion x^2")
	m.handleEnter()
	if len(m.objs) == 0 {
		t.Fatal("la expresión no se declaró")
	}

	m.cmdInput.SetValue("/limpiar")
	if cmd := m.handleEnter(); cmd != nil {
		t.Fatal("/limpiar no debe ejecutar nada antes de confirmar")
	}
	if !m.clearConfirm {
		t.Fatal("debe abrirse la confirmación")
	}
	if len(m.objs) == 0 {
		t.Fatal("no debe borrar nada antes de confirmar")
	}
	if !strings.Contains(m.View(), "¿Limpiar todo?") {
		t.Fatal("el modal no muestra el título")
	}

	m.width, m.height = 40, 12
	m.updateLayout()
	checkView(t, &m, [2]int{40, 12}, "confirmar-limpiar")

	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}); !handled {
		t.Fatal("n debe manejarse")
	}
	if m.clearConfirm {
		t.Fatal("n debe cerrar la confirmación")
	}
	if len(m.objs) == 0 {
		t.Fatal("n no debe borrar los objetos")
	}

	// Esc también cancela.
	m.cmdInput.SetValue("/limpiar")
	m.handleEnter()
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc}); !handled {
		t.Fatal("Esc debe manejarse")
	}
	if m.clearConfirm || len(m.objs) == 0 {
		t.Fatal("Esc no debe borrar los objetos")
	}
}

func TestLimpiarConfirmaYBorra(t *testing.T) {
	m := newFlowModel()
	declararPuntosYEjecutarLagrange(t, &m)
	confirmarLimpieza(t, &m)
	comprobarEstadoLimpio(t, &m)
}

func confirmarLimpieza(t *testing.T, m *Model) {
	t.Helper()
	m.cmdInput.SetValue("/limpiar")
	m.handleEnter()
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}); !handled {
		t.Fatal("s debe manejarse")
	}
	if m.clearConfirm {
		t.Fatal("el modal debe cerrarse")
	}
}

func comprobarEstadoLimpio(t *testing.T, m *Model) {
	t.Helper()
	if len(m.objs) != 0 || m.lastObj != "" {
		t.Fatalf("quedaron objetos: %v / %q", m.objs, m.lastObj)
	}
	if len(m.counters) != 0 {
		t.Fatalf("quedaron contadores: %v", m.counters)
	}
	if m.declValid || m.focus != FocusDecl {
		t.Fatalf("estado inválido: declValid=%v focus=%v", m.declValid, m.focus)
	}
	if m.declInput.Value() != "" || m.cmdInput.Value() != "" {
		t.Fatalf("inputs no limpios: %q / %q", m.declInput.Value(), m.cmdInput.Value())
	}
	if len(m.history) != 0 || m.histIdx != -1 {
		t.Fatalf("historial no limpio: %v/%d", m.history, m.histIdx)
	}
	if len(m.undoStack) != 0 || len(m.redoStack) != 0 {
		t.Fatalf("undo/redo no limpios: %d/%d", len(m.undoStack), len(m.redoStack))
	}
	if !strings.Contains(m.renderLog(), "Espacio de trabajo limpiado") {
		t.Fatal("falta el mensaje de limpieza")
	}
}

func TestCtrlLAbreConfirmacionYLimpia(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/expresion x^2")
	m.handleEnter()
	if len(m.objs) == 0 {
		t.Fatal("la expresión no se declaró")
	}

	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlL}); !handled {
		t.Fatal("Ctrl+L debe manejarse")
	}
	if !m.clearConfirm {
		t.Fatal("Ctrl+L debe abrir la confirmación")
	}
	if len(m.objs) == 0 {
		t.Fatal("Ctrl+L no debe borrar antes de confirmar")
	}

	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}); !handled {
		t.Fatal("s debe manejarse")
	}
	if m.clearConfirm {
		t.Fatal("la confirmación debe cerrarse")
	}
	if len(m.objs) != 0 {
		t.Fatal("Ctrl+L + s debe limpiar los objetos")
	}
}

func TestCtrlLDesdePanel2Bloqueado(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 28
	m.updateLayout()
	m.focus = FocusCmd
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlL}); !handled {
		t.Fatal("Ctrl+L debe manejarse con el panel 2 bloqueado")
	}
	if !m.clearConfirm {
		t.Fatal("Ctrl+L debe abrir la confirmación")
	}
}

func TestLimpiarDesdePanel2Bloqueado(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 28
	m.updateLayout()
	m.focus = FocusCmd
	m.cmdInput.SetValue("/limpiar")
	m.handleEnter()
	if !m.clearConfirm {
		t.Fatal("el comando debe funcionar aunque el panel 2 esté bloqueado")
	}
}

func TestLimpiarGuardaSesionVacia(t *testing.T) {
	withTempSessions(t)

	m := newFlowModel()
	m.sessionEnabled = true
	m.sessionName = "trabajo"
	m.declInput.SetValue("/expresion x^2")
	m.handleEnter()
	m.saveSessionNow()

	m.cmdInput.SetValue("/limpiar")
	m.handleEnter()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if m.sessionName != "trabajo" {
		t.Fatalf("la sesión cambió a %q", m.sessionName)
	}
	if !SessionExists("trabajo") {
		t.Fatal("la sesión no existe tras limpiar")
	}
	data, err := LoadSession("trabajo")
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if len(data.Objects) != 0 {
		t.Fatalf("la sesión guardada conserva objetos: %+v", data.Objects)
	}
	if data.DeclInput != "" || data.CmdInput != "" {
		t.Fatalf("la sesión guardada conserva inputs: %q / %q", data.DeclInput, data.CmdInput)
	}
}
