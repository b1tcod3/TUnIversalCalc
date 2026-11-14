package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"tuniversalcalc/internal/engine"
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
