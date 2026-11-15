package ui

import (
	"errors"
	"fmt"
	"strings"
	"tuniversalcalc/internal/engine"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:gocyclo // despachador central de mensajes del bucle Elm
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		if m.helpOpen {
			m.layoutHelp()
		}
		return m, nil

	case ResultMsg:
		m.applyResult(msg)
		return m, nil

	case previewTickMsg:
		return m, m.handlePreviewTick(msg)

	case previewResultMsg:
		m.applyPreviewResult(msg)
		return m, nil

	case autosaveMsg:
		cmds := []tea.Cmd{m.autosaveTick()}
		if m.dirty {
			if cmd := m.saveSessionCmd(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case sessionSavedMsg:
		if msg.name != m.sessionName {
			return m, nil
		}
		if msg.err != nil {
			m.addLog(LevelError, "No se pudo guardar la sesión: "+msg.err.Error())
		} else if sessionFingerprint(m.sessionData()) == msg.fingerprint {
			m.dirty = false
		}
		return m, nil

	case tea.MouseMsg:
		if cmd, handled := m.handleMouse(msg); handled {
			return m, cmd
		}

	case tea.KeyMsg:
		if cmd, handled := m.handleKey(msg); handled {
			return m, cmd
		}
	}

	// Enrutar el mensaje solo al input enfocado (el otro devuelve nil)
	var cmd tea.Cmd
	if m.picker.open {
		m.picker.query, cmd = m.picker.query.Update(msg)
		return m, cmd
	}
	oldValue := m.focusedInputValue()
	if m.focus == FocusDecl {
		m.declInput, cmd = m.declInput.Update(msg)
	} else {
		m.cmdInput, cmd = m.cmdInput.Update(msg)
	}
	if m.focusedInputValue() != oldValue {
		m.dirty = true
		if m.focus == FocusCmd {
			if previewCmd := m.schedulePreview(); previewCmd != nil {
				cmd = tea.Batch(cmd, previewCmd)
			}
		}
	}
	m.refreshPalette()
	return m, cmd
}

// focusedInputValue devuelve el texto del panel activo.
func (m *Model) focusedInputValue() string {
	if m.focus == FocusDecl {
		return m.declInput.Value()
	}
	return m.cmdInput.Value()
}

// invalidatePending descarta resultados de operaciones lanzadas antes de un
// cambio de estado (nueva declaración, undo, etc.).
func (m *Model) invalidatePending() {
	m.pendingReq = 0
}

// applyResult vuelca el resultado del motor en el log de salida. Los
// resultados de peticiones antiguas se ignoran.
func (m *Model) applyResult(res ResultMsg) {
	if res.ReqID != 0 && res.ReqID != m.pendingReq {
		return
	}
	m.pendingReq = 0
	if res.Err != nil {
		m.addLog(LevelError, res.Err.Error())
		return
	}
	for _, s := range res.Steps {
		m.addLog(LevelStep, s)
	}
	m.addLog(LevelSuccess, "Listo")
}

// handleKey procesa una tecla. El segundo valor indica si el mensaje quedó
// manejado; si no, cae al input enfocado.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) { //nolint:gocyclo // un case por atajo de teclado, sin lógica anidada
	// El selector de sesiones es modal y tiene prioridad.
	if m.picker.open {
		return m.handleSessionPickerKey(msg)
	}

	// La ayuda modal tiene prioridad sobre paneles y paleta.
	if m.helpOpen {
		return m.handleHelpKey(msg)
	}

	// Paleta abierta: navegar, insertar o cerrar tiene prioridad
	if m.palette.Open {
		return m.handlePaletteKey(msg)
	}

	// Popup de autocompletado de expresiones.
	if m.completion.Open {
		return m.handleCompletionKey(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m.quit(), true

	case tea.KeyCtrlAt:
		m.openCompletions()
		return nil, true

	case tea.KeyCtrlH:
		m.toggleHelp()
		return nil, true

	case tea.KeyCtrlO:
		m.openSessionPicker()
		return nil, true

	case tea.KeyCtrlZ:
		m.undo()
		return nil, true

	case tea.KeyCtrlY:
		m.redo()
		return nil, true

	case tea.KeyCtrlS:
		return m.manualSave(), true

	case tea.KeyF2:
		m.copyOutput()
		return nil, true

	// Navegación entre paneles
	case tea.KeyTab:
		m.switchPanel()
		return nil, true

	// Salto de campo dentro de una plantilla (ambos paneles)
	case tea.KeyShiftTab:
		m.jumpNextTemplateField()
		return nil, true

	case tea.KeyEnter:
		return m.handleEnter(), true

	// Historial navegable en el panel de comando
	case tea.KeyUp:
		if m.focus == FocusCmd {
			m.histPrev()
			return nil, true
		}

	case tea.KeyDown:
		if m.focus == FocusCmd {
			m.histNext()
			return nil, true
		}

	// Scroll del panel de salida
	case tea.KeyPgUp, tea.KeyPgDown:
		var cmd tea.Cmd
		m.output, cmd = m.output.Update(msg)
		return cmd, true
	}
	return nil, false
}

// handlePaletteKey gestiona las teclas con la paleta abierta. Las teclas
// no reconocidas devuelven handled=false para que caigan al input y
// refiltren la lista.
func (m *Model) handlePaletteKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m.quit(), true
	case tea.KeyCtrlS:
		return m.manualSave(), true
	case tea.KeyUp, tea.KeyShiftTab:
		m.paletteUp()
		return nil, true
	case tea.KeyDown, tea.KeyTab:
		m.paletteDown()
		return nil, true
	case tea.KeyEnter:
		m.paletteSelect()
		return nil, true
	case tea.KeyEsc:
		m.closePalette()
		return nil, true
	}
	return nil, false
}

