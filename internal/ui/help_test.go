package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpSeAbreYCierra(t *testing.T) {
	m := newFlowModel()
	if !strings.Contains(helpPlain(), "TUnIversalCalc") {
		t.Fatal("help.txt está vacío o no se embebió")
	}

	m.toggleHelp()
	if !m.helpOpen {
		t.Fatal("la ayuda debe abrirse")
	}
	view := m.View()
	if !strings.Contains(view, "Ayuda de TUnIversalCalc") {
		t.Fatal("la vista de ayuda no contiene el título")
	}
	checkView(t, &m, [2]int{100, 28}, "ayuda")

	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc}); !handled {
		t.Fatal("Esc debe manejarse con la ayuda abierta")
	}
	if m.helpOpen {
		t.Fatal("Esc debe cerrar la ayuda")
	}
}

func TestHelpScroll(t *testing.T) {
	m := newFlowModel()
	m.toggleHelp()
	before := m.helpViewport.YOffset
	m.helpViewport.ScrollDown(5)
	if m.helpViewport.YOffset <= before {
		t.Fatal("la ayuda debe poder desplazarse")
	}
	if _, handled := m.handleKey(tea.KeyMsg{Type: tea.KeyDown}); !handled {
		t.Fatal("la flecha abajo debe manejarse en la ayuda")
	}
}

func TestHelpEnTerminalPequena(t *testing.T) {
	m := NewModel()
	m.width, m.height = 40, 12
	m.updateLayout()
	m.toggleHelp()
	checkView(t, &m, [2]int{40, 12}, "ayuda-pequeña")
}
