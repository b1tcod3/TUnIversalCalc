package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type pickerMode int

const (
	pickerBrowse pickerMode = iota
	pickerRename
	pickerConfirmDelete
)

// sessionPicker es el selector modal de sesiones (Ctrl+O).
type sessionPicker struct {
	open      bool
	mode      pickerMode
	query     textinput.Model
	nameInput textinput.Model
	items     []SessionMeta
	sel       int
	status    string
}

func newSessionPicker() sessionPicker {
	q := textinput.New()
	q.Prompt = "> "
	q.Placeholder = "filtra por nombre…"
	q.CharLimit = maxSessionNameLen

	name := textinput.New()
	name.Prompt = ""
	name.CharLimit = maxSessionNameLen

	return sessionPicker{query: q, nameInput: name}
}

// filtered devuelve las sesiones que coinciden con la consulta, conservando
// el orden por fecha de modificación.
func (p *sessionPicker) filtered() []SessionMeta {
	q := strings.TrimSpace(p.query.Value())
	if q == "" {
		return p.items
	}
	out := make([]SessionMeta, 0, len(p.items))
	for _, s := range p.items {
		if fuzzy.MatchFold(q, s.Name) {
			out = append(out, s)
		}
	}
	return out
}

func (p *sessionPicker) selected() (SessionMeta, bool) {
	list := p.filtered()
	if p.sel < 0 || p.sel >= len(list) {
		return SessionMeta{}, false
	}
	return list[p.sel], true
}

func (p *sessionPicker) clamp() {
	list := p.filtered()
	if p.sel >= len(list) {
		p.sel = len(list) - 1
	}
	if p.sel < 0 {
		p.sel = 0
	}
}

func (p *sessionPicker) reload() {
	p.items = ListSessions()
	p.clamp()
}

func (m *Model) openSessionPicker() {
	if !m.sessionEnabled {
		m.addLog(LevelInfo, "Sesión no habilitada: inicia sin --no-save para usar sesiones")
		return
	}
	m.saveSessionNow()
	items := ListSessions()
	if len(items) == 0 {
		m.addLog(LevelInfo, "No hay sesiones guardadas todavía")
		return
	}
	m.closePalette()
	m.helpOpen = false
	m.closeCompletion()
	m.clearPreview()
	m.picker.open = true
	m.picker.mode = pickerBrowse
	m.picker.items = items
	m.picker.sel = 0
	m.picker.status = ""
	m.picker.query.SetValue("")
	m.picker.query.Focus()
	m.picker.nameInput.Blur()
}

func (m *Model) closeSessionPicker() {
	m.picker.open = false
	m.picker.mode = pickerBrowse
	m.picker.status = ""
	m.picker.query.Blur()
	m.picker.nameInput.Blur()
}

func (m *Model) handleSessionPickerKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch m.picker.mode {
	case pickerRename:
		return m.handlePickerRenameKey(msg)
	case pickerConfirmDelete:
		return m.handlePickerDeleteKey(msg)
	}
	return m.handlePickerBrowseKey(msg)
}

func (m *Model) handlePickerBrowseKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m.quit(), true
	case tea.KeyEsc, tea.KeyCtrlO:
		m.closeSessionPicker()
		return nil, true
	case tea.KeyUp:
		if m.picker.sel > 0 {
			m.picker.sel--
		}
		return nil, true
	case tea.KeyDown:
		if m.picker.sel < len(m.picker.filtered())-1 {
			m.picker.sel++
		}
		return nil, true
	case tea.KeyEnter:
		s, ok := m.picker.selected()
		if !ok {
			return nil, true
		}
		m.switchToSession(s.Name)
		return nil, true
	case tea.KeyCtrlR:
		return m.startRename(), true
	case tea.KeyCtrlU:
		m.duplicateSelected()
		return nil, true
	case tea.KeyCtrlD:
		if _, ok := m.picker.selected(); ok {
			m.picker.mode = pickerConfirmDelete
			m.picker.status = ""
			m.picker.query.Blur()
		}
		return nil, true
	}

	var cmd tea.Cmd
	m.picker.query, cmd = m.picker.query.Update(msg)
	m.picker.status = ""
	m.picker.clamp()
	return cmd, true
}

func (m *Model) startRename() tea.Cmd {
	s, ok := m.picker.selected()
	if !ok {
		return nil
	}
	m.picker.mode = pickerRename
	m.picker.status = ""
	m.picker.query.Blur()
	m.picker.nameInput.SetValue(s.Name)
	m.picker.nameInput.CursorEnd()
	m.picker.nameInput.Focus()
	return textinput.Blink
}

func (m *Model) handlePickerRenameKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m.quit(), true
	case tea.KeyEsc:
		m.picker.mode = pickerBrowse
		m.picker.nameInput.Blur()
		m.picker.query.Focus()
		return nil, true
	case tea.KeyEnter:
		m.acceptRename()
		return nil, true
	}

	var cmd tea.Cmd
	m.picker.nameInput, cmd = m.picker.nameInput.Update(msg)
	return cmd, true
}

