package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"tuniversalcalc/internal/engine"
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
		m.refreshExpr()
	} else {
		m.cmdInput, cmd = m.cmdInput.Update(msg)
	}
	return m, cmd
}

// handleEnter decide según el panel enfocado:
//   - Expr: valida y guarda la expresión como f (foco salta al panel 2).
//   - Cmd: si matchea una plantilla la inserta; si no, ejecuta el comando
//     sobre la variable f guardada.
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.focus {
	case FocusExpr:
		expr := strings.TrimSpace(m.exprInput.Value())
		if expr == "" {
			m.addLog(LevelInfo, "Escribe una expresión matemática en el panel 1")
			return m, nil
		}
		if err := engine.Validate(expr); err != nil {
			m.exprValid = false
			m.addLog(LevelError, "Expresión inválida: "+err.Error())
			return m, nil
		}
		m.vars["f"] = expr
		m.exprValid = true
		m.addLog(LevelSuccess, "f = "+expr+" guardada")
		return m.switchPanel(), nil

	case FocusCmd:
		if !m.exprValid {
			return m, nil
		}
		if tpl, ok := MatchTemplate(m.cmdInput.Value()); ok {
			val, pos := JumpNextField(tpl.InsertText, 0)
			m.cmdInput.SetValue(val)
			m.cmdInput.SetCursor(pos)
			m.addLog(LevelInfo, "Plantilla insertada: "+tpl.Trigger+" (Shift+Tab para saltar campos)")
			return m, nil
		}

		val := strings.TrimSpace(m.cmdInput.Value())
		expr := m.vars["f"]
		if val == "" || expr == "" {
			return m, nil
		}
		m.pushHistory(val)
		m.addLog(LevelInfo, "Comando: "+val+"  (f = "+expr+")")
		return m, evaluateCmd(expr, val)
	}
	return m, nil
}

func (m Model) switchPanel() Model {
	switch m.focus {
	case FocusExpr:
		// El panel 2 solo se habilita con una expresión válida guardada
		if !m.exprValid || m.vars["f"] == "" {
			if !m.tabHinted {
				m.tabHinted = true
				m.addLog(LevelInfo, "Panel 2 bloqueado: escribe y valida una expresión en el panel 1 (Enter)")
			}
			return m
		}
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

// refreshExpr reevalúa la validez de la expresión en vivo. Si el usuario
// rompe la expresión estando en el panel 2, vuelve automáticamente al 1.
func (m *Model) refreshExpr() {
	expr := strings.TrimSpace(m.exprInput.Value())
	m.exprValid = engine.Validate(expr) == nil
	if m.exprValid {
		m.tabHinted = false
	}
	if !m.exprValid && m.focus == FocusCmd {
		m.focus = FocusExpr
		m.cmdInput.Blur()
		m.exprInput.Focus()
	}
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

// Chrome (borde + padding) que lipgloss suma POR ENCIMA de Width/Height:
//   - inputPanel: Padding(0,1) → 4 cols (2 borde + 2 padX), 2 filas (borde)
//   - outputPanel: Padding(1)   → 4 cols, 4 filas (borde + 2 padY)
// Width() y Height() fijan el ÁREA DE CONTENIDO, así que los paneles se
// dimensionan restando el chrome para que el total sea exacto.

const (
	chromeInW  = 4
	chromeInH  = 2
	chromeOutW = 4
	chromeOutH = 4
)

// updateLayout recalcula las dimensiones de los componentes según el tamaño
// de la terminal. Solo se aplica si ya recibimos WindowSizeMsg.
func (m *Model) updateLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	lw, rw := m.leftWidth(), m.rightWidth()
	_, rh := m.leftHeight(), m.rightHeight()
	contentW := lw - chromeInW - chromeInW // área interior de los paneles 1/2

	m.exprInput.Width = contentW - 7 // margen para el prompt "ƒ(x) = "
	if m.exprInput.Width < 10 {
		m.exprInput.Width = 10
	}
	m.cmdInput.Width = contentW - 3 // margen para el prompt "❯ "
	if m.cmdInput.Width < 10 {
		m.cmdInput.Width = 10
	}

	// El viewport ocupa el contenido del panel 3 tras la cabecera
	w := rw - chromeOutW - chromeOutW
	h := (rh - chromeOutH) - m.panel3HeaderRows()
	if w < 6 {
		w = 6
	}
	if h < 2 {
		h = 2
	}
	m.output.Width = w
	m.output.Height = h
}

func (m Model) leftWidth() int {
	w := m.width * 50 / 100
	if w < 45 {
		w = 45
	}
	return w
}

func (m Model) rightWidth() int {
	return m.width - m.leftWidth() - 3
}

func (m Model) leftHeight() int {
	return m.gridAvailable() / 2
}

func (m Model) rightHeight() int {
	return m.gridAvailable()
}

// gridAvailable es la altura que queda para la cuadrícula tras restar
// las 2 líneas de chrome (ayuda + estado). Mantener esto exacto evita
// que la View desborde la terminal y corte la parte superior.
func (m Model) gridAvailable() int {
	if h := m.height - 2; h > 0 {
		return h
	}
	return 1
}
