package ui

import (
	_ "embed"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed help.txt
var helpText string

// toggleHelp abre o cierra la ayuda contextual.
func (m *Model) toggleHelp() {
	if m.helpOpen {
		m.helpOpen = false
		return
	}
	m.closePalette()
	m.closeCompletion()
	m.clearPreview()
	m.layoutHelp()
	m.helpViewport.SetContent(helpText)
	m.helpViewport.GotoTop()
	m.helpOpen = true
}

// layoutHelp ajusta el viewport de ayuda al tamaño actual de la terminal.
func (m *Model) layoutHelp() {
	w := m.width - 6
	if w < 20 {
		w = 20
	}
	h := m.height - 8
	if h < 3 {
		h = 3
	}
	m.helpViewport.Width = w
	m.helpViewport.Height = h
}

// handleHelpKey procesa las teclas mientras la ayuda está abierta.
func (m *Model) handleHelpKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m.quit(), true
	case tea.KeyEsc, tea.KeyCtrlH:
		m.helpOpen = false
		return nil, true
	case tea.KeyUp:
		m.helpViewport.LineUp(1)
		return nil, true
	case tea.KeyDown:
		m.helpViewport.LineDown(1)
		return nil, true
	case tea.KeyPgUp:
		m.helpViewport.HalfViewUp()
		return nil, true
	case tea.KeyPgDown:
		m.helpViewport.HalfViewDown()
		return nil, true
	}
	switch msg.String() {
	case "j":
		m.helpViewport.LineDown(1)
		return nil, true
	case "k":
		m.helpViewport.LineUp(1)
		return nil, true
	case "q":
		m.helpOpen = false
		return nil, true
	}
	return nil, true
}

// helpView renderiza la ayuda como una pantalla completa centrada.
func (m *Model) helpView() string {
	if m.width < 40 {
		m.width = 40
	}
	if m.height < 12 {
		m.height = 12
	}
	m.layoutHelp()

	title := helpTitleStyle.Render("Ayuda de TUnIversalCalc")
	footer := helpFooterStyle.Render("↑/↓ línea · PgUp/PgDn página · Esc o Ctrl+H cerrar")
	content := title + "\n\n" + m.helpViewport.View() + "\n\n" + footer
	box := helpBoxStyle.Width(m.width - 4).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// helpPlain devuelve el texto de ayuda sin formato, útil para tests.
func helpPlain() string {
	return strings.TrimSpace(helpText)
}
