package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMouseClickCambiaPanel(t *testing.T) {
	m := newFlowModel()
	declararEcuacion(t, &m) // deja foco en panel 2
	m.switchPanel()         // vuelve al panel 1
	if m.focus != FocusDecl {
		t.Fatalf("foco inicial = %v", m.focus)
	}

	// Clic en la mitad inferior de la columna izquierda → panel 2
	msg := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      5,
		Y:      m.leftHeight() + 2,
	}
	if _, handled := m.handleMouse(msg); !handled {
		t.Fatal("el clic debe manejarse")
	}
	if m.focus != FocusCmd {
		t.Fatalf("el clic no cambió al panel 2: %v", m.focus)
	}

	// Clic en la mitad superior → panel 1
	msg.Y = 2
	if _, handled := m.handleMouse(msg); !handled {
		t.Fatal("el clic superior debe manejarse")
	}
	if m.focus != FocusDecl {
		t.Fatalf("el clic no volvió al panel 1: %v", m.focus)
	}
}

func TestMouseClickPanel2Bloqueado(t *testing.T) {
	m := newFlowModel()
	msg := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      5,
		Y:      m.leftHeight() + 2,
	}
	if _, handled := m.handleMouse(msg); !handled {
		t.Fatal("el clic debe manejarse")
	}
	if m.focus != FocusDecl {
		t.Fatalf("el panel 2 no debe enfocarse sin objetos: %v", m.focus)
	}
}

// TestMouseClickCoincideConLayoutEstrecho verifica que el hit-testing usa el
// mismo leftWidth/leftHeight que el render en una terminal de 40x12.
func TestMouseClickCoincideConLayoutEstrecho(t *testing.T) {
	m := NewModel()
	m.width, m.height = 40, 12
	m.updateLayout()
	declararEcuacion(t, &m) // foco en panel 2
	m.switchPanel()         // volver al panel 1

	click := func(x, y int) bool {
		t.Helper()
		_, handled := m.handleMouse(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      x,
			Y:      y,
		})
		return handled
	}

	if !click(5, m.leftHeight()) || m.focus != FocusCmd {
		t.Fatalf("el clic en la mitad inferior no enfocó el panel 2: %v", m.focus)
	}
	if !click(5, m.leftHeight()-1) || m.focus != FocusDecl {
		t.Fatalf("el clic en la mitad superior no enfocó el panel 1: %v", m.focus)
	}
	if click(5, m.gridAvailable()) {
		t.Fatal("un clic en la barra de estado no debe manejarse como panel")
	}
	if m.focus != FocusDecl {
		t.Fatalf("el clic fuera de la cuadrícula cambió el foco: %v", m.focus)
	}
}

func TestMouseWheelScrollSalida(t *testing.T) {
	m := newFlowModel()
	for range 40 {
		m.addLog(LevelInfo, strings.Repeat("linea ", 5))
	}
	m.output.GotoTop()
	before := m.output.YOffset

	msg := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      80,
		Y:      10,
	}
	if _, handled := m.handleMouse(msg); !handled {
		t.Fatal("la rueda debe manejarse")
	}
	if m.output.YOffset == before {
		t.Fatal("la rueda no desplazó el viewport")
	}
}

func TestPlainOutputSinANSI(t *testing.T) {
	m := newFlowModel()
	m.addLog(LevelSuccess, "resultado")
	text := m.plainOutput()
	if strings.Contains(text, "\x1b[") {
		t.Fatalf("plainOutput contiene ANSI: %q", text)
	}
	if !strings.Contains(text, "resultado") {
		t.Fatalf("plainOutput perdió contenido: %q", text)
	}
}