func (m *Model) acceptRename() {
	s, ok := m.picker.selected()
	if !ok {
		return
	}
	newName := sanitizeSessionName(m.picker.nameInput.Value())
	if newName == "" {
		m.picker.status = "el nombre no puede estar vacío"
		return
	}
	if newName != s.Name {
		if err := RenameSession(s.Name, newName); err != nil {
			m.picker.status = "no se pudo renombrar: " + err.Error()
			return
		}
		if m.sessionName == s.Name {
			m.sessionName = newName
		}
		m.picker.status = "renombrada como " + newName
	}
	m.picker.mode = pickerBrowse
	m.picker.nameInput.Blur()
	m.picker.query.Focus()
	m.picker.reload()
}

func (m *Model) duplicateSelected() {
	s, ok := m.picker.selected()
	if !ok {
		return
	}
	copyName, err := DuplicateSession(s.Name)
	if err != nil {
		m.picker.status = "no se pudo duplicar: " + err.Error()
		return
	}
	m.picker.status = "duplicada como " + copyName
	m.picker.reload()
}

func (m *Model) handlePickerDeleteKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.Type == tea.KeyCtrlC {
		return m.quit(), true
	}
	switch strings.ToLower(msg.String()) {
	case "y":
		m.deleteSelectedSession()
	default:
		m.picker.mode = pickerBrowse
		m.picker.query.Focus()
	}
	return nil, true
}

func (m *Model) deleteSelectedSession() {
	s, ok := m.picker.selected()
	if !ok {
		m.picker.mode = pickerBrowse
		return
	}
	if err := DeleteSession(s.Name); err != nil {
		m.picker.status = "no se pudo borrar: " + err.Error()
		m.picker.mode = pickerBrowse
		return
	}
	m.picker.status = "borrada " + s.Name
	m.picker.mode = pickerBrowse
	m.picker.reload()

	if m.sessionName != s.Name {
		m.picker.query.Focus()
		return
	}

	// La sesión abierta desapareció: saltar a la siguiente más reciente o
	// crear una nueva vacía.
	m.sessionName = ""
	m.picker.open = false
	if next, exists := LastSessionName(); exists {
		m.switchToSession(next)
		return
	}
	m.applySession(NewSessionName(), SessionData{})
}

// switchToSession guarda la sesión actual y carga la seleccionada.
func (m *Model) switchToSession(name string) {
	if name == m.sessionName {
		m.closeSessionPicker()
		return
	}
	m.saveSessionNow()
	data, err := LoadSession(name)
	if err != nil {
		m.addLog(LevelError, "No se pudo abrir la sesión: "+err.Error())
		return
	}
	m.applySession(name, data)
	m.closeSessionPicker()
	m.addLog(LevelInfo, "Sesión cargada: "+name)
	m.dirty = false
}

// pickerView renderiza el selector como pantalla completa.
func (m *Model) pickerView() string {
	if m.width < 40 {
		m.width = 40
	}
	if m.height < 12 {
		m.height = 12
	}

	width := m.width - 4
	if width < 24 {
		width = 24
	}
	m.picker.nameInput.Width = width - 18
	if m.picker.nameInput.Width < 10 {
		m.picker.nameInput.Width = 10
	}

	title := helpTitleStyle.Render("Sesiones")
	header := m.picker.query.View()
	switch m.picker.mode {
	case pickerRename:
		header = "Renombrar a: " + m.picker.nameInput.View()
	case pickerConfirmDelete:
		s, _ := m.picker.selected()
		header = lipgloss.NewStyle().Foreground(colorError).Render(
			"¿Eliminar \"" + truncateVisual(s.Name, width-24) + "\"? (y/N)")
	}

	list := m.picker.filtered()
	maxRows := m.height - 10
	if maxRows < 2 {
		maxRows = 2
	}
	start := m.picker.sel - maxRows/2
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > len(list) {
		end = len(list)
		start = end - maxRows
		if start < 0 {
			start = 0
		}
	}

	var rows []string
	if len(list) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(colorInfo).Render("  (sin coincidencias)"))
	}
	for i := start; i < end; i++ {
		s := list[i]
		line := padOrTrimVisual(s.Name, width-22) + "  " + s.Modified.Format("2006-01-02 15:04")
		if i == m.picker.sel {
			rows = append(rows, paletteSelStyle.Render("▶ "+line))
		} else {
			rows = append(rows, paletteRowStyle.Render("  "+line))
		}
	}

	footer := "↑/↓ seleccionar · Enter abrir · ^R renombrar · ^U duplicar · ^D borrar · Esc cerrar"
	switch m.picker.mode {
	case pickerRename:
		footer = "Enter guardar · Esc cancelar"
	case pickerConfirmDelete:
		footer = "y confirmar · cualquier otra tecla cancelar"
	}
	if m.picker.status != "" {
		footer = m.picker.status
	}

	content := title + "\n" + header + "\n\n" + strings.Join(rows, "\n") +
		"\n\n" + helpFooterStyle.Render(truncateVisual(footer, width))
	box := helpBoxStyle.Width(width).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
