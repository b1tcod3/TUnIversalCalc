package ui

import (
	"strings"
	"tuniversalcalc/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// isClearCommand reconoce /limpiar y /limpiar todo, con o sin paréntesis.
func isClearCommand(raw string) bool {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "()")
	s = strings.Join(strings.Fields(s), " ")
	return s == "limpiar" || s == "limpiar todo"
}

// openClearConfirm abre el modal de confirmación sin borrar nada todavía.
func (m *Model) openClearConfirm() {
	m.clearConfirm = true
	m.helpOpen = false
	m.closeSessionPicker()
	m.closePalette()
	m.closeCompletion()
	m.clearPreview()
	m.invalidatePending()
	m.reqSeq++
}

// handleClearConfirmKey procesa el modal de /limpiar. s/y confirman y
// n/Esc cancelan; el resto de teclas se ignoran.
func (m *Model) handleClearConfirmKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m.quit(), true
	case tea.KeyEsc:
		m.clearConfirm = false
		return nil, true
	}
	switch strings.ToLower(msg.String()) {
	case "s", "y":
		m.applyClearWorkspace()
		return nil, true
	case "n":
		m.clearConfirm = false
		return nil, true
	}
	return nil, true
}

// applyClearWorkspace borra el espacio de trabajo de la sesión actual y la
// guarda vacía. Mantiene el nombre de sesión y la persistencia configurada.
func (m *Model) applyClearWorkspace() {
	m.clearConfirm = false
	m.helpOpen = false
	m.closeSessionPicker()
	m.closePalette()
	m.closeCompletion()
	m.clearPreview()
	m.invalidatePending()
	m.reqSeq = 0

	m.focus = FocusDecl
	m.declInput.SetValue("")
	m.declInput.SetCursor(0)
	m.cmdInput.SetValue("")
	m.cmdInput.SetCursor(0)
	m.cmdInput.Blur()
	m.declInput.Focus()

	m.objs = map[string]engine.Decl{}
	m.lastObj = ""
	m.counters = map[engine.DeclType]int{}
	m.declValid = false
	m.tabHinted = false

	m.history = nil
	m.histIdx = -1
	m.outputLog = welcomeLog()
	m.undoStack = nil
	m.redoStack = nil
	m.dirty = true

	m.addLog(LevelSuccess, "Espacio de trabajo limpiado")
	m.saveSessionNow()
}

// clearConfirmView dibuja el modal de confirmación centrado.
func (m *Model) clearConfirmView() string {
	boxWidth := max(m.width-4, 1)
	innerW := max(boxWidth-2, 1)
	textW := max(innerW-2, 1)

	title := clampWidth(confirmTitleStyle.Render("¿Limpiar todo?"), textW)
	bodyLines := []string{
		"Se borrarán objetos, historial,",
		"salida, inputs y undo/redo",
		"de la sesión actual.",
		"",
		"El autosave guardará la sesión",
		"vacía.",
		"No se puede deshacer.",
	}
	if m.height < 13 {
		bodyLines = []string{
			"Borra objetos, historial,",
			"salida e inputs.",
			"El autosave la guardará vacía.",
			"No se puede deshacer.",
		}
	}
	body := clampLines(strings.Join(bodyLines, "\n"), textW)
	footer := clampWidth("s = confirmar · n/Esc = cancelar", textW)

	content := title + "\n\n" + body + "\n\n" + footer
	box := confirmBoxStyle.Width(boxWidth).Render(content)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}
