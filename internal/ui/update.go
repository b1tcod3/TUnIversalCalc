package ui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"tuniversalcalc/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, handled := m.handleMessage(msg); handled {
		return m, cmd
	}
	return m, m.updateFocusedInput(msg)
}

// handleMessage atiende los mensajes propios del ciclo Elm y los eventos de
// terminal. Devuelve handled=false para que el mensaje llegue al input activo.
func (m *Model) handleMessage(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg), true
	case ResultMsg:
		m.applyResult(msg)
		return nil, true
	case previewTickMsg:
		return m.handlePreviewTick(msg), true
	case previewResultMsg:
		m.applyPreviewResult(msg)
		return nil, true
	case autosaveMsg:
		return m.handleAutosave(), true
	case sessionSavedMsg:
		m.handleSessionSaved(msg)
		return nil, true
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return nil, false
}

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	m.width = msg.Width
	m.height = msg.Height
	m.updateLayout()
	if m.helpOpen {
		m.layoutHelp()
	}
	return nil
}

func (m *Model) handleAutosave() tea.Cmd {
	cmds := []tea.Cmd{autosaveTick()}
	if m.dirty {
		if cmd := m.saveSessionCmd(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) handleSessionSaved(msg sessionSavedMsg) {
	if msg.name != m.sessionName {
		return
	}
	if msg.err != nil {
		m.addLog(LevelError, "No se pudo guardar la sesión: "+msg.err.Error())
		return
	}
	// Un guardado antiguo puede llegar cuando el estado ya cambió: en ese
	// caso hay que volver a guardar.
	m.dirty = sessionFingerprint(m.sessionData()) != msg.fingerprint
}

// updateFocusedInput enruta el mensaje solo al input enfocado (el otro
// devuelve nil) y programa la previsualización si el texto cambió.
func (m *Model) updateFocusedInput(msg tea.Msg) tea.Cmd {
	if m.picker.open {
		var cmd tea.Cmd
		m.picker.query, cmd = m.picker.query.Update(msg)
		return cmd
	}
	oldValue := m.focusedInputValue()
	ti := &m.cmdInput
	if m.focus == FocusDecl {
		ti = &m.declInput
	}
	var cmd tea.Cmd
	*ti, cmd = ti.Update(msg)
	if m.focusedInputValue() != oldValue {
		m.dirty = true
		if m.focus == FocusCmd {
			if previewCmd := m.schedulePreview(); previewCmd != nil {
				cmd = tea.Batch(cmd, previewCmd)
			}
		}
	}
	m.refreshPalette()
	return cmd
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
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if cmd, handled, modal := m.handleModalKey(msg); modal {
		return cmd, handled
	}
	return m.handleShortcut(msg)
}

// handleModalKey da prioridad a los modales abiertos. El tercer valor indica
// si algún modal consumió la tecla, manejada o no.
func (m *Model) handleModalKey(msg tea.KeyMsg) (tea.Cmd, bool, bool) {
	// La confirmación de /limpiar es el modal de mayor prioridad.
	if m.clearConfirm {
		cmd, handled := m.handleClearConfirmKey(msg)
		return cmd, handled, true
	}

	// Ctrl+L abre la misma confirmación que /limpiar desde cualquier panel.
	if msg.Type == tea.KeyCtrlL {
		m.openClearConfirm()
		return nil, true, true
	}

	// El selector de sesiones es modal y tiene prioridad.
	if m.picker.open {
		cmd, handled := m.handleSessionPickerKey(msg)
		return cmd, handled, true
	}

	// La ayuda modal tiene prioridad sobre paneles y paleta.
	if m.helpOpen {
		cmd, handled := m.handleHelpKey(msg)
		return cmd, handled, true
	}

	// Paleta abierta: navegar, insertar o cerrar tiene prioridad.
	if m.palette.Open {
		cmd, handled := m.handlePaletteKey(msg)
		return cmd, handled, true
	}

	// Popup de autocompletado de expresiones.
	if m.completion.Open {
		cmd, handled := m.handleCompletionKey(msg)
		return cmd, handled, true
	}
	return nil, false, false
}

func (m *Model) handleShortcut(msg tea.KeyMsg) (tea.Cmd, bool) {
	if cmd, handled := m.handleGlobalShortcut(msg); handled {
		return cmd, handled
	}
	if cmd, handled := m.handlePanelShortcut(msg); handled {
		return cmd, handled
	}
	return m.handleOutputScroll(msg)
}

func (m *Model) handleGlobalShortcut(msg tea.KeyMsg) (tea.Cmd, bool) {
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
	}
	return nil, false
}

func (m *Model) handlePanelShortcut(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyTab:
		m.switchPanel()
		return nil, true
	case tea.KeyShiftTab:
		m.jumpNextTemplateField()
		return nil, true
	case tea.KeyEnter:
		return m.handleEnter(), true
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
	}
	return nil, false
}

