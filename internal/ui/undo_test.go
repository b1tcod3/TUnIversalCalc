package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUndoRedoDeclaracion(t *testing.T) {
	m := newFlowModel()

	m.declInput.SetValue("/funcion 2x+2")
	m.handleEnter()
	if len(m.objs) != 1 || !m.declValid {
		t.Fatalf("declaración no registrada: %+v", m.objs)
	}
	if !m.undo() {
		t.Fatal("undo debe tener un snapshot")
	}
	if len(m.objs) != 0 || m.declValid {
		t.Fatalf("undo no eliminó la declaración: %+v", m.objs)
	}
	if m.declInput.Value() != "/funcion 2x+2" {
		t.Fatalf("undo no restauró el texto: %q", m.declInput.Value())
	}
	if !m.redo() {
		t.Fatal("redo debe tener un snapshot")
	}
	if len(m.objs) != 1 || m.lastObj != "f1" {
		t.Fatalf("redo no restauró la declaración: %+v", m.objs)
	}
}

func TestUndoComandoNoRestauraResultado(t *testing.T) {
	m := newFlowModel()
	declararPuntos(t, &m)

	m.cmdInput.SetValue("/lagrange t1")
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("se esperaba comando")
	}
	m.applyResult(mustResultMsg(t, cmd))
	if len(m.history) != 1 {
		t.Fatalf("historial = %v", m.history)
	}

	if !m.undo() {
		t.Fatal("undo debe tener snapshot del comando")
	}
	if len(m.history) != 0 {
		t.Fatalf("undo no limpió el historial: %v", m.history)
	}
	if m.lastObj != "t1" || len(m.objs) != 1 {
		t.Fatalf("undo no debe borrar objetos: %+v", m.objs)
	}
	if m.plainOutput() == "" {
		t.Fatal("el log anterior debe seguir visible")
	}
}

func TestUndoLimite(t *testing.T) {
	m := newFlowModel()
	for range maxUndoLevels + 10 {
		m.pushUndo()
	}
	if len(m.undoStack) != maxUndoLevels {
		t.Fatalf("undoStack = %d, se esperaba %d", len(m.undoStack), maxUndoLevels)
	}
}

func TestAtajosUndoRedo(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/funcion x")
	m.handleEnter()

	// Ctrl+Z
	if cmd, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlZ}); !handled || cmd != nil {
		t.Fatalf("Ctrl+Z no manejado: cmd=%v handled=%v", cmd, handled)
	}
	if len(m.objs) != 0 {
		t.Fatal("Ctrl+Z no deshizo")
	}
	// Ctrl+Y
	if cmd, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlY}); !handled || cmd != nil {
		t.Fatalf("Ctrl+Y no manejado: cmd=%v handled=%v", cmd, handled)
	}
	if len(m.objs) != 1 {
		t.Fatal("Ctrl+Y no rehízo")
	}
}
