package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHistorialNavegacionConFlechas(t *testing.T) {
	m := newFlowModel()
	m.focus = FocusCmd
	m.history = []string{"/lagrange t1", "/newton t1"}
	m.histIdx = len(m.history)

	press := func(key tea.KeyType) {
		t.Helper()
		if _, handled := m.handleKey(tea.KeyMsg{Type: key}); !handled {
			t.Fatalf("la tecla %v no se manejó", key)
		}
	}

	press(tea.KeyUp)
	if got := m.cmdInput.Value(); got != "/newton t1" {
		t.Fatalf("primera flecha arriba = %q", got)
	}
	press(tea.KeyUp)
	if got := m.cmdInput.Value(); got != "/lagrange t1" {
		t.Fatalf("segunda flecha arriba = %q", got)
	}
	press(tea.KeyUp)
	if got := m.cmdInput.Value(); got != "/lagrange t1" {
		t.Fatalf("la flecha arriba no debe salir del historial: %q", got)
	}

	press(tea.KeyDown)
	if got := m.cmdInput.Value(); got != "/newton t1" {
		t.Fatalf("primera flecha abajo = %q", got)
	}
	press(tea.KeyDown)
	if got := m.cmdInput.Value(); got != "" {
		t.Fatalf("la flecha abajo final debe limpiar el input: %q", got)
	}
	press(tea.KeyDown)
	if got := m.cmdInput.Value(); got != "" {
		t.Fatalf("la flecha abajo no debe salir del historial: %q", got)
	}
}

func TestHistorialSoloEnPanelComando(t *testing.T) {
	m := newFlowModel()
	m.focus = FocusDecl
	m.history = []string{"/lagrange t1"}
	m.histIdx = 1

	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyUp}); handled {
		t.Fatal("la flecha arriba en el panel 1 no debe manejarse")
	}
	if m.cmdInput.Value() != "" {
		t.Fatalf("el historial no debe tocarse desde el panel 1: %q", m.cmdInput.Value())
	}
}

func TestPaletaNavegacionConTeclas(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/")
	m.refreshPalette()
	if !m.palette.Open || len(m.palette.List) < 2 {
		t.Fatalf("la paleta no se abrió: %+v", m.palette)
	}

	press := func(key tea.KeyType) {
		t.Helper()
		if _, handled := m.handleKey(tea.KeyMsg{Type: key}); !handled {
			t.Fatalf("la tecla %v no se manejó con la paleta abierta", key)
		}
	}

	press(tea.KeyDown)
	if m.palette.Sel != 1 {
		t.Fatalf("flecha abajo: selección = %d", m.palette.Sel)
	}
	press(tea.KeyTab)
	if m.palette.Sel != 2 {
		t.Fatalf("Tab: selección = %d", m.palette.Sel)
	}
	press(tea.KeyUp)
	if m.palette.Sel != 1 {
		t.Fatalf("flecha arriba: selección = %d", m.palette.Sel)
	}
	press(tea.KeyShiftTab)
	if m.palette.Sel != 0 {
		t.Fatalf("Shift+Tab: selección = %d", m.palette.Sel)
	}

	press(tea.KeyEsc)
	if m.palette.Open {
		t.Fatal("Esc debe cerrar la paleta")
	}
}

func TestPaletaInsertaPlantillaConEnter(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/puntos")
	if m.palette.Open {
		t.Fatal("con el comando completo la paleta debe estar cerrada")
	}
	m.declInput.SetValue("/pun")
	m.refreshPalette()
	if !m.palette.Open || len(m.palette.List) == 0 {
		t.Fatalf("la paleta no filtró /pun: %+v", m.palette)
	}
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter}); !handled {
		t.Fatal("Enter debe manejarse con la paleta abierta")
	}
	if m.palette.Open {
		t.Fatal("Enter debe cerrar la paleta")
	}
	if got := m.declInput.Value(); got != "/puntos " {
		t.Fatalf("plantilla insertada = %q", got)
	}
	if pos := m.declInput.Position(); pos != len("/puntos ") {
		t.Fatalf("cursor = %d, se esperaba %d", pos, len("/puntos "))
	}
}

func TestShiftTabSaltaCampoDePlantilla(t *testing.T) {
	m := newFlowModel()
	m.declInput.SetValue("/puntos ?")
	m.declInput.SetCursor(len("/puntos ?"))

	if !HasFields(m.declInput.Value()) {
		t.Fatal("la plantilla debe tener un campo pendiente")
	}
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyShiftTab}); !handled {
		t.Fatal("Shift+Tab debe manejarse")
	}
	if got := m.declInput.Value(); got != "/puntos " {
		t.Fatalf("valor tras saltar el campo = %q", got)
	}
	if HasFields(m.declInput.Value()) {
		t.Fatal("no quedan campos por saltar")
	}

	before := m.declInput.Value()
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyShiftTab}); !handled {
		t.Fatal("Shift+Tab sin campos también debe manejarse")
	}
	if m.declInput.Value() != before {
		t.Fatalf("Shift+Tab sin campos cambió el valor: %q", m.declInput.Value())
	}
}

func TestHasFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"campo pendiente", "/puntos ?", true},
		{"varios campos", "/funcion ? ?", true},
		{"sin campos", "/puntos (1,2) (3,4)", false},
		{"vacio", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasFields(tt.in); got != tt.want {
				t.Fatalf("HasFields(%q) = %v, se esperaba %v", tt.in, got, tt.want)
			}
		})
	}
}