// jumpNextTemplateField salta al siguiente marcador "?" de la plantilla
// del input enfocado (Shift+Tab).
func (m *Model) jumpNextTemplateField() {
	var ti *textinput.Model
	if m.focus == FocusDecl {
		ti = &m.declInput
	} else {
		ti = &m.cmdInput
	}
	if !HasFields(ti.Value()) {
		return
	}
	m.pushUndo()
	m.invalidatePending()
	val, pos := JumpNextField(ti.Value(), ti.Position())
	ti.SetValue(val)
	ti.SetCursor(pos)
	m.dirty = true
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
			m.addLog(LevelInfo, "Declara un objeto: /puntos (1,2) (3,4), /funcion 2x+2 o /ecuacion 2x+2=0")
			return nil
		}
		if tpl, ok := matchTemplate(declRegistry, raw); ok {
			m.pushUndo()
			m.invalidatePending()
			val, pos := JumpNextField(tpl.InsertText, 0)
			m.declInput.SetValue(val)
			m.declInput.SetCursor(pos)
			m.dirty = true
			m.addLog(LevelInfo, "Plantilla insertada: "+tpl.Trigger+" (Shift+Tab para saltar campos)")
			return nil
		}
		d, err := engine.ParseDecl(raw)
		if err != nil {
			m.addLog(LevelError, err.Error())
			return nil
		}
		m.pushUndo()
		m.invalidatePending()
		m.clearPreview()
		d.Nombre = m.autoName(d.Tipo)
		m.objs[d.Nombre] = d
		m.lastObj = d.Nombre
		m.declValid = true
		m.tabHinted = false
		m.addLog(LevelSuccess, d.Nombre+" declarada: "+d.Latex)
		m.addLog(LevelStep, d.Pretty)
		m.switchPanel()

		// Detección automática: una expresión que sea un entero literal se
		// factoriza sin que el usuario tenga que pedirlo.
		if d.Tipo == engine.DeclExpresion {
			if _, ok := engine.ParseIntegerLiteral(d.Raw); ok {
				m.reqSeq++
				m.pendingReq = m.reqSeq
				return evaluateCmd(d, "/factorizar", "x", m.pendingReq)
			}
		}
		return nil

	case FocusCmd:
		if !m.declValid {
			if _, ok := m.literalIntegerDecl(m.cmdInput.Value()); !ok {
				return nil
			}
		}
		if tpl, ok := matchTemplate(opsRegistry, m.cmdInput.Value()); ok {
			m.pushUndo()
			m.invalidatePending()
			val, pos := JumpNextField(tpl.InsertText, 0)
			m.cmdInput.SetValue(val)
			m.cmdInput.SetCursor(pos)
			m.dirty = true
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
		m.pushUndo()
		m.pushHistory(val)
		m.addLog(LevelInfo, fmt.Sprintf("Comando: %s  (objeto: %s = %s)", val, decl.Nombre, decl.Latex))
		m.clearPreview()
		m.reqSeq++
		m.pendingReq = m.reqSeq
		return evaluateCmd(decl, val, varName, m.pendingReq)
	}
	return nil
}

