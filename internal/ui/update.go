package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case ResultMsg:
		if msg.Err != nil {
			m.addLog(LevelError, msg.Err.Error())
			return m, nil
		}
		for _, s := range msg.Steps {
			m.addLog(LevelStep, s)
		}
		m.addLog(LevelSuccess, "Listo")
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		// Navegación entre paneles
		case tea.KeyTab:
			return m.switchPanel(), nil

		// Salto de campo dentro de una plantilla
		case tea.KeyShiftTab:
			if m.focus != FocusCmd || !HasFields(m.cmdInput.Value()) {
				return m, nil
			}
			val, pos := JumpNextField(m.cmdInput.Value(), m.cmdInput.Position())
			m.cmdInput.SetValue(val)
			m.cmdInput.SetCursor(pos)
			return m, nil

		case tea.KeyEnter:
			return m.handleEnter()

		// Historial navegable en el panel de comando
		case tea.KeyUp:
			if m.focus == FocusCmd {
				m.histPrev()
				return m, nil
			}
		case tea.KeyDown:
			if m.focus == FocusCmd {
				m.histNext()
				return m, nil
			}

		// Scroll del panel de salida
		case tea.KeyPgUp, tea.KeyPgDown:
			var cmd tea.Cmd
			m.output, cmd = m.output.Update(msg)
			return m, cmd
		}
	}

	// Enrutar el mensaje solo al input enfocado (el otro devuelve nil)
	var cmd tea.Cmd
	if m.focus == FocusExpr {
		m.exprInput, cmd = m.exprInput.Update(msg)
	} else {
		m.cmdInput, cmd = m.cmdInput.Update(msg)
	}
	return m, cmd
}

// handleEnter decide según el panel enfocado:
//   - Expr: evalúa la expresión directamente.
//   - Cmd: si matchea una plantilla la inserta; si no, ejecuta el comando.
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.focus {
	case FocusExpr:
		expr := strings.TrimSpace(m.exprInput.Value())
		if expr == "" {
			return m, nil
		}
		m.addLog(LevelInfo, "Evaluando: "+expr)
		return m, evaluateCmd(expr, "")

	case FocusCmd:
		if tpl, ok := MatchTemplate(m.cmdInput.Value()); ok {
			val, pos := JumpNextField(tpl.InsertText, 0)
			m.cmdInput.SetValue(val)
			m.cmdInput.SetCursor(pos)
			m.addLog(LevelInfo, "Plantilla insertada: "+strings.TrimPrefix(tpl.Trigger, "/")+" (Shift+Tab para saltar campos)")
			return m, nil
		}

		val := strings.TrimSpace(m.cmdInput.Value())
		expr := strings.TrimSpace(m.exprInput.Value())
		if val == "" || expr == "" {
			return m, nil
		}
		m.pushHistory(val)
		m.addLog(LevelInfo, "Comando: "+val)
		return m, evaluateCmd(expr, val)
	}
	return m, nil
}

func (m Model) switchPanel() Model {
	switch m.focus {
	case FocusExpr:
		m.focus = FocusCmd
		m.exprInput.Blur()
		m.cmdInput.Focus()
	default:
		m.focus = FocusExpr
		m.cmdInput.Blur()
		m.exprInput.Focus()
	}
	return m
}

// --- Historial del panel de comando ---

func (m *Model) pushHistory(v string) {
	m.history = append(m.history, v)
	m.histIdx = len(m.history)
}

func (m *Model) histPrev() {
	if len(m.history) == 0 {
		return
	}
	if m.histIdx > 0 {
		m.histIdx--
	}
	m.cmdInput.SetValue(m.history[m.histIdx])
	m.cmdInput.CursorEnd()
}

func (m *Model) histNext() {
	if len(m.history) == 0 {
		return
	}
	if m.histIdx < len(m.history)-1 {
		m.histIdx++
		m.cmdInput.SetValue(m.history[m.histIdx])
		m.cmdInput.CursorEnd()
	} else if m.histIdx == len(m.history)-1 {
		m.histIdx++
		m.cmdInput.SetValue("")
	}
}

// --- Log del panel de salida ---

func (m *Model) addLog(level LogLevel, content string) {
	m.outputLog = append(m.outputLog, LogEntry{Level: level, Content: content})
	if len(m.outputLog) > maxLogEntries {
		m.outputLog = m.outputLog[len(m.outputLog)-maxLogEntries:]
	}
	m.output.SetContent(m.renderLog())
	m.output.GotoBottom()
}

func (m Model) renderLog() string {
	var b strings.Builder
	for i, e := range m.outputLog {
		if i > 0 {
			b.WriteByte('\n')
		}
		switch e.Level {
		case LevelSuccess:
			b.WriteString(levelSuccessStyle.Render("✔ " + e.Content))
		case LevelError:
			b.WriteString(levelErrorStyle.Render("✗ " + e.Content))
		case LevelStep:
			b.WriteString(levelStepStyle.Render("⚙ " + e.Content))
		default:
			b.WriteString(levelInfoStyle.Render("ℹ " + e.Content))
		}
	}
	return b.String()
}

// --- Layout dinámico ---

// updateLayout recalcula las dimensiones de los componentes según el tamaño
// de la terminal. Solo se aplica si ya recibimos WindowSizeMsg.
func (m *Model) updateLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	lw, rw := m.leftWidth(), m.rightWidth()
	_, rh := m.leftHeight(), m.rightHeight()
	contentW := m.panelInner(lw)

	m.exprInput.Width = contentW - 7 // margen para el prompt "ƒ(x) = "
	if m.exprInput.Width < 10 {
		m.exprInput.Width = 10
	}
	m.cmdInput.Width = contentW - 3 // margen para el prompt "❯ "
	if m.cmdInput.Width < 10 {
		m.cmdInput.Width = 10
	}

	w, h := m.panelInner(rw), m.panelInner(rh)
	if w < 20 {
		w = 20
	}
	if h < 3 {
		h = 3
	}
	m.output.Width = w
	m.output.Height = h
}

// panelInner = ancho interior que queda tras borde (2) + padding (2).
func (m Model) panelInner(outer int) int {
	if inner := outer - 4; inner > 0 {
		return inner
	}
	return 1
}

func (m Model) leftWidth() int {
	w := m.width * 45 / 100
	if w < 40 {
		w = 40
	}
	return w
}

func (m Model) rightWidth() int {
	return m.width - m.leftWidth() - 3
}

func (m Model) leftHeight() int {
	return (m.height - 3) / 2
}

func (m Model) rightHeight() int {
	return m.height - 3
}
