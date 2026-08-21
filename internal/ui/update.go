package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"tuniversalcalc/internal/engine"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Paleta abierta: navegar, insertar o cerrar tiene prioridad
		if m.palette.Open {
			switch msg.Type {
			case tea.KeyUp, tea.KeyShiftTab:
				m.paletteUp()
				return m, nil
			case tea.KeyDown, tea.KeyTab:
				m.paletteDown()
				return m, nil
			case tea.KeyEnter:
				m.paletteSelect()
				return m, nil
			case tea.KeyEsc:
				m.closePalette()
				return m, nil
			}
			// Las demás teclas caen al input y refiltran la lista
		}

		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		// Navegación entre paneles
		case tea.KeyTab:
			m.switchPanel()
			return m, nil

		// Salto de campo dentro de una plantilla (ambos paneles)
		case tea.KeyShiftTab:
			var ti *textinput.Model
			if m.focus == FocusDecl {
				ti = &m.declInput
			} else {
				ti = &m.cmdInput
			}
			if !HasFields(ti.Value()) {
				return m, nil
			}
			val, pos := JumpNextField(ti.Value(), ti.Position())
			ti.SetValue(val)
			ti.SetCursor(pos)
			return m, nil

		case tea.KeyEnter:
			cmd := m.handleEnter()
			return m, cmd

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
	if m.focus == FocusDecl {
		m.declInput, cmd = m.declInput.Update(msg)
	} else {
		m.cmdInput, cmd = m.cmdInput.Update(msg)
	}
	m.refreshPalette()
	return m, cmd
}

// handleEnter decide según el panel enfocado:
//   - Declaración: si matchea una plantilla la inserta; si no, valida la
//     declaración, la registra con auto-nombre (e1, f1, p1...) y el foco
//     salta al panel 2.
//   - Comando: si matchea una plantilla la inserta; si no, resuelve el
//     objeto (por nombre o el último) y ejecuta el motor.
func (m *Model) handleEnter() tea.Cmd {
	switch m.focus {
	case FocusDecl:
		raw := strings.TrimSpace(m.declInput.Value())
		if raw == "" {
			m.addLog(LevelInfo, "Declara un objeto: /ecuacion 2x+2=0, /funcion 2x+2 o /expresion x^2+1")
			return nil
		}
		if tpl, ok := matchTemplate(declRegistry, raw); ok {
			val, pos := JumpNextField(tpl.InsertText, 0)
			m.declInput.SetValue(val)
			m.declInput.SetCursor(pos)
			m.addLog(LevelInfo, "Plantilla insertada: "+tpl.Trigger+" (Shift+Tab para saltar campos)")
			return nil
		}
		d, err := engine.ParseDecl(raw)
		if err != nil {
			m.addLog(LevelError, err.Error())
			return nil
		}
		d.Nombre = m.autoName(d.Tipo)
		m.objs[d.Nombre] = d
		m.lastObj = d.Nombre
		m.declValid = true
		m.tabHinted = false
		m.addLog(LevelSuccess, d.Nombre+" declarada: "+d.Latex)
		m.addLog(LevelStep, d.Pretty)
		m.switchPanel()
		return nil

	case FocusCmd:
		if !m.declValid {
			return nil
		}
		if tpl, ok := matchTemplate(opsRegistry, m.cmdInput.Value()); ok {
			val, pos := JumpNextField(tpl.InsertText, 0)
			m.cmdInput.SetValue(val)
			m.cmdInput.SetCursor(pos)
			m.addLog(LevelInfo, "Plantilla insertada: "+tpl.Trigger+" (Shift+Tab para saltar campos)")
			return nil
		}

		val := strings.TrimSpace(m.cmdInput.Value())
		if val == "" {
			return nil
		}
		decl, varName, ok := m.resolveTarget(val)
		if !ok {
			return nil
		}
		m.pushHistory(val)
		m.addLog(LevelInfo, fmt.Sprintf("Comando: %s  (objeto: %s = %s)", val, decl.Nombre, decl.Latex))
		return evaluateCmd(decl, val, varName)
	}
	return nil
}

// resolveTarget averigua sobre qué objeto y variable opera el comando:
//   - /resolver x     → último objeto, variable x
//   - /resolver e1 x  → objeto e1, variable x
//   - /resolver e1    → objeto e1 (variable por defecto x)
//   - /resolver g1    → error si "g1" no es un objeto declarado
func (m *Model) resolveTarget(cmd string) (engine.Decl, string, bool) {
	norm := strings.NewReplacer("(", "", ")", "", ",", "").Replace(cmd)
	toks := strings.Fields(norm)
	if len(toks) == 0 {
		return engine.Decl{}, "", false
	}
	objName, varName := "", "x"
	if len(toks) >= 2 {
		name := toks[1]
		if _, ok := m.objs[name]; ok {
			objName = name
			if len(toks) >= 3 {
				varName = toks[2]
			}
		} else if len(name) == 1 {
			varName = name
		} else {
			m.addLog(LevelError, fmt.Sprintf("objeto %q no declarado — declarados: %s", name, m.objList()))
			return engine.Decl{}, "", false
		}
	}
	if objName == "" {
		objName = m.lastObj
	}
	if objName == "" {
		m.addLog(LevelError, "no hay objetos declarados: declara uno en el panel 1 (/ecuacion 2x+2=0)")
		return engine.Decl{}, "", false
	}
	return m.objs[objName], varName, true
}

// autoName asigna el siguiente nombre automático para el tipo: e1, e2...
func (m *Model) autoName(t engine.DeclType) string {
	m.counters[t]++
	return fmt.Sprintf("%s%d", t.Prefix(), m.counters[t])
}

func (m *Model) switchPanel() {
	m.closePalette() // cambiar de panel cierra la paleta
	switch m.focus {
	case FocusDecl:
		// El panel 2 solo se habilita con al menos un objeto declarado
		if !m.declValid {
			if !m.tabHinted {
				m.tabHinted = true
				m.addLog(LevelInfo, "Panel 2 bloqueado: declara un objeto en el panel 1 y pulsa Enter")
			}
			return
		}
		m.focus = FocusCmd
		m.declInput.Blur()
		m.cmdInput.Focus()
	default:
		m.focus = FocusDecl
		m.cmdInput.Blur()
		m.declInput.Focus()
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

	m.declInput.Width = contentW - 3 // margen para el prompt "❯ "
	if m.declInput.Width < 10 {
		m.declInput.Width = 10
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

	// Filas visibles de la paleta: contenido del panel menos título,
	// blanco e input. 2..7 para que nunca desborde el layout.
	m.palette.Max = m.leftHeight() - chromeInH - 4
	if m.palette.Max < 2 {
		m.palette.Max = 2
	}
	if m.palette.Max > 7 {
		m.palette.Max = 7
	}
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
