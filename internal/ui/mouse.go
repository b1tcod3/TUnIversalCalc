package ui

import tea "github.com/charmbracelet/bubbletea"

// handleMouse gestiona rueda y clics. Devuelve handled=false solo para
// eventos que deben seguir su curso normal.
func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Cmd, bool) {
	if m.picker.open {
		return nil, true
	}
	if m.helpOpen {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.helpViewport.LineUp(3)
		case tea.MouseButtonWheelDown:
			m.helpViewport.LineDown(3)
		}
		return nil, true
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		var cmd tea.Cmd
		m.output, cmd = m.output.Update(msg)
		return cmd, true
	}

	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return nil, false
	}
	if msg.X < 0 || msg.Y < 0 || msg.X >= m.width {
		return nil, false
	}

	if msg.X < m.leftWidth() {
		if msg.Y >= m.leftHeight() {
			m.focusPanel(FocusCmd)
		} else {
			m.focusPanel(FocusDecl)
		}
		return nil, true
	}
	return nil, true
}

// focusPanel cambia el foco por clic. El panel 2 sigue bloqueado hasta que
// exista al menos un objeto declarado.
func (m *Model) focusPanel(f Focus) {
	if f == FocusCmd && !m.declValid {
		if !m.tabHinted {
			m.tabHinted = true
			m.addLog(LevelInfo, "Panel 2 bloqueado: declara un objeto en el panel 1 y pulsa Enter")
		}
		return
	}
	m.closePalette()
	m.closeCompletion()
	m.clearPreview()
	m.focus = f
	m.declInput.Blur()
	m.cmdInput.Blur()
	if f == FocusDecl {
		m.declInput.Focus()
	} else {
		m.cmdInput.Focus()
	}
}
