package ui

import (
	"strings"
	"testing"
	"time"
	"tuniversalcalc/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

func seedSessions(t *testing.T) {
	t.Helper()
	withTempSessions(t)
	for _, name := range []string{"alfa", "beta"} {
		data := SessionData{
			Objects:  map[string]engine.Decl{},
			Counters: map[engine.DeclType]int{},
		}
		if err := SaveSession(name, data); err != nil {
			t.Fatalf("SaveSession(%q): %v", name, err)
		}
	}
}

func TestSessionPickerAbreYFiltra(t *testing.T) {
	seedSessions(t)
	m := newFlowModel()
	m.sessionName = "alfa"
	m.sessionEnabled = true

	m.openSessionPicker()
	if !m.picker.open {
		t.Fatal("el selector debe abrirse")
	}
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlO}); !handled {
		t.Fatal("Ctrl+O debe manejarse con el selector abierto")
	}
	if m.picker.open {
		t.Fatal("Ctrl+O debe cerrar el selector")
	}

	m.openSessionPicker()
	m.picker.query.SetValue("bet")
	list := m.picker.filtered()
	if len(list) != 1 || list[0].Name != "beta" {
		t.Fatalf("filtro incorrecto: %+v", list)
	}
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter}); !handled {
		t.Fatal("Enter debe manejarse")
	}
	if m.SessionName() != "beta" {
		t.Fatalf("sesión activa = %q, se esperaba beta", m.SessionName())
	}
	if m.picker.open {
		t.Fatal("el selector debe cerrarse al abrir una sesión")
	}
}

func TestSessionPickerViewCabe(t *testing.T) {
	seedSessions(t)
	m := newFlowModel()
	m.sessionName = "alfa"
	m.sessionEnabled = true
	m.openSessionPicker()
	checkView(t, &m, [2]int{100, 28}, "selector-sesiones")
	if !strings.Contains(m.View(), "Sesiones") {
		t.Fatal("la vista no contiene el título del selector")
	}

	small := NewModel()
	small.width, small.height = 40, 12
	small.sessionName = "alfa"
	small.sessionEnabled = true
	small.picker.open = true
	small.picker.items = []SessionMeta{{Name: "alfa"}}
	checkView(t, &small, [2]int{40, 12}, "selector-sesiones-pequeño")
}

func TestSessionPickerModalesCaben(t *testing.T) {
	m := NewModel()
	m.width, m.height = 40, 12
	m.updateLayout()
	m.picker.open = true
	m.picker.items = []SessionMeta{{
		Name:     "una-sesion-con-nombre-larguisimo",
		Modified: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}}
	m.picker.mode = pickerRename
	m.picker.nameInput.SetValue("nuevo-nombre-igual-de-largo")
	checkView(t, &m, [2]int{40, 12}, "selector-renombrar")

	m.picker.mode = pickerConfirmDelete
	checkView(t, &m, [2]int{40, 12}, "selector-confirmar-borrado")
}

func TestSessionPickerRenombrar(t *testing.T) {
	seedSessions(t)
	m := newFlowModel()
	m.sessionName = "alfa"
	m.sessionEnabled = true
	m.openSessionPicker()
	m.picker.query.SetValue("alfa")
	m.picker.clamp()

	if cmd := m.startRename(); cmd == nil {
		t.Fatal("startRename debe devolver el blink del input")
	}
	m.picker.nameInput.SetValue("gamma")
	m.acceptRename()

	if m.sessionName != "gamma" {
		t.Fatalf("sesión activa = %q", m.sessionName)
	}
	if !SessionExists("gamma") || SessionExists("alfa") {
		t.Fatal("el archivo no se renombró")
	}
	if last, ok := LastSessionName(); !ok || last != "gamma" {
		t.Fatalf("última sesión = %q, %v", last, ok)
	}
}

func TestSessionPickerDuplicar(t *testing.T) {
	seedSessions(t)
	m := newFlowModel()
	m.sessionName = "alfa"
	m.sessionEnabled = true
	m.openSessionPicker()
	m.picker.query.SetValue("alfa")
	m.picker.clamp()
	m.duplicateSelected()

	if !SessionExists("alfa copia") {
		t.Fatal("la sesión duplicada no existe")
	}
	if !strings.Contains(m.picker.status, "duplicada") {
		t.Fatalf("estado = %q", m.picker.status)
	}
}

func TestSessionPickerBorrar(t *testing.T) {
	seedSessions(t)
	m := newFlowModel()
	m.sessionName = "beta"
	m.sessionEnabled = true
	m.openSessionPicker()
	m.picker.query.SetValue("alfa")
	m.picker.clamp()
	m.deleteSelectedSession()

	if SessionExists("alfa") {
		t.Fatal("alfa no se borró")
	}
	if m.SessionName() != "beta" {
		t.Fatalf("no debe cambiar la sesión activa: %q", m.SessionName())
	}

	// Borrar la sesión activa cae en la siguiente o crea una nueva.
	m.picker.query.SetValue("beta")
	m.picker.clamp()
	m.picker.mode = pickerBrowse
	m.deleteSelectedSession()
	if SessionExists("beta") {
		t.Fatal("beta no se borró")
	}
	if m.SessionName() == "" || m.SessionName() == "beta" {
		t.Fatalf("sesión de respaldo inesperada: %q", m.SessionName())
	}
}