// resolveTarget averigua sobre qué objeto y variable opera el comando:
//   - /lagrange x     → último objeto, variable x
//   - /lagrange t1    → objeto t1
//   - /newton g1      → error si "g1" no es un objeto declarado
func (m *Model) resolveTarget(cmd string) (engine.Decl, string, bool) {
	toks := splitCommandTokens(normalizeCommandArgs(cmd))
	if d, ok := m.literalIntegerDecl(cmd); ok {
		return d, "x", true
	}
	decl, varName, err := m.resolveTargetTokens(toks)
	if err != nil {
		m.addLog(LevelError, err.Error())
		return engine.Decl{}, "", false
	}
	return decl, varName, true
}

// resolveTargetTokens resuelve objeto y variable a partir de los tokens de un
// comando. Comparte la lógica entre la ejecución real y la previsualización.
func (m *Model) resolveTargetTokens(toks []string) (engine.Decl, string, error) {
	if len(toks) == 0 {
		return engine.Decl{}, "", errors.New("comando vacío")
	}
	objName, varName, err := m.parseTargetName(toks)
	if err != nil {
		return engine.Decl{}, "", err
	}
	decl, ok := m.objs[objName]
	if !ok {
		return engine.Decl{}, "", fmt.Errorf("objeto %q no declarado — declarados: %s", objName, m.objList())
	}
	return decl, varName, nil
}

// parseTargetName interpreta el objeto y la variable de los tokens del
// comando, aplicando el último objeto declarado como valor por defecto.
func (m *Model) parseTargetName(toks []string) (objName, varName string, err error) {
	if len(toks) < 2 {
		objName, err = m.defaultTargetName()
		return objName, "x", err
	}
	name := toks[1]
	if _, ok := m.objs[name]; ok {
		varName = "x"
		if len(toks) >= 3 {
			varName = toks[2]
		}
		return name, varName, nil
	}
	if len(name) > 1 {
		return "", "", fmt.Errorf("objeto %q no declarado — declarados: %s", name, m.objList())
	}
	objName, err = m.defaultTargetName()
	return objName, name, err
}

// defaultTargetName devuelve el último objeto declarado o un error si aún no
// hay ninguno.
func (m *Model) defaultTargetName() (string, error) {
	if m.lastObj == "" {
		return "", errors.New("no hay objetos declarados: declara uno en el panel 1 (/puntos (1,2) (3,4))")
	}
	return m.lastObj, nil
}

// literalIntegerDecl permite usar /factorizar 120 sin declarar el número
// como objeto. Solo aplica al comando de factorización.
func (m *Model) literalIntegerDecl(cmd string) (engine.Decl, bool) {
	toks := splitCommandTokens(normalizeCommandArgs(cmd))
	if len(toks) < 2 || !isFactorCommand(toks[0]) {
		return engine.Decl{}, false
	}
	n, ok := engine.ParseIntegerLiteral(toks[1])
	if !ok {
		return engine.Decl{}, false
	}
	return engine.IntegerDecl(n), true
}

func isFactorCommand(tok string) bool {
	tok = strings.ToLower(strings.TrimPrefix(tok, "/"))
	return strings.HasPrefix(tok, "factor")
}

// autoName asigna el siguiente nombre automático para el tipo: e1, e2...
func (m *Model) autoName(t engine.DeclType) string {
	m.counters[t]++
	return fmt.Sprintf("%s%d", t.Prefix(), m.counters[t])
}

func (m *Model) switchPanel() {
	m.closePalette() // cambiar de panel cierra la paleta
	m.closeCompletion()
	m.clearPreview()
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
	m.dirty = true
}

func (m *Model) renderLog() string {
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

	// margen para el prompt "❯ "
	m.declInput.Width = max(contentW-3, 10)
	m.cmdInput.Width = max(contentW-3, 10)

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
	m.palette.Max = min(max(m.leftHeight()-chromeInH-4, 2), 7)
	m.completion.Max = min(max(m.leftHeight()-chromeInH-4, 2), 7)
	if m.completion.Open {
		m.clampCompletion()
	}
}

func (m *Model) leftWidth() int {
	w := max(m.width*50/100, 45)
	return w
}

func (m *Model) rightWidth() int {
	return m.width - m.leftWidth() - 3
}

func (m *Model) leftHeight() int {
	return m.gridAvailable() / 2
}

func (m *Model) rightHeight() int {
	return m.gridAvailable()
}

// gridAvailable es la altura que queda para la cuadrícula tras restar
// las 2 líneas de chrome (ayuda + estado). Mantener esto exacto evita
// que la View desborde la terminal y corte la parte superior.
func (m *Model) gridAvailable() int {
	if h := m.height - 2; h > 0 {
		return h
	}
	return 1
}