func (m *Model) handleOutputScroll(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
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
	ti := &m.cmdInput
	if m.focus == FocusDecl {
		ti = &m.declInput
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
	if isClearCommand(m.focusedInputValue()) {
		m.openClearConfirm()
		return nil
	}
	switch m.focus {
	case FocusDecl:
		return m.handleEnterDecl()
	case FocusCmd:
		return m.handleEnterCmd()
	}
	return nil
}

func (m *Model) handleEnterDecl() tea.Cmd {
	raw := strings.TrimSpace(m.declInput.Value())
	if raw == "" {
		m.addLog(LevelInfo, "Declara un objeto: /puntos (1,2) (3,4), /funcion 2x+2 o /ecuacion 2x+2=0")
		return nil
	}
	if tpl, ok := matchTemplate(declRegistry, raw); ok {
		m.insertTemplate(tpl)
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
}

func (m *Model) handleEnterCmd() tea.Cmd {
	if !m.declValid {
		if _, ok := literalIntegerDecl(m.cmdInput.Value()); !ok {
			return nil
		}
	}
	if tpl, ok := matchTemplate(opsRegistry, m.cmdInput.Value()); ok {
		// /potencia y /radical no necesitan argumentos: si se escriben
		// completos, se ejecutan directamente sin insertar la plantilla.
		full := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(m.cmdInput.Value(), "/")))
		if !isObjectOnlyCommand(tpl.Trigger) || full != tpl.Trigger {
			m.insertTemplate(tpl)
			return nil
		}
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

// insertTemplate rellena el input enfocado con la plantilla elegida.
func (m *Model) insertTemplate(tpl Template) {
	m.pushUndo()
	m.invalidatePending()
	ti := m.focusedInput()
	val, pos := JumpNextField(tpl.InsertText, 0)
	ti.SetValue(val)
	ti.SetCursor(pos)
	m.dirty = true
	m.addLog(LevelInfo, "Plantilla insertada: "+tpl.Trigger+" (Shift+Tab para saltar campos)")
}

// resolveTarget averigua sobre qué objeto y variable opera el comando:
//   - /lagrange x     → último objeto, variable x
//   - /lagrange t1    → objeto t1
//   - /newton g1      → error si "g1" no es un objeto declarado
func (m *Model) resolveTarget(command string) (engine.Decl, string, bool) {
	toks := splitCommandTokens(normalizeCommandArgs(command))
	if d, ok := literalIntegerDecl(command); ok {
		return d, "x", true
	}
	if len(toks) > 0 && isObjectOnlyCommand(toks[0]) {
		decl, err := m.resolveObjectTarget(toks)
		if err != nil {
			m.addLog(LevelError, err.Error())
			return engine.Decl{}, "", false
		}
		return decl, "x", true
	}
	decl, varName, err := m.resolveTargetTokens(toks)
	if err != nil {
		m.addLog(LevelError, err.Error())
		return engine.Decl{}, "", false
	}
	return decl, varName, true
}

// isObjectOnlyCommand marca las operaciones que actúan sobre la expresión
// completa y no sobre una variable concreta.
func isObjectOnlyCommand(tok string) bool {
	name := strings.ToLower(strings.TrimPrefix(tok, "/"))
	return strings.HasPrefix(name, "potencia") ||
		strings.HasPrefix(name, "radical") ||
		strings.HasPrefix(name, "desarrollar") ||
		strings.HasPrefix(name, "factorizar")
}

// resolveObjectTarget resuelve /potencia o /potencia p1 sin admitir un
// segundo token como variable.
func (m *Model) resolveObjectTarget(toks []string) (engine.Decl, error) {
	if len(toks) == 1 {
		name, err := m.defaultTargetName()
		if err != nil {
			return engine.Decl{}, err
		}
		return m.objs[name], nil
	}
	if len(toks) == 2 {
		name := strings.ToLower(toks[1])
		decl, ok := m.objs[name]
		if ok {
			return decl, nil
		}
		return engine.Decl{}, fmt.Errorf("objeto %q no declarado — declarados: %s", name, m.objList())
	}
	cmd := strings.TrimPrefix(toks[0], "/")
	return engine.Decl{}, fmt.Errorf("/%s opera sobre una expresión completa: usa /%s o /%s p1", cmd, cmd, cmd)
}

// resolveTargetTokens resuelve objeto y variable a partir de los tokens de un
// comando. Comparte la lógica entre la ejecución real y la previsualización.
func (m *Model) resolveTargetTokens(toks []string) (engine.Decl, string, error) {
	if len(toks) == 0 {
		return engine.Decl{}, "", errors.New("comando vacío")
	}
	objName, varName, explicit, err := m.parseTargetName(toks)
	if err != nil {
		return engine.Decl{}, "", err
	}
	decl, ok := m.objs[objName]
	if !ok {
		return engine.Decl{}, "", fmt.Errorf("objeto %q no declarado — declarados: %s", objName, m.objList())
	}
	if explicit {
		if err := validateTargetVar(decl, varName); err != nil {
			if looksLikeObjectName(varName) {
				return engine.Decl{}, "", fmt.Errorf("objeto %q no declarado — declarados: %s", varName, m.objList())
			}
			return engine.Decl{}, "", err
		}
	}
	return decl, varName, nil
}

// looksLikeObjectName reconoce los auto-nombres e1, f1, p1, t1... para dar
// un error de objeto en vez de uno de variable cuando el usuario se equivoca.
func looksLikeObjectName(s string) bool {
	if len(s) < 2 || !isASCIIDigit(s[1]) {
		return false
	}
	for i := 2; i < len(s); i++ {
		if !isASCIIDigit(s[i]) {
			return false
		}
	}
	return true
}

func isASCIIDigit(c byte) bool { return c >= '0' && c <= '9' }

// parseTargetName interpreta el objeto y la variable de los tokens del
// comando, aplicando el último objeto declarado como valor por defecto.
// El tercer valor indica si la variable se escribió explícitamente.
func (m *Model) parseTargetName(toks []string) (objName, varName string, explicit bool, err error) {
	if len(toks) < 2 {
		objName, err = m.defaultTargetName()
		return objName, "x", false, err
	}
	name := strings.ToLower(toks[1])
	if _, ok := m.objs[name]; ok {
		varName = "x"
		if len(toks) >= 3 {
			varName = strings.ToLower(toks[2])
			explicit = true
		}
		return name, varName, explicit, nil
	}
	objName, err = m.defaultTargetName()
	return objName, name, true, err
}

// validateTargetVar comprueba que la variable citada exista en el objeto.
// Los puntos no tienen variables propias: cualquier identificador vale.
func validateTargetVar(decl engine.Decl, name string) error {
	if !engine.ValidIdentifier(name) {
		return fmt.Errorf("variable %q no válida: usa letras y dígitos", name)
	}
	if decl.Tipo == engine.DeclPuntos {
		return nil
	}
	vars := engine.VariablesOf(decl)
	if slices.Contains(vars, name) {
		return nil
	}
	if len(vars) == 0 {
		return fmt.Errorf("el objeto %s no tiene variables; no se puede usar %q", decl.Nombre, name)
	}
	return fmt.Errorf("la variable %q no aparece en %s — variables: %s", name, decl.Nombre, strings.Join(vars, ", "))
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
func literalIntegerDecl(command string) (engine.Decl, bool) {
	toks := splitCommandTokens(normalizeCommandArgs(command))
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
		b.WriteString(renderLogEntry(e.Level, e.Content, m.output.Width))
	}
	return b.String()
}

// --- Layout dinámico ---

// En Lipgloss v1 Width/Height incluyen el padding pero no el borde.
// Los paneles restan solo el borde al fijar Width/Height; para calcular el
// área de texto hay que restar además el padding correspondiente.

const (
	panelBorder = 2 // borde de cualquier panel
	inputPadX   = 2 // Padding(0,1) del panel de entrada
	outputPadX  = 2 // Padding(1) horizontal del panel de salida
	outputPadY  = 2 // Padding(1) vertical del panel de salida
)

// updateLayout recalcula las dimensiones de los componentes según el tamaño
// de la terminal. Solo se aplica si ya recibimos WindowSizeMsg.
func (m *Model) updateLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	lw, rw := m.leftWidth(), m.rightWidth()
	_, rh := m.leftHeight(), m.rightHeight()
	contentW := lw - panelBorder - inputPadX // área de texto de los paneles 1/2

	// margen para el prompt "❯ "
	m.declInput.Width = max(contentW-3, 10)
	m.cmdInput.Width = max(contentW-3, 10)

	// El viewport ocupa el contenido del panel 3 tras la cabecera
	w := rw - panelBorder - outputPadX
	h := (rh - panelBorder - outputPadY) - m.panel3HeaderRows()
	if w < 6 {
		w = 6
	}
	if h < 2 {
		h = 2
	}
	// Al cambiar el ancho hay que volver a envolver la salida, porque el
	// viewport recorta las líneas que no caben. Si el usuario estaba al
	// final, se mantiene allí tras el reajuste.
	atBottom := m.output.AtBottom()
	m.output.Width = w
	m.output.Height = h
	m.output.SetContent(m.renderLog())
	if atBottom {
		m.output.GotoBottom()
	}

	// Filas visibles de la paleta: contenido del panel menos título,
	// blanco e input. 2..7 para que nunca desborde el layout.
	m.palette.Max = min(max(m.leftHeight()-panelBorder-4, 2), 7)
	m.completion.Max = min(max(m.leftHeight()-panelBorder-4, 2), 7)
	if m.completion.Open {
		m.clampCompletion()
	}
}

func (m *Model) leftWidth() int {
	w := max(m.width*50/100, 45)
	// En terminales estrechas reserva al menos 12 columnas para el panel de
	// salida en lugar de dejar que rightWidth() se vuelva negativo.
	return min(w, max(m.width-12, 1))
}

func (m *Model) rightWidth() int {
	return m.width - m.leftWidth()
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
