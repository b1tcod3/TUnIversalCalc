package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// sessionPicker es el selector modal de sesiones (Ctrl+O).
type sessionPicker struct {
	open  bool
	query textinput.Model
	items []SessionMeta
	sel   int
}

func newSessionPicker() sessionPicker {
	q := textinput.New()
	q.Prompt = "> "
	q.Placeholder = "filtra por nombre…"
	q.CharLimit = maxSessionNameLen
	return sessionPicker{query: q}
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
	m.picker.open = true
	m.picker.items = items
	m.picker.sel = 0
	m.picker.query.SetValue("")
	m.picker.query.Focus()
}

func (m *Model) closeSessionPicker() {
	m.picker.open = false
	m.picker.query.Blur()
}

func (m *Model) handleSessionPickerKey(msg tea.KeyMsg) (tea.Cmd, bool) {
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
	}

	var cmd tea.Cmd
	m.picker.query, cmd = m.picker.query.Update(msg)
	m.picker.clamp()
	return cmd, true
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
	title := helpTitleStyle.Render("Sesiones (Ctrl+O)")
	query := m.picker.query.View()

	list := m.picker.filtered()
	maxRows := m.height - 8
	if maxRows < 3 {
		maxRows = 3
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
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo)).Render("  (sin coincidencias)"))
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

	content := title + "\n" + query + "\n\n" + strings.Join(rows, "\n") +
		"\n\n" + helpFooterStyle.Render("↑/↓ seleccionar · Enter abrir · Esc cerrar")
	box := helpBoxStyle.Width(width).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
